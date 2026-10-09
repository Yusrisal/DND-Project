# Development

## Prerequisites

| Tool  | Version | Notes                                  |
| ----- | ------- | -------------------------------------- |
| Go    | 1.25+   |                                        |
| Node  | 20+     | Developed on 22                        |
| MySQL | 8+      | Your own instance; no container needed |

## First run

```bash
cp .env.example .env
```

Create the database and a user for it:

```sql
CREATE DATABASE dnd CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'dnd'@'localhost' IDENTIFIED BY 'dnd';
GRANT ALL PRIVILEGES ON dnd.* TO 'dnd'@'localhost';
```

Point `DB_DSN` in `.env` at it, then apply the schema and start both sides in
two shells:

```bash
make migrate-up
make api
```

```bash
make web
```

The API listens on `:8080`, Vite on `:5173`. Vite proxies `/api` and `/ws`
through to the API, so the frontend always uses same-origin paths and behaves in
development the way it does behind the tunnel.

Open <http://localhost:5173>, pick **Use an invite**, and register with the code
field left empty: on an empty database the first account is created as the DM.
Everyone after that needs a code the DM mints from the Invites panel. See
[auth.md](auth.md).

Without `make`:

```bash
cd backend && go run ./cmd/migrate up && go run ./cmd/server
```

```bash
cd frontend && npm install && npm run dev
```

## Configuration

Both binaries read `.env` from the repository root, then from the working
directory. Real environment variables always win, which is how systemd supplies
them in production.

| Variable       | Default                 | Purpose                                                      |
| -------------- | ----------------------- | ------------------------------------------------------------ |
| `APP_ENV`      | `development`           | `production` marks the session cookie `Secure`                |
| `HTTP_ADDR`    | `:8080`                 | Listen address                                                |
| `DB_DSN`       | local `dnd` database    | Must include `parseTime=true`                                 |
| `CORS_ORIGINS` | `http://localhost:5173` | Comma-separated; also gates websocket `Origin`                |

| `WEB_DIR`      | empty                   | Production only: path to `frontend/dist`. Empty disables SPA serving |

`DB_DSN` needs `parseTime=true`, or scanning `DATETIME` into `time.Time` fails.

## Checks

```bash
make check    # go vet + tsc --noEmit
make tidy     # go mod tidy + gofmt -w
```

The tests that exist cover the pieces with real logic and no database: the dice
evaluator (`internal/service/dice_test.go`), the ability modifier rule
(`internal/model/character_model_test.go`), and attack expressions plus critical
damage doubling (`internal/service/combat_service_test.go`). Run them with
`go test ./...`.

The service layer is the useful seam for more: it holds the validation and
permission rules and depends only on a repository.

## Layout

```
backend/
  cmd/server/      HTTP + websocket entrypoint
  cmd/migrate/     Migration CLI
  internal/
    config/        Environment parsing
    database/      Connection pool and the migration runner
    middleware/    Recover, Logger, CORS
    handler/       HTTP handlers, websocket handler, route guards, router, SPA serving
    service/       Validation and business rules
    repository/    SQL
    model/         Structs and shared list-query parsing
    ws/            Domain-free websocket hub
  migrations/      Embedded .sql files
frontend/
  src/
    api/           Fetch client and shared types
    components/    DMSidebar, PlayerSidebar, MapCanvas, ChatPanel,
                   CharacterSheetPanel, InitiativePanel, BestiaryPanel,
                   AuthForm, InvitePanel, AccountBar
    hooks/         useAuth (session), useSession (tokens, chat, rolls)
docs/
deploy/
```

## Layout

Three columns, two arrangements, chosen by role in `App.tsx`:

| Column | Player            | DM                                           |
| ------ | ----------------- | -------------------------------------------- |
| Left   | Their sheet       | Tabs: board, party, bestiary, setup          |
| Centre | The board         | The board                                    |
| Right  | Initiative + chat | Initiative + chat                            |

The rule is: **left is what you manage, centre is the board, right is the
table.** A player manages a character, so the sheet gets the whole column
instead of a strip. The DM manages the world, so it goes behind tabs and only
the active one takes space — the board stays visible either way, which is why
tabs rather than a drawer.

Nothing lives in the right column except initiative and chat. That is
deliberate: with three panels sharing it, chat collapsed to a sliver exactly
when a fight filled the initiative list.

Role gating in the UI is a convenience. Every rule is enforced again in the
service layer — see [auth.md](auth.md).

## Conventions

- Model files are `snake_case` with a type suffix: `campaign_model.go`.
- Dependencies point one way: handler to service to repository to model.
  Services never import `net/http` or `internal/ws`.
- Anything interpolated into SQL — only `sort_by` and `sort_dir` — must be
  checked against an allowlist first. Everything else is a bound parameter.
- Repositories return `sql.ErrNoRows`; services translate it to
  `service.ErrNotFound`; handlers turn that into a 404.
- Every list endpoint supports pagination, `search` and sorting. See
  [api.md](api.md).
- Permission is decided in the router, by wrapping each route in `g.user` or
  `g.dm`. A new write route is DM only unless there is a reason otherwise, and
  anything the UI hides must also be enforced server-side.
- No seeders.

## Where to look

| Question                            | File                                            |
| ----------------------------------- | ----------------------------------------------- |
| How does a request reach the DB?     | [architecture.md](architecture.md)              |
| What does this endpoint accept?      | [api.md](api.md)                                |
| What does the socket send?           | [websocket.md](websocket.md)                    |
| Who is allowed to do what?           | [auth.md](auth.md)                              |
| How do sheets and dice work?         | [characters-and-dice.md](characters-and-dice.md) |
| How does a fight work?               | [combat.md](combat.md)                          |
| How do I change the schema?          | [database.md](database.md)                      |
| How do I ship this to the homeserver?| [deployment.md](deployment.md)                  |
