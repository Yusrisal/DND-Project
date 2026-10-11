import { useCallback, useEffect, useState } from 'react'
import { api } from '../api/client'
import type {
  Character,
  CompendiumItem,
  CompendiumSpell,
  Environment,
  EnvironmentKind,
  LootRollResult,
  Monster,
  MonsterDrop,
  PendingLootDrop,
  StatusEffect,
  StatusEffectKind,
} from '../api/types'

type DashboardTab = 'environments' | 'items' | 'drops' | 'spells' | 'statuses' | 'loot_review' | 'progression_control'

type Props = {
  campaignId: number | null
  currentMapId: number | null
  characters: Character[]
  onRoll: (expression: string, label: string, characterId: number | null) => void
  onError: (error: unknown) => void
}

const ENV_KINDS: EnvironmentKind[] = ['hazard', 'weather', 'terrain', 'trap', 'aura']
const ITEM_CATEGORIES = ['all', 'weapon', 'armor', 'potion', 'scroll', 'wondrous', 'gear']
const STATUS_KINDS: StatusEffectKind[] = ['condition', 'buff', 'debuff', 'hazard']

const CORE_CONDITIONS_5E = [
  { name: 'blinded', desc: 'Auto fails sight checks. Attack rolls against have advantage, attacker rolls have disadvantage.' },
  { name: 'charmed', desc: 'Cannot attack charmer. Charmer has advantage on social ability checks.' },
  { name: 'deafened', desc: 'Auto fails hearing checks.' },
  { name: 'exhaustion', desc: 'Levels 1–6 (Disadvantage on checks, speed halved, disadv on attacks/saves, HP max halved, speed 0, death).' },
  { name: 'frightened', desc: 'Disadvantage on ability checks & attack rolls while source of fear is in line of sight. Cannot willingly move closer.' },
  { name: 'grappled', desc: 'Speed becomes 0. Ends if grappler is incapacitated.' },
  { name: 'incapacitated', desc: 'Cannot take actions or reactions.' },
  { name: 'invisible', desc: 'Impossible to see without special sense. Attacks against have disadvantage, attacker has advantage.' },
  { name: 'paralyzed', desc: 'Incapacitated, cannot move/speak. Auto fails STR & DEX saves. Attacks against have advantage, hits within 5ft are criticals.' },
  { name: 'petrified', desc: 'Transformed into solid stone. Incapacitated, resistance to all damage, immune to poison/disease.' },
  { name: 'poisoned', desc: 'Disadvantage on attack rolls and ability checks.' },
  { name: 'prone', desc: 'Crawling (costs extra move). Disadvantage on attack rolls. Attacks within 5ft have advantage, ranged attacks have disadvantage.' },
  { name: 'restrained', desc: 'Speed 0. Attacks against have advantage, attacker has disadvantage. Disadvantage on DEX saves.' },
  { name: 'stunned', desc: 'Incapacitated, cannot move, speaks falteringly. Auto fails STR & DEX saves. Attacks against have advantage.' },
  { name: 'unconscious', desc: 'Incapacitated, drops what holding, falls prone. Auto fails STR & DEX saves. Attacks against have advantage, hits within 5ft are criticals.' },
]

