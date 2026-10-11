import { useCallback, useEffect, useMemo, useState } from 'react'
import { api } from '../api/client'
import type {
  ClassSkill,
  CompendiumItem,
  CompendiumSpell,
  Monster,
  MonsterAction,
  StatusEffect,
} from '../api/types'
import { matches, price, sellPrice } from './wiki-format'
import { CORE_CONDITIONS, GLOSSARY, HOW_TO_PLAY } from './wiki-rules'

type Props = {
  campaignId: number | null
  onClose: () => void
  onError: (error: unknown) => void
}

type WikiSection = 'basics' | 'glossary' | 'equipment' | 'skills' | 'spells' | 'bestiary' | 'conditions'

const SECTIONS: { id: WikiSection; label: string; blurb: string }[] = [
  { id: 'basics', label: 'How to play', blurb: 'Start here' },
  { id: 'glossary', label: 'Glossary', blurb: 'Every term, plainly' },
  { id: 'equipment', label: 'Equipment', blurb: 'Gear, weapons, prices' },
  { id: 'skills', label: 'Skills', blurb: 'Class trees, tier by tier' },
  { id: 'spells', label: 'Spells', blurb: 'What casters can do' },
  { id: 'bestiary', label: 'Bestiary', blurb: 'Everything that fights back' },
  { id: 'conditions', label: 'Conditions', blurb: 'Poisoned, prone, and worse' },
]

/**
 * The wiki is a full page rather than a sidebar panel: reference reading wants
 * the width, and nobody consults a rulebook and the battle map at the same
 * moment. The catalogue sections read the same endpoints the DM tools write to,
 * so anything edited there shows up here. Only the prose, the glossary and the
 * core conditions are hand-written, in `wiki-rules.ts`.
 */
