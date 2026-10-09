# Database and migrations

MySQL 8, `utf8mb4_unicode_ci`, InnoDB. The service talks to it through
`database/sql` and hand-written queries — there is no ORM.

## Schema

```
users ──1:N──► sessions
  │  └─1:N──► invites (created_by)
  └──────────────────────────┐
                             │ owner_user_id (nullable)
campaigns ──1:N──► maps ──1:N──► tokens ──1:N──► token_conditions
    │                  │           │
    │                  │           ├──► characters  (character_id, nullable)
    │                  │           └──► monsters    (monster_id, nullable)
    │                  └─1:N──► encounters ──1:N──► combatants
    ├─1:N──► chat_messages
    ├─1:N──► dice_rolls
    ├─1:N──► shops ──1:N──► shop_items
    └─1:N──► characters ──1:N──► character_items
                        ├─1:N──► character_spells
                        ├─1:N──► character_spell_slots
                        └─1:N──► character_proficiencies

monsters ──1:N──► monster_actions
         ├─1:N──► monster_proficiencies
         └─1:N──► monster_drops

campaigns ──1:N──► environments
          ├─1:N──► compendium_items
          └─1:N──► compendium_spells
```

### `campaigns`

| Column        | Type           | Notes                       |
| ------------- | -------------- | --------------------------- |
| `id`          | BIGINT UNSIGNED | Primary key, auto-increment |
| `name`        | VARCHAR(120)   | Indexed, searched           |
| `description` | TEXT           |                             |
| `created_at`  | DATETIME       | Defaults to now             |
| `updated_at`  | DATETIME       | Updated by MySQL on write   |

### `users`

| Column          | Type            | Notes                            |
| --------------- | --------------- | -------------------------------- |
| `id`            | BIGINT UNSIGNED | Primary key                      |
| `email`         | VARCHAR(191)    | Unique, stored lowercased        |
| `display_name`  | VARCHAR(80)     | Shown on tokens and the roster   |
| `password_hash` | VARCHAR(255)    | Argon2id, PHC format             |
| `role`          | ENUM            | `dm` or `player`, default `player` |

### `sessions`

| Column       | Type            | Notes                                   |
| ------------ | --------------- | --------------------------------------- |
| `token_hash` | CHAR(64)        | Primary key: SHA-256 of the cookie value |
| `user_id`    | BIGINT UNSIGNED | FK to `users`, cascade delete           |
| `expires_at` | DATETIME        | Indexed; the lookup filters on it       |

The raw token is never stored. Expired rows are ignored by the lookup and swept
every six hours.

### `invites`

| Column       | Type            | Notes                                |
| ------------ | --------------- | ------------------------------------ |
| `code`       | VARCHAR(32)     | Primary key, ten characters          |
| `role`       | ENUM            | The role the code grants             |
| `created_by` | BIGINT UNSIGNED | FK to `users`, cascade delete        |
| `used_by`    | BIGINT UNSIGNED | NULL until spent; FK set null        |
| `used_at`    | DATETIME        | NULL until spent                     |
| `expires_at` | DATETIME        | Seven days out                       |

`used_by IS NULL` is the single-use guard: claiming is an `UPDATE` with that
condition, so two registrations racing on one code cannot both win.

### `campaigns`

| Column              | Type            | Notes                            |
| ------------------- | --------------- | -------------------------------- |
| `id`                | BIGINT UNSIGNED | Primary key                      |
| `name`              | VARCHAR(120)    |                                  |
| `description`       | TEXT            |                                  |
| `is_paused`         | TINYINT(1)      | Session pause / standby state    |
| `pause_message`     | VARCHAR(255)    | Banner text shown to players     |
| `chat_retention_days`| INT            | 0 = all time, 1 = 24h, 7 = 1w, etc|

### `campaign_savepoints`

Game state snapshots/checkpoints. Stores captured token positions, character HP/vitals, and active encounter state for quick restoring.

### `maps`