export function DMDashboardPanel({ campaignId, currentMapId, characters, onRoll, onError }: Props) {
  const [tab, setTab] = useState<DashboardTab>('environments')

  // Environments state
  const [environments, setEnvironments] = useState<Environment[]>([])
  const [envSearch, setEnvSearch] = useState('')

  // Compendium Items state
  const [items, setItems] = useState<CompendiumItem[]>([])
  const [itemCategory, setItemCategory] = useState('all')
  const [itemSearch, setItemSearch] = useState('')

  // Monster Drops state
  const [monsters, setMonsters] = useState<Monster[]>([])
  const [selectedMonsterId, setSelectedMonsterId] = useState<number | null>(null)
  const [monsterDrops, setMonsterDrops] = useState<MonsterDrop[]>([])
  const [lootResult, setLootResult] = useState<LootRollResult | null>(null)

  // Spells state
  const [spells, setSpells] = useState<CompendiumSpell[]>([])
  const [spellLevel, setSpellLevel] = useState<number>(-1)
  const [spellSearch, setSpellSearch] = useState('')

  // Status effects state
  const [statusEffects, setStatusEffects] = useState<StatusEffect[]>([])
  const [pendingDrops, setPendingDrops] = useState<PendingLootDrop[]>([])
  const [statusKind, setStatusKind] = useState<string>('')
  const [statusSearch, setStatusSearch] = useState('')

  // Load environments
  const refreshEnvironments = useCallback(() => {
    if (!campaignId) return
    api
      .listEnvironments(campaignId, null, { search: envSearch, page_size: 50, sort_by: 'name', sort_dir: 'asc' })
      .then((page) => setEnvironments(page.items))
      .catch(onError)
  }, [campaignId, envSearch, onError])

  useEffect(() => {
    if (tab === 'environments') {
      const timer = setTimeout(refreshEnvironments, 200)
      return () => clearTimeout(timer)
    }
  }, [tab, refreshEnvironments])

  // Load compendium items
  const refreshItems = useCallback(() => {
    api
      .listCompendiumItems(campaignId, itemCategory === 'all' ? undefined : itemCategory, {
        search: itemSearch,
        page_size: 100,
        sort_by: 'name',
        sort_dir: 'asc',
      })
      .then((page) => setItems(page.items))
      .catch(onError)
  }, [campaignId, itemCategory, itemSearch, onError])

  useEffect(() => {
    if (tab === 'items') {
      const timer = setTimeout(refreshItems, 200)
      return () => clearTimeout(timer)
    }
  }, [tab, refreshItems])

  // Load monsters for drops
  useEffect(() => {
    if (tab === 'drops') {
      api
        .listMonsters()
        .then((page) => {
          setMonsters(page.items)
          setSelectedMonsterId((curr) => curr ?? page.items[0]?.id ?? null)
        })
        .catch(onError)
    }
  }, [tab, onError])

  // Load monster drops
  const refreshMonsterDrops = useCallback(() => {
    if (!selectedMonsterId) {
      setMonsterDrops([])
      return
    }
    api.listMonsterDrops(selectedMonsterId).then(setMonsterDrops).catch(onError)
  }, [selectedMonsterId, onError])

  useEffect(() => {
    if (tab === 'drops' && selectedMonsterId) {
      refreshMonsterDrops()
    }
  }, [tab, selectedMonsterId, refreshMonsterDrops])

  // Load spells
  const refreshSpells = useCallback(() => {
    api
      .listCompendiumSpells(campaignId, spellLevel, { search: spellSearch, page_size: 100, sort_by: 'name', sort_dir: 'asc' })
      .then((page) => setSpells(page.items))
      .catch(onError)
  }, [campaignId, spellLevel, spellSearch, onError])

  useEffect(() => {
    if (tab === 'spells') {
      const timer = setTimeout(refreshSpells, 200)
      return () => clearTimeout(timer)
    }
  }, [tab, refreshSpells])

  // Load status effects
  const refreshStatusEffects = useCallback(() => {
    api
      .listStatusEffects(campaignId, statusKind, { search: statusSearch, page_size: 100, sort_by: 'name', sort_dir: 'asc' })
      .then((page) => setStatusEffects(page.items))
      .catch(onError)
  }, [campaignId, statusKind, statusSearch, onError])

  useEffect(() => {
    if (tab === 'statuses') {
      const timer = setTimeout(refreshStatusEffects, 200)
      return () => clearTimeout(timer)
    }
  }, [tab, refreshStatusEffects])

  // Environment actions
  const addEnvironment = async () => {
    if (!campaignId) return
    const name = prompt('Environment name (e.g. Lava Pool, Blinding Fog, Pit Trap)')
    if (!name) return
    const kind = (prompt(`Kind (${ENV_KINDS.join(', ')})`, 'hazard') ?? 'hazard') as EnvironmentKind
    const dc = Number(prompt('Difficulty DC (e.g. 14, or 0 if none)', '13') ?? '0') || 0
    const damage = (prompt('Damage expression if any (e.g. 2d6, 3d10)', '') ?? '').trim()
    const damageType = damage ? (prompt('Damage type (fire, poison, acid)', 'fire') ?? '') : ''
    const effectDescription = prompt('Effect description / rules', 'Creatures entering this area must make a DC save or take damage.') ?? ''

    try {
      await api.createEnvironment({
        campaign_id: campaignId,
        map_id: currentMapId,
        name,
        kind,
        difficulty_dc: dc,
        damage,
        damage_type: damageType,
        effect_description: effectDescription,
        active: true,
      })
      refreshEnvironments()
    } catch (err) {
      onError(err)
    }
  }

  const toggleEnvironment = async (env: Environment) => {
    try {
      await api.updateEnvironment(env.id, { ...env, active: !env.active })
      refreshEnvironments()
    } catch (err) {
      onError(err)
    }
  }

  const deleteEnvironment = async (id: number) => {
    try {
      await api.deleteEnvironment(id)
      refreshEnvironments()
    } catch (err) {
      onError(err)
    }
  }

  // Item actions
  const addCompendiumItem = async () => {
    const name = prompt('Item name')
    if (!name) return
    const category = prompt('Category (weapon, armor, potion, scroll, wondrous, gear)', 'weapon') ?? 'gear'
    const rarity = prompt('Rarity (common, uncommon, rare, very rare, legendary)', 'common') ?? 'common'
    const costGP = Number(prompt('Cost in Gold (GP)', '15') ?? '0') || 0
    const weight = Number(prompt('Weight (lbs)', '2') ?? '0') || 0
    const damage = (prompt('Damage expression if weapon (e.g. 1d8+2)', '') ?? '').trim()
    const attackBonus = damage ? Number(prompt('Attack bonus', '0') ?? '0') || 0 : 0
    const damageType = damage ? (prompt('Damage type (slashing, piercing, etc.)', 'slashing') ?? '') : ''
    const description = prompt('Description', '') ?? ''

    try {
      await api.createCompendiumItem({
        campaign_id: campaignId,
        name,
        category,
        rarity,
        cost_cp: Math.round(costGP * 100),
        weight,
        attack_bonus: attackBonus,
        damage,
        damage_type: damageType,
        description,
      })
      refreshItems()
    } catch (err) {
      onError(err)
    }
  }

  const deleteCompendiumItem = async (id: number) => {
    try {
      await api.deleteCompendiumItem(id)
      refreshItems()
    } catch (err) {
      onError(err)
    }
  }

  const giveItemToCharacter = async (item: CompendiumItem) => {
    if (characters.length === 0) {
      alert('No characters available')
      return
    }
    const roster = characters.map((c, i) => `${i + 1}. ${c.name}`).join('\n')
    const choice = prompt(`Give ${item.name} to which character?\n${roster}`)
    if (!choice) return
    const char = characters[Number(choice) - 1]
    if (!char) return

    try {
      await api.addItem(char.id, {
        name: item.name,
        quantity: 1,
        weight: item.weight,
        cost_cp: item.cost_cp,
        attack_bonus: item.attack_bonus,
        damage: item.damage,
        damage_type: item.damage_type,
        equipped: Boolean(item.damage),
        description: item.description,
      })
      alert(`Given ${item.name} to ${char.name}!`)
    } catch (err) {
      onError(err)
    }
  }

  // Monster Drop actions
  const addMonsterDrop = async () => {
    if (!selectedMonsterId) return
    const itemName = prompt('Item name dropped (e.g. Goblin Dagger, Potion of Healing, Wolf Pelt)')
    if (!itemName) return
    const rate = Number(prompt('Drop chance (1 - 100 %)', '50') ?? '50') || 50
    const minQty = Number(prompt('Min quantity', '1') ?? '1') || 1
    const maxQty = Number(prompt('Max quantity', '1') ?? '1') || minQty
    const gp = Number(prompt('Gold (GP) dropped with this item', '0') ?? '0') || 0
    const sp = Number(prompt('Silver (SP) dropped with this item', '0') ?? '0') || 0
    const cp = Number(prompt('Copper (CP) dropped with this item', '0') ?? '0') || 0

    try {
      await api.createMonsterDrop(selectedMonsterId, {
        item_name: itemName,
        drop_rate: rate,
        min_quantity: minQty,
        max_quantity: maxQty,
        gp,
        sp,
        cp,
      })
      refreshMonsterDrops()
    } catch (err) {
      onError(err)
    }
  }

  const deleteMonsterDrop = async (id: number) => {
    try {
      await api.deleteMonsterDrop(id)
      refreshMonsterDrops()
    } catch (err) {
      onError(err)
    }
  }

  const rollLoot = async () => {
    if (!selectedMonsterId) return
    try {
      const res = await api.rollMonsterLoot(selectedMonsterId)
      setLootResult(res)
    } catch (err) {
      onError(err)
    }
  }

  // Spell actions
  const addSpell = async () => {
    const name = prompt('Spell or Skill name')
    if (!name) return
    const level = Number(prompt('Level (0 for cantrip, 1-9)', '1') ?? '1') || 0
    const school = prompt('School / Type (Evocation, Martial Skill, etc.)', 'Evocation') ?? ''
    const castingTime = prompt('Casting time / Action cost', '1 action') ?? '1 action'
    const rangeArea = prompt('Range / Area (e.g. 60 ft, 15 ft cone)', '60 ft') ?? ''
    const duration = prompt('Duration', 'Instantaneous') ?? 'Instantaneous'
    const damage = (prompt('Damage / Healing roll (e.g. 3d6, 1d8+3)', '') ?? '').trim()
    const damageType = damage ? (prompt('Damage type (fire, radiant, necrotic)', 'fire') ?? '') : ''
    const saveAbility = prompt('Save Ability if any (DEX, CON, WIS)', '') ?? ''
    const description = prompt('Description', '') ?? ''

    try {
      await api.createCompendiumSpell({
        campaign_id: campaignId,
        name,
        level,
        school,
        casting_time: castingTime,
        range_area: rangeArea,
        duration,
        damage,
        damage_type: damageType,
        save_ability: saveAbility,
        description,
      })
      refreshSpells()
    } catch (err) {
      onError(err)
    }
  }

  const editSpell = async (spell: CompendiumSpell) => {
    const name = prompt('Spell or Skill name', spell.name)
    if (!name) return
    const level = Number(prompt('Level (0 for cantrip, 1-9)', String(spell.level)) ?? '') || 0
    const school = prompt('School / Type', spell.school) ?? spell.school
    const damage = (prompt('Damage / Healing roll (blank for none)', spell.damage) ?? '').trim()
    const damageType = damage
      ? (prompt('Damage type', spell.damage_type || 'fire') ?? '')
      : ''
    const description = prompt('Description', spell.description) ?? spell.description

    try {
      await api.updateCompendiumSpell(spell.id, {
        ...spell,
        name,
        level,
        school,
        damage,
        damage_type: damageType,
        description,
      })
      refreshSpells()
    } catch (err) {
      onError(err)
    }
  }

  const deleteSpell = async (id: number, name: string) => {
    if (!confirm(`Delete "${name}" from the catalog?`)) return
    try {
      await api.deleteCompendiumSpell(id)
      refreshSpells()
    } catch (err) {
      onError(err)
    }
  }

  // Refresh Pending Loot drops
  const refreshPendingDrops = useCallback(() => {
    if (!campaignId) return
    api.listPendingLoot(campaignId).then(setPendingDrops).catch(onError)
  }, [campaignId, onError])

  useEffect(() => {
    if (tab === 'loot_review') {
      refreshPendingDrops()
    }
  }, [tab, refreshPendingDrops])

  // Resolve pending loot (approve or reject with reason)
  const resolveDrop = async (dropId: number, approve: boolean) => {
    let reason = ''
    if (!approve) {
      reason = prompt('Enter rejection reason (e.g. Broken item, Monster fled):', 'DM veto') || 'Rejected by DM'
    }
    try {
      await api.resolvePendingLoot(dropId, approve, reason)
      refreshPendingDrops()
    } catch (err) {
      onError(err)
    }
  }

  // Progression controls (XP, Level lock, Grant points)
  const handleAwardXP = async (charId: number, charName: string) => {
    const amt = Number(prompt(`Award XP to ${charName}:`, '500'))
    if (!amt) return
    try {
      await api.addXP(charId, amt)
      alert(`Awarded ${amt} XP to ${charName}!`)
    } catch (err) {
      onError(err)
    }
  }

  const handleToggleLevelBlock = async (charId: number, currentBlocked: boolean, charName: string) => {
    let reason = ''
    if (!currentBlocked) {
      reason = prompt(`Reason to lock ${charName}'s level up:`, 'Milestone not reached') || ''
    }
    try {
      await api.setLevelBlock(charId, !currentBlocked, reason)
      alert(`${charName} level up ${!currentBlocked ? 'BLOCKED' : 'UNLOCKED'}!`)
    } catch (err) {
      onError(err)
    }
  }

  const handleGrantPoints = async (charId: number, charName: string) => {
    const stats = Number(prompt(`Grant STAT points to ${charName}:`, '1')) || 0
    const skills = Number(prompt(`Grant SKILL points to ${charName}:`, '1')) || 0
    try {
      await api.grantPoints(charId, stats, skills)
      alert(`Granted +${stats} Stats, +${skills} Skills to ${charName}!`)
    } catch (err) {
      onError(err)
    }
  }

  // Status effect actions
  const addStatusEffect = async () => {
    const name = prompt('Status effect name (e.g. Blessed, Raging, Haste, Hex, Burning)')
    if (!name) return
    const kind = (prompt(`Kind (${STATUS_KINDS.join(', ')})`, 'buff') ?? 'buff') as StatusEffectKind
    const duration = Number(prompt('Default duration in rounds (0 for indefinite)', '10') ?? '0') || 0
    const color = prompt('Color hex (e.g. #27ae60, #e67e22, #8e44ad)', '#e67e22') ?? '#e67e22'
    const desc = prompt('Effect description & mechanics', '+1d4 to attack rolls and saving throws.') ?? ''

    try {
      await api.createStatusEffect({
        campaign_id: campaignId,
        name,
        kind,
        default_duration_rounds: duration,
        color,
        effect_description: desc,
      })
      refreshStatusEffects()
    } catch (err) {
      onError(err)
    }
  }

  const deleteStatusEffect = async (id: number) => {
    try {
      await api.deleteStatusEffect(id)
      refreshStatusEffects()
    } catch (err) {
      onError(err)
    }
  }

  return (
    <div className="space-y-3 text-xs">
      {/* 8-bit Retro Tool Subtabs */}
      <div className="flex flex-wrap gap-1 p-1 bg-[#08090d] border-2 border-[#2b3145]">
        <button
          className={`flex-1 min-w-[70px] py-1.5 px-1 text-[9px] font-bold uppercase transition-all ${
            tab === 'environments'
              ? 'bg-[#1f5156] text-[#66fcf1] border border-[#66fcf1]'
              : 'bg-[#12141d] text-[#8e98af] border border-[#2b3145] hover:text-white'
          }`}
          onClick={() => setTab('environments')}
        >
          ENV
        </button>
        <button
          className={`flex-1 min-w-[70px] py-1.5 px-1 text-[9px] font-bold uppercase transition-all ${
            tab === 'items'
              ? 'bg-[#1f5156] text-[#66fcf1] border border-[#66fcf1]'
              : 'bg-[#12141d] text-[#8e98af] border border-[#2b3145] hover:text-white'
          }`}
          onClick={() => setTab('items')}
        >
          ITEMS
        </button>
        <button
          className={`flex-1 min-w-[70px] py-1.5 px-1 text-[9px] font-bold uppercase transition-all ${
            tab === 'drops'
              ? 'bg-[#1f5156] text-[#66fcf1] border border-[#66fcf1]'
              : 'bg-[#12141d] text-[#8e98af] border border-[#2b3145] hover:text-white'
          }`}
          onClick={() => setTab('drops')}
        >
          DROPS
        </button>
        <button
          className={`flex-1 min-w-[70px] py-1.5 px-1 text-[9px] font-bold uppercase transition-all ${
            tab === 'spells'
              ? 'bg-[#1f5156] text-[#66fcf1] border border-[#66fcf1]'
              : 'bg-[#12141d] text-[#8e98af] border border-[#2b3145] hover:text-white'
          }`}
          onClick={() => setTab('spells')}
        >
          SPELLS
        </button>
        <button
          className={`flex-1 min-w-[70px] py-1.5 px-1 text-[9px] font-bold uppercase transition-all ${
            tab === 'loot_review'
              ? 'bg-[#1f5156] text-[#66fcf1] border border-[#66fcf1]'
              : 'bg-[#12141d] text-[#8e98af] border border-[#2b3145] hover:text-white'
          }`}
          onClick={() => setTab('loot_review')}
        >
          LOOT QUEUE
        </button>
        <button
          className={`flex-1 min-w-[70px] py-1.5 px-1 text-[9px] font-bold uppercase transition-all ${
            tab === 'progression_control'
              ? 'bg-[#1f5156] text-[#66fcf1] border border-[#66fcf1]'
              : 'bg-[#12141d] text-[#8e98af] border border-[#2b3145] hover:text-white'
          }`}
          onClick={() => setTab('progression_control')}
        >
          XP/LEVELS
        </button>
      </div>

      {/* 1. Environments Tab */}
      {tab === 'environments' && (
        <section className="space-y-2">
          <header className="flex items-center justify-between pb-1 border-b border-[#2b3145]">
            <h2 className="text-[10px] font-bold uppercase tracking-wider text-[#e0e2ec]">Environments</h2>
            <button
              onClick={addEnvironment}
              className="px-2 py-0.5 text-[9px] font-bold bg-[#1f5156] border border-[#66fcf1] text-[#66fcf1] hover:bg-[#45a29e] hover:text-black"
            >
              + NEW
            </button>
          </header>

          <input
            value={envSearch}
            onChange={(e) => setEnvSearch(e.target.value)}
            placeholder="SEARCH HAZARDS / WEATHER..."
            className="w-full bg-[#090a0f] border-2 border-[#2b3145] p-1.5 text-[10px] text-[#e0e2ec] placeholder-[#57606f] focus:outline-none focus:border-[#66fcf1]"
          />

          {environments.length === 0 && <p className="text-[10px] text-[#57606f] italic text-center py-2">NO HAZARDS DEFINED.</p>}
          <div className="space-y-1.5">
            {environments.map((env) => (
              <div key={env.id} className="p-2 bg-[#12141d] border-2 border-[#2b3145] space-y-1">
                <div className="flex items-center justify-between">
                  <span className="font-bold text-[#66fcf1] text-[11px] uppercase">{env.name}</span>
                  <span className="text-[9px] text-[#ffa502] border border-[#454d66] px-1 bg-black">
                    [{env.kind.toUpperCase()}]
                  </span>
                </div>
                {env.effect_description && (
                  <p className="text-[10px] text-[#a4b0be] m-0 leading-tight">{env.effect_description}</p>
                )}
                <div className="flex items-center gap-2 text-[9px] text-[#747d8c] pt-1 border-t border-[#1e2333]">
                  {env.difficulty_dc > 0 && <span>DC: {env.difficulty_dc}</span>}
                  {env.damage && <span>DMG: {env.damage}</span>}
                  <div className="ml-auto flex gap-1.5 items-center">
                    <button
                      className={`px-1 text-[9px] border ${env.active ? 'bg-[#1f5156] text-[#66fcf1] border-[#66fcf1]' : 'bg-[#1e2333] text-[#747d8c] border-[#343b52]'}`}
                      onClick={() => toggleEnvironment(env)}
                    >
                      {env.active ? 'ON' : 'OFF'}
                    </button>
                    {env.damage && (
                      <button
                        className="px-1 bg-[#2f3542] text-white hover:bg-[#66fcf1] hover:text-black"
                        onClick={() => onRoll(env.damage, `${env.name} damage`, null)}
                      >
                        ROLL
                      </button>
                    )}
                    <button
                      className="text-[#ff4757] hover:underline bg-transparent border-none cursor-pointer"
                      onClick={() => deleteEnvironment(env.id)}
                    >
                      [DEL]
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </section>
      )}

      {/* 2. Items Compendium Tab */}
      {tab === 'items' && (
        <section>
          <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <h2>Item Compendium</h2>
            <button className="btn-add" onClick={addCompendiumItem}>
              + Add Item
            </button>
          </header>

          <div style={{ display: 'flex', gap: '6px', margin: '6px 0' }}>
            <input
              value={itemSearch}
              onChange={(e) => setItemSearch(e.target.value)}
              placeholder="Search items..."
              style={{ flex: 1 }}
            />
            <select value={itemCategory} onChange={(e) => setItemCategory(e.target.value)}>
              {ITEM_CATEGORIES.map((cat) => (
                <option key={cat} value={cat}>
                  {cat}
                </option>
              ))}
            </select>
          </div>

          {items.length === 0 && <p className="hint">No compendium items found.</p>}
          <div className="list">
            <ul>
              {items.map((item) => (
                <li key={item.id} className="list-row" style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <div>
                    <span className="list-name">
                      <strong>{item.name}</strong> <em style={{ fontSize: '0.85em' }}>({item.rarity} {item.category})</em>
                      {item.cost_cp > 0 && <em> · {item.cost_cp >= 100 ? `${item.cost_cp / 100}gp` : `${item.cost_cp}cp`}</em>}
                      {item.damage && <em> · {item.damage} {item.damage_type}</em>}
                    </span>
                    {item.description && (
                      <p className="hint" style={{ margin: '2px 0', fontSize: '0.8em' }}>
                        {item.description}
                      </p>
                    )}
                  </div>

                  <div style={{ display: 'flex', gap: '4px', alignItems: 'center' }}>
                    <button className="pill" onClick={() => giveItemToCharacter(item)} title="Give item to a character">
                      give
                    </button>
                    <button className="link" onClick={() => deleteCompendiumItem(item.id)}>
                      remove
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          </div>
        </section>
      )}

      {/* 3. Monster Drops Tab */}
      {tab === 'drops' && (
        <section>
          <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <h2>Monster Drops & Loot</h2>
            <button className="btn-add" onClick={addMonsterDrop} disabled={!selectedMonsterId}>
              + Add Drop
            </button>
          </header>

          <div style={{ margin: '6px 0' }}>
            <label style={{ fontSize: '0.85em', marginRight: '6px' }}>Monster:</label>
            <select
              value={selectedMonsterId ?? ''}
              onChange={(e) => setSelectedMonsterId(Number(e.target.value) || null)}
            >
              {monsters.map((m) => (
                <option key={m.id} value={m.id}>
                  {m.name} (CR {m.challenge_rating})
                </option>
              ))}
            </select>
          </div>

          <div style={{ margin: '8px 0', display: 'flex', gap: '8px', alignItems: 'center' }}>
            <button className="pill on" onClick={rollLoot} disabled={!selectedMonsterId}>
              🎲 Roll Loot Drop
            </button>
          </div>

          {lootResult && (
            <div
              style={{
                padding: '8px',
                background: 'rgba(39, 174, 96, 0.15)',
                borderRadius: '4px',
                margin: '6px 0',
                fontSize: '0.9em',
              }}
            >
              <strong>🎁 Loot Dropped ({lootResult.monster_name}):</strong>
              <div style={{ marginTop: '4px' }}>
                {lootResult.items.length === 0 && <span>No items dropped. </span>}
                {lootResult.items.map((it, idx) => (
                  <span key={idx} style={{ marginRight: '8px' }}>
                    • {it.name} ×{it.quantity}
                  </span>
                ))}
                {(lootResult.gp > 0 || lootResult.sp > 0 || lootResult.cp > 0) && (
                  <div>
                    💰 Coins: {lootResult.gp > 0 && `${lootResult.gp} GP `}
                    {lootResult.sp > 0 && `${lootResult.sp} SP `}
                    {lootResult.cp > 0 && `${lootResult.cp} CP`}
                  </div>
                )}
              </div>
            </div>
          )}

          {monsterDrops.length === 0 && <p className="hint">No loot rules defined for this monster.</p>}
          <div className="list">
            <ul>
              {monsterDrops.map((drop) => (
                <li key={drop.id} className="list-row" style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span className="list-name">
                    <strong>{drop.item_name}</strong>
                    <em> · {drop.drop_rate}% chance</em>
                    <em>
                      {' '}
                      · qty {drop.min_quantity === drop.max_quantity ? drop.min_quantity : `${drop.min_quantity}-${drop.max_quantity}`}
                    </em>
                    {(drop.gp > 0 || drop.sp > 0 || drop.cp > 0) && (
                      <em>
                        {' '}
                        · +{drop.gp > 0 ? `${drop.gp}gp ` : ''}
                        {drop.sp > 0 ? `${drop.sp}sp ` : ''}
                      </em>
                    )}
                  </span>
                  <button className="link" onClick={() => deleteMonsterDrop(drop.id)}>
                    remove
                  </button>
                </li>
              ))}
            </ul>
          </div>
        </section>
      )}

      {/* 4. Skills & Spells Compendium Tab */}
      {tab === 'spells' && (
        <section>
          <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <h2>Spells &amp; Skills Catalog</h2>
            <button className="btn-add" onClick={addSpell}>
              + Add Spell / Skill
            </button>
          </header>

          <div style={{ display: 'flex', gap: '6px', margin: '6px 0' }}>
            <input
              value={spellSearch}
              onChange={(e) => setSpellSearch(e.target.value)}
              placeholder="Search spells, skills..."
              style={{ flex: 1 }}
            />
            <select
              value={spellLevel}
              onChange={(e) => setSpellLevel(parseInt(e.target.value, 10))}
            >
              <option value={-1}>All levels</option>
              <option value={0}>Cantrips (0)</option>
              {Array.from({ length: 9 }, (_, i) => i + 1).map((lvl) => (
                <option key={lvl} value={lvl}>
                  Level {lvl}
                </option>
              ))}
            </select>
          </div>

          {spells.length === 0 && <p className="hint">No spells/skills found.</p>}
          <div className="list">
            <ul>
              {spells.map((sp) => (
                <li key={sp.id} className="list-row" style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <div>
                    <span className="list-name">
                      <strong>{sp.name}</strong>{' '}
                      <em style={{ fontSize: '0.85em' }}>
                        ({sp.level === 0 ? 'Cantrip' : `Level ${sp.level}`} {sp.school})
                      </em>
                      {sp.damage && <em> · {sp.damage} {sp.damage_type}</em>}
                      {sp.save_ability && <em> · {sp.save_ability} Save</em>}
                    </span>
                    {sp.description && (
                      <p className="hint" style={{ margin: '2px 0', fontSize: '0.8em' }}>
                        {sp.description}
                      </p>
                    )}
                  </div>

                  <div style={{ display: 'flex', gap: '4px', alignItems: 'center' }}>
                    {sp.damage && (
                      <button
                        className="pill on"
                        onClick={() => onRoll(sp.damage, `${sp.name} roll`, null)}
                        title="Roll damage/effect"
                      >
                        roll
                      </button>
                    )}
                    <button className="pill" onClick={() => editSpell(sp)} title="Edit this entry">
                      edit
                    </button>
                    <button
                      className="btn-danger"
                      onClick={() => deleteSpell(sp.id, sp.name)}
                      title="Delete this entry"
                    >
                      delete
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          </div>
        </section>
      )}

      {/* 5. Status Effects & Conditions Tab */}
      {tab === 'statuses' && (
        <section className="space-y-2">
          <header className="flex items-center justify-between pb-1 border-b border-[#2b3145]">
            <h2 className="text-[10px] font-bold uppercase tracking-wider text-[#e0e2ec]">Status Effects</h2>
            <button
              onClick={addStatusEffect}
              className="px-2 py-0.5 text-[9px] font-bold bg-[#1f5156] border border-[#66fcf1] text-[#66fcf1] hover:bg-[#45a29e] hover:text-black"
            >
              + NEW
            </button>
          </header>

          <div className="flex gap-1.5">
            <input
              value={statusSearch}
              onChange={(e) => setStatusSearch(e.target.value)}
              placeholder="SEARCH STATUSES..."
              className="flex-1 bg-[#090a0f] border-2 border-[#2b3145] p-1.5 text-[10px] text-[#e0e2ec] placeholder-[#57606f] focus:outline-none focus:border-[#66fcf1]"
            />
            <select
              value={statusKind}
              onChange={(e) => setStatusKind(e.target.value)}
              className="bg-[#090a0f] border-2 border-[#2b3145] px-1 text-[10px] text-[#e0e2ec]"
            >
              <option value="">ALL KINDS</option>
              {STATUS_KINDS.map((k) => (
                <option key={k} value={k}>
                  {k.toUpperCase()}
                </option>
              ))}
            </select>
          </div>

          {/* Custom buffs / debuffs */}
          {statusEffects.length > 0 && (
            <div className="space-y-1">
              <h3 className="text-[10px] font-bold text-[#ffa502] uppercase tracking-wider">Custom Effects</h3>
              <div className="space-y-1">
                {statusEffects.map((eff) => (
                  <div key={eff.id} className="p-2 bg-[#12141d] border-2 border-[#2b3145] space-y-1">
                    <div className="flex items-center justify-between">
                      <span className="font-bold text-[#66fcf1] uppercase text-[10px]" style={{ color: eff.color }}>
                        {eff.name}
                      </span>
                      <span className="text-[9px] text-[#747d8c] border border-[#2b3145] px-1">
                        [{eff.kind.toUpperCase()}] {eff.default_duration_rounds > 0 && `· ${eff.default_duration_rounds}R`}
                      </span>
                    </div>
                    {eff.effect_description && (
                      <p className="text-[10px] text-[#a4b0be] m-0">{eff.effect_description}</p>
                    )}
                    <button
                      className="text-[#ff4757] hover:underline text-[9px] bg-transparent border-none cursor-pointer"
                      onClick={() => deleteStatusEffect(eff.id)}
                    >
                      [DEL]
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Core 5e Conditions Reference */}
          <div className="space-y-1 pt-2 border-t border-[#2b3145]">
            <h3 className="text-[10px] font-bold text-[#e0e2ec] uppercase tracking-wider">5e Conditions</h3>
            <div className="space-y-1 max-h-[320px] overflow-y-auto pr-1">
              {CORE_CONDITIONS_5E.filter(
                (c) =>
                  !statusSearch ||
                  c.name.toLowerCase().includes(statusSearch.toLowerCase()) ||
                  c.desc.toLowerCase().includes(statusSearch.toLowerCase()),
              ).map((cond) => (
                <div key={cond.name} className="p-2 bg-[#12141d] border-2 border-[#1e2333] space-y-0.5">
                  <span className="font-bold text-[#66fcf1] text-[10px] tracking-wide block">
                    [{cond.name.toUpperCase()}]
                  </span>
                  <p className="text-[10px] text-[#8e98af] m-0 leading-snug">
                    {cond.desc}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </section>
      )}

      {/* 6. Pending Loot Queue (DM Review) */}
      {tab === 'loot_review' && (
        <section className="space-y-2">
          <header className="flex items-center justify-between pb-1 border-b border-[#2b3145]">
            <h2 className="text-[10px] font-bold uppercase tracking-wider text-[#e0e2ec]">Pending Loot Queue</h2>
            <button className="link" onClick={refreshPendingDrops}>
              refresh
            </button>
          </header>
          {pendingDrops.length === 0 && (
            <p className="text-[10px] text-slate-500 italic text-center py-4">No pending monster loot drops waiting for approval.</p>
          )}
          <div className="space-y-1.5">
            {pendingDrops.map((drop) => (
              <div key={drop.id} className="p-2.5 bg-[#12151e] border border-[#2b3145] rounded space-y-1.5">
                <div className="flex justify-between items-start">
                  <div>
                    <span className="font-bold text-white text-[11px]">{drop.item_name} ×{drop.quantity}</span>
                    <span className="text-[9px] text-slate-400 block">From: {drop.monster_name}</span>
                    {(drop.gp > 0 || drop.sp > 0 || drop.cp > 0) && (
                      <span className="text-[9px] text-amber-300">
                        Coin: {drop.gp}gp {drop.sp}sp {drop.cp}cp
                      </span>
                    )}
                  </div>
                  <div className="flex gap-1">
                    <button
                      onClick={() => resolveDrop(drop.id, true)}
                      className="px-2 py-0.5 bg-[#1e4a38] border border-[#2ed573] text-[#2ed573] text-[9px] font-bold uppercase rounded cursor-pointer hover:bg-[#2ed573] hover:text-black"
                    >
                      Approve
                    </button>
                    <button
                      onClick={() => resolveDrop(drop.id, false)}
                      className="px-2 py-0.5 bg-[#3a1518] border border-[#ff4757] text-[#ff4757] text-[9px] font-bold uppercase rounded cursor-pointer hover:bg-[#ff4757] hover:text-white"
                    >
                      Reject
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </section>
      )}

      {/* 7. XP & Level Gating Controls */}
      {tab === 'progression_control' && (
        <section className="space-y-2">
          <header className="flex items-center justify-between pb-1 border-b border-[#2b3145]">
            <h2 className="text-[10px] font-bold uppercase tracking-wider text-[#e0e2ec]">Character Progression & Level Lock</h2>
          </header>
          {characters.length === 0 && (
            <p className="text-[10px] text-slate-500 italic">No characters in this campaign.</p>
          )}
          <div className="space-y-2">
            {characters.map((char) => (
              <div key={char.id} className="p-2.5 bg-[#12151e] border border-[#2b3145] rounded space-y-1.5">
                <div className="flex justify-between items-center">
                  <div>
                    <span className="font-bold text-[#66fcf1] text-[11px]">{char.name}</span>
                    <span className="text-[10px] text-slate-300 ml-2">
                      Level {char.level} · {char.xp} XP
                    </span>
                    <div className="text-[9px] text-slate-400">
                      Unspent: {char.stat_points} Stats, {char.skill_points} Skills
                    </div>
                  </div>
                  <div className="flex gap-1 flex-wrap justify-end">
                    <button
                      onClick={() => handleAwardXP(char.id, char.name)}
                      className="px-2 py-0.5 bg-[#1f5156] border border-[#66fcf1] text-[#66fcf1] text-[9px] font-bold uppercase rounded hover:bg-[#45a29e] hover:text-black cursor-pointer"
                    >
                      + XP
                    </button>
                    <button
                      onClick={() => handleGrantPoints(char.id, char.name)}
                      className="px-2 py-0.5 bg-[#1e2333] border border-[#3a445d] text-slate-200 text-[9px] font-bold uppercase rounded hover:border-[#66fcf1] cursor-pointer"
                    >
                      Grant Pts
                    </button>
                    <button
                      onClick={() => handleToggleLevelBlock(char.id, char.level_blocked, char.name)}
                      className={`px-2 py-0.5 text-[9px] font-bold uppercase rounded border cursor-pointer ${
                        char.level_blocked
                          ? 'bg-[#4a181c] border-[#ff4757] text-[#ff4757]'
                          : 'bg-[#1e272e] border-slate-600 text-slate-300'
                      }`}
                    >
                      {char.level_blocked ? 'Blocked' : 'Unlocked'}
                    </button>
                  </div>
                </div>
                {char.level_blocked && char.block_reason && (
                  <div className="text-[9px] text-red-400 italic">
                    Lock reason: {char.block_reason}
                  </div>
                )}
              </div>
            ))}
          </div>
        </section>
      )}
    </div>
  )
}
