# Combat, monsters and conditions

What the software decides, and what stays with the DM.

| Resolved by the server | Still yours |
| --- | --- |
| Initiative order and whose turn it is | Narrating the scene |
| Attack rolls against armour class | Range, cover, line of sight |
| Critical hits and misses | Reactions, opportunity attacks, movement rules |
| Damage applied to hit points | Death saves, concentration, spell effects |
| Who is allowed to attack with what | Which conditions a given effect imposes |

Everything below happens over the websocket and lands in the shared feed, so
nobody has to read out numbers.

## The bestiary

A monster is a **template**, not a token. The DM writes the stat block once, and
spawning copies its vitals onto each token, so six goblins take damage
separately while sharing one block for their actions.

A stat block holds the usual: size, type, alignment, AC, average hit points and
the hit point formula, speed, the six abilities, challenge rating, senses,
languages and free-text traits. It also carries the colour and size its tokens
spawn with.

**Skills and saves are flat modifiers** here, not proficiency flags — a stat
block prints `Perception +4`, and that is what gets stored. Character sheets
work the other way round, deriving the modifier from proficiency; see
[characters-and-dice.md](characters-and-dice.md).

**Actions** are rows with a name, an attack bonus, a damage expression and a
damage type. An action with a damage expression is an attack the DM can make;
one without is a trait to read out. Damage expressions are validated when the
stat block is written, so a bad one fails at the desk rather than mid-fight.

### Spawning

```bash
curl -s localhost:8080/api/spawn -b cookies.txt -H "Content-Type: application/json" \
  -d "{\"monster_id\":3,\"map_id\":1,\"count\":6,\"roll_hp\":true}"
```

Up to 20 at a time. With `roll_hp` the hit point formula is rolled per token, so
each goblin differs; without it they all take the average. Several spawns are
numbered (`Goblin 1`, `Goblin 2`) so the initiative order can tell them apart,
and staggered across the grid rather than stacked in one pile.

Each spawned token is broadcast as it appears, so the party watches them arrive.

## Initiative

One encounter per map, started and ended by the DM. Ending it clears the flag
rather than deleting the row, so the fight stays in the record.

```
POST   /api/maps/{id}/encounter                        start
POST   /api/maps/{id}/encounter/turn                   advance
POST   /api/maps/{id}/encounter/combatants             add
PUT    /api/maps/{id}/encounter/combatants/{combatant} set initiative
DELETE /api/maps/{id}/encounter/combatants/{combatant} drop
DELETE /api/maps/{id}/encounter                        end
```

Adding a combatant with an initiative of zero **rolls it**: `1d20` plus the
dexterity modifier of the linked character sheet, or flat for anything without
one. Adding a token with no name borrows the token name. The DM can also add a
bare name with no token, for a lair action or something off the board.

The order sorts by initiative descending, with the id as the tiebreak so it
stays stable across reads. Advancing past the last combatant rolls into the next
round.

Dropping a combatant keeps the pointer on the same creature, so removing a dead
monster does not skip whoever was next.

Every change broadcasts the whole encounter, so each screen shows the same turn
without polling.

## Attacks

Two sources, one resolution path:

| Source | Who may use it | Fields |
| --- | --- | --- |
| A character weapon | Whoever owns that sheet | `item_id` |
| A monster action | The DM only | `action_id`, optionally `attacker_token_id` |

An inventory item **is a weapon** when it has a damage expression. That is the
only thing that marks it, which keeps the inventory one list rather than two
that can disagree.

Sent over the socket:

```js
socket.send(JSON.stringify({
  type: "attack",
  payload: { item_id: 9, target_token_id: 14, advantage: true },
}))
```

The server then:

1. Resolves and authorizes the source. A weapon means owning its sheet; an
   action means being the DM, and `attacker_token_id` must be a token that
   actually uses that stat block.
2. Rolls `1d20` plus the attack bonus — `2d20kh1` with advantage, `2d20kl1` with
   disadvantage, and the two cancel when both are set.
