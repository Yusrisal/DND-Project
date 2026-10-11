export type Role = 'dm' | 'player'

export type User = {
  id: number
  email: string
  display_name: string
  role: Role
  created_at: string
  updated_at: string
}

export type Invite = {
  code: string
  role: Role
  created_by: number
  used_by: number | null
  used_at: string | null
  expires_at: string
  created_at: string
}

export type Campaign = {
  id: number
  name: string
  description: string
  is_paused?: boolean
  pause_message?: string
  chat_retention_days?: number
  created_at: string
  updated_at: string
}

export type PauseInput = {
  is_paused: boolean
  pause_message?: string
}

export type RetentionInput = {
  chat_retention_days: number
}

export type Savepoint = {
  id: number
  campaign_id: number
  map_id: number | null
  name: string
  notes: string
  snapshot_data: string
  created_at: string
}

export type SavepointInput = {
  campaign_id: number
  map_id?: number | null
  name: string
  notes?: string
}

export type GameMap = {
  id: number
  campaign_id: number
  name: string
  image_url: string
  grid_size: number
  width: number
  height: number
  fog_enabled?: boolean
  fog_revealed?: string
  weather?: string
  created_at: string
  updated_at: string
}

export type WeatherInput = {
  weather: string
}

export type MapObjectKind = 'door' | 'chest' | 'trap' | 'marker'

export type MapObject = {
  id: number
  map_id: number
  kind: MapObjectKind
  name: string
  x: number
  y: number
  state: string
  data: Record<string, unknown>
  created_at: string
  updated_at: string
}

export type MapObjectInput = {
  map_id: number
  kind: MapObjectKind
  name: string
  x: number
  y: number
  state: string
  data?: Record<string, unknown>
}

export type MapPing = {
  map_id: number
  x: number
  y: number
  author_name: string
  color: string
}

export type FogInput = {
  fog_enabled: boolean
  fog_revealed: string
}

export type AoETemplateShape = 'circle' | 'cone' | 'line' | 'cube'

export type AoETemplate = {
  id: number
  map_id: number
  name: string
  shape: AoETemplateShape
  size_ft: number
  color: string
  x: number
  y: number
  created_at: string
}

export type AoETemplateInput = {
  name: string
  shape: AoETemplateShape
  size_ft: number
  color: string
  x: number
  y: number
}

export type Handout = {
  id: number
  campaign_id: number
  title: string
  content: string
  image_url: string
  is_revealed: boolean
  created_at: string
  updated_at: string
}

export type HandoutInput = {
  campaign_id: number
  title: string
  content: string
  image_url: string
  is_revealed: boolean
}

export type Token = {
  id: number
  map_id: number
  owner_user_id: number | null
  character_id: number | null
  monster_id: number | null
  hp_current: number
  hp_max: number
  armor_class: number
  conditions: string[]
  name: string
  image_url: string
  color: string
  size: number
  x: number
  y: number
  created_at: string
  updated_at: string
}

export type Page<T> = {
  items: T[]
  page: number
  page_size: number
  total: number
  total_pages: number
}

export type ListParams = {
  page?: number
  page_size?: number
  search?: string
  sort_by?: string
  sort_dir?: 'asc' | 'desc'
}

export type ChatMessage = {
  id: number
  campaign_id: number
  user_id: number | null
  author_name: string
  body: string
  created_at: string
}

export type DiceGroup = {
  expression: string
  rolls: number[]
  dropped: number[]
  total: number
}

export type RollDetail = {
  expression: string
  groups: DiceGroup[]
  modifier: number
  total: number
}

export type DiceRoll = {
  id: number
  campaign_id: number
  user_id: number | null
  character_id: number | null
  author_name: string
  label: string
  expression: string
  detail: RollDetail
  total: number
  created_at: string
}

/** A chat message and a roll share the feed, tagged so the panel can render both. */
export type FeedEntry =
  | { kind: 'message'; at: string; message: ChatMessage }
  | { kind: 'roll'; at: string; roll: DiceRoll }

export type Abilities = {
  strength: number
  dexterity: number
  constitution: number
  intelligence: number
  wisdom: number
  charisma: number
}

export const ABILITY_KEYS: (keyof Abilities)[] = [
  'strength',
  'dexterity',
  'constitution',
  'intelligence',
  'wisdom',
  'charisma',
]

export type CharacterStatus = 'active' | 'tavern' | 'dead' | 'retired'

export type Character = {
  id: number
  campaign_id: number
  user_id: number | null
  status: CharacterStatus
  name: string
  class: string
  race: string
  level: number
  xp: number
  stat_points: number
  skill_points: number
  level_blocked: boolean
  block_reason: string
  hp_current: number
  hp_max: number
  hp_temp: number
  armor_class: number
  speed: number
  proficiency_bonus: number
  cp: number
  sp: number
  ep: number
  gp: number
  pp: number
  abilities: Abilities
  modifiers: Abilities
  notes: string
  death_save_successes: number
  death_save_failures: number
  hit_dice_total: number
  hit_dice_used: number
  hit_dice_die: string
  damage_resistances: string
  damage_vulnerabilities: string
  damage_immunities: string
  condition_immunities: string
  created_at: string
  updated_at: string
}

export type CharacterItem = {
  id: number
  character_id: number
  name: string
  quantity: number
  weight: number
  cost_cp: number
  equipped: boolean
  /** An item with a damage expression is a weapon it can attack with. */
  attack_bonus: number
  damage: string
  damage_type: string
  description: string
  req_level: number
  req_class: string
  req_str: number
  req_dex: number
  req_con: number
  req_int: number
  req_wis: number
  req_cha: number
  teaches_skill_id?: number | null
}

