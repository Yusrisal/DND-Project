/**
 * The only hand-written content in the wiki. Everything else on the page is
 * read from the API, so a monster or item the DM edits is never duplicated
 * here. These rules describe what this server actually does — the numbers come
 * from `service/dice.go`, `service/combat_service.go`,
 * `service/progression_rules.go` and `service/shop_repository.go`, so they are
 * worth re-checking if those change.
 */

export type GlossaryEntry = {
  term: string
  tag: string
  body: string
}

/**
 * The fifteen conditions the rules themselves define. The `status_effects`
 * table starts empty and only holds what a DM adds for their own campaign, so
 * without these the conditions page would be blank on a fresh database.
 */
export const CORE_CONDITIONS: { name: string; body: string }[] = [
  { name: 'Blinded', body: 'You cannot see, and automatically fail any check that needs sight. Attacks against you have advantage, and your own attacks have disadvantage.' },
  { name: 'Charmed', body: 'You cannot attack whoever charmed you. They have advantage on social checks against you.' },
  { name: 'Deafened', body: 'You cannot hear, and automatically fail any check that needs hearing.' },
  { name: 'Exhaustion', body: 'Stacks in six levels, each worse than the last: disadvantage on checks, then half speed, then disadvantage on attacks and saves, then half hit points, then speed 0, then death.' },
  { name: 'Frightened', body: 'Disadvantage on checks and attacks while the thing scaring you is in sight, and you cannot willingly move closer to it.' },
  { name: 'Grappled', body: 'Your speed is 0. It ends if the grappler is incapacitated or you are moved out of reach.' },
  { name: 'Incapacitated', body: 'You can take no actions and no reactions.' },
  { name: 'Invisible', body: 'You cannot be seen without special senses. Your attacks have advantage, attacks against you have disadvantage.' },
  { name: 'Paralyzed', body: 'Incapacitated, cannot move or speak, and you fail Strength and Dexterity saves automatically. Attacks against you have advantage, and any hit from within 5 feet is a critical.' },
  { name: 'Petrified', body: 'Turned to stone. Incapacitated, unaware of your surroundings, resistant to all damage, and immune to poison and disease.' },
  { name: 'Poisoned', body: 'Disadvantage on attack rolls and ability checks.' },
  { name: 'Prone', body: 'You can only crawl, and it costs extra movement to stand. Disadvantage on your attacks. Attacks against you from within 5 feet have advantage, from further away disadvantage.' },
  { name: 'Restrained', body: 'Speed 0, disadvantage on your attacks and on Dexterity saves. Attacks against you have advantage.' },
  { name: 'Stunned', body: 'Incapacitated, cannot move, and can barely speak. You fail Strength and Dexterity saves automatically, and attacks against you have advantage.' },
  { name: 'Unconscious', body: 'Incapacitated, unaware, you drop what you hold and fall prone. You fail Strength and Dexterity saves automatically, attacks against you have advantage, and any hit from within 5 feet is a critical.' },
]

