import type {
  AoETemplate,
  AoETemplateInput,
  Campaign,
  Character,
  CharacterSheet as Sheet,
  CharacterItem,
  CharacterSheet,
  CharacterSpell,
  ClassSkill,
  CompendiumItem,
  CompendiumSpell,
  Environment,
  FogInput,
  GameMap,
  Encounter,
  Handout,
  HandoutInput,
  Invite,
  ListParams,
  LootRollResult,
  MapObject,
  MapObjectInput,
  Monster,
  MonsterAction,
  MonsterDrop,
  MonsterStatBlock,
  Page,
  PauseInput,
  PendingLootDrop,
  RetentionInput,
  Role,
  Savepoint,
  SavepointInput,
  Shop,
  ShopDetails,
  ShopItem,
  SpellSlot,
  StatusEffect,
  Token,
  User,
} from './types'

export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly code?: string,
    readonly field?: string,
  ) {
    super(message)
  }

  get isUnauthorized(): boolean {
    return this.status === 401
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`/api${path}`, {
    ...init,
    // The session is an HttpOnly cookie, so every call has to carry it.
    credentials: 'include',
    headers: init?.body ? { 'Content-Type': 'application/json' } : undefined,
  })

  if (!response.ok) {
    const body = await response.json().catch(() => null)
    throw new ApiError(body?.error ?? response.statusText, response.status, body?.code, body?.field)
  }
  return response.status === 204 ? (undefined as T) : (response.json() as Promise<T>)
}

function query(params: Record<string, string | number | undefined>): string {
  const search = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '') search.set(key, String(value))
  }
  const qs = search.toString()
  return qs ? `?${qs}` : ''
}

const json = (body: unknown): RequestInit => ({ method: 'POST', body: JSON.stringify(body) })

// Multipart sets its own boundary, so this cannot go through request(), which
// forces a JSON content type whenever there is a body.
async function upload(file: File): Promise<{ url: string }> {
  const form = new FormData()
  form.append('file', file)
  const response = await fetch('/api/uploads', { method: 'POST', body: form, credentials: 'include' })
  if (!response.ok) {
    const body = await response.json().catch(() => null)
    throw new ApiError(body?.error ?? response.statusText, response.status, body?.code, body?.field)
  }
  return response.json()
}