3. Compares against the target armour class. A natural 20 always hits, a natural
   1 always misses.
4. On a hit, rolls damage and calculates modifiers. Damage resistance halves the damage (`/ 2`),
   vulnerability doubles it (`* 2`), and immunity reduces it to `0`.
   **A critical doubles the dice, not the modifier**, which is the 5e rule: `1d8+3` becomes `2d8+3`,
   so 5 to 19 rather than 22. There is a test for exactly this.
   Damage absorbs `hp_temp` first before reducing `hp_current`.
5. Broadcasts the whole resolution and the updated token.

### Where the hit points live

`tokens.character_id` and `tokens.monster_id` are both nullable, and they decide
where damage goes:

- **Linked to a character** — the sheet owns hit points and armour class, and
  the token columns are ignored. Damage lands on the sheet.
- **Linked to a monster, or nothing** — the token carries its own vitals.
  Damage lands on the token.

This is what lets several goblins share one stat block and still die one at a
time.

Hit points are clamped between zero and the maximum by the `UPDATE` itself, so
nothing goes negative and no race can push a creature past full health.

### Result shape

An attack is stored as a roll whose detail carries the whole resolution, tagged
`"kind": "attack"`. It therefore lands in the same feed as every other roll and
survives a reconnect the same way.

```json
{
  "kind": "attack",
  "attacker_name": "Ireena", "target_name": "Goblin 2", "weapon_name": "Rapier",
  "attack_roll": { "expression": "1d20+6", "total": 17 },
  "target_ac": 15, "hit": true, "critical": false, "fumble": false,
  "damage": { "expression": "1d8+4", "total": 9 }, "damage_type": "piercing",
  "target": { "hp_current": 2, "hp_max": 11, "armor_class": 15 }
}
```

The panel renders that as a card: who hit what, the roll against the AC, the
damage, and where the target is now.

## Conditions

The 5e list, validated server-side so the board cannot collect typos:

> blinded, charmed, deafened, exhaustion, frightened, grappled, incapacitated,
> invisible, paralyzed, petrified, poisoned, prone, restrained, stunned,
> unconscious

```
PUT /api/tokens/{id}/conditions   { "condition": "prone", "active": true }
```

DM only. The updated token is broadcast, so markers appear on every board at
once. Conditions are **labels** — nothing enforces what being `prone` does to a
roll. That stays at the table.

## On the board

Tokens show a health bar once they have a maximum, coloured green through amber
to red, and go translucent with a cross at zero hit points. Conditions print
above as abbreviations. The targeted token wears a dashed ring; click a token to
target it, click empty canvas to clear.

Targeting is what the attack buttons act on: pick a target, then press the
attack button next to a weapon on your sheet, or next to a monster action in the
bestiary.

## Skills and saves

Character sheets carry proficiencies, and the server derives every skill and
save modifier from them:

```
PUT /api/characters/{id}/proficiencies   { "kind": "skill", "name": "stealth", "trained": true }
```

The modifier is the ability modifier, plus the proficiency bonus when trained,
doubled with expertise. Derivation happens server-side so no client reimplements
the rule and then disagrees about it. The sheet renders a roll button per line.

The eighteen skills and their abilities are in
[`internal/model/rules_model.go`](../backend/internal/model/rules_model.go).

## What this does not do

- **No movement rules.** Nothing checks range, speed, cover or line of sight.
  The grid is there to look at.
- **No reactions or opportunity attacks.**
- **No death saves, concentration or exhaustion levels.** A creature at zero hit
  points is marked down and left there.
- **No spell effects.** Casting spends a slot; what the spell *does* is
  narrated, and damage goes through an ordinary roll.
- **No saving throws against anything.** A save is a roll button; comparing it
  to a DC is the DM reading the number.
- **No conditions that modify rolls.** They are labels.
- **No automatic initiative at the start of combat** — the DM presses add all.
- **No SRD monster data.** The bestiary starts empty and you fill it in.
