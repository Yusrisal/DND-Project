# Installing and deploying

Target: a Debian 13 homeserver running the service under systemd, behind the
existing Cloudflare Tunnel. No Docker, no Dokploy. MySQL or MariaDB is whatever
is already on the box.

Everything goes through [`deploy/dnd-deploy`](../deploy/dnd-deploy), a POSIX
shell script with no dependencies beyond what Debian ships.

```bash
./deploy/dnd-deploy --help
```

## What gets deployed

One Go binary serves the API, the websocket **and** the built frontend, so there
is a single unit and a single tunnel ingress rule. There is no separate
frontend server, and no second port to configure.

Releases sit side by side and are switched with a symlink:

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
`WEB_DIR` and the unit both point through `current`, so neither changes when a
release does.

Migrations are embedded in both binaries, so `backend/migrations/` is never
copied anywhere.

## Two ways to run it

The script works in either direction, and `SSH_HOST` is the switch.

| | **Local** — repo is on the server | **Remote** — push from a workstation |
| --- | --- | --- |
| `SSH_HOST` | empty | `192.168.1.15` |
| Builds on | the server | the workstation, cross-compiled |
| Needs on the server | Go 1.25+, Node 20+ | nothing but systemd and curl |
| Transfer | none, installed in place | tar over scp |

Local is simplest when you already edit on the server over SSH. Remote keeps
build tools off the server. Nothing else differs — same commands, same result.

## Install from scratch

### 1. Prerequisites

On whichever machine builds:

```bash
go version; node --version; npm --version
```

Go 1.25+ and Node 20+. Debian 13 packages Go 1.24, which is fine — Go
downloads the newer toolchain itself when `go.mod` asks for it — but a server
with no outbound internet needs the tarball from go.dev instead.

On the server:

```bash
apt update && apt install -y curl
```

`curl` is what the health check uses. Add `mariadb-server` too if the database
is not already there.

### 2. Database

Skip if you have one. Otherwise:

```sql
CREATE DATABASE dnd CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'dnd'@'localhost' IDENTIFIED BY '<password>';
GRANT ALL PRIVILEGES ON dnd.* TO 'dnd'@'localhost';
```

`GRANT ALL`, not just read-write: migrations create and alter tables. A
read-write grant fails on migration 1 with `Error 1142`.

On Debian, MariaDB authenticates root over a unix socket, so `sudo mariadb`
needs no password.

### 3. Get the code

```bash
git clone <your remote> /var/www/html/DND-Project
cd /var/www/html/DND-Project
```

### 4. Configure the deploy

`deploy/deploy.env` is **gitignored**, so it never arrives with a pull. Create
it on the machine you deploy from:

```bash
cp deploy/deploy.env.example deploy/deploy.env
```

For local mode on the server, as root:

```
SSH_HOST=
SSH_USER=
REMOTE_DIR=/opt/dnd
SERVICE=dnd-api
SERVICE_USER=dnd
SUDO=
HEALTH_URL=http://localhost:8101/api/health
KEEP_RELEASES=5
```

`SERVICE_USER` is the unprivileged account the service runs as, created in the
next step. It is not the account you deploy as.