export const api = {
  me: () => request<User>('/auth/me'),

  uploadImage: upload,

  login: (input: { email: string; password: string }) => request<User>('/auth/login', json(input)),

  register: (input: { code: string; email: string; display_name: string; password: string }) =>
    request<User>('/auth/register', json(input)),

  logout: () => request<void>('/auth/logout', { method: 'POST' }),

  listUsers: () => request<Page<User>>(`/users${query({ page_size: 100, sort_dir: 'asc' })}`),

  listInvites: () => request<Page<Invite>>(`/invites${query({ page_size: 50 })}`),

  createInvite: (role: Role) => request<Invite>('/invites', json({ role })),

  deleteInvite: (code: string) => request<void>(`/invites/${code}`, { method: 'DELETE' }),

  listCampaigns: (params: ListParams = {}) =>
    request<Page<Campaign>>(`/campaigns${query({ ...params })}`),

  createCampaign: (input: { name: string; description: string }) =>
    request<Campaign>('/campaigns', json(input)),

  deleteCampaign: (id: number) => request<void>(`/campaigns/${id}`, { method: 'DELETE' }),

  setCampaignPause: (campaignId: number, input: PauseInput) =>
    request<Campaign>(`/campaigns/${campaignId}/pause`, { method: 'PUT', body: JSON.stringify(input) }),

  setCampaignRetention: (campaignId: number, input: RetentionInput) =>
    request<Campaign>(`/campaigns/${campaignId}/retention`, { method: 'PUT', body: JSON.stringify(input) }),

  // Savepoints & Session Checkpoints
  listSavepoints: (campaignId: number, params: ListParams = {}) =>
    request<Page<Savepoint>>(`/campaigns/${campaignId}/savepoints${query({ ...params })}`),

  createSavepoint: (campaignId: number, input: SavepointInput) =>
    request<Savepoint>(`/campaigns/${campaignId}/savepoints`, json(input)),

  restoreSavepoint: (id: number) =>
    request<Savepoint>(`/savepoints/${id}/restore`, { method: 'POST' }),

  deleteSavepoint: (id: number) =>
    request<void>(`/savepoints/${id}`, { method: 'DELETE' }),

  listMaps: (campaignId: number, params: ListParams = {}) =>
    request<Page<GameMap>>(`/maps${query({ campaign_id: campaignId, ...params })}`),

  createMap: (input: {
    campaign_id: number
    name: string
    image_url?: string
    grid_size?: number
    width?: number
    height?: number
  }) => request<GameMap>('/maps', json(input)),

  updateMap: (
    id: number,
    input: { name: string; image_url: string; grid_size: number; width: number; height: number },
  ) => request<GameMap>(`/maps/${id}`, { method: 'PUT', body: JSON.stringify(input) }),

  deleteMap: (id: number) => request<void>(`/maps/${id}`, { method: 'DELETE' }),

  board: (mapId: number) => request<Token[]>(`/maps/${mapId}/tokens`),

  createToken: (input: Partial<Token> & { map_id: number; name: string }) =>
    request<Token>('/tokens', json(input)),

  updateToken: (id: number, input: Partial<Token> & { name: string }) =>
    request<Token>(`/tokens/${id}`, { method: 'PUT', body: JSON.stringify(input) }),

  deleteToken: (id: number) => request<void>(`/tokens/${id}`, { method: 'DELETE' }),

  listCharacters: (campaignId: number) =>
    request<Page<Character>>(`/characters${query({ campaign_id: campaignId, page_size: 50, sort_dir: 'asc' })}`),

  sheet: (id: number) => request<CharacterSheet>(`/characters/${id}`),

  createCharacter: (input: Partial<Character> & { campaign_id: number; name: string }) =>
    request<Character>('/characters', json(input)),

  updateCharacter: (id: number, input: Partial<Character>) =>
    request<Character>(`/characters/${id}`, { method: 'PUT', body: JSON.stringify(input) }),

  deleteCharacter: (id: number) => request<void>(`/characters/${id}`, { method: 'DELETE' }),

  longRest: (characterId: number) =>
    request<CharacterSheet>(`/characters/${characterId}/long-rest`, { method: 'POST' }),

  shortRest: (characterId: number, input: { hit_dice_spent: number; hp_healed: number }) =>
    request<CharacterSheet>(`/characters/${characterId}/short-rest`, json(input)),

  deathSave: (characterId: number, roll: number) =>
    request<CharacterSheet>(`/characters/${characterId}/death-save`, json({ roll })),

  splitLoot: (input: { campaign_id: number; cp: number; sp: number; ep: number; gp: number; pp: number }) =>
    request<void>('/loot/split', json(input)),

  addItem: (characterId: number, input: Partial<CharacterItem> & { name: string }) =>
    request<CharacterItem>(`/characters/${characterId}/items`, json(input)),

  updateItem: (id: number, input: Partial<CharacterItem> & { name: string }) =>
    request<CharacterItem>(`/items/${id}`, { method: 'PUT', body: JSON.stringify(input) }),

  deleteItem: (id: number) => request<void>(`/items/${id}`, { method: 'DELETE' }),

  addSpell: (characterId: number, input: Partial<CharacterSpell> & { name: string }) =>
    request<CharacterSpell>(`/characters/${characterId}/spells`, json(input)),

  updateSpell: (id: number, input: Partial<CharacterSpell> & { name: string }) =>
    request<CharacterSpell>(`/spells/${id}`, { method: 'PUT', body: JSON.stringify(input) }),

  deleteSpell: (id: number) => request<void>(`/spells/${id}`, { method: 'DELETE' }),

  setSlot: (characterId: number, level: number, input: { total: number; used: number }) =>
    request<SpellSlot[]>(`/characters/${characterId}/slots/${level}`, {
      method: 'PUT',
      body: JSON.stringify(input),
    }),

  setProficiency: (
    characterId: number,
    input: { kind: 'skill' | 'save'; name: string; expertise?: boolean; trained: boolean },
  ) => request<Sheet>(`/characters/${characterId}/proficiencies`, {
    method: 'PUT',
    body: JSON.stringify(input),
  }),

  // ARPG Progression & Skills
  allocateStat: (characterId: number, ability: string) =>
    request<CharacterSheet>(`/characters/${characterId}/allocate-stat`, json({ ability })),

  listSkillTrees: (className?: string) =>
    request<ClassSkill[]>(`/skill-trees${query({ class_name: className })}`),

  learnSkill: (characterId: number, skillId: number) =>
    request<CharacterSheet>(`/characters/${characterId}/learn-skill`, json({ skill_id: skillId })),

  toggleSkillActive: (characterId: number, skillId: number, active: boolean) =>
    request<CharacterSheet>(`/characters/${characterId}/toggle-skill`, json({ skill_id: skillId, active })),

  useSkillScroll: (characterId: number, itemId: number) =>
    request<CharacterSheet>(`/characters/${characterId}/use-scroll`, json({ item_id: itemId })),

  transferItem: (characterId: number, input: { to_character_id: number; item_id: number; quantity?: number }) =>
    request<CharacterSheet>(`/characters/${characterId}/transfer-item`, json(input)),

  addXP: (characterId: number, amount: number) =>
    request<Character>(`/characters/${characterId}/xp`, json({ amount })),

  setLevelBlock: (characterId: number, blocked: boolean, reason?: string) =>
    request<Character>(`/characters/${characterId}/level-block`, {
      method: 'PUT',
      body: JSON.stringify({ blocked, reason: reason ?? '' }),
    }),

  grantPoints: (characterId: number, statPoints: number, skillPoints: number) =>
    request<Character>(`/characters/${characterId}/grant-points`, json({ stat_points: statPoints, skill_points: skillPoints })),

  listPendingLoot: (campaignId: number) =>
    request<PendingLootDrop[]>(`/pending-loot${query({ campaign_id: campaignId })}`),

  resolvePendingLoot: (dropId: number, approve: boolean, reason?: string) =>
    request<void>(`/pending-loot/${dropId}/resolve`, json({ approve, reason: reason ?? '' })),

  // Bestiary

  listMonsters: (search = '') =>
    request<Page<Monster>>(`/monsters${query({ search, page_size: 100, sort_dir: 'asc' })}`),

  statBlock: (id: number) => request<MonsterStatBlock>(`/monsters/${id}`),

  createMonster: (input: Partial<Monster> & { name: string }) =>
    request<Monster>('/monsters', json(input)),

  updateMonster: (id: number, input: Partial<Monster> & { name: string }) =>
    request<Monster>(`/monsters/${id}`, { method: 'PUT', body: JSON.stringify(input) }),

  deleteMonster: (id: number) => request<void>(`/monsters/${id}`, { method: 'DELETE' }),

  addAction: (monsterId: number, input: Partial<MonsterAction> & { name: string }) =>
    request<MonsterAction>(`/monsters/${monsterId}/actions`, json(input)),

  deleteAction: (id: number) => request<void>(`/actions/${id}`, { method: 'DELETE' }),

  spawn: (input: { monster_id: number; map_id: number; count: number; roll_hp: boolean }) =>
    request<Token[]>('/spawn', json(input)),

  // Combat

  encounter: (mapId: number) => request<Encounter>(`/maps/${mapId}/encounter`),

  startEncounter: (mapId: number) => request<Encounter>(`/maps/${mapId}/encounter`, { method: 'POST' }),

  endEncounter: (mapId: number) => request<Encounter>(`/maps/${mapId}/encounter`, { method: 'DELETE' }),

  nextTurn: (mapId: number) =>
    request<Encounter>(`/maps/${mapId}/encounter/turn`, { method: 'POST' }),

  addCombatant: (mapId: number, input: { token_id?: number | null; name?: string; initiative?: number }) =>
    request<Encounter>(`/maps/${mapId}/encounter/combatants`, json(input)),

  removeCombatant: (mapId: number, combatantId: number) =>
    request<Encounter>(`/maps/${mapId}/encounter/combatants/${combatantId}`, { method: 'DELETE' }),

  setCondition: (tokenId: number, condition: string, active: boolean, durationRounds = 0, notes = '') =>
    request<Token>(`/tokens/${tokenId}/conditions`, {
      method: 'PUT',
      body: JSON.stringify({ condition, active, duration_rounds: durationRounds, notes }),
    }),

  // Shops & Merchants

  listShops: (campaignId: number, params: ListParams = {}) =>
    request<Page<Shop>>(`/shops${query({ campaign_id: campaignId, ...params })}`),

  getShop: (id: number) => request<ShopDetails>(`/shops/${id}`),

  createShop: (input: { campaign_id: number; name: string; description?: string }) =>
    request<Shop>('/shops', json(input)),

  updateShop: (id: number, input: { name: string; description?: string }) =>
    request<Shop>(`/shops/${id}`, { method: 'PUT', body: JSON.stringify(input) }),

  deleteShop: (id: number) => request<void>(`/shops/${id}`, { method: 'DELETE' }),

  addShopItem: (shopId: number, input: Partial<ShopItem> & { name: string }) =>
    request<ShopItem>(`/shops/${shopId}/items`, json(input)),

  updateShopItem: (id: number, input: Partial<ShopItem> & { name: string }) =>
    request<ShopItem>(`/shop-items/${id}`, { method: 'PUT', body: JSON.stringify(input) }),

  deleteShopItem: (id: number) => request<void>(`/shop-items/${id}`, { method: 'DELETE' }),

  buyShopItem: (shopItemId: number, input: { character_id: number; quantity?: number }) =>
    request<CharacterItem>(`/shop-items/${shopItemId}/buy`, json(input)),

  sellItem: (input: { character_id: number; item_id: number; quantity?: number }) =>
    request<{
      item_name: string
      quantity_sold: number
      refund_cp: number
      cp: number
      sp: number
      ep: number
      gp: number
      pp: number
    }>('/shops/sell', json(input)),

  // Environments

  listEnvironments: (campaignId: number, mapId?: number | null, params: ListParams = {}) =>
    request<Page<Environment>>(`/environments${query({ campaign_id: campaignId, map_id: mapId ?? undefined, ...params })}`),

  getEnvironment: (id: number) => request<Environment>(`/environments/${id}`),

  createEnvironment: (input: Partial<Environment> & { campaign_id: number; name: string }) =>
    request<Environment>('/environments', json(input)),

  updateEnvironment: (id: number, input: Partial<Environment>) =>
    request<Environment>(`/environments/${id}`, { method: 'PUT', body: JSON.stringify(input) }),

  deleteEnvironment: (id: number) => request<void>(`/environments/${id}`, { method: 'DELETE' }),

  // Compendium Items

  listCompendiumItems: (campaignId?: number | null, category?: string, params: ListParams = {}) =>
    request<Page<CompendiumItem>>(
      `/compendium/items${query({ campaign_id: campaignId ?? undefined, category: category || undefined, ...params })}`,
    ),

  getCompendiumItem: (id: number) => request<CompendiumItem>(`/compendium/items/${id}`),

  createCompendiumItem: (input: Partial<CompendiumItem> & { name: string }) =>
    request<CompendiumItem>('/compendium/items', json(input)),

  updateCompendiumItem: (id: number, input: Partial<CompendiumItem>) =>
    request<CompendiumItem>(`/compendium/items/${id}`, { method: 'PUT', body: JSON.stringify(input) }),

  deleteCompendiumItem: (id: number) => request<void>(`/compendium/items/${id}`, { method: 'DELETE' }),

  // Monster Drops & Loot

  listMonsterDrops: (monsterId: number) => request<MonsterDrop[]>(`/monsters/${monsterId}/drops`),

  createMonsterDrop: (monsterId: number, input: Partial<MonsterDrop> & { item_name: string }) =>
    request<MonsterDrop>(`/monsters/${monsterId}/drops`, json(input)),

  deleteMonsterDrop: (id: number) => request<void>(`/monster-drops/${id}`, { method: 'DELETE' }),

  rollMonsterLoot: (monsterId: number) => request<LootRollResult>(`/monsters/${monsterId}/roll-loot`, { method: 'POST' }),

  // Compendium Spells & Skills

  listCompendiumSpells: (campaignId?: number | null, level?: number, params: ListParams = {}) =>
    request<Page<CompendiumSpell>>(
      `/compendium/spells${query({ campaign_id: campaignId ?? undefined, level: level !== undefined && level >= 0 ? level : undefined, ...params })}`,
    ),

  getCompendiumSpell: (id: number) => request<CompendiumSpell>(`/compendium/spells/${id}`),

  createCompendiumSpell: (input: Partial<CompendiumSpell> & { name: string }) =>
    request<CompendiumSpell>('/compendium/spells', json(input)),

  updateCompendiumSpell: (id: number, input: Partial<CompendiumSpell>) =>
    request<CompendiumSpell>(`/compendium/spells/${id}`, { method: 'PUT', body: JSON.stringify(input) }),

  deleteCompendiumSpell: (id: number) => request<void>(`/compendium/spells/${id}`, { method: 'DELETE' }),

  // Status Effects

  listStatusEffects: (campaignId?: number | null, kind?: string, params: ListParams = {}) =>
    request<Page<StatusEffect>>(
      `/status-effects${query({ campaign_id: campaignId ?? undefined, kind: kind || undefined, ...params })}`,
    ),

  getStatusEffect: (id: number) => request<StatusEffect>(`/status-effects/${id}`),

  createStatusEffect: (input: Partial<StatusEffect> & { name: string }) =>
    request<StatusEffect>('/status-effects', json(input)),

  updateStatusEffect: (id: number, input: Partial<StatusEffect>) =>
    request<StatusEffect>(`/status-effects/${id}`, { method: 'PUT', body: JSON.stringify(input) }),

  deleteStatusEffect: (id: number) => request<void>(`/status-effects/${id}`, { method: 'DELETE' }),

  // Fog of War & AoE Templates
  updateMapFog: (mapId: number, input: FogInput) =>
    request<GameMap>(`/maps/${mapId}/fog`, { method: 'PUT', body: JSON.stringify(input) }),

  listAoETemplates: (mapId: number) => request<AoETemplate[]>(`/maps/${mapId}/aoe-templates`),

  createAoETemplate: (mapId: number, input: AoETemplateInput) =>
    request<AoETemplate>(`/maps/${mapId}/aoe-templates`, json(input)),

  deleteAoETemplate: (id: number) => request<void>(`/aoe-templates/${id}`, { method: 'DELETE' }),

  // Weather & Map Objects (Doors, Chests, Traps)
  setMapWeather: (mapId: number, input: { weather: string }) =>
    request<GameMap>(`/maps/${mapId}/weather`, { method: 'PUT', body: JSON.stringify(input) }),

  listMapObjects: (mapId: number) =>
    request<MapObject[]>(`/maps/${mapId}/objects`),

  createMapObject: (mapId: number, input: MapObjectInput) =>
    request<MapObject>(`/maps/${mapId}/objects`, json(input)),

  updateMapObject: (id: number, input: Partial<MapObjectInput>) =>
    request<MapObject>(`/map-objects/${id}`, { method: 'PUT', body: JSON.stringify(input) }),

  deleteMapObject: (id: number) =>
    request<void>(`/map-objects/${id}`, { method: 'DELETE' }),

  // Campaign Handouts & Lore
  listHandouts: (campaignId: number, params: ListParams = {}) =>
    request<Page<Handout>>(`/campaigns/${campaignId}/handouts${query({ ...params })}`),

  getHandout: (id: number) => request<Handout>(`/handouts/${id}`),

  createHandout: (campaignId: number, input: HandoutInput) =>
    request<Handout>(`/campaigns/${campaignId}/handouts`, json(input)),

  updateHandout: (id: number, input: Partial<HandoutInput>) =>
    request<Handout>(`/handouts/${id}`, { method: 'PUT', body: JSON.stringify(input) }),

  deleteHandout: (id: number) => request<void>(`/handouts/${id}`, { method: 'DELETE' }),
}
