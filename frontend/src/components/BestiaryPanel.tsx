import { useCallback, useEffect, useState } from 'react'
import { api } from '../api/client'
import type { Monster, MonsterStatBlock } from '../api/types'

type Props = {
  mapId: number | null
  /** The token currently targeted, which a monster action attacks. */
  targetTokenId: number | null
  onAttack: (request: { action_id: number; attacker_token_id?: number; target_token_id: number }) => void
  attackerTokenId: number | null
  onError: (error: unknown) => void
}

const signed = (value: number) => (value >= 0 ? `+${value}` : String(value))

/**
 * The bestiary: stat blocks the DM writes once and spawns from. A spawned token
 * copies the vitals, so each goblin takes its own damage while sharing the
 * block for its actions.
 */
export function BestiaryPanel({ mapId, targetTokenId, onAttack, attackerTokenId, onError }: Props) {
  const [monsters, setMonsters] = useState<Monster[]>([])
  const [search, setSearch] = useState('')
  const [openId, setOpenId] = useState<number | null>(null)
  const [block, setBlock] = useState<MonsterStatBlock | null>(null)

  const refresh = useCallback(() => {
    api
      .listMonsters(search)
      .then((page) => setMonsters(page.items))
      .catch(onError)
  }, [search, onError])

  useEffect(() => {
    const timer = setTimeout(refresh, 250)
    return () => clearTimeout(timer)
  }, [refresh])

  useEffect(() => {
    if (openId === null) {
      setBlock(null)
      return
    }
    api.statBlock(openId).then(setBlock).catch(onError)
  }, [openId, onError])

  const create = async () => {
    const name = prompt('Monster name')
    if (!name) return
    const ac = Number(prompt('Armour class', '12') ?? '12') || 12
    const hp = Number(prompt('Average hit points', '7') ?? '7') || 7
    try {
      const created = await api.createMonster({ name, armor_class: ac, hp_average: hp })
      refresh()
      setOpenId(created.id)
    } catch (err) {
      onError(err)
    }
  }

  const addAction = async (monsterId: number) => {
    const name = prompt('Action name, for example Scimitar')
    if (!name) return
    const bonus = Number(prompt('Attack bonus', '4') ?? '0') || 0
    const damage = prompt('Damage expression, for example 1d6+2', '1d6+2') ?? ''
    const damageType = prompt('Damage type', 'slashing') ?? ''
    try {
      await api.addAction(monsterId, {
        name,
        attack_bonus: bonus,
        damage,
        damage_type: damageType,
      })
      setBlock(await api.statBlock(monsterId))
    } catch (err) {
      onError(err)
    }
  }

  const spawn = async (monster: Monster) => {
    if (mapId === null) return
    const count = Number(prompt(`How many ${monster.name}?`, '1') ?? '1') || 1
    try {
      // The socket echo puts them on the board, so nothing to set here.
      await api.spawn({ monster_id: monster.id, map_id: mapId, count, roll_hp: true })
    } catch (err) {
      onError(err)
    }
  }

  return (
    <section>
      <header>
        <h2>Bestiary</h2>
        <button className="btn-add" onClick={create}>
          + Add Monster
        </button>
      </header>

      <input
        value={search}
        onChange={(event) => setSearch(event.target.value)}
        placeholder="Search monsters"
      />

      <ul>
        {monsters.map((monster) => (
          <li key={monster.id}>
            <div className="monster-row">
              <button
                className={monster.id === openId ? 'row active' : 'row'}
                onClick={() => setOpenId(monster.id === openId ? null : monster.id)}
              >
                {monster.name}
                <em className="owner">
                  AC {monster.armor_class} · {monster.hp_average} hp · CR {monster.challenge_rating}
                </em>
              </button>
              {mapId !== null && (
                <button className="link" onClick={() => spawn(monster)}>
                  spawn
                </button>
              )}
            </div>

            {monster.id === openId && block && (
              <div className="statblock">
                <p className="hint">
                  {[block.monster.size, block.monster.type, block.monster.alignment]
                    .filter(Boolean)
                    .join(' · ')}
                </p>

                <div className="abilities small">
                  {(['strength', 'dexterity', 'constitution', 'intelligence', 'wisdom', 'charisma'] as const).map(
                    (key) => (
                      <span key={key} className="ability">
                        <span className="ability-name">{key.slice(0, 3).toUpperCase()}</span>
                        <span className="ability-mod">{signed(block.monster.modifiers[key])}</span>
                      </span>
                    ),
                  )}
                </div>

                {block.proficiencies.length > 0 && (
                  <p className="hint">
                    {block.proficiencies
                      .map((p) => `${p.name.replace(/_/g, ' ')} ${signed(p.modifier)}`)
                      .join(', ')}
                  </p>
                )}

                <header>
                  <h3>Actions</h3>
                  <button className="btn-add" onClick={() => addAction(monster.id)}>
                    + Add Action
                  </button>
                </header>

                {block.actions.length === 0 && <p className="hint">No actions yet.</p>}

                <ul>
                  {block.actions.map((action) => (
                    <li key={action.id} className="list-row">
                      <span className="list-name">
                        {action.name}
                        {action.damage && (
                          <em>
                            {signed(action.attack_bonus)} to hit, {action.damage} {action.damage_type}
                          </em>
                        )}
                      </span>
                      {action.damage && targetTokenId !== null && (
                        <button
                          className="pill on"
                          title="Attack the targeted token"
                          onClick={() =>
                            onAttack({
                              action_id: action.id,
                              attacker_token_id: attackerTokenId ?? undefined,
                              target_token_id: targetTokenId,
                            })
                          }
                        >
                          attack
                        </button>
                      )}
                      <button
                        className="link"
                        onClick={() =>
                          api
                            .deleteAction(action.id)
                            .then(() => api.statBlock(monster.id).then(setBlock))
                            .catch(onError)
                        }
                      >
                        remove
                      </button>
                    </li>
                  ))}
                </ul>

                {block.monster.traits && <p className="hint">{block.monster.traits}</p>}
              </div>
            )}
          </li>
        ))}
      </ul>

      {monsters.length === 0 && <p className="hint">No monsters yet.</p>}
    </section>
  )
}