export const HOW_TO_PLAY: { heading: string; paragraphs: string[] }[] = [
  {
    heading: 'What this is',
    paragraphs: [
      'One person is the Dungeon Master (DM). They describe the world, control the monsters and decide what happens. Everyone else plays one character and decides only what that character does.',
      'There is no winning. The DM is not your opponent — they are running the world you are poking at. When you want to do something, you say it in plain words ("I check the chest for traps"), and the DM either tells you what happens or asks you to roll dice.',
    ],
  },
  {
    heading: 'Your character',
    paragraphs: [
      'A character has six ability scores: Strength, Dexterity, Constitution, Intelligence, Wisdom and Charisma. Each score gives a modifier, which is the number actually added to rolls: floor((score - 10) / 2). A score of 10 gives +0, a 16 gives +3, an 8 gives -1.',
      'Hit points (HP) are how much damage you can take before you drop. Armour class (AC) is how hard you are to hit. Speed is how far you can move on your turn, in feet. One grid square on the map is 5 feet.',
      'Your class decides which skills you can learn. Your level, from 1 to 50, decides how much you can do. Open the SHEET tab to see all of it.',
    ],
  },
  {
    heading: 'Rolling dice',
    paragraphs: [
      'Dice are written NdM: 2d6 means "roll two six-sided dice and add them". You can add a flat number: 1d20+5. The server rolls everything and shows each die, so nobody has to be trusted.',
      'Almost every uncertain action is a d20 roll plus a modifier, compared against a target number the DM sets (the DC, or difficulty class). Roll equal or higher and you succeed.',
      'Advantage means roll two d20 and keep the higher (written 2d20kh1). Disadvantage means keep the lower (2d20kl1). You get these from a good position, a helpful spell, or a condition working for or against you.',
    ],
  },
  {
    heading: 'Attacking',
    paragraphs: [
      'Pick a target token on the map, then press the attack button on a weapon, a monster action or a skill. The server rolls 1d20 plus your attack bonus against the target AC.',
      'A natural 20 on the die always hits and is a critical: the damage dice are rolled twice. A natural 1 always misses. Otherwise you hit if the total is equal to or above the target AC.',
      'On a hit, the damage expression is rolled and taken off the target hit points. If the target resists that damage type it takes half, if it is vulnerable it takes double, and if it is immune it takes none.',
    ],
  },
  {
    heading: 'Turn order',
    paragraphs: [
      'When a fight starts the DM opens an encounter. Everyone rolls initiative (a Dexterity-based d20 roll) and acts in order from highest to lowest, then the round repeats.',
      'On your turn you can move up to your speed and take one action: attack, cast a spell, use a skill, drink a potion, or something you talk the DM into. The current turn is highlighted in the initiative panel.',
    ],
  },
  {
    heading: 'Dying',
    paragraphs: [
      'At 0 hit points you fall unconscious and start making death saving throws: a d20 with no modifier each turn. Ten or higher is a success, nine or lower a failure. A natural 1 counts as two failures.',
      'Three successes and you stabilise. Three failures and the character is dead. A natural 20 wakes you up with 1 hit point. Healing of any amount also stops the countdown.',
    ],
  },
  {
    heading: 'Levelling up',
    paragraphs: [
      'The DM awards XP. Reaching the next threshold levels you up automatically, up to level 50. Early levels come quickly (300 XP for level 2, 900 for level 3) and then the curve gets steep.',
      'Each level gives 3 stat points and 1 skill point. Spend stat points to raise ability scores, and skill points in the skill tree. A skill needs your class, a minimum level, and its parent skill learnt first. Learning the same skill again raises its rank, up to its maximum.',
      'You can keep up to 10 skills active at once — that is your combat loadout. Learned skills you have not activated stay available to swap in.',
    ],
  },
  {
    heading: 'Money and loot',
    paragraphs: [
      'Coins are copper (cp), silver (sp), electrum (ep), gold (gp) and platinum (pp). 10 cp is 1 sp, 100 cp is 1 gp. Prices in the shop and the wiki are shown in whichever unit reads most cleanly.',
      'Shops buy your things back at half the listed price, rounded down, with a minimum of 1 cp for anything that is worth something at all.',
      'Monsters drop loot by chance. Drops go into a queue the DM approves or rejects before they reach a character, so a bad roll does not silently hand out a boss weapon.',
    ],
  },
  {
    heading: 'Resting',
    paragraphs: [
      'A short rest lets you spend hit dice to heal. You only have so many between long rests, so they run out.',
      'A long rest restores hit points, hit dice and spell slots. The DM decides when the party can safely take one.',
    ],
  },
]

