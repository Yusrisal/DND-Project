# Character sheets and dice

Sheets are 5e-shaped: six ability scores, HP, AC, speed, proficiency bonus, an
inventory, a spell list and spell slots. Items and spells are free text — there
is no SRD catalogue in this project to reference, so a spell is a name, a level
and whatever description you paste in.

## Who sees what

| | Player | DM |
| --- | --- | --- |
| List sheets | Only their own | Every sheet in the campaign |
| Read, edit, delete a sheet | Their own | Any |
| Reassign a sheet to another account | No | Yes |
| Roll on behalf of a character | Their own | Any |

Ownership is one gate, `CharacterService.authorize`, which every mutating call
goes through. Items and spells resolve to their character first, so editing an
item is exactly as restricted as editing the sheet it belongs to.

Creating a sheet always assigns it to the caller unless the DM says otherwise,
so a player cannot make a character for someone else.

## Ability modifiers

The server sends both the raw scores and the derived modifiers:

```json
{
  "abilities": { "strength": 16, "dexterity": 9, "constitution": 14, "…": 10 },
  "modifiers": { "strength": 3,  "dexterity": -1, "constitution": 2,  "…": 0  }
}
```

`modifiers` is computed from `floor((score - 10) / 2)`. It is sent rather than
left to the client because Go integer division rounds toward zero, so a score of
9 gives `-1` only if the rule is written carefully — and every client agreeing
with the server matters more than saving six numbers. The rule lives in
`model.AbilityModifier`, with the negative half covered by a test.

## Spell slots

One row per spell level, 1 to 9. Cantrips are level 0 spells and have no slots.

Casting is `PUT /api/characters/{id}/slots/{level}` with `used` raised by one,
so the server is the only place the count changes and it refuses `used > total`.

## Rest & Death saves

- **Short Rest**: Spends hit dice (`1d8 + CON` modifier) to heal HP and increments `hit_dice_used`.
- **Long Rest**: Resets HP to full, clears `hp_temp`, resets all spell slots (`used: 0`), resets death saves, and recovers half of maximum hit dice.
- **Death Saves**: Tracked when HP is 0 (3 successes / 3 failures). Natural 20 restores 1 HP; natural 1 adds 2 failures. 3 failures marks the character `dead`.

## Wallet & Currency

Characters track `cp`, `sp`, `ep`, `gp`, and `pp`. Shops support automatic conversion and deduction when purchasing goods. DMs can split party loot across all active characters.

## Dice

Rolls are evaluated on the server. The client sends an expression, never a
result, and the stored breakdown shows every die so a roll can be checked
afterwards rather than taken on faith. `crypto/rand` supplies the numbers, so
there is no question about how the generator was seeded.

### Expressions

A sum of terms, where a term is a flat number or `NdM` with an optional keep
suffix:

| Expression | Meaning                          |
| ---------- | -------------------------------- |
| `1d20+5`   | A d20 plus five                  |
| `d8`       | Count defaults to one            |
| `2d6+1d4+3`| Several dice groups and a modifier |
| `2d20kh1`  | Advantage: roll two, keep highest |
| `2d20kl1`  | Disadvantage: keep lowest        |
| `4d6kh3`   | Ability generation: drop the lowest |
| `1d4-1`    | Negative modifiers work          |

Bounds: at most 20 terms, 100 dice per term, and 2 to 1000 sides. Anything else
is rejected with a message the player can act on, such as
`die must have between 2 and 1000 sides`.

### Rolling

Three ways in, all producing the same broadcast:

- The ability buttons on the sheet, which roll `1d20` plus that modifier and
  label it, for example `STR check`.
- The quick-roll row under the feed: `1d20`, advantage, disadvantage and the
  standard dice.
- `/roll 2d6 sneak attack` or `/r 1d20+5` typed in the chat box. The first word
  after the command is the expression; the rest becomes the label.

Everyone at the table sees every roll, the roller included — there are no hidden
rolls, including for the DM. If you want those, that is a `private` flag on
`dice_rolls` and a second broadcast path, not a config change.

A roll attributed to a character requires owning that sheet; the server checks
before it records anything.

### Result shape

```json
{
  "expression": "2d20kh1+3",
  "groups": [
    { "expression": "2d20kh1", "rolls": [7, 18], "dropped": [7], "total": 18 }
  ],
  "modifier": 3,
  "total": 21
}
```

The panel renders this as the total in large type with the arithmetic beneath,
so a dropped die is visible rather than implied.

## Chat

One feed per campaign, carrying messages and rolls together, ordered by when the
server recorded them. Messages are at most 2000 characters.

The author name is copied from the account at the time of writing and kept on
the row, so history stays readable after an account is deleted — the foreign key
sets `user_id` to NULL but the name remains.

Joining a session delivers the last 50 messages and the last 50 rolls. Further
back is `GET /api/chat?campaign_id=N`, which paginates the usual way.

Chat lives on the campaign rather than the map, so it survives the DM switching
maps mid-session. That is also why the websocket room is keyed by campaign —
see [websocket.md](websocket.md).

## What this does not do

- **No hidden or whispered rolls.** Everything is public to the table.
- **No spell catalogue.** Entries are free text; nothing validates that
  `Fireball` is a level 3 evocation.
- **No encumbrance, skills, saves, conditions or initiative tracking.** Weight is
  recorded per item but nothing sums or enforces it.
- **No editing or deleting a chat message** once it is sent.
- **No automation between sheets and tokens.** A token and a character are
  separate rows; moving one does not touch the other.