export function WikiPage({ campaignId, onClose, onError }: Props) {
  const [section, setSection] = useState<WikiSection>('basics')
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(false)

  const [items, setItems] = useState<CompendiumItem[]>([])
  const [skills, setSkills] = useState<ClassSkill[]>([])
  const [spells, setSpells] = useState<CompendiumSpell[]>([])
  const [monsters, setMonsters] = useState<Monster[]>([])
  const [statuses, setStatuses] = useState<StatusEffect[]>([])
  const [openId, setOpenId] = useState<number | null>(null)

  // Escape closes the page, which is what a full-screen overlay should do.
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  // One fetch per section, the first time it is opened and on refresh.
  const load = useCallback(() => {
    const done = () => setLoading(false)
    const fail = (err: unknown) => {
      setLoading(false)
      onError(err)
    }

    switch (section) {
      case 'equipment':
        setLoading(true)
        api
          .listCompendiumItems(campaignId, undefined, { page_size: 250, sort_by: 'name', sort_dir: 'asc' })
          .then((page) => setItems(page.items ?? []))
          .then(done, fail)
        break
      case 'skills':
        setLoading(true)
        api
          .listSkillTrees()
          .then((res) => setSkills(res ?? []))
          .then(done, fail)
        break
      case 'spells':
        setLoading(true)
        api
          .listCompendiumSpells(campaignId, undefined, { page_size: 250, sort_by: 'name', sort_dir: 'asc' })
          .then((page) => setSpells(page.items ?? []))
          .then(done, fail)
        break
      case 'bestiary':
        setLoading(true)
        api
          .listMonsters()
          .then((page) => setMonsters(page.items ?? []))
          .then(done, fail)
        break
      case 'conditions':
        setLoading(true)
        api
          .listStatusEffects(campaignId, undefined, { page_size: 250, sort_by: 'name', sort_dir: 'asc' })
          .then((page) => setStatuses(page.items ?? []))
          .then(done, fail)
        break
    }
  }, [section, campaignId, onError])

  useEffect(load, [load])

  const visibleGlossary = useMemo(
    () => GLOSSARY.filter((entry) => matches(search, entry.term, entry.body, entry.tag)),
    [search],
  )

  const visibleConditions = useMemo(
    () => CORE_CONDITIONS.filter((entry) => matches(search, entry.name, entry.body)),
    [search],
  )

  const current = SECTIONS.find((entry) => entry.id === section)

  return (
    <div className="wiki-page">
      <header className="wiki-topbar">
        <div className="wiki-brand">
          <h1>Wiki</h1>
          <span className="wiki-blurb">{current?.blurb}</span>
        </div>

        <div className="wiki-topbar-actions">
          {section !== 'basics' && (
            <input
              className="wiki-search"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder={`Search ${current?.label.toLowerCase()}`}
              aria-label={`Search ${current?.label.toLowerCase()}`}
            />
          )}
          <button className="btn-add" onClick={onClose}>
            back to the table
          </button>
        </div>
      </header>

      <div className="wiki-body">
        <nav className="wiki-sidenav" aria-label="Wiki sections">
          {SECTIONS.map(({ id, label }) => (
            <button
              key={id}
              className={section === id ? 'wiki-navlink active' : 'wiki-navlink'}
              aria-current={section === id ? 'page' : undefined}
              onClick={() => {
                setSection(id)
                setSearch('')
                setOpenId(null)
              }}
            >
              {label}
            </button>
          ))}
        </nav>

        <main className="wiki-content">
          {section === 'basics' && (
            <div className="wiki-prose">
              {HOW_TO_PLAY.map((chapter) => (
                <article key={chapter.heading}>
                  <h2>{chapter.heading}</h2>
                  {chapter.paragraphs.map((paragraph, index) => (
                    <p key={index}>{paragraph}</p>
                  ))}
                </article>
              ))}
            </div>
          )}

          {section === 'glossary' && (
            <Cards empty={visibleGlossary.length === 0} emptyText="Nothing matches that.">
              {visibleGlossary.map((entry) => (
                <Card key={entry.term} title={entry.term} tag={entry.tag}>
                  <p>{entry.body}</p>
                </Card>
              ))}
            </Cards>
          )}

          {section === 'equipment' && (
            <Cards
              loading={loading}
              empty={items.length === 0}
              emptyText="No equipment in the compendium yet. The DM adds it under TOOLS."
            >
              {items
                .filter((item) => matches(search, item.name, item.category, item.rarity, item.damage))
                .map((item) => (
                  <Card
                    key={item.id}
                    title={item.name}
                    tag={[item.category, item.rarity].filter(Boolean).join(' · ') || 'gear'}
                  >
                    <p className="wiki-stats">
                      {price(item.cost_cp)}
                      {item.cost_cp > 0 && ` · sells for ${price(sellPrice(item.cost_cp))}`}
                      {item.weight > 0 && ` · ${item.weight} lb`}
                      {item.damage && ` · ${item.damage} ${item.damage_type}`}
                      {item.attack_bonus !== 0 && ` · ${item.attack_bonus > 0 ? '+' : ''}${item.attack_bonus} to hit`}
                    </p>
                    {item.description && <p>{item.description}</p>}
                  </Card>
                ))}
            </Cards>
          )}

          {section === 'skills' && (
            <Cards loading={loading} empty={skills.length === 0} emptyText="No class skills found.">
              {skills
                .filter((skill) => matches(search, skill.name, skill.class_name, skill.damage, skill.description))
                .map((skill) => (
                  <Card
                    key={skill.id}
                    title={skill.name}
                    tag={`${skill.class_name} · tier ${skill.tier} · level ${skill.req_level}`}
                  >
                    <p className="wiki-stats">
                      {`max rank ${skill.max_rank}`}
                      {skill.damage && ` · ${skill.damage} ${skill.damage_type}`}
                      {skill.cooldown_rounds > 0 && ` · ${skill.cooldown_rounds} round cooldown`}
                      {skill.cost_resource && ` · costs ${skill.cost_resource}`}
                      {skill.parent_skill_id !== null && ' · needs its parent skill first'}
                    </p>
                    {skill.description && <p>{skill.description}</p>}
                  </Card>
                ))}
            </Cards>
          )}

          {section === 'spells' && (
            <Cards
              loading={loading}
              empty={spells.length === 0}
              emptyText="No spells in the compendium yet. The DM adds them under TOOLS."
            >
              {spells
                .filter((spell) => matches(search, spell.name, spell.school, spell.damage, spell.description))
                .map((spell) => (
                  <Card
                    key={spell.id}
                    title={spell.name}
                    tag={`${spell.level === 0 ? 'cantrip' : `level ${spell.level}`}${spell.school ? ` · ${spell.school}` : ''}`}
                  >
                    <p className="wiki-stats">
                      {[
                        spell.casting_time && `cast ${spell.casting_time}`,
                        spell.range_area && `range ${spell.range_area}`,
                        spell.duration && `lasts ${spell.duration}`,
                        spell.damage && `${spell.damage} ${spell.damage_type}`,
                        spell.save_ability && `${spell.save_ability.toUpperCase()} save`,
                      ]
                        .filter(Boolean)
                        .join(' · ')}
                    </p>
                    {spell.description && <p>{spell.description}</p>}
                  </Card>
                ))}
            </Cards>
          )}

          {section === 'bestiary' && (
            <Cards loading={loading} empty={monsters.length === 0} emptyText="No monsters found.">
              {monsters
                .filter((monster) => matches(search, monster.name, monster.type, monster.challenge_rating))
                .map((monster) => (
                  <Card
                    key={monster.id}
                    title={monster.name}
                    tag={`CR ${monster.challenge_rating || '?'} · AC ${monster.armor_class} · ${monster.hp_average} hp`}
                    onToggle={() => setOpenId(openId === monster.id ? null : monster.id)}
                    open={openId === monster.id}
                  >
                    {openId === monster.id && <MonsterDetail monster={monster} onError={onError} />}
                  </Card>
                ))}
            </Cards>
          )}

          {section === 'conditions' && (
            <>
              <Cards empty={visibleConditions.length === 0} emptyText="Nothing matches that.">
                {visibleConditions.map((entry) => (
                  <Card key={entry.name} title={entry.name} tag="condition">
                    <p>{entry.body}</p>
                  </Card>
                ))}
              </Cards>

              {statuses.length > 0 && (
                <>
                  <h2 className="wiki-subhead">Added by your DM</h2>
                  <Cards empty={false}>
                    {statuses
                      .filter((status) => matches(search, status.name, status.kind, status.effect_description))
                      .map((status) => (
                        <Card key={status.id} title={status.name} tag={status.kind}>
                          <p className="wiki-stats">
                            {status.default_duration_rounds > 0
                              ? `lasts ${status.default_duration_rounds} rounds by default`
                              : 'lasts until removed'}
                          </p>
                          {status.effect_description && <p>{status.effect_description}</p>}
                        </Card>
                      ))}
                  </Cards>
                </>
              )}
            </>
          )}
        </main>
      </div>
    </div>
  )
}