export const GLOSSARY: GlossaryEntry[] = [
  { term: 'AC (Armour Class)', tag: 'combat', body: 'How hard you are to hit. An attack roll must equal or beat it.' },
  { term: 'Ability check', tag: 'rolls', body: 'A d20 plus an ability modifier against a DC, for anything uncertain that is not an attack.' },
  { term: 'Ability modifier', tag: 'character', body: 'The number added to rolls, floor((score - 10) / 2). A 10 gives +0, a 16 gives +3.' },
  { term: 'Ability score', tag: 'character', body: 'One of the six raw numbers: Strength, Dexterity, Constitution, Intelligence, Wisdom, Charisma.' },
  { term: 'Action', tag: 'combat', body: 'The one meaningful thing you do on your turn: attack, cast, use a skill, drink a potion.' },
  { term: 'Advantage', tag: 'rolls', body: 'Roll two d20 and keep the higher. Written 2d20kh1.' },
  { term: 'AoE (Area of Effect)', tag: 'combat', body: 'A shape on the map — circle, cone, line or cube — that hits everything inside it.' },
  { term: 'Attack bonus', tag: 'combat', body: 'Added to the d20 when you attack. Comes from the weapon, the monster action or the skill.' },
  { term: 'Bestiary', tag: 'tools', body: 'The DM list of monster stat blocks. Spawning one copies its vitals onto a new token.' },
  { term: 'Boss', tag: 'monsters', body: 'A high-CR monster meant as the end of something. Hits harder and drops better.' },
  { term: 'Cantrip', tag: 'magic', body: 'A level 0 spell. Small, and castable without using up a spell slot.' },
  { term: 'Campaign', tag: 'play', body: 'The ongoing game. Maps, characters, shops and chat all belong to one.' },
  { term: 'Character sheet', tag: 'character', body: 'Everything about your character in one panel: vitals, abilities, gear, spells and skills.' },
  { term: 'Condition', tag: 'status', body: 'A named state like poisoned or prone that changes what you can do. See the conditions tab.' },
  { term: 'Cooldown', tag: 'skills', body: 'Rounds you must wait before using a skill again.' },
  { term: 'CR (Challenge Rating)', tag: 'monsters', body: 'Rough difficulty of a monster. Higher is nastier.' },
  { term: 'Critical hit', tag: 'combat', body: 'A natural 20 on the attack die. Always hits, and the damage dice are rolled twice.' },
  { term: 'd20', tag: 'dice', body: 'A twenty-sided die. The one most rolls are built on.' },
  { term: 'Damage expression', tag: 'dice', body: 'What a hit rolls for damage, such as 1d8+3. The server evaluates it, never the client.' },
  { term: 'Damage type', tag: 'combat', body: 'Fire, slashing, poison and so on. Decides whether resistance, vulnerability or immunity applies.' },
  { term: 'DC (Difficulty Class)', tag: 'rolls', body: 'The number to beat on a check or save. The DM sets it.' },
  { term: 'Death saving throw', tag: 'combat', body: 'A bare d20 while at 0 HP. Three successes stabilise, three failures kill.' },
  { term: 'Disadvantage', tag: 'rolls', body: 'Roll two d20 and keep the lower. Written 2d20kl1.' },
  { term: 'DM (Dungeon Master)', tag: 'play', body: 'The person running the world, the monsters and the rules calls.' },
  { term: 'Encounter', tag: 'combat', body: 'A fight with a turn order. The DM starts and ends it.' },
  { term: 'Environment', tag: 'tools', body: 'A hazard, weather, terrain, trap or aura attached to a map.' },
  { term: 'Fog of war', tag: 'map', body: 'Hidden parts of the map the players cannot see until the DM reveals them.' },
  { term: 'Fumble', tag: 'combat', body: 'A natural 1 on an attack roll. Always a miss, whatever your bonus.' },
  { term: 'Hit dice', tag: 'rest', body: 'A pool you spend on a short rest to heal. Refilled by a long rest.' },
  { term: 'HP (Hit points)', tag: 'character', body: 'Damage you can take before you drop. At 0 you start making death saves.' },
  { term: 'Immunity', tag: 'combat', body: 'Takes no damage at all of that type.' },
  { term: 'Initiative', tag: 'combat', body: 'A Dexterity-based d20 roll that sets turn order, highest first.' },
  { term: 'Level', tag: 'progression', body: 'How far along a character is, 1 to 50. Each level grants 3 stat points and 1 skill point.' },
  { term: 'Long rest', tag: 'rest', body: 'Restores hit points, hit dice and spell slots.' },
  { term: 'Loadout', tag: 'skills', body: 'The at most 10 skills you have activated for combat.' },
  { term: 'Loot queue', tag: 'loot', body: 'Drops waiting on DM approval before they reach a character.' },
  { term: 'Natural roll', tag: 'dice', body: 'The number on the die itself, before any modifier. What decides crits and fumbles.' },
  { term: 'Proficiency bonus', tag: 'character', body: 'Added to rolls you are trained in. Grows with level.' },
  { term: 'Rank', tag: 'skills', body: 'How many times you have invested in one skill, up to its maximum.' },
  { term: 'Resistance', tag: 'combat', body: 'Takes half damage of that type, rounded down.' },
  { term: 'Round', tag: 'combat', body: 'One pass through the whole turn order. Roughly six seconds in the fiction.' },
  { term: 'Saving throw', tag: 'rolls', body: 'A roll to resist something happening to you, rather than to make something happen.' },
  { term: 'Savepoint', tag: 'tools', body: 'A snapshot of the campaign the DM can restore.' },
  { term: 'Short rest', tag: 'rest', body: 'A breather where you can spend hit dice to heal.' },
  { term: 'Skill point', tag: 'progression', body: 'Spent in the skill tree to learn a new skill or raise one a rank. One per level.' },
  { term: 'Skill tree', tag: 'progression', body: 'Your class skills in tiers. Later ones need a level and their parent skill learnt.' },
  { term: 'Speed', tag: 'character', body: 'How far you move on your turn, in feet. One grid square is 5 feet.' },
  { term: 'Spell slot', tag: 'magic', body: 'A use of magic at a given level. Spent to cast, refilled by a long rest.' },
  { term: 'Stat point', tag: 'progression', body: 'Spent to raise an ability score. Three per level.' },
  { term: 'Stat block', tag: 'monsters', body: 'A monster template: its vitals, defences and actions.' },
  { term: 'Token', tag: 'map', body: 'A piece on the map. A spawned monster token has its own hit points, copied from the template.' },
  { term: 'Vulnerability', tag: 'combat', body: 'Takes double damage of that type.' },
  { term: 'XP (Experience points)', tag: 'progression', body: 'Awarded by the DM. Crossing a threshold levels you up.' },
]
