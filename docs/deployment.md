# Deployment runbook

Target: the Debian homeserver at `192.168.1.15`, running the service directly
under systemd and exposed through the existing Cloudflare Tunnel. No Docker, no
Dokploy. MySQL is the instance already on the box.

## What gets deployed

One Go binary serves the API, the websocket and the built frontend, so there is
a single unit and a single tunnel ingress rule.

```
/opt/dnd/
  bin/dnd-api         the service
  bin/dnd-migrate     the migration runner
  dist/               frontend build output
  .env                configuration, owned by the dnd user
```

## Prerequisites

- A `dnd` system user and group on the server.
- A MySQL database and user, with the DSN ready for `.env`.
- Go 1.25+ and Node 20+ on whichever machine builds. Building on your Windows
  box and copying the artefacts is fine, as long as the binary is built for
  Linux.

## Build

On a Linux build host:

```bash
make build
```

Cross-compiling from Windows:

```bash
cd backend
GOOS=linux GOARCH=amd64 go build -o bin/dnd-api ./cmd/server
GOOS=linux GOARCH=amd64 go build -o bin/dnd-migrate ./cmd/migrate
cd ../frontend && npm ci && npm run build
```

Migrations are embedded in both binaries, so `backend/migrations/` does not need
to be copied.

## First install

1. Create the user and directory:

   ```bash
   sudo useradd --system --home /opt/dnd --shell /usr/sbin/nologin dnd
   sudo mkdir -p /opt/dnd/bin
   sudo chown -R dnd:dnd /opt/dnd
   ```

2. Copy `backend/bin/dnd-api`, `backend/bin/dnd-migrate` and `frontend/dist`
   into `/opt/dnd`.

3. Write `/opt/dnd/.env`:

   ```
   APP_ENV=production
   HTTP_ADDR=:8080
   DB_DSN=dnd:<password>@tcp(127.0.0.1:3306)/dnd?parseTime=true&charset=utf8mb4&loc=Local
   CORS_ORIGINS=https://<your-hostname>
   WEB_DIR=/opt/dnd/dist
   ```

   It holds the database password, so lock it down:

   ```bash
   sudo chown dnd:dnd /opt/dnd/.env && sudo chmod 600 /opt/dnd/.env
   ```

   `CORS_ORIGINS` must name the real hostname. It gates the websocket `Origin`
   check, so a stale value here means the board silently fails to connect while
   the page itself loads fine.

   `APP_ENV=production` also marks the session cookie `Secure`. Leave it as
   `development` and the cookie travels over plain HTTP; set it while serving
   plain HTTP and nobody can sign in at all.

4. Apply the schema:

   ```bash
   sudo -u dnd /opt/dnd/bin/dnd-migrate up
   ```

5. Install the unit:

   ```bash
   sudo cp deploy/dnd-api.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable --now dnd-api
   ```

6. Add one ingress rule to the tunnel config, pointing the hostname at
   `http://localhost:8080`, and restart `cloudflared`. The tunnel carries the
   websocket upgrade on the same rule; no separate route is needed.

7. Verify:

   ```bash
   curl -s localhost:8080/api/health
   curl -s https://<your-hostname>/api/health
   ```

   Both should return `{"status":"ok"}`.

7. **Register the first account immediately.** While the `users` table is empty,
   registration takes no invite code and creates a DM. Open the app, choose
   **Use an invite**, leave the code field blank, and create your account before
   anyone else can reach the hostname. Once one user exists the door closes and
   every registration needs a code.

8. Confirm the board works: create a campaign, a map and a token, then check
   that the status dot next to the map name is green.

## Updating

```bash
sudo systemctl stop dnd-api
# copy the new bin/ and dist/ into /opt/dnd
sudo -u dnd /opt/dnd/bin/dnd-migrate up
sudo systemctl start dnd-api
sudo systemctl status dnd-api
```

Migrate before starting, never after: the new binary may expect columns the old
schema does not have.

## Rollback

```bash
sudo systemctl stop dnd-api
# restore the previous bin/ and dist/
sudo -u dnd /opt/dnd/bin/dnd-migrate down 1   # only if the update added one
sudo systemctl start dnd-api
```

Keep the previous `bin/` and `dist/` until the new version has been played on
once — that is the whole rollback plan.

`migrate down` is destructive when the migration dropped or rewrote data. If the
release added a column, rolling the binary back without rolling the schema back
is usually the safer move, since the old binary simply ignores the new column.

## Troubleshooting

| Symptom                                       | Check                                                                     |
| --------------------------------------------- | ------------------------------------------------------------------------- |
| Unit will not start                           | `journalctl -u dnd-api -n 50`; usually a bad `DB_DSN` or an unreadable `.env` |
| `/api/health` returns `degraded`              | MySQL is down or the credentials are wrong                                |
| App loads, status dot stays red               | `CORS_ORIGINS` does not match the hostname, so the websocket upgrade is refused |
| Blank page, API responds                      | `WEB_DIR` is unset or points at the wrong directory                       |
| Tokens move locally but others do not see it  | Players are on different maps, or the socket reconnect loop is running — check the dot |
| Signed in, then immediately signed out again  | `APP_ENV=production` while serving plain HTTP, so the browser drops the `Secure` cookie |
| A player cannot drag their token              | It has no `owner_user_id`, or it belongs to someone else; the DM reassigns it |
| Registration rejects every code               | Codes are single-use and expire after seven days; mint a fresh one          |
| Nobody can register and there is no DM        | A user already exists, so the bootstrap is closed. Mint a code as the existing DM, or delete the stray row and register again |
| `Error 1045` at startup                       | MySQL user or password wrong in `DB_DSN`                                  |
| Migration fails halfway                       | MySQL does not roll DDL back; inspect the table, fix it by hand, then re-run |

Logs are structured and go to the journal:

```bash
journalctl -u dnd-api -f
journalctl -u dnd-api --since "10 minutes ago" | grep -i error
```

## Exposure

The service has no rate limiting on `/api/auth/login`. Argon2id makes each
attempt expensive, but nothing caps the attempts. If the hostname is reachable
from the open internet rather than only over the tunnel to people you have
invited, put Cloudflare Access or a rate limit in front of it. The full list of
what the auth layer does not cover is in [auth.md](auth.md#what-this-does-not-do).

## Notes on the unit

[`deploy/dnd-api.service`](../deploy/dnd-api.service) runs as `dnd`, restarts on
failure after 5s, and is sandboxed with `ProtectSystem=strict` plus a single
`ReadWritePaths=/opt/dnd`. If you move the install directory, that line has to
move with it or the service will not be able to write anything.
