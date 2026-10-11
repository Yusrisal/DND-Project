# Deployment runbook

Target: the Debian homeserver at `192.168.1.15`, running the service directly
under systemd and exposed through the existing Cloudflare Tunnel. No Docker, no
Dokploy. MySQL is the instance already on the box.

Everything here goes through [`deploy/dnd-deploy`](../deploy/dnd-deploy), a
POSIX shell script that needs only `ssh`, `scp` and `tar` locally, so it runs
from Git Bash on Windows as well as from Linux.

```
./deploy/dnd-deploy --help
```

## What gets deployed

One Go binary serves the API, the websocket and the built frontend, so there is
a single unit and a single tunnel ingress rule. Releases sit side by side and
are switched with a symlink:

```
/opt/dnd/
  .env                                configuration, never touched by a deploy
  current -> releases/20261011-143000
  releases/20261011-143000/
    bin/dnd-api                       the service
    bin/dnd-migrate                   the migration runner
    dist/                             frontend build output
```

The symlink is what makes a rollback a swap and a restart rather than a rebuild.
`WEB_DIR` and the unit both point through `current`, so neither has to change
when a release does.

Migrations are embedded in both binaries, so `backend/migrations/` is never
copied to the server.

## Prerequisites

- Go 1.25+ and Node 20+ on the machine you deploy from. The binary is
  cross-compiled with `CGO_ENABLED=0`, so it is static and the server needs
  nothing installed.
- SSH access to the homeserver, with `sudo` for the account you log in as.
- A MySQL database and user on the server.
- `curl` on the server, used for the health check.

## First install

1. **Configure the deploy script.**

   ```bash
   cp deploy/deploy.env.example deploy/deploy.env
   ```

   Fill in `SSH_HOST`, `SSH_USER` and the hostname. The file is gitignored
   because it names your server; it holds no passwords.

2. **Create the database.**

   ```sql
   CREATE DATABASE dnd CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   CREATE USER 'dnd'@'localhost' IDENTIFIED BY '<password>';
   GRANT ALL PRIVILEGES ON dnd.* TO 'dnd'@'localhost';
   ```

3. **Bootstrap the server.**

   ```bash
   ./deploy/dnd-deploy setup
   ```

   This creates the `dnd` system account, makes `/opt/dnd` and
   `/opt/dnd/releases`, installs the systemd unit and enables it. The first run
   stops and tells you to write `/opt/dnd/.env`, because it will not invent
   your database password:

   ```bash
   scp deploy/env.production.example <user>@<host>:/tmp/dnd.env
   ssh <user>@<host> "sudo install -o dnd -g dnd -m 600 /tmp/dnd.env /opt/dnd/.env && sudo nano /opt/dnd/.env"
   ```

   [`deploy/env.production.example`](../deploy/env.production.example) explains
   each setting. Two that bite:

   - `CORS_ORIGINS` must be the real hostname. It gates both CORS and the
     websocket `Origin` check, so a stale value here means the page loads fine
     and the board silently never connects.
   - `APP_ENV=production` marks the session cookie `Secure`. Correct behind the
     tunnel, but it means you cannot sign in over plain `http://192.168.1.15:8080`.

   Then run `setup` again.

4. **Deploy.**

   ```bash
   ./deploy/dnd-deploy release
   ```

5. **Point the tunnel** at `http://localhost:8080` with one ingress rule, and
   restart `cloudflared`. The tunnel carries the websocket upgrade on the same
   rule; no separate route is needed.

6. **Register the first account immediately.** While the `users` table is empty,
   registration takes no invite code and creates a DM. Open the app, choose
   **Use an invite**, leave the code field blank, and create your account before
   anyone else can reach the hostname. Once one user exists the door closes and
   every registration needs a code.

7. **Confirm the board works**: create a campaign, a map and a token, and check
   the status dot next to the map name is green.

## Deploying an update

```bash
./deploy/dnd-deploy release
```

In order, that:

1. Cross-compiles both binaries and builds the frontend into `.deploy-stage`.
2. Tars it up, uploads it, and unpacks it into a new timestamped release.
3. Runs `migrate up` **with the new binary, before the switch** — the new code
   may expect columns the old schema does not have.
4. Repoints `current` and restarts the service.
5. Polls `/api/health` for 30 seconds.

If the health check fails it puts `current` back, restarts, prints the last 40
log lines and exits non-zero. **The schema is not rolled back**, because
automatically reversing a migration is how data goes missing. See below.

Old releases beyond `KEEP_RELEASES` (default 5) are pruned, never the live one.

## Rollback

```bash
./deploy/dnd-deploy rollback
```

Points `current` at the previous release, restarts, and verifies health.

The schema is deliberately left alone. If the release you backed out of added
migrations, decide case by case:

