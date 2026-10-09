# Architecture

## What this is

One Go process serves three things: a JSON API, a websocket gateway, and (in
production) the built React app. The browser draws the battle map with Konva and
keeps it in sync with everyone else at the table over the socket.

```
Browser                         Go service                    MySQL
┌───────────────────┐          ┌────────────────────────┐    ┌──────────┐
│ React + Konva     │  /api    │ middleware             │    │ campaigns│
│                   │ ───────► │   ↓                    │    │ maps     │
│  useBoard()       │          │ handler ──► service ──►│───►│ tokens   │
│       ▲           │  /ws     │   ↓          repository│    └──────────┘
│       └───────────│ ◄──────► │ ws.Hub (room per map)  │
└───────────────────┘          └────────────────────────┘
```

## Layering

Dependencies point one way only. A layer may call the one below it and nothing
above.

| Layer        | Owns                                                      | Must not                             |
| ------------ | --------------------------------------------------------- | ------------------------------------ |
| `middleware` | Cross-cutting HTTP concerns: panic recovery, logging, CORS | Know about any resource              |
| `handler`    | HTTP and websocket shape: parsing, status codes, broadcast, route guards | Build SQL, hold business rules       |
| `service`    | Validation, defaults, business rules                       | Touch `http` or `ws`                 |
| `repository` | SQL, scanning, pagination queries                          | Validate input                       |
| `model`      | Structs and the shared list-query parsing                  | Depend on any other internal package |

Two consequences worth remembering:

- **Services never import `net/http` or `internal/ws`.** They return domain
  errors; handlers turn those into status codes and broadcasts. That is why
  `TokenService.Move` persists but does not broadcast — the caller does, because
  only the caller knows whether a websocket client should be skipped.
- **`internal/ws` is domain-free.** Rooms are opaque `int64` keys and messages
  are opaque bytes. It has no idea what a token is, so nothing in it has to
  change when the domain grows.

## Request lifecycle

`POST /api/tokens`:

1. `middleware.Recover` → `middleware.Logger` → `middleware.CORS`
2. `handler.TokenHandler.Create` decodes the body (capped at 1 MiB)
3. `service.TokenService.Create` trims, applies defaults, rejects a blank name
4. `repository.TokenRepository.Create` inserts, then the service re-reads the row
5. The handler broadcasts `token.created` to the map's room and answers `201`

A websocket `token.moved` follows the same path from step 3, except the handler
broadcasts to everyone in the room *except* the sender.

Authenticated routes get one more step at the front: `guard.user` or `guard.dm`
resolves the session cookie to a `model.User`, puts it in the request context,
and answers 401 or 403 itself when the caller does not qualify. The guards live
in `handler` rather than `middleware` because they resolve a domain object, and
they are applied per route so the router reads as the permission table. See
[auth.md](auth.md).

## Key decisions

### No HTTP framework

Go 1.22 `net/http` mux does method-and-pattern routing (`GET /api/maps/{id}`),
which is the only thing a router was needed for. Four direct dependencies in
total: the MySQL driver, the websocket library, godotenv, and
`golang.org/x/crypto` for argon2id.

### `github.com/coder/websocket`, not `nhooyr.io/websocket`

Same library — the author moved it. The old import path is frozen.

### A small embedded migrator instead of golang-migrate

golang-migrate pulls Docker and OpenTelemetry into the module graph and raises
the required toolchain. [`internal/database/migrate.go`](../backend/internal/database/migrate.go)
does what this project needs in well under 200 lines and keeps golang-migrate's
file naming, so the CLI stays usable if the runner is ever outgrown.

### One origin in production

With `WEB_DIR` set, the binary also serves `frontend/dist` with an SPA fallback.
The API, the socket and the app share a hostname, so the Cloudflare Tunnel needs
one ingress rule and the browser never makes a cross-origin request. CORS is
then a development-only concern — but `CORS_ORIGINS` still gates which `Origin`
may open a websocket, so it must list the real hostname in production.

### Roles on the account, not on the membership

