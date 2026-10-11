# HTTP API

Base path `/api`. Requests and responses are JSON.

Every route except health and the auth endpoints needs a session cookie, and the
write routes need the `dm` role. Each section below states its scope; the full
model is in [auth.md](auth.md).

## Conventions

### Pagination, search and sorting

Every list endpoint accepts the same query parameters.

| Parameter   | Default      | Notes                                           |
| ----------- | ------------ | ----------------------------------------------- |
| `page`      | `1`          | 1-based; anything below 1 is ignored            |
| `page_size` | `20`         | Capped at 100                                   |
| `search`    | —            | `LIKE` match on `name`                          |
| `sort_by`   | per resource | Ignored silently unless allowlisted (see below) |
| `sort_dir`  | `desc`       | `asc` or `desc`                                 |

Allowlisted `sort_by` values are `id`, `name`, `created_at` and `updated_at` for
every resource. The default is `created_at` for campaigns and maps, `id` for
tokens.

Every list answers with the same envelope:

```json
{
  "items": [],
  "page": 1,
  "page_size": 20,
  "total": 0,
  "total_pages": 0
}
```

`items` is always an array, never `null`.

### Errors

```json
{ "error": "is required", "code": "validation_failed", "field": "name" }
```

| Status | `code`              | Cause                                            |
| ------ | ------------------- | ------------------------------------------------ |
| 400    | `bad_request`       | Malformed JSON, non-numeric id, missing `map_id` |
| 401    | `unauthorized`      | No session, or it expired                        |
| 403    | `forbidden`         | Signed in, but the route is DM only              |
| 404    | `not_found`         | No such row                                      |
| 422    | `validation_failed` | A rule failed; `field` names the offender        |
| 500    | `internal`          | Logged server-side, details not exposed          |

`field` is present only on 422. A 401 also clears the session cookie, so a stale
cookie does not keep failing.

## Auth

Public, except `GET /api/auth/me`.

| Method | Path                 | Success | Notes                                 |
| ------ | -------------------- | ------- | ------------------------------------- |
| POST   | `/api/auth/register` | 201     | Claims an invite code and signs in    |
| POST   | `/api/auth/login`    | 200     | Sets the session cookie               |
| POST   | `/api/auth/logout`   | 204     | Revokes the session, clears the cookie |
| GET    | `/api/auth/me`       | 200     | Current account; 401 when signed out  |

All four answer with the `User` object, which never carries a password hash:

```json
{
  "id": 1,
  "email": "dm@example.com",
  "display_name": "Yus",
  "role": "dm",
  "created_at": "2026-10-09T14:02:11Z",
  "updated_at": "2026-10-09T14:02:11Z"
}
```