- **It added a column or a table.** Leave the schema. The old binary ignores
  what it does not know about, which is the safer outcome.
- **It dropped or rewrote something.** Roll it back explicitly, knowing it may
  destroy data:

  ```bash
  ./deploy/dnd-deploy migrate down 1
  ```

## Day to day

```bash
./deploy/dnd-deploy status        # service, live release, health, migration version
./deploy/dnd-deploy releases      # what is on disk, and which is current
./deploy/dnd-deploy logs 100      # last 100 journal lines
./deploy/dnd-deploy follow        # tail the journal
./deploy/dnd-deploy migrate status
./deploy/dnd-deploy restart
./deploy/dnd-deploy ssh           # a shell in /opt/dnd
```

`make deploy`, `make deploy-status`, `make deploy-logs` and
`make deploy-rollback` wrap the common ones, if you have `make`.

## Troubleshooting

| Symptom                                       | Check                                                                     |
| --------------------------------------------- | ------------------------------------------------------------------------- |
| `release` hangs with no output                | `sudo` is prompting for a password; the script allocates a tty, so answer it |
| Permission denied writing `/opt/dnd`          | The login account needs `sudo`, or set `SUDO=` empty if you log in as root |
| Health check fails but logs look fine         | `HEALTH_URL` is wrong, or `curl` is missing on the server                 |
| Unit will not start                           | `./deploy/dnd-deploy logs 50`; usually bad database credentials or an unreadable `.env` |
| `/api/health` returns `degraded`              | MySQL is down or the credentials are wrong                                |
| App loads, status dot stays red               | `CORS_ORIGINS` does not match the hostname, so the websocket upgrade is refused |
| Blank page, API responds                      | `WEB_DIR` is unset or does not point at `/opt/dnd/current/dist`           |
| Signed in, then immediately signed out again  | `APP_ENV=production` while serving plain HTTP, so the browser drops the `Secure` cookie |
| Tokens move locally but others do not see it  | Players are on different maps, or the socket reconnect loop is running — check the dot |
| A player cannot drag their token              | It has no `owner_user_id`, or it belongs to someone else; the DM reassigns it |
| Registration rejects every code               | Codes are single-use and expire after seven days; mint a fresh one          |
| Nobody can register and there is no DM        | A user already exists, so the bootstrap is closed. Mint a code as the existing DM, or delete the stray row and register again |
| `Error 1045` at startup                       | MySQL user or password wrong in `/opt/dnd/.env`                           |
| Migration fails halfway                       | MySQL does not roll DDL back; inspect the table, fix it by hand, then re-run |

Logs are structured and go to the journal:

```bash
./deploy/dnd-deploy logs 200
ssh <user>@<host> "journalctl -u dnd-api --since '10 minutes ago' | grep -i error"
```

## Deploying by hand

If the script is unavailable or you want to see what it does, the whole thing is
four steps:

```bash
# 1. build
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o bin/dnd-api ./cmd/server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o bin/dnd-migrate ./cmd/migrate
cd ../frontend && npm ci && npm run build

# 2. upload into a new release directory
ssh <host> "sudo install -d -o dnd -g dnd /opt/dnd/releases/manual"
scp -r backend/bin frontend/dist <host>:/tmp/rel/
ssh <host> "sudo cp -r /tmp/rel/* /opt/dnd/releases/manual/ && sudo chown -R dnd:dnd /opt/dnd/releases/manual && sudo chmod +x /opt/dnd/releases/manual/bin/*"

# 3. migrate, then switch
ssh <host> "cd /opt/dnd && sudo -u dnd /opt/dnd/releases/manual/bin/dnd-migrate up"
ssh <host> "sudo ln -sfn /opt/dnd/releases/manual /opt/dnd/current && sudo systemctl restart dnd-api"

# 4. verify
ssh <host> "curl -s localhost:8080/api/health"
```

The `chmod +x` matters when you build on Windows: the exec bit does not survive,
so the binary arrives unrunnable.

## Exposure

The service has no rate limiting on `/api/auth/login`. Argon2id makes each
attempt expensive, but nothing caps the attempts. If the hostname is reachable
from the open internet rather than only over the tunnel to people you have
invited, put Cloudflare Access or a rate limit in front of it. The full list of
what the auth layer does not cover is in [auth.md](auth.md#what-this-does-not-do).

## Notes on the unit

[`deploy/dnd-api.service`](../deploy/dnd-api.service) runs as `dnd`, restarts on
failure after 5s, and is sandboxed with `ProtectSystem=strict` plus a single
`ReadWritePaths=/opt/dnd`. `ExecStart` goes through `current`, which is what
lets a rollback be a symlink swap. If you move the install directory, both
`ReadWritePaths` and `WEB_DIR` have to move with it.
