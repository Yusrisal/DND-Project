# Authentication and roles

Two roles, held on the account itself:

| Role     | Can                                                                      |
| -------- | ------------------------------------------------------------------------ |
| `dm`     | Everything: create and delete campaigns, maps and tokens, mint invite codes, move any token |
| `player` | Read the world, and drag the tokens assigned to them                     |

The role is a column on `users`, not a per-campaign membership. One account is
either the Dungeon Master or a player, everywhere.

## Sessions

Login creates a row in `sessions` and returns a cookie:

```
Set-Cookie: dnd_session=<43 random chars>; Path=/; HttpOnly; SameSite=Lax; Secure
```

- **HttpOnly**, so no script can read it. There is no token in `localStorage`.
- **Secure** only when `APP_ENV=production`, because a Secure cookie is never
  sent over plain `http://localhost`.
- **SameSite=Lax**, which is sufficient because the app and the API share an
  origin.
- The cookie rides along on the websocket upgrade for free, which is why `/ws`
  needs no token in the query string.

Only the SHA-256 of the token is stored. The raw value exists in the cookie and
nowhere else, so a database dump does not hand over live sessions. The token is
32 bytes from `crypto/rand`, so there is nothing a slow KDF would protect here —
unlike a password, it cannot be guessed.

Sessions last 30 days and are revocable: logout deletes the row, and the lookup
query filters on `expires_at`, so an expired session reads exactly like a
missing one. A background ticker clears expired rows every six hours.

## Passwords

Argon2id, 64 MiB, one pass, four lanes, with a 16-byte random salt. Hashes are
stored in PHC format:

```
$argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
```

The parameters travel with the hash, and verification uses the recorded ones
rather than the current constants, so they can be raised later without
invalidating existing passwords. Minimum length is 10 characters.

Login hashes a password even when the email does not exist, so a missing account
and a wrong password take the same time to reject.

## Getting accounts in

### The first DM

While the `users` table is empty, registration accepts no invite code and
creates the account as `dm`. That is the bootstrap, and it is also a window:
**register the first account before the hostname is reachable by anyone else.**
Once one user exists, every registration needs a code.

### Everyone after that

The DM mints a single-use code, sends it over whatever chat the group uses, and
the invitee registers with it.

```bash
curl -s -X POST localhost:8080/api/invites -b cookies.txt \
  -H "Content-Type: application/json" -d "{\"role\":\"player\"}"
```

```json
{ "code": "q7fk3mwptz", "role": "player", "expires_at": "2026-10-16T14:02:11Z" }
```

Codes are ten characters from a 31-character alphabet with `0/O` and `1/I/l`
left out, so they survive being read aloud or retyped. They expire after seven
days and are spent on first use.

Registering and claiming the code happen in one transaction. A signup that fails
on a duplicate email does not burn the code, and two people racing on the same
code cannot both win — the `UPDATE` matching zero rows is the guard.

Registration also signs the new account in, so an invited player lands on the
board rather than on a login form.

## Route permissions

The full table is [`internal/handler/router.go`](../backend/internal/handler/router.go);
every route is wrapped in `g.user` or `g.dm`, or neither when it is public.

| Scope           | Routes                                                                     |
| --------------- | -------------------------------------------------------------------------- |
| Public          | `GET /api/health`, `POST /api/auth/{register,login,logout}`                 |
| Any account     | All `GET` routes, plus `GET /ws`                                            |
| DM only         | Every `POST`, `PUT` and `DELETE` on campaigns, maps, tokens and invites     |

A request without a session gets 401 and a cleared cookie, so a stale cookie
does not keep failing. A player hitting a DM route gets 403.

## Moving tokens

`tokens.owner_user_id` is nullable. A player may move a token only if it is
theirs; the DM may move anything, including unowned monsters.

The check reads the stored row before writing, so the client cannot claim
ownership by sending a different id. A rejected move gets an `error` frame back,
and the frontend reloads the board from `GET /api/maps/{id}/tokens` rather than
leaving the optimistic drag on screen.

The frontend disables dragging and dims tokens the viewer cannot move, but that
is a convenience. The rule is enforced server-side in `TokenService.Move`.

## What this does not do

Worth knowing before you expose the hostname:

- **No rate limiting on login.** Argon2id makes each attempt expensive, but
  nothing stops repeated tries. Cloudflare Access or a rate limit in front of
  `/api/auth/login` is the obvious place to add it.
- **No password reset.** The DM can create a new account; there is no email
  flow, because there is no mail configured.
- **No email verification.** The address is an identifier, not a proof.
- **No CSRF token.** `SameSite=Lax` plus a same-origin deployment covers the
  cross-site POST case. If the API is ever served from a different origin than
  the app, this needs revisiting.
- **No per-campaign roles.** A DM is a DM everywhere. If someone needs to play
  in a campaign they do not run, that is a schema change, not a config flag.
