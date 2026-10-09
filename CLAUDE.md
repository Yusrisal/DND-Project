# CLAUDE.md

Monorepo for a shared DnD battle map. Go API + websocket gateway in `backend/`,
React + Konva board in `frontend/`. Read [docs/architecture.md](docs/architecture.md)
before changing how the layers fit together.

## Commands

```bash
make api              # run the API on :8080
make web              # run Vite on :5173 (proxies /api and /ws)
make migrate-up       # apply migrations
make check            # go vet + tsc --noEmit — run before handing work back
make tidy             # gofmt -w + go mod tidy
```

Ask before starting the server, the frontend, or anything that touches MySQL.
`make check`, `go build` and `npm run build` are static and need no approval.

## Rules

- **Layering is one-directional:** handler → service → repository → model.
  Services must not import `net/http` or `internal/ws`. Handlers do the
  broadcasting, because only they know which websocket client to skip.
- **`internal/ws` stays domain-free.** Opaque room ids, opaque bytes. If a
  change there needs to know what a token is, it belongs in `handler`.
- **Socket rooms are keyed by campaign, not map.** Token payloads carry `map_id`
  and the client filters. Do not add a second room per client.
- **Dice are evaluated server-side.** The client sends an expression, never a
  result. Keep it that way. The same goes for attacks: hit, damage and the new
  hit points are all decided in `CombatService`.
- **A monster is a template, a token is an instance.** Spawning copies vitals
  onto the token. Do not make several tokens share one hit point pool.
- **Model files** are `snake_case` with a type suffix: `campaign_model.go`.
- **Every list endpoint** gets pagination, `search` and sorting via
  `model.ParseQuery`.
- **Only `sort_by` and `sort_dir` are interpolated into SQL,** and only after
  passing a per-resource allowlist. Everything else is a bound parameter.
- **Errors:** repositories return `sql.ErrNoRows`, services translate to
  `service.ErrNotFound` or `service.ValidationError`, handlers map those to
  404 / 422 in `writeError`.
- **Permission lives in the router.** Every route is wrapped in `g.user` or
  `g.dm`, so the route table is the permission model. A new write route is DM
  only unless there is a reason otherwise.
- **Never trust the client for identity.** Role and ownership come from the
  session and the stored row. Anything the UI hides must also be enforced in the
  service layer — see `TokenService.Move` and `CharacterService.authorize`.
- **Password hashes leave the database in one place only,**
  `UserRepository.GetCredentials`. Responses are built from `model.User`, which
  has no hash field. Keep it that way.
- **Schema changes always go through a migration.** Never edit an applied one.
  See [docs/database.md](docs/database.md).
- **No seeders** unless explicitly asked for.
- **Keep dependencies few.** There are four direct Go dependencies and the
  routing is stdlib. A new one needs a reason that a few lines cannot cover.
- **This is one group playing one campaign at a time.** Do not add
  multi-tenancy, sharding, caching layers or background workers for scale that
  is not there.

## Adding a resource

The walkthrough is in [docs/database.md](docs/database.md#adding-a-resource):
migration, model, repository, service, handler, router — in that order.

## Documentation

Docs live in `docs/` and the README links to them rather than repeating them.
When behaviour changes, update the doc that covers it: endpoints in
[api.md](docs/api.md), frames in [websocket.md](docs/websocket.md), schema in
[database.md](docs/database.md), permissions in [auth.md](docs/auth.md), sheets
and dice in [characters-and-dice.md](docs/characters-and-dice.md), fights in
[combat.md](docs/combat.md), operations in [deployment.md](docs/deployment.md).
