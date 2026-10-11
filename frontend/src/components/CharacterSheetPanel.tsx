import { useCallback, useEffect, useState } from 'react'
import { api } from '../api/client'
import { ABILITY_KEYS } from '../api/types'
import type { Abilities, CharacterItem, CharacterSheet, CharacterStatus } from '../api/types'
import { SkillTreeTab } from './SkillTreeTab'

type Props = {
  characterId: number
  /** Rolls go out over the socket so the whole table sees them. */
  onRoll: (expression: string, label: string, characterId: number | null) => void
  /** The token an attack lands on, picked by clicking it on the board. */
  targetTokenId: number | null
  onAttack: (request: { item_id: number; target_token_id: number }) => void
  onError: (error: unknown) => void
}

type Tab = 'stats' | 'progression' | 'skills' | 'inventory' | 'spells'

const signed = (value: number) => (value >= 0 ? `+${value}` : String(value))

const SHORT: Record<keyof Abilities, string> = {
  strength: 'STR',
  dexterity: 'DEX',
  constitution: 'CON',
  intelligence: 'INT',
  wisdom: 'WIS',
  charisma: 'CHA',
}

const STATUS_OPTIONS: { id: CharacterStatus; label: string }[] = [
  { id: 'active', label: 'Active' },
  { id: 'tavern', label: 'Tavern' },
  { id: 'dead', label: 'Dead' },
  { id: 'retired', label: 'Retired' },
]