`users.role` is `dm` or `player` for everything. A per-campaign membership table
would let one person DM one campaign and play in another, which is the more
general model — but it puts a join on every permission check, and this is one
group with one Dungeon Master. Changing the mind later is a migration plus a
different `guard`, not a rewrite.

### Sessions in the database, not JWTs

Everything is one origin, so an HttpOnly cookie is simply available — including
on the websocket upgrade, which is where bearer tokens get awkward. A session
row is also revocable, which a JWT is not without building the revocation list
that the session table already is. The cost is one indexed lookup per request.

### One room per campaign

The hub keys rooms by campaign, not by map. Chat and rolls belong to the table
rather than to whichever map is open, and one room means one connection and no
subscription protocol.

The cost is that a client receives token events for maps it is not looking at,
and filters them on `map_id`. For a group running one campaign at a time that is
a handful of extra frames, which is cheaper than joining two rooms per client or
fanning out across every map in a campaign.

### Sheets are one repository, not four

`CharacterRepository` covers the character row, its items, its spells and its
slots, and one service sits on top. The same screen reads and writes all four,
so splitting them would add files without adding a seam. The ownership check
lives in one place, `CharacterService.authorize`, which every mutating call
routes through, including item and spell edits, which resolve to their character
first.

### A monster is a template, a token is an instance

Spawning copies the stat block vitals onto each token rather than reading them
through the link, so six goblins take damage separately while sharing one block
for their actions. `tokens.character_id` and `tokens.monster_id` are both
nullable, and which one is set decides where damage lands: a linked sheet owns
its hit points, and anything else carries its own.

### Dice are evaluated server-side

The client sends an expression, never a result. The breakdown is stored, so a
roll can be checked afterwards, and `crypto/rand` supplies the numbers so nobody
has to reason about how the generator was seeded.

## Concurrency

- **One writer per connection.** Everything a connection sends — the initial
  board, broadcasts, error replies — goes through that client's queue, and a
  single `writePump` goroutine drains it. Nothing else writes to the socket.
- **Send is never blocking.** A client's queue holds 32 messages. Past that,
  updates are dropped rather than stalling the room; the client notices the
  close, reconnects after two seconds, and reloads the board.
- **Closing is lock-protected.** `Leave` closes the queue under the hub's write
  lock, and `Broadcast` and `SendTo` take the read lock, so a send can never
  race with a close.
- **No `WriteTimeout` on the server.** It would cut long-lived websockets.
  Individual frame writes are bounded by a 10 second context instead.

## Trust boundaries

- `sort_by` and `sort_dir` are interpolated into SQL, so both are checked
  against a per-resource allowlist in `model.ParseQuery` before the repository
  sees them. Everything else is a bound parameter.
- Permission is decided from the session, never from the request body. The
  client cannot name its own role or claim a token it does not own.
- A player may move only a token whose stored `owner_user_id` is theirs. The
  check reads the row before writing, in `TokenService.Move`, so disabling the
  drag in the UI is a convenience rather than the rule.
- After a websocket move, the server re-reads the token and compares its
  `map_id` to the room. A client cannot write into a map it is not watching.
- Password hashes leave the database in exactly one place,
  `UserRepository.GetCredentials`, and never reach a response: every response is
  built from `model.User`, which has no hash field.
- A player reaches only their own character sheet, and only their own items and
  spells, which resolve to their character before the check.
- Dice results come from the server. A client cannot report a roll it made
  itself, and cannot attribute a roll to a sheet it does not own.
- An attack is resolved entirely server-side: hit, critical, damage and the new
  hit points. A client names a weapon and a target, nothing more. Attacking with
  a weapon needs the sheet it is on; a monster action needs the DM, and the
  attacking token must really use that stat block.
- Hit points are clamped in the `UPDATE` itself, so nothing goes negative and
  no race pushes a creature past full health.
- Request bodies are capped at 1 MiB; `page_size` is capped at 100; chat
  messages at 2000 characters; dice expressions at 20 terms and 100 dice.