`HEALTH_URL` must match `HTTP_ADDR` in the server `.env`, and must point at
`localhost` rather than the public hostname — see
[Changing the port](#changing-the-port).

### 5. Bootstrap the server

```bash
./deploy/dnd-deploy setup
```

Creates the `dnd` system account, makes `/opt/dnd` and `/opt/dnd/releases`,
installs the systemd unit and enables it.

The first run **stops and asks for `/opt/dnd/.env`**. That is the intended path,
not a failure: the script will not invent your database password.

### 6. Server configuration

```bash
cp deploy/env.production.example /opt/dnd/.env
chown dnd:dnd /opt/dnd/.env && chmod 600 /opt/dnd/.env
nano /opt/dnd/.env
```

[`deploy/env.production.example`](../deploy/env.production.example) explains
every setting. The ones that must be right:

```
APP_ENV=production
HTTP_ADDR=:8101
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=appuser
DB_PASSWORD=<password>
DB_NAME=dnd
CORS_ORIGINS=https://dnd.example.online
WEB_DIR=/opt/dnd/current/dist
```

Three that bite:

- **`CORS_ORIGINS` must be the real hostname.** It gates both CORS and the
  websocket `Origin` check, so a stale value means the page loads fine and the
  board silently never connects — which reads as "realtime is broken" rather
  than "config is wrong".
- **`APP_ENV=production` marks the session cookie `Secure`.** Correct behind the
  tunnel, but you then cannot sign in over plain `http://<ip>:8101`. Use
  `development` temporarily if you want to test before the tunnel is up.
- **`WEB_DIR` points at `current`, not at a release.** Point it at a release
  directory and the app keeps serving that one forever.

Then run setup again:

```bash
./deploy/dnd-deploy setup
```

### 7. Deploy

```bash
./deploy/dnd-deploy release
```

### 8. Tunnel

One ingress rule, hostname → `http://localhost:8101`. The tunnel carries the
websocket upgrade on the same rule; there is no second route to add.

```bash
systemctl restart cloudflared
```

### 9. Claim the DM account

**Do this before anyone else can reach the hostname.** While the `users` table
is empty, registration takes no invite code and creates a DM. Open the site,
choose **Use an invite**, leave the code blank. Once one user exists the door
closes and every registration needs a code from the DM.

Then create a campaign, a map and a token, and check the status dot beside the
map name is green. Green means the websocket is up.

## Deploying an update

```bash
git pull && ./deploy/dnd-deploy release
```

In order, `release`:

1. Builds both binaries and the frontend into `.deploy-stage`.
2. Installs them into a new timestamped release directory.
3. Runs `migrate up` **with the new binary, before the switch** — new code may
   expect columns the old schema lacks.
4. Repoints `current` and restarts the service.
5. Polls `/api/health` for 30 seconds.

If the health check fails it puts `current` back, restarts, prints the last 40
log lines and exits non-zero. **The schema is not rolled back**, because
automatically reversing a migration is how data goes missing.

Releases beyond `KEEP_RELEASES` are pruned. The live one is never pruned.

## Rollback

```bash
./deploy/dnd-deploy rollback
```

Points `current` at the previous release, restarts, verifies health.

The schema is left alone deliberately. If the release you backed out of added
migrations:

- **It added a column or table** — leave it. The old binary ignores what it does
  not know about, which is the safer outcome.
- **It dropped or rewrote something** — roll it back explicitly, knowing it may
  destroy data:

  ```bash
  ./deploy/dnd-deploy migrate down 1
  ```

## Day to day

```bash
./deploy/dnd-deploy status        # service, live release, health, migration version
./deploy/dnd-deploy releases      # what is on disk, and which is current
./deploy/dnd-deploy logs 100
./deploy/dnd-deploy follow        # tail the journal
./deploy/dnd-deploy migrate status
./deploy/dnd-deploy restart
./deploy/dnd-deploy build         # build without installing
./deploy/dnd-deploy ssh           # a shell in /opt/dnd
```

`make deploy`, `make deploy-status`, `make deploy-logs` and
`make deploy-rollback` wrap the common ones.

## Changing the port

Three places, then a restart. No rebuild: the frontend uses same-origin
relative paths, so it neither knows nor cares what port the server listens on.

1. `HTTP_ADDR` in `/opt/dnd/.env`
2. `HEALTH_URL` in `deploy/deploy.env`
3. The tunnel ingress rule

```bash
./deploy/dnd-deploy restart
```

Getting 1 and 2 out of step is the nastiest failure mode available here: every
future `release` looks like a failed deploy and auto-rolls-back a release that
was actually fine. `HEALTH_URL` stays on `localhost` so that a DNS or tunnel
problem cannot masquerade as a bad release.

## Troubleshooting

| Symptom | Check |
| --- | --- |
| `Permission denied` running the script | `chmod +x deploy/dnd-deploy`. Git stores the mode, but a checkout on a filesystem without exec bits loses it |
| `bad interpreter: /bin/sh^M` | The script got CRLF line endings. [`.gitattributes`](../.gitattributes) pins LF; re-clone or run `dos2unix deploy/dnd-deploy` |
| `no deploy/deploy.env` | It is gitignored on purpose. Create it on this machine from the example |
| Health check fails, logs look fine | `HEALTH_URL` does not match `HTTP_ADDR`, or `curl` is missing on the server |
| `release` hangs with no output | `sudo` is waiting for a password. Set `SUDO=` empty if you are root |
| `Error 1142` during migration | The database user lacks `CREATE`/`ALTER`. `GRANT ALL PRIVILEGES ON dnd.* TO ...` |
| `Error 1045` at startup | Wrong database user or password in `/opt/dnd/.env` |
| Unit will not start | `./deploy/dnd-deploy logs 50`; usually bad credentials or an unreadable `.env` |
| `/api/health` returns `degraded` | The database is down or the credentials are wrong |
| Blank page, API responds | `WEB_DIR` unset or not pointing at `/opt/dnd/current/dist` |
| App loads, status dot stays red | `CORS_ORIGINS` does not match the hostname, so the websocket upgrade is refused |
| Signed in, then immediately signed out | `APP_ENV=production` while serving plain HTTP, so the browser drops the `Secure` cookie |
| 502 at the hostname, `status` says healthy | The tunnel points at the wrong port. `status` checks localhost by design |
| A player cannot drag their token | It has no `owner_user_id`, or belongs to someone else; the DM reassigns it |
| Registration rejects every code | Codes are single-use and expire after seven days; mint a fresh one |
| Nobody can register and there is no DM | A user already exists, so the bootstrap is closed. Mint a code as the existing DM, or delete the stray row |
| Migration fails halfway | MySQL does not roll DDL back. Inspect the table, fix it by hand, re-run |

```bash
journalctl -u dnd-api --since "10 minutes ago" | grep -i error
```

## Deploying by hand

The whole script is four steps, if you ever need to do it without it:

```bash
# 1. build
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o bin/dnd-api ./cmd/server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o bin/dnd-migrate ./cmd/migrate
cd ../frontend && npm ci && npm run build

# 2. install into a new release directory
install -d -o dnd -g dnd /opt/dnd/releases/manual
cp -r backend/bin frontend/dist /opt/dnd/releases/manual/
chown -R dnd:dnd /opt/dnd/releases/manual
chmod +x /opt/dnd/releases/manual/bin/*

# 3. migrate, then switch
cd /opt/dnd && sudo -u dnd /opt/dnd/releases/manual/bin/dnd-migrate up
ln -sfn /opt/dnd/releases/manual /opt/dnd/current && systemctl restart dnd-api

# 4. verify
curl -s localhost:8101/api/health
```

Run the migration from `/opt/dnd`, not from the release directory: the binary
looks for `.env` relative to the working directory, the same way the unit does.

`chmod +x` matters when the build came from Windows — the exec bit does not
survive the trip, and the binary arrives unrunnable.

## Exposure

There is no rate limiting on `/api/auth/login`. Argon2id makes each attempt
expensive, but nothing caps the attempts. If the hostname is reachable from the
open internet rather than only by people you have invited, put Cloudflare Access
or a rate limit in front of it. The full list of what the auth layer does not
cover is in [auth.md](auth.md#what-this-does-not-do).

## Notes on the unit

[`deploy/dnd-api.service`](../deploy/dnd-api.service) runs as `dnd`, restarts on
failure after 5s, and is sandboxed with `ProtectSystem=strict` plus a single
`ReadWritePaths=/opt/dnd`. `ExecStart` goes through `current`, which is what
lets a rollback be a symlink swap. Move the install directory and both
`ReadWritePaths` and `WEB_DIR` have to move with it.

`After=` names both `mysql.service` and `mariadb.service`; systemd ignores
whichever is not installed.