| Column        | Type            | Notes                            |
| ------------- | --------------- | -------------------------------- |
| `id`          | BIGINT UNSIGNED | Primary key                      |
| `campaign_id` | BIGINT UNSIGNED | FK to `campaigns`, cascade delete |
| `name`        | VARCHAR(120)    |                                  |
| `image_url`   | VARCHAR(512)    | Empty string when unset          |
| `grid_size`   | INT             | Pixels per cell, default 50      |
| `width`       | INT             | Default 1920                     |
| `height`      | INT             | Default 1080                     |
| `fog_enabled` | TINYINT(1)      | Toggle for Fog of War shroud     |
| `fog_revealed`| LONGTEXT        | JSON array of revealed box coords |

### `tokens`

| Column      | Type            | Notes                        |
| ----------- | --------------- | ---------------------------- |
| `id`            | BIGINT UNSIGNED | Primary key                              |
| `map_id`        | BIGINT UNSIGNED | FK to `maps`, cascade delete             |
| `owner_user_id` | BIGINT UNSIGNED | NULL for monsters; FK to `users`, set null |
| `character_id`  | BIGINT UNSIGNED | NULL unless this token is a player character; the sheet then owns hp and AC |
| `monster_id`    | BIGINT UNSIGNED | NULL unless spawned from the bestiary     |
| `hp_current` / `hp_max` / `armor_class` | INT | Used only when there is no linked character |
| `name`      | VARCHAR(120)    |                              |
| `image_url` | VARCHAR(512)    | Empty string when unset      |
| `color`     | VARCHAR(16)     | Hex, default `#c0392b`       |
| `size`      | INT             | In grid cells, default 1     |
| `x`, `y`    | DOUBLE          | Pixel centre of the token    |

Deleting a campaign removes its maps and their tokens; the cascades are enforced
by the foreign keys, not by application code. Deleting a user leaves their
tokens in place, unowned, so a departing player does not wipe the board.

### `characters`

The 5e sheet: `name`, `status` (`active`, `tavern`, `dead`, `retired`), `class`, `race`, `level`, `hp_current` / `hp_max` /
`hp_temp`, `armor_class`, `speed`, `proficiency_bonus`, currency (`cp`, `sp`, `ep`, `gp`, `pp`),
the six ability scores as their own columns, `notes`, death saves (`death_save_successes`, `death_save_failures`),
hit dice (`hit_dice_total`, `hit_dice_used`, `hit_dice_die`), and damage modifiers (`damage_resistances`,
`damage_vulnerabilities`, `damage_immunities`, `condition_immunities`). `campaign_id` cascades on delete;
`user_id` is nullable and set to NULL when the account goes, so the DM keeps the
sheet.

