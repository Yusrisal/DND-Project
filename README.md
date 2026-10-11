# DnD Project

Monorepo for a shared DnD battle map: a Go API with a websocket gateway, and a
React + Konva board that stays in sync across everyone at the table.

```
backend/     Go service (API + websocket + migrations)
frontend/    React + TypeScript + Vite + Konva
docs/        Architecture, API, protocol, schema, deployment
deploy/      systemd unit for the homeserver
```

## Stack

| Layer    | Choice                                                           |
| -------- | ---------------------------------------------------------------- |
| Frontend | React + TypeScript (Vite), Konva for the map                     |
| Realtime | WebSocket (`github.com/coder/websocket`): tokens, chat and dice  |
| Backend  | Go, layered: middleware → handler → service → repository → model |
| Auth     | Session cookies in MySQL, argon2id passwords, DM and player roles |
| Database | MySQL, migrations embedded in the binary                         |
| Hosting  | systemd on the homeserver, behind the Cloudflare Tunnel          |

Four direct Go dependencies: the MySQL driver, the websocket library, godotenv,
and `golang.org/x/crypto` for argon2id. Routing is `net/http`, so there is no
framework to keep up with.

## Quick start

Needs Go 1.25+, Node 20+, and a MySQL you already run.

```bash
cp .env.example .env
```

Create the database, then point `DB_DSN` in `.env` at it:

```sql
CREATE DATABASE dnd CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

Apply the schema and start both sides in two shells:

```bash
make migrate-up && make api
```

```bash
make web
```

The app is on <http://localhost:5173>, the API on `:8080`. Vite proxies `/api`
and `/ws`, so the frontend uses same-origin paths in development exactly as it
does in production.

Register the first account with the invite-code field left empty: on an empty
database the first account becomes the Dungeon Master. Add a campaign, a map and
a token, then open the same page in a second window and drag the token — both
boards move.

Players join with a single-use code the DM mints from the Invites panel, and can
only drag the tokens assigned to them.

Each player gets a character sheet with stats, skills, inventory, spells and
slots, and rolls from it. Dice are evaluated server-side and land in the shared
feed beside the chat, so the whole table sees the breakdown. Type `/roll 1d20+5`
in the chat box, tap an ability or skill on the sheet, or use the quick-roll row.

Combat is resolved rather than narrated: the DM writes a stat block once and
spawns from it, initiative tracks whose turn it is, and attacks roll against
armour class and subtract damage from hit points. Click a token to target it,
then press attack next to a weapon. What the software does **not** decide is
listed in [combat.md](docs/combat.md).

## Documentation

| Document                               | For                                                        |
| -------------------------------------- | ---------------------------------------------------------- |
| [development.md](docs/development.md)  | Setup, configuration, project layout, conventions          |
| [architecture.md](docs/architecture.md) | How the layers fit together and why                       |
| [api.md](docs/api.md)                  | REST reference: endpoints, pagination, errors              |
| [auth.md](docs/auth.md)                | Sessions, passwords, invites, who may do what              |
| [characters-and-dice.md](docs/characters-and-dice.md) | Sheets, inventory, spells, slots, dice |
| [combat.md](docs/combat.md)            | Bestiary, initiative, attacks, conditions                   |
| [websocket.md](docs/websocket.md)      | Realtime protocol and delivery guarantees                  |
| [database.md](docs/database.md)        | Schema, migration workflow, adding a resource              |
| [deployment.md](docs/deployment.md)    | Build, install, update, rollback, troubleshooting          |

## Make targets

```
make api              run the API              make migrate-up      apply migrations
make web              run the frontend         make migrate-down    roll back one
make build            release build            make migrate-status  version + pending
make check            vet + typecheck          make tidy            gofmt + go mod tidy
```

## Deploying

```bash
cp deploy/deploy.env.example deploy/deploy.env   # name your server
./deploy/dnd-deploy setup                        # one time
./deploy/dnd-deploy release                      # every time after that
```

`release` builds, installs a new timestamped release, migrates, switches a
symlink, restarts and checks health — rolling itself back if the health check
fails. `rollback`, `status`, `logs` and `migrate` are the other commands worth
knowing.

It runs either on the server itself, or from a workstation that pushes over
ssh; `SSH_HOST` in `deploy/deploy.env` is the switch. Note that file is
gitignored, so create it wherever you deploy from.

Full runbook, including the by-hand version: [deployment.md](docs/deployment.md).