function Cards({
  children,
  loading,
  empty,
  emptyText,
}: {
  children: React.ReactNode
  loading?: boolean
  empty: boolean
  emptyText?: string
}) {
  if (loading) return <p className="hint">Loading…</p>
  if (empty) return <p className="hint">{emptyText}</p>
  return <div className="wiki-grid">{children}</div>
}

function Card({
  title,
  tag,
  children,
  onToggle,
  open,
}: {
  title: string
  tag: string
  children: React.ReactNode
  onToggle?: () => void
  open?: boolean
}) {
  return (
    <article className={open ? 'wiki-card open' : 'wiki-card'}>
      {onToggle ? (
        <button className="wiki-card-head" onClick={onToggle} aria-expanded={open}>
          <h3>{title}</h3>
          <span className="wiki-tag">{tag}</span>
        </button>
      ) : (
        <div className="wiki-card-head">
          <h3>{title}</h3>
          <span className="wiki-tag">{tag}</span>
        </div>
      )}
      {children}
    </article>
  )
}

/** Actions live on a second endpoint, so they are only fetched when expanded. */
function MonsterDetail({ monster, onError }: { monster: Monster; onError: (error: unknown) => void }) {
  const [actions, setActions] = useState<MonsterAction[] | null>(null)

  useEffect(() => {
    api
      .statBlock(monster.id)
      .then((block) => setActions(block.actions ?? []))
      .catch(onError)
  }, [monster.id, onError])

  const tags = [
    monster.size,
    monster.type,
    monster.alignment,
    monster.speed > 0 && `${monster.speed} ft speed`,
    monster.senses,
    monster.languages,
  ].filter(Boolean)

  return (
    <div className="wiki-detail">
      {tags.length > 0 && <p className="wiki-stats">{tags.join(' · ')}</p>}

      <div className="abilities small">
        {(['strength', 'dexterity', 'constitution', 'intelligence', 'wisdom', 'charisma'] as const).map((key) => (
          <span key={key} className="ability">
            <span className="ability-name">{key.slice(0, 3).toUpperCase()}</span>
            <span className="ability-mod">
              {monster.modifiers[key] >= 0 ? `+${monster.modifiers[key]}` : monster.modifiers[key]}
            </span>
          </span>
        ))}
      </div>

      {monster.damage_resistances && <p className="wiki-stats">Resists: {monster.damage_resistances}</p>}
      {monster.damage_immunities && <p className="wiki-stats">Immune: {monster.damage_immunities}</p>}
      {monster.damage_vulnerabilities && <p className="wiki-stats">Vulnerable: {monster.damage_vulnerabilities}</p>}
      {monster.traits && <p>{monster.traits}</p>}

      {actions === null && <p className="hint">Loading actions…</p>}
      {actions !== null && actions.length === 0 && <p className="hint">No actions recorded.</p>}
      {actions?.map((action) => (
        <p key={action.id}>
          <strong>{action.name}</strong>
          {action.damage &&
            ` — ${action.attack_bonus >= 0 ? '+' : ''}${action.attack_bonus} to hit, ${action.damage} ${action.damage_type}`}
          {action.description && ` — ${action.description}`}
        </p>
      ))}
    </div>
  )
}