export type ClassSkill = {
  id: number
  class_name: string
  name: string
  tier: number
  req_level: number
  parent_skill_id: number | null
  damage: string
  damage_type: string
  cooldown_rounds: number
  cost_resource: string
  description: string
}

export type CharacterLearnedSkill = {
  id: number
  character_id: number
  skill_id: number
  is_active: boolean
  learned_at: string
  skill: ClassSkill
}

export type PendingLootDrop = {
  id: number
  campaign_id: number
  map_id: number
  monster_name: string
  item_name: string
  item_id: number | null
  quantity: number
  cp: number
  sp: number
  gp: number
  status: 'pending' | 'approved' | 'rejected'
  reject_reason: string
  assigned_to_character_id: number | null
  created_at: string
}

export type CharacterSpell = {
  id: number
  character_id: number
  name: string
  level: number
  school: string
  prepared: boolean
  description: string
}

export type SpellSlot = {
  level: number
  total: number
  used: number
}

export type CharacterSheet = {
  character: Character
  items: CharacterItem[]
  spells: CharacterSpell[]
  skills_loaded?: CharacterLearnedSkill[]
  slots: SpellSlot[]
  proficiencies: Proficiency[]
  /** Derived by the server so no client reimplements the proficiency rule. */
  skills: CheckModifier[]
  saves: CheckModifier[]
}

export type Combatant = {
  id: number
  token_id: number | null
  name: string
  initiative: number
}

export type Encounter = {
  id: number
  map_id: number
  round: number
  turn_index: number
  active: boolean
  combatants: Combatant[]
}

export type Vitals = {
  token_id: number
  character_id: number | null
  name: string
  hp_current: number
  hp_max: number
  hp_temp: number
  armor_class: number
  damage_resistances: string
  damage_vulnerabilities: string
  damage_immunities: string
  condition_immunities: string
}

export type DiceSummary = {
  expression: string
  total: number
  detail: RollDetail
}

/** The detail of a roll whose kind is "attack": a full resolution. */
export type AttackDetail = {
  kind: 'attack'
  attacker_name: string
  target_name: string
  weapon_name: string
  attack_roll: DiceSummary
  target_ac: number
  hit: boolean
  critical: boolean
  fumble: boolean
  damage: DiceSummary | null
  damage_type: string
  target: Vitals
}

export type CheckModifier = {
  name: string
  ability: string
  modifier: number
  trained: boolean
}

export type Proficiency = {
  kind: 'skill' | 'save'
  name: string
  expertise: boolean
}

export type Monster = {
  id: number
  name: string
  size: string
  type: string
  alignment: string
  armor_class: number
  hp_average: number
  hp_formula: string
  speed: number
  abilities: Abilities
  modifiers: Abilities
  challenge_rating: string
  senses: string
  languages: string
  traits: string
  damage_resistances: string
  damage_vulnerabilities: string
  damage_immunities: string
  condition_immunities: string
  color: string
  image_url: string
  token_size: number
}

export type MonsterProficiency = {
  kind: 'skill' | 'save'
  name: string
  modifier: number
}

export type MonsterAction = {
  id: number
  monster_id: number
  name: string
  attack_bonus: number
  damage: string
  damage_type: string
  description: string
}

export type MonsterStatBlock = {
  monster: Monster
  proficiencies: MonsterProficiency[]
  actions: MonsterAction[]
}

export type Shop = {
  id: number
  campaign_id: number
  name: string
  description: string
  created_at: string
  updated_at: string
}

export type ShopItem = {
  id: number
  shop_id: number
  name: string
  cost_cp: number
  quantity: number
  weight: number
  attack_bonus: number
  damage: string
  damage_type: string
  description: string
  created_at: string
  updated_at: string
}

export type ShopDetails = {
  shop: Shop
  items: ShopItem[]
}

export type EnvironmentKind = 'hazard' | 'weather' | 'terrain' | 'trap' | 'aura'

export type Environment = {
  id: number
  campaign_id: number
  map_id: number | null
  name: string
  kind: EnvironmentKind
  difficulty_dc: number
  damage: string
  damage_type: string
  effect_description: string
  active: boolean
  created_at: string
  updated_at: string
}

export type CompendiumItem = {
  id: number
  campaign_id: number | null
  name: string
  category: string
  rarity: string
  cost_cp: number
  weight: number
  attack_bonus: number
  damage: string
  damage_type: string
  description: string
  created_at: string
  updated_at: string
}

export type MonsterDrop = {
  id: number
  monster_id: number
  item_name: string
  item_id: number | null
  drop_rate: number
  min_quantity: number
  max_quantity: number
  cp: number
  sp: number
  gp: number
  created_at: string
  updated_at: string
}

export type CompendiumSpell = {
  id: number
  campaign_id: number | null
  name: string
  level: number
  school: string
  casting_time: string
  range_area: string
  duration: string
  damage: string
  damage_type: string
  save_ability: string
  description: string
  created_at: string
  updated_at: string
}

export type StatusEffectKind = 'condition' | 'buff' | 'debuff' | 'hazard'

export type StatusEffect = {
  id: number
  campaign_id: number | null
  name: string
  kind: StatusEffectKind
  default_duration_rounds: number
  color: string
  effect_description: string
  created_at: string
  updated_at: string
}

export type DroppedItem = {
  name: string
  quantity: number
}

export type LootRollResult = {
  monster_name: string
  items: DroppedItem[]
  cp: number
  sp: number
  gp: number
}

export const CONDITIONS = [
  'blinded',
  'charmed',
  'deafened',
  'exhaustion',
  'frightened',
  'grappled',
  'incapacitated',
  'invisible',
  'paralyzed',
  'petrified',
  'poisoned',
  'prone',
  'restrained',
  'stunned',
  'unconscious',
] as const