Ability modifiers are derived rather than stored — see
[characters-and-dice.md](characters-and-dice.md#ability-modifiers).

### `character_items` and `character_spells`

Free-text rows hanging off a character, both cascading on delete. Items carry
`quantity`, `weight`, `cost_cp` and `equipped`; spells carry `level` (0 is a cantrip),
`school` and `prepared`.

### `character_spell_slots`

`(character_id, level)` as the primary key, with `total` and `used`. Written
through an upsert, so the caller never has to know whether the row exists.

### `chat_messages`

| Column        | Type            | Notes                                     |
| ------------- | --------------- | ----------------------------------------- |
| `campaign_id` | BIGINT UNSIGNED | FK, cascade delete                        |
| `user_id`     | BIGINT UNSIGNED | NULL once the account goes                |
| `author_name` | VARCHAR(80)     | Copied at write time so history stays readable |
| `body`        | VARCHAR(2000)   |                                           |

Indexed on `(campaign_id, id)`, which is what both the tail query and the
paginated history walk.

### `dice_rolls`

The same shape plus `character_id`, `label`, `expression`, `total`, and `detail`
as JSON holding the per-die breakdown the server produced. Keeping the
breakdown is what makes a result checkable after the fact.

### `monsters`

The bestiary: a template, not an instance. Size, type, alignment, AC, average
hit points and a hit point formula, speed, the six abilities, challenge rating,
senses, languages and traits, plus the colour and size its tokens spawn with.

### `monster_proficiencies` and `monster_actions`

Skills and saves are stored as **flat modifiers**, the way a stat block prints
them, unlike `character_proficiencies` which stores a trained flag and derives
the number. An action with a damage expression is an attack; one without is a
trait. Both cascade on delete.

### `encounters` and `combatants`

One active encounter per map, holding `round`, `turn_index` and `active`.
Ending a fight clears the flag rather than deleting the row. Combatants carry an
optional `token_id`, a name and an initiative, indexed on
`(encounter_id, initiative)` which is the order they are read in.

### `token_conditions`

`(token_id, condition_name)` as the primary key, cascading on delete. Names are
validated against the 5e list before they are written.

### `character_proficiencies`

`(character_id, kind, name)` with an `expertise` flag, where `kind` is `skill`
or `save`. The modifier itself is derived, never stored.

### `shops` and `shop_items`

Merchants and items for sale per campaign. `shops` holds `name` and `description`. `shop_items`
carries `name`, `cost_cp`, `quantity` (-1 for unlimited), `weight`, weapon stats and `description`.
Purchasing deducts currency from the character and adds the item to `character_items`.

### `environments`

Weather, hazards, difficult terrain, and traps per campaign/map. Holds `difficulty_dc`, `damage`, `damage_type`, and `effect_description`.

### `compendium_items` and `compendium_spells`

Compendium catalogues for equipment, magic items, weapons, armor, spells, and special skills.

### `monster_drops`

Loot tables per monster: item drops with `drop_rate` (1-100%), `min_quantity`, `max_quantity`, and coin drops (`cp`, `sp`, `gp`).

### `aoe_templates`

Persistent or interactive spell & hazard area templates on maps (`circle`, `cone`, `line`, `cube`) with size in feet and custom color.

### `handouts`

Campaign lore, letters, clues, and image handouts. DM toggles `is_revealed` to make visible to players.

### `schema_migrations`

Bookkeeping for the migration runner: one row per applied migration, holding
`version`, `name` and `applied_at`.

## Running migrations

Migration files live in `backend/migrations` and are embedded in both binaries,
so a deploy carries its own schema.

```bash
make migrate-up       # apply everything pending
make migrate-down     # roll back the newest migration
make migrate-status   # current version and pending count
```

Directly, without `make`:

```bash
cd backend
go run ./cmd/migrate up
go run ./cmd/migrate down 2
go run ./cmd/migrate status
```

In production, use the compiled runner instead:

```bash
/opt/dnd/bin/dnd-migrate up
```

## Adding a migration

Files are named `NNNNNN_description.up.sql` and `NNNNNN_description.down.sql` —
the layout golang-migrate uses, so its CLI stays an option if this runner is
ever outgrown.

1. Create the next numbered pair, e.g. `000004_add_token_initiative.up.sql` and
   its `.down.sql`.
2. Write forward DDL in the `up` file and the exact inverse in the `down` file.
   A migration without a `down` file fails the rollback instead of being
   skipped.
3. Run `make migrate-up`, then `make migrate-down` and `make migrate-up` again
   to prove the rollback works.

Rules worth keeping:

- **Never edit an applied migration.** Its version is already recorded; the
  change would not be picked up anywhere it has run. Add a new one.
- **Multiple statements per file are fine.** The runner splits on `;` and
  executes them in order.
- **MySQL does not roll DDL back in a transaction.** A migration that fails
  halfway leaves the earlier statements applied. Keep each migration small, and
  prefer one DDL statement per file.
- **There is no seeder,** by design. Create fixtures through the API.

## Adding a resource

The layers each take one small, mechanical addition. Using an `encounters`
table as the example:

1. **Migration** — `000004_create_encounters.up.sql` plus its `.down.sql`.
2. **Model** — `internal/model/encounter_model.go` with `Encounter` and
   `EncounterInput`.
3. **Repository** — `internal/repository/encounter_repository.go` with
   `List`, `Get`, `Create`, `Update`, `Delete`, plus an `EncounterSortColumns()`
   allowlist. Copy the shape from `campaign_repository.go`; `List` returns
   `(items, total, error)`.
4. **Service** — validation and defaults in a `normalizeEncounter` helper, and
   `sql.ErrNoRows` translated to `service.ErrNotFound`.
5. **Handler** — thin methods that parse, call the service, and write JSON.
6. **Router** — five `mux.HandleFunc` lines in `internal/handler/router.go`.

If the resource should sync live, broadcast from the handler the way
`token_handler.go` does. Do not reach for `internal/ws` from a service.