export function CharacterSheetPanel({ characterId, onRoll, targetTokenId, onAttack, onError }: Props) {
  const [sheet, setSheet] = useState<CharacterSheet | null>(null)
  const [tab, setTab] = useState<Tab>('stats')

  const refresh = useCallback(() => {
    api.sheet(characterId).then(setSheet).catch(onError)
  }, [characterId, onError])

  useEffect(refresh, [refresh])

  if (!sheet) return <section className="sheet"><p className="hint">Loading sheet…</p></section>

  const { character, items, spells, slots, skills, saves } = sheet

  const handleAllocateStat = async (ability: string) => {
    try {
      setSheet(await api.allocateStat(character.id, ability))
    } catch (err) {
      onError(err)
    }
  }

  const handleUseScroll = async (item: CharacterItem) => {
    try {
      setSheet(await api.useSkillScroll(character.id, item.id))
      alert(`Learned skill from scroll: ${item.name}!`)
    } catch (err) {
      onError(err)
    }
  }

  const handleTransferItem = async (item: CharacterItem) => {
    const toCharInput = prompt('Enter recipient Character ID to transfer item to:')
    if (!toCharInput) return
    const toCharId = Number(toCharInput)
    if (!toCharId) return
    const qty = Number(prompt(`Quantity to send (max ${item.quantity}):`, '1')) || 1
    try {
      setSheet(await api.transferItem(character.id, {
        to_character_id: toCharId,
        item_id: item.id,
        quantity: qty,
      }))
      alert(`Sent ${qty}x ${item.name} to character #${toCharId}!`)
    } catch (err) {
      onError(err)
    }
  }

  const adjustHP = async (delta: number) => {
    try {
      await api.updateCharacter(character.id, {
        ...character,
        hp_current: Math.max(0, Math.min(character.hp_max, character.hp_current + delta)),
      })
      refresh()
    } catch (err) {
      onError(err)
    }
  }

  const setStatus = async (status: CharacterStatus) => {
    try {
      await api.updateCharacter(character.id, {
        ...character,
        status,
      })
      refresh()
    } catch (err) {
      onError(err)
    }
  }

  const handleLongRest = async () => {
    if (!confirm('Take a long rest? Restores HP, reset spell slots & death saves, recovers hit dice.')) return
    try {
      setSheet(await api.longRest(character.id))
    } catch (err) {
      onError(err)
    }
  }

  const handleShortRest = async () => {
    const available = character.hit_dice_total - character.hit_dice_used
    if (available <= 0) {
      alert('No hit dice available to spend!')
      return
    }
    const countStr = prompt(`How many hit dice to spend? (Available: ${available} ${character.hit_dice_die})`, '1')
    if (!countStr) return
    const count = parseInt(countStr, 10)
    if (isNaN(count) || count <= 0 || count > available) {
      alert('Invalid hit dice count')
      return
    }

    const conMod = character.modifiers.constitution
    const rollExpr = `${count}${character.hit_dice_die}${signed(count * conMod)}`
    onRoll(rollExpr, `Short Rest Hit Dice (${character.name})`, character.id)

    const healStr = prompt(`Enter HP rolled to heal (${rollExpr}):`, String(count * (4 + conMod)))
    if (!healStr) return
    const heal = Math.max(0, parseInt(healStr, 10) || 0)

    try {
      setSheet(await api.shortRest(character.id, { hit_dice_spent: count, hp_healed: heal }))
    } catch (err) {
      onError(err)
    }
  }

  const handleDeathSave = async () => {
    const d20 = Math.floor(Math.random() * 20) + 1
    onRoll('1d20', `Death Save (${character.name})`, character.id)
    try {
      setSheet(await api.deathSave(character.id, d20))
    } catch (err) {
      onError(err)
    }
  }

  const editCoins = async () => {
    const gpStr = prompt('Gold (GP):', String(character.gp))
    if (gpStr === null) return
    const spStr = prompt('Silver (SP):', String(character.sp))
    if (spStr === null) return
    const cpStr = prompt('Copper (CP):', String(character.cp))
    if (cpStr === null) return
    const epStr = prompt('Electrum (EP):', String(character.ep))
    if (epStr === null) return
    const ppStr = prompt('Platinum (PP):', String(character.pp))
    if (ppStr === null) return

    try {
      await api.updateCharacter(character.id, {
        ...character,
        gp: Math.max(0, parseInt(gpStr, 10) || 0),
        sp: Math.max(0, parseInt(spStr, 10) || 0),
        cp: Math.max(0, parseInt(cpStr, 10) || 0),
        ep: Math.max(0, parseInt(epStr, 10) || 0),
        pp: Math.max(0, parseInt(ppStr, 10) || 0),
      })
      refresh()
    } catch (err) {
      onError(err)
    }
  }

  const addItem = async () => {
    const name = prompt('Item name')
    if (!name) return
    const quantity = Number(prompt('Quantity', '1') ?? '1') || 1
    const costGP = Number(prompt('Cost in GP', '0') ?? '0') || 0
    const damage = (prompt('Damage, for example 1d8+3. Leave empty for gear.', '') ?? '').trim()
    const attackBonus = damage ? Number(prompt('Attack bonus', '5') ?? '0') || 0 : 0
    const damageType = damage ? (prompt('Damage type', 'slashing') ?? '').trim() : ''
    try {
      await api.addItem(character.id, {
        name,
        quantity,
        cost_cp: costGP * 100,
        damage,
        attack_bonus: attackBonus,
        damage_type: damageType,
        equipped: Boolean(damage),
      })
      refresh()
    } catch (err) {
      onError(err)
    }
  }

  const toggleProficiency = async (kind: 'skill' | 'save', name: string, trained: boolean) => {
    try {
      setSheet(await api.setProficiency(character.id, { kind, name, trained: !trained }))
    } catch (err) {
      onError(err)
    }
  }

  const toggleEquipped = async (id: number, name: string, quantity: number, equipped: boolean) => {
    try {
      await api.updateItem(id, { name, quantity, equipped: !equipped })
      refresh()
    } catch (err) {
      onError(err)
    }
  }

  const addSpell = async () => {
    const name = prompt('Spell name')
    if (!name) return
    const level = Number(prompt('Spell level (0 for a cantrip)', '1') ?? '1') || 0
    try {
      await api.addSpell(character.id, { name, level, prepared: true })
      refresh()
    } catch (err) {
      onError(err)
    }
  }

  const togglePrepared = async (id: number, name: string, level: number, prepared: boolean) => {
    try {
      await api.updateSpell(id, { name, level, prepared: !prepared })
      refresh()
    } catch (err) {
      onError(err)
    }
  }

  const spendSlot = async (level: number, total: number, used: number) => {
    try {
      await api.setSlot(character.id, level, { total, used: Math.min(total, used + 1) })
      refresh()
    } catch (err) {
      onError(err)
    }
  }

  const restoreSlots = async (level: number, total: number) => {
    try {
      await api.setSlot(character.id, level, { total, used: 0 })
      refresh()
    } catch (err) {
      onError(err)
    }
  }

  const setSlotTotal = async (level: number, currentTotal: number) => {
    const next = Number(prompt(`Slots at level ${level}`, String(currentTotal)) ?? '')
    if (Number.isNaN(next)) return
    try {
      await api.setSlot(character.id, level, { total: Math.max(0, next), used: 0 })
      refresh()
    } catch (err) {
      onError(err)
    }
  }

  return (
    <section className="sheet">
      <header className="sheet-header">
        <div>
          <strong>{character.name}</strong>
          <span className="hint">
            {[character.race, character.class, `level ${character.level}`].filter(Boolean).join(' · ')}
          </span>
        </div>
        <div className="status-selector">
          <select
            value={character.status || 'active'}
            onChange={(e) => setStatus(e.target.value as CharacterStatus)}
            className={`status-badge ${character.status}`}
          >
            {STATUS_OPTIONS.map((opt) => (
              <option key={opt.id} value={opt.id}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>
      </header>

      <div className="rest-bar" style={{ display: 'flex', gap: '8px', margin: '4px 0' }}>
        <button className="link" onClick={handleShortRest} title="Spend Hit Dice to regain HP">
          Short Rest ({character.hit_dice_total - character.hit_dice_used}/{character.hit_dice_total} {character.hit_dice_die})
        </button>
        <button className="link" onClick={handleLongRest} title="Reset HP, slots, death saves, recover HD">
          Long Rest
        </button>
      </div>

      <div className="vitals">
        <div className="vital">
          <span className="vital-label">HP</span>
          <span className="vital-value">
            {character.hp_current}
            <em>/{character.hp_max}</em>
            {character.hp_temp > 0 && <em className="temp">+{character.hp_temp}</em>}
          </span>
          <span className="vital-actions">
            <button className="pill" title="Lose 1 hit point" onClick={() => adjustHP(-1)}>
              − HP
            </button>
            <button className="pill" title="Heal 1 hit point" onClick={() => adjustHP(1)}>
              + HP
            </button>
          </span>
        </div>
        <div className="vital">
          <span className="vital-label">AC</span>
          <span className="vital-value">{character.armor_class}</span>
        </div>
        <div className="vital">
          <span className="vital-label">Speed</span>
          <span className="vital-value">{character.speed}</span>
        </div>
        <div className="vital">
          <span className="vital-label">Prof</span>
          <span className="vital-value">{signed(character.proficiency_bonus)}</span>
        </div>
      </div>

      {character.hp_current === 0 && (
        <div className="death-saves" style={{ padding: '8px', background: 'rgba(192, 57, 43, 0.1)', borderRadius: '4px', margin: '6px 0' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <strong>Death Saves</strong>
            <button className="pill on" onClick={handleDeathSave}>
              Roll Save
            </button>
          </div>
          <div style={{ display: 'flex', gap: '16px', marginTop: '4px', fontSize: '0.85em' }}>
            <span>
              Success: {'●'.repeat(character.death_save_successes)}{'○'.repeat(Math.max(0, 3 - character.death_save_successes))}
            </span>
            <span>
              Failure: {'●'.repeat(character.death_save_failures)}{'○'.repeat(Math.max(0, 3 - character.death_save_failures))}
            </span>
          </div>
        </div>
      )}

      <div className="wallet" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', margin: '6px 0', fontSize: '0.85em' }}>
        <span>
          💰 <strong>{character.gp}</strong> GP · <strong>{character.sp}</strong> SP · <strong>{character.cp}</strong> CP
          {(character.ep > 0 || character.pp > 0) && ` · ${character.ep} EP · ${character.pp} PP`}
        </span>
        <button className="link" onClick={editCoins}>
          edit coins
        </button>
      </div>

      <div className="tabs">
        {(['stats', 'progression', 'skills', 'inventory', 'spells'] as Tab[]).map((name) => (
          <button
            key={name}
            className={tab === name ? 'tab active' : 'tab'}
            onClick={() => setTab(name)}
          >
            {name === 'progression' ? 'tree' : name}
          </button>
        ))}
      </div>

      {tab === 'stats' && (
        <>
          {character.stat_points > 0 && (
            <div className="p-2 mb-2 bg-[#1f5156]/50 border-2 border-[#66fcf1] rounded text-[11px] text-[#66fcf1] flex items-center justify-between font-bold">
              <span>✨ {character.stat_points} Stat Points Available!</span>
              <span className="text-[10px] text-slate-300">Click [+] below to allocate</span>
            </div>
          )}

          {character.level_blocked && (
            <div className="p-2 mb-2 bg-red-950/60 border-2 border-red-500 rounded text-[11px] text-red-300 font-bold">
              🛑 Level Up Blocked by DM: {character.block_reason || 'Milestone locked'}
            </div>
          )}

          <div className="abilities">
            {ABILITY_KEYS.map((key) => (
              <div key={key} className="flex flex-col items-center">
                <button
                  className="ability w-full"
                  title={`Roll 1d20 ${signed(character.modifiers[key])}`}
                  onClick={() =>
                    onRoll(`1d20${signed(character.modifiers[key])}`, `${SHORT[key]} check`, character.id)
                  }
                >
                  <span className="ability-name">{SHORT[key]}</span>
                  <span className="ability-mod">{signed(character.modifiers[key])}</span>
                  <span className="ability-score">{character.abilities[key]}</span>
                </button>
                {character.stat_points > 0 && (
                  <button
                    onClick={() => handleAllocateStat(key)}
                    className="mt-1 px-2 py-0.5 bg-[#1f5156] border border-[#66fcf1] text-[#66fcf1] hover:bg-[#45a29e] hover:text-black text-[9px] font-bold rounded uppercase cursor-pointer"
                    title={`Add 1 point to ${SHORT[key]}`}
                  >
                    +1 {SHORT[key]}
                  </button>
                )}
              </div>
            ))}
          </div>

          {(character.damage_resistances || character.damage_vulnerabilities || character.damage_immunities || character.condition_immunities) && (
            <div className="modifiers-box" style={{ marginTop: '10px', fontSize: '0.85em', opacity: 0.9 }}>
              {character.damage_resistances && <div>🛡️ <strong>Resistances:</strong> {character.damage_resistances}</div>}
              {character.damage_vulnerabilities && <div>⚠️ <strong>Vulnerabilities:</strong> {character.damage_vulnerabilities}</div>}
              {character.damage_immunities && <div>✨ <strong>Immunities:</strong> {character.damage_immunities}</div>}
              {character.condition_immunities && <div>🚫 <strong>Condition Immunities:</strong> {character.condition_immunities}</div>}
            </div>
          )}
        </>
      )}

      {tab === 'progression' && (
        <SkillTreeTab sheet={sheet} onRefresh={refresh} onError={onError} />
      )}

      {tab === 'skills' && (
        <div className="list">
          <header>
            <h3>Saves</h3>
          </header>
          <ul>
            {saves.map((save) => (
              <li key={save.name} className="list-row">
                <button
                  className={save.trained ? 'pill on' : 'pill'}
                  title="Toggle proficiency"
                  onClick={() => toggleProficiency('save', save.name, save.trained)}
                >
                  {save.name.slice(0, 3).toUpperCase()}
                </button>
                <span className="list-name">{save.name}</span>
                <button
                  className="check-roll"
                  onClick={() =>
                    onRoll(`1d20${signed(save.modifier)}`, `${save.name} save`, character.id)
                  }
                >
                  {signed(save.modifier)}
                </button>
              </li>
            ))}
          </ul>

          <header>
            <h3>Skills</h3>
          </header>
          <ul>
            {skills.map((skill) => (
              <li key={skill.name} className="list-row">
                <button
                  className={skill.trained ? 'pill on' : 'pill'}
                  title="Toggle proficiency"
                  onClick={() => toggleProficiency('skill', skill.name, skill.trained)}
                >
                  {skill.ability.slice(0, 3).toUpperCase()}
                </button>
                <span className="list-name">{skill.name.replace(/_/g, ' ')}</span>
                <button
                  className="check-roll"
                  onClick={() =>
                    onRoll(
                      `1d20${signed(skill.modifier)}`,
                      skill.name.replace(/_/g, ' '),
                      character.id,
                    )
                  }
                >
                  {signed(skill.modifier)}
                </button>
              </li>
            ))}
          </ul>
        </div>
      )}

      {tab === 'inventory' && (
        <div className="list">
          <header>
            <h3>Carried</h3>
            <button className="btn-add" onClick={addItem}>
              + Add Item
            </button>
          </header>
          {items.length === 0 && <p className="hint">Nothing carried.</p>}
          <ul>
            {items.map((item) => (
              <li key={item.id} className="list-row">
                <button
                  className={item.equipped ? 'pill on' : 'pill'}
                  title={item.equipped ? 'Equipped' : 'Stowed'}
                  onClick={() => toggleEquipped(item.id, item.name, item.quantity, item.equipped)}
                >
                  {item.equipped ? 'equipped' : 'stowed'}
                </button>
                <span className="list-name">
                  {item.name}
                  {item.quantity > 1 && <em> ×{item.quantity}</em>}
                  {item.cost_cp > 0 && <em> · {item.cost_cp >= 100 ? `${item.cost_cp / 100}gp` : `${item.cost_cp}cp`}</em>}
                  {item.damage && (
                    <em>
                      {' '}· {signed(item.attack_bonus)} to hit, {item.damage} {item.damage_type}
                    </em>
                  )}
                </span>
                {item.damage && targetTokenId !== null && (
                  <button
                    className="pill on"
                    title="Attack the targeted token"
                    onClick={() => onAttack({ item_id: item.id, target_token_id: targetTokenId })}
                  >
                    attack
                  </button>
                )}
                {item.teaches_skill_id && (
                  <button
                    className="pill on"
                    title="Learn skill from this scroll"
                    onClick={() => handleUseScroll(item)}
                  >
                    learn
                  </button>
                )}
                <button
                  className="link"
                  title="Transfer item to another character"
                  onClick={() => handleTransferItem(item)}
                >
                  send
                </button>
                <button className="link" onClick={() => api.deleteItem(item.id).then(refresh).catch(onError)}>
                  remove
                </button>
              </li>
            ))}
          </ul>
        </div>
      )}

      {tab === 'spells' && (
        <div className="list">
          <header>
            <h3>Slots</h3>
          </header>
          <div className="slots">
            {Array.from({ length: 9 }, (_, index) => index + 1).map((level) => {
              const slot = slots.find((s) => s.level === level) ?? { level, total: 0, used: 0 }
              return (
                <div key={level} className={slot.total > 0 ? 'slot' : 'slot empty'}>
                  <span className="slot-level">{level}</span>
                  <button
                    className="slot-count"
                    title="Spend a slot"
                    onClick={() => spendSlot(level, slot.total, slot.used)}
                    disabled={slot.total === 0 || slot.used >= slot.total}
                  >
                    {slot.total - slot.used}/{slot.total}
                  </button>
                  <span className="slot-actions">
                    <button className="link" onClick={() => restoreSlots(level, slot.total)}>
                      rest
                    </button>
                    <button className="link" onClick={() => setSlotTotal(level, slot.total)}>
                      set
                    </button>
                  </span>
                </div>
              )
            })}
          </div>

          <header>
            <h3>Known</h3>
            <button className="btn-add" onClick={addSpell}>
              + Add Spell
            </button>
          </header>
          {spells.length === 0 && <p className="hint">No spells.</p>}
          <ul>
            {spells.map((spell) => (
              <li key={spell.id} className="list-row">
                <button
                  className={spell.prepared ? 'pill on' : 'pill'}
                  onClick={() => togglePrepared(spell.id, spell.name, spell.level, spell.prepared)}
                >
                  {spell.level === 0 ? 'cantrip' : `lvl ${spell.level}`}
                </button>
                <span className="list-name">{spell.name}</span>
                <button className="link" onClick={() => api.deleteSpell(spell.id).then(refresh).catch(onError)}>
                  remove
                </button>
              </li>
            ))}
          </ul>
        </div>
      )}
    </section>
  )
}