**Register body**: `code`, `email`, `display_name`, `password`. The password must
be at least 10 characters. `code` is required unless the database has no users
yet, in which case the first account is created as the DM — see
[auth.md](auth.md#the-first-dm).

**Login body**: `email` and `password`. A wrong password and an unknown address
both answer 401 with the same message and take the same time.

Registration signs the new account in, so the response already carries a
session cookie.

## Users

Any signed-in account.

| Method | Path         | Success |
| ------ | ------------ | ------- |
| GET    | `/api/users` | 200     |

Paginated the usual way, searching on display name and email. It backs the owner
picker: the DM needs names to assign tokens to.

## Invites

DM only.

| Method | Path                   | Success |
| ------ | ---------------------- | ------- |
| GET    | `/api/invites`         | 200     |
| POST   | `/api/invites`         | 201     |
| DELETE | `/api/invites/{code}`  | 204     |

```bash
curl -s localhost:8080/api/invites -b cookies.txt -H "Content-Type: application/json" -d "{\"role\":\"player\"}"
```

```json
{
  "code": "q7fk3mwptz",
  "role": "player",
  "created_by": 1,
  "used_by": null,
  "used_at": null,
  "expires_at": "2026-10-16T14:02:11Z",
  "created_at": "2026-10-09T14:02:11Z"
}
```

**Body**: `role` is `dm` or `player`, defaulting to `player`. Codes are
single-use and expire after seven days. `DELETE` revokes one that has not been
spent.

## Health

```
GET /api/health
```

Returns `200 {"status":"ok"}`, or `503 {"status":"degraded"}` when the database
ping fails. Suitable as a systemd or tunnel health probe.

## Campaigns

| Method | Path                               | Success | Notes                                |
| ------ | ---------------------------------- | ------- | ------------------------------------ |
| GET    | `/api/campaigns`                   | 200     |                                      |
| POST   | `/api/campaigns`                   | 201     |                                      |
| GET    | `/api/campaigns/{id}`              | 200     |                                      |
| PUT    | `/api/campaigns/{id}`              | 200     |                                      |
| PUT    | `/api/campaigns/{id}/pause`        | 200     | Toggle live session pause & message  |
| PUT    | `/api/campaigns/{id}/retention`    | 200     | Set lingering visibility window (days)|
| DELETE | `/api/campaigns/{id}`              | 204     |                                      |
| GET    | `/api/campaigns/{id}/savepoints`   | 200     | List session checkpoints             |
| POST   | `/api/campaigns/{id}/savepoints`   | 201     | Create checkpoint snapshot           |
| POST   | `/api/savepoints/{id}/restore`     | 200     | Revert game state back to snapshot   |
| DELETE | `/api/savepoints/{id}`             | 204     | Delete savepoint                     |

Reading is open to any signed-in account; `POST`, `PUT` and `DELETE` are DM
only. Deleting a campaign cascades to its maps and their tokens.

```bash
curl -s "localhost:8080/api/campaigns?search=curse&sort_by=name&sort_dir=asc"
```

```bash
curl -s localhost:8080/api/campaigns -H "Content-Type: application/json" -d "{\"name\":\"Curse of Strahd\",\"description\":\"Barovia\"}"
```

```json
{
  "id": 1,
  "name": "Curse of Strahd",
  "description": "Barovia",
  "created_at": "2026-10-09T14:02:11Z",
  "updated_at": "2026-10-09T14:02:11Z"
}
```

**Body** (`POST`, `PUT`): `name` is required and at most 120 characters;
`description` is optional. Both are trimmed.

## Maps

| Method | Path                           | Success |
| ------ | ------------------------------ | ------- |
| GET    | `/api/maps`                    | 200     |
| POST   | `/api/maps`                    | 201     |
| GET    | `/api/maps/{id}`               | 200     |
| PUT    | `/api/maps/{id}`               | 200     |
| PUT    | `/api/maps/{id}/fog`           | 200     |
| DELETE | `/api/maps/{id}`               | 204     |
| GET    | `/api/maps/{id}/tokens`        | 200     |
| GET    | `/api/maps/{id}/aoe-templates` | 200     |
| POST   | `/api/maps/{id}/aoe-templates` | 201     |
| DELETE | `/api/aoe-templates/{id}`      | 204     |

Reading is open to any signed-in account; writing is DM only.
`GET /api/maps` filters by campaign with `?campaign_id=1`.

`GET /api/maps/{id}/tokens` returns the whole board as a bare array, not a page
— a map holds few enough tokens that paginating it only adds round trips. Use
`GET /api/tokens?map_id=1` when you do want a page.

```bash
curl -s localhost:8080/api/maps -H "Content-Type: application/json" -d "{\"campaign_id\":1,\"name\":\"Village of Barovia\"}"
```

**Body**: `campaign_id` and `name` are required. `image_url` is optional.
`grid_size`, `width` and `height` fall back to `50`, `1920` and `1080` when
omitted or non-positive. `campaign_id` is ignored on `PUT` — a map does not move
between campaigns.

## Uploads

| Method | Path            | Success |
| ------ | --------------- | ------- |
| POST   | `/api/uploads`  | 201     |
| GET    | `/uploads/{name}` | 200   |

Posting is DM only; reading is open to any signed-in account. Note that the
files are served from `/uploads/`, outside the `/api` prefix.

The request is `multipart/form-data` with one `file` field, at most 8 MB. The
kind is sniffed from the first bytes and must be PNG, JPEG, GIF or WebP — the
filename and the `Content-Type` header are both ignored, so renaming a script
to `.png` is rejected with a 422 and nothing is written.

```json
{ "url": "/uploads/3f2a…4c.png" }
```

Store that `url` in an `image_url` field. Files are written to `UPLOAD_DIR`,
which in production must sit outside the release directory or the next deploy
deletes them. See [deployment.md](deployment.md).

## Tokens

| Method | Path               | Success | Broadcasts      |
| ------ | ------------------ | ------- | --------------- |
| GET    | `/api/tokens`      | 200     | —               |
| POST   | `/api/tokens`      | 201     | `token.created` |
| GET    | `/api/tokens/{id}` | 200     | —               |
| PUT    | `/api/tokens/{id}` | 200     | `token.updated` |
| DELETE | `/api/tokens/{id}` | 204     | `token.deleted` |

Reading is open to any signed-in account; creating, editing and deleting are DM
only. Players move their own tokens over the websocket — see
[websocket.md](websocket.md).

`GET /api/tokens` filters by map with `?map_id=1`.

Mutations are broadcast to everyone watching that map, so a change made over
HTTP reaches other players exactly like one made over the socket. See
[websocket.md](websocket.md).

```bash
curl -s localhost:8080/api/tokens -H "Content-Type: application/json" -d "{\"map_id\":1,\"name\":\"Ireena\",\"color\":\"#2980b9\",\"size\":1,\"x\":125,\"y\":75}"
```

**Body**: `map_id` and `name` are required. `owner_user_id` assigns the token to
a player and may be `null` for a monster or an NPC, which only the DM can move.
`color` defaults to `#c0392b` and `size` to `1`, measured in grid cells. `x` and
`y` are the pixel coordinates of the token centre and default to `0`.

`character_id` links the token to a sheet, and it changes where the vitals live.
A linked token does not own its hit points: `hp_current`, `hp_max` and
`armor_class` are read from the character on the way out and the columns on the
token row are ignored, so damage dealt on the board and healing applied on the
sheet are the same number. An unlinked token keeps its own. A `character_id`
that matches no sheet is a 422 rather than a silently unlinked token.

The two fields answer different questions and are set independently:
`owner_user_id` decides **who may drag it**, `character_id` decides **whose hit
points it spends**. A player's own figure normally sets both.

Position is also writable through `PUT`, but a drag should use the socket — it
is one frame instead of a round trip, and it skips the echo back to the mover.

## Characters

Any signed-in account. What comes back depends on the caller: the DM reaches
every sheet in the campaign, a player only their own. See
[characters-and-dice.md](characters-and-dice.md).

| Method | Path                                    | Success | Notes                          |
| ------ | --------------------------------------- | ------- | ------------------------------ |
| GET    | `/api/characters`                       | 200     | Filter with `?campaign_id=`    |
| POST   | `/api/characters`                       | 201     | Assigned to the caller unless the DM says otherwise |
| GET    | `/api/characters/{id}`                  | 200     | The whole sheet, see below     |
| PUT    | `/api/characters/{id}`                  | 200     |                                |
| DELETE | `/api/characters/{id}`                  | 204     | Cascades to items and spells   |
| POST   | `/api/characters/{id}/long-rest`        | 200     | Reset HP, slots, death saves   |
| POST   | `/api/characters/{id}/short-rest`       | 200     | Spend hit dice, heal HP        |
| POST   | `/api/characters/{id}/death-save`       | 200     | Record d20 death save roll     |
| POST   | `/api/characters/{id}/items`            | 201     |                                |
| POST   | `/api/characters/{id}/spells`           | 201     |                                |
| PUT    | `/api/characters/{id}/slots/{level}`    | 200     | Returns every slot row         |
| PUT    | `/api/characters/{id}/proficiencies`    | 200     | Returns the whole sheet        |
| PUT    | `/api/items/{id}`                       | 200     |                                |
| DELETE | `/api/items/{id}`                       | 204     |                                |
| PUT    | `/api/spells/{id}`                      | 200     |                                |
| DELETE | `/api/spells/{id}`                      | 204     |                                |

`GET /api/characters/{id}` answers with the sheet in one object, which is the
single call the character panel makes:

```json
{
  "character": { "id": 3, "name": "Ireena", "level": 4,
                 "abilities": { "strength": 16, "…": 10 },
                 "modifiers": { "strength": 3, "…": 0 } },
  "items":  [{ "id": 9, "name": "Rapier", "quantity": 1, "equipped": true }],
  "spells": [{ "id": 4, "name": "Shield", "level": 1, "prepared": true }],
  "slots":  [{ "level": 1, "total": 4, "used": 1 }]
}
```

`modifiers` is derived from `abilities` by the server so clients cannot disagree
about the rounding rule.

**Character body**: `campaign_id` and `name` are required. `level` defaults to
1 and is capped at 20, ability scores run 1 to 30, `armor_class` defaults to 10
and `speed` to 30. `proficiency_bonus` is derived from the level when omitted.
`user_id` is honoured only for the DM.

**Item body**: `name` is required; `quantity` defaults to 1 and `weight` cannot
be negative. An item with a `damage` expression is a weapon it can attack with,
alongside `attack_bonus` and `damage_type`.

**Proficiency body**: `kind` (`skill` or `save`), `name`, `expertise` and
`trained`. The response carries the recalculated `skills` and `saves`.

**Spell body**: `name` is required and `level` runs 0 to 9, where 0 is a cantrip.

**Slot body**: `total` and `used`, where `used` may not exceed `total`. Casting
is this call with `used` raised by one; a long rest is `used: 0`.

Items and spells resolve to their character before the permission check, so
`PUT /api/items/{id}` is exactly as restricted as editing the sheet it hangs off.

## Compendium & Environments

| Method | Path                           | Scope | Notes                               |
| ------ | ------------------------------ | ----- | ----------------------------------- |
| GET    | `/api/environments`            | any   | Filter with `?campaign_id=`         |
| GET    | `/api/environments/{id}`       | any   | Environment hazard details          |
| POST   | `/api/environments`            | DM    | Create hazard, weather, or trap     |
| PUT    | `/api/environments/{id}`       | DM    | Update hazard                       |
| DELETE | `/api/environments/{id}`       | DM    | Delete hazard                       |
| GET    | `/api/compendium/items`        | any   | Global / campaign item catalogue    |
| GET    | `/api/compendium/items/{id}`   | any   | Item details                        |
| POST   | `/api/compendium/items`        | DM    | Create compendium item              |
| PUT    | `/api/compendium/items/{id}`   | DM    | Update compendium item              |
| DELETE | `/api/compendium/items/{id}`   | DM    | Delete compendium item              |
| GET    | `/api/compendium/spells`       | any   | Global / campaign spell catalogue   |
| GET    | `/api/compendium/spells/{id}`  | any   | Spell details                       |
| POST   | `/api/compendium/spells`       | DM    | Create spell/skill in library       |
| PUT    | `/api/compendium/spells/{id}`  | DM    | Update spell/skill                  |
| DELETE | `/api/compendium/spells/{id}`  | DM    | Delete spell/skill                  |
| GET    | `/api/monsters/{id}/drops`     | any   | Monster loot table                  |
| POST   | `/api/monsters/{id}/drops`     | DM    | Add drop rule (% rate, qty, coins)  |
| DELETE | `/api/monster-drops/{id}`      | DM    | Delete drop rule                    |
| POST   | `/api/monsters/{id}/roll-loot` | DM    | Roll drop table on demand           |

## Shops and Loot

| Method | Path                         | Scope | Notes                               |
| ------ | ---------------------------- | ----- | ----------------------------------- |
| GET    | `/api/shops`                 | any   | Filter with `?campaign_id=`         |
| GET    | `/api/shops/{id}`            | any   | Details with items for sale         |
| POST   | `/api/shops`                 | DM    | Create merchant / shop              |
| PUT    | `/api/shops/{id}`            | DM    | Edit merchant / shop                |
| DELETE | `/api/shops/{id}`            | DM    | Delete shop                         |
| POST   | `/api/shops/{id}/items`      | DM    | Add ware to shop                    |
| PUT    | `/api/shop-items/{id}`       | DM    | Edit ware                           |
| DELETE | `/api/shop-items/{id}`       | DM    | Delete ware                         |
| POST   | `/api/shop-items/{id}/buy`   | any   | Buy ware with character coins       |
| POST   | `/api/loot/split`            | DM    | Split coin loot across active party |

## Monsters and combat

Reading the bestiary is open to any signed-in account; writing it, spawning and
running the fight are DM only. Attacks go over the websocket rather than here.
The whole model is in [combat.md](combat.md).

| Method | Path                                                     | Scope |
| ------ | -------------------------------------------------------- | ----- |
| GET    | `/api/monsters`                                          | any   |
| GET    | `/api/monsters/{id}`                                     | any   |
| POST   | `/api/monsters`                                          | DM    |
| PUT    | `/api/monsters/{id}`                                     | DM    |
| DELETE | `/api/monsters/{id}`                                     | DM    |
| PUT    | `/api/monsters/{id}/proficiencies`                       | DM    |
| POST   | `/api/monsters/{id}/actions`                             | DM    |
| PUT    | `/api/actions/{id}`                                      | DM    |
| DELETE | `/api/actions/{id}`                                      | DM    |
| POST   | `/api/spawn`                                             | DM    |
| GET    | `/api/maps/{id}/encounter`                               | any   |
| POST   | `/api/maps/{id}/encounter`                               | DM    |
| DELETE | `/api/maps/{id}/encounter`                               | DM    |
| POST   | `/api/maps/{id}/encounter/turn`                          | DM    |
| POST   | `/api/maps/{id}/encounter/combatants`                    | DM    |
| PUT    | `/api/maps/{id}/encounter/combatants/{combatant}`        | DM    |
| DELETE | `/api/maps/{id}/encounter/combatants/{combatant}`        | DM    |
| PUT    | `/api/tokens/{id}/conditions`                            | DM    |

`GET /api/monsters/{id}` answers with the stat block, its skills and saves, and
its actions in one object.

**Spawn body**: `monster_id`, `map_id`, `count` (at most 20) and `roll_hp`.
It answers with the created tokens and broadcasts each one.

**Combatant body**: `token_id`, `name` and `initiative`. An initiative of zero
is rolled as `1d20` plus the dexterity modifier of the linked sheet.

**Condition body**: `condition` and `active`, checked against the 5e list.

## Chat

Any signed-in account.

| Method | Path        | Success | Notes                       |
| ------ | ----------- | ------- | --------------------------- |
| GET    | `/api/chat` | 200     | Requires `?campaign_id=`    |

This is the scrollback. Live messages and rolls arrive on the websocket, which
also delivers the last 50 of each on connect — see
[websocket.md](websocket.md). Rolls are made over the socket, not here.

## Handouts and Lore

| Method | Path                          | Scope | Notes                             |
| ------ | ----------------------------- | ----- | --------------------------------- |
| GET    | `/api/campaigns/{id}/handouts`| any   | Player sees only revealed items   |
| GET    | `/api/handouts/{id}`          | any   | Fetch single handout              |
| POST   | `/api/campaigns/{id}/handouts`| DM    | Create lore/handout/clue          |
| PUT    | `/api/handouts/{id}`          | DM    | Edit or toggle `is_revealed`      |
| DELETE | `/api/handouts/{id}`          | DM    | Delete handout                    |
