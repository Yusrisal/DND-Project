import { useCallback, useEffect, useMemo, useState } from 'react'
import { api } from '../api/client'
import type {
  ClassSkill,
  CompendiumItem,
  CompendiumSpell,
  Monster,
  StatusEffect,
} from '../api/types'
import { matches, price, sellPrice } from './wiki-format'
import { GLOSSARY, HOW_TO_PLAY } from './wiki-rules'
import type { GlossaryEntry } from './wiki-rules'

type Props = {
  campaignId: number | null
  onError: (error: unknown) => void
}

type WikiSection = 'basics' | 'glossary' | 'equipment' | 'skills' | 'spells' | 'bestiary' | 'statuses'

const SECTIONS: { id: WikiSection; label: string }[] = [
  { id: 'basics', label: 'how to play' },
  { id: 'glossary', label: 'glossary' },
  { id: 'equipment', label: 'equipment' },
  { id: 'skills', label: 'skills' },
  { id: 'spells', label: 'spells' },
  { id: 'bestiary', label: 'bestiary' },
  { id: 'statuses', label: 'conditions' },
]


/**
 * The wiki is a read-only reference built from the same endpoints the DM tools
 * write to, so a monster or skill edited in the dashboard shows up here on the
 * next visit to the tab. Nothing is duplicated into the frontend except the
 * rules prose itself, which has no table behind it.
 */
export function WikiPanel({ campaignId, onError }: Props) {
  const [section, setSection] = useState<WikiSection>('basics')
  const [search, setSearch] = useState('')

  const [items, setItems] = useState<CompendiumItem[]>([])
  const [skills, setSkills] = useState<ClassSkill[]>([])
  const [spells, setSpells] = useState<CompendiumSpell[]>([])
  const [monsters, setMonsters] = useState<Monster[]>([])
  const [statuses, setStatuses] = useState<StatusEffect[]>([])
  const [openId, setOpenId] = useState<string | null>(null)

  // One fetch per section the first time it is opened. The catalogues only
  // change when the DM edits them, and switching back re-runs this effect.
  const load = useCallback(() => {
    switch (section) {
      case 'equipment':
        api
          .listCompendiumItems(campaignId, undefined, { page_size: 250, sort_by: 'name', sort_dir: 'asc' })
          .then((page) => setItems(page.items ?? []))
          .catch(onError)
        break
      case 'skills':
        api
          .listSkillTrees()
          .then((res) => setSkills(res ?? []))
          .catch(onError)
        break
      case 'spells':
        api
          .listCompendiumSpells(campaignId, undefined, { page_size: 250, sort_by: 'name', sort_dir: 'asc' })
          .then((page) => setSpells(page.items ?? []))
          .catch(onError)
        break
      case 'bestiary':
        api
          .listMonsters()
          .then((page) => setMonsters(page.items ?? []))
          .catch(onError)
        break
      case 'statuses':
        api
          .listStatusEffects(campaignId, undefined, { page_size: 250, sort_by: 'name', sort_dir: 'asc' })
          .then((page) => setStatuses(page.items ?? []))
          .catch(onError)
        break
    }
  }, [section, campaignId, onError])

  useEffect(load, [load])

  const visibleGlossary = useMemo(
    () => GLOSSARY.filter((entry) => matches(search, entry.term, entry.body, entry.tag)),
    [search],
  )

  const toggle = (key: string) => setOpenId((current) => (current === key ? null : key))

  return (
    <section className="space-y-2">
      <header>
        <h2>Wiki</h2>
        <button className="btn-add" onClick={load} title="Reload from the server">
          refresh
        </button>
      </header>

      <div className="wiki-nav">
        {SECTIONS.map(({ id, label }) => (
          <button
            key={id}
            className={section === id ? 'tab active' : 'tab'}
            onClick={() => {
              setSection(id)
              setOpenId(null)
            }}
          >
            {label}
          </button>
        ))}
      </div>

      {section !== 'basics' && (
        <div className="filter-bar">
          <input
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder={`Search ${section}`}
          />
        </div>
      )}

      {section === 'basics' && (
        <div className="wiki-prose">
          {HOW_TO_PLAY.map((chapter) => (
            <article key={chapter.heading}>
              <h3>{chapter.heading}</h3>
              {chapter.paragraphs.map((paragraph, index) => (
                <p key={index}>{paragraph}</p>
              ))}
            </article>
          ))}
        </div>
      )}

      {section === 'glossary' && (
        <div className="list">
          {visibleGlossary.length === 0 && <p className="hint">Nothing matches that.</p>}
          <ul>
            {visibleGlossary.map((entry: GlossaryEntry) => (
              <li key={entry.term} className="list-row wiki-entry">
                <div>
                  <span className="list-name">
                    <strong>{entry.term}</strong> <em className="wiki-tag">[{entry.tag}]</em>
                  </span>
                  <p className="hint wiki-body">{entry.body}</p>
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}

      {section === 'equipment' && (
        <div className="list">
          {items.length === 0 && <p className="hint">No equipment in the compendium yet.</p>}
          <ul>
            {items
              .filter((item) => matches(search, item.name, item.category, item.rarity, item.damage))
              .map((item) => (
                <li key={item.id} className="list-row wiki-entry">
                  <div>
                    <span className="list-name">
                      <strong>{item.name}</strong>{' '}
                      <em className="wiki-tag">
                        [{[item.category, item.rarity].filter(Boolean).join(' · ') || 'gear'}]
                      </em>
                    </span>
                    <p className="hint wiki-body">
                      {price(item.cost_cp)}
                      {item.cost_cp > 0 && ` · sells for ${price(sellPrice(item.cost_cp))}`}
                      {item.weight > 0 && ` · ${item.weight} lb`}
                      {item.damage && ` · ${item.damage} ${item.damage_type}`}
                      {item.attack_bonus !== 0 && ` · ${item.attack_bonus > 0 ? '+' : ''}${item.attack_bonus} to hit`}
                    </p>
                    {item.description && <p className="hint wiki-body">{item.description}</p>}
                  </div>
                </li>
              ))}
          </ul>
        </div>
      )}

      {section === 'skills' && (
        <div className="list">
          {skills.length === 0 && <p className="hint">No class skills seeded.</p>}
          <ul>
            {skills
              .filter((skill) => matches(search, skill.name, skill.class_name, skill.damage, skill.description))
              .map((skill) => (
                <li key={skill.id} className="list-row wiki-entry">
                  <div>
                    <span className="list-name">
                      <strong>{skill.name}</strong>{' '}
                      <em className="wiki-tag">
                        [{skill.class_name} · tier {skill.tier} · needs level {skill.req_level}]
                      </em>
                    </span>
                    <p className="hint wiki-body">
                      max rank {skill.max_rank}
                      {skill.damage && ` · ${skill.damage} ${skill.damage_type}`}
                      {skill.cooldown_rounds > 0 && ` · ${skill.cooldown_rounds} round cooldown`}
                      {skill.cost_resource && ` · costs ${skill.cost_resource}`}
                      {skill.parent_skill_id !== null && ' · needs its parent skill first'}
                    </p>
                    {skill.description && <p className="hint wiki-body">{skill.description}</p>}
                  </div>
                </li>
              ))}
          </ul>
        </div>
      )}

      {section === 'spells' && (
        <div className="list">
          {spells.length === 0 && <p className="hint">No spells in the compendium yet.</p>}
          <ul>
            {spells
              .filter((spell) => matches(search, spell.name, spell.school, spell.damage, spell.description))
              .map((spell) => (
                <li key={spell.id} className="list-row wiki-entry">
                  <div>
                    <span className="list-name">
                      <strong>{spell.name}</strong>{' '}
                      <em className="wiki-tag">
                        [{spell.level === 0 ? 'cantrip' : `level ${spell.level}`}
                        {spell.school && ` · ${spell.school}`}]
                      </em>
                    </span>
                    <p className="hint wiki-body">
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
                    {spell.description && <p className="hint wiki-body">{spell.description}</p>}
                  </div>
                </li>
              ))}
          </ul>
        </div>
      )}

      {section === 'bestiary' && (
        <div className="list">
          {monsters.length === 0 && <p className="hint">No monsters yet.</p>}
          <ul>
            {monsters
              .filter((monster) => matches(search, monster.name, monster.type, monster.challenge_rating))
              .map((monster) => (
                <li key={monster.id} className="list-row wiki-entry">
                  <div>
                    <button className="wiki-disclosure" onClick={() => toggle(`monster-${monster.id}`)}>
                      <strong>{monster.name}</strong>{' '}
                      <em className="wiki-tag">
                        [CR {monster.challenge_rating || '?'} · {monster.armor_class} AC · {monster.hp_average} hp]
                      </em>
                    </button>
                    {openId === `monster-${monster.id}` && (
                      <MonsterDetail monster={monster} onError={onError} />
                    )}
                  </div>
                </li>
              ))}
          </ul>
        </div>
      )}

      {section === 'statuses' && (
        <div className="list">
          {statuses.length === 0 && <p className="hint">No status effects yet.</p>}
          <ul>
            {statuses
              .filter((status) => matches(search, status.name, status.kind, status.effect_description))
              .map((status) => (
                <li key={status.id} className="list-row wiki-entry">
                  <div>
                    <span className="list-name">
                      <strong>{status.name}</strong> <em className="wiki-tag">[{status.kind}]</em>
                    </span>
                    <p className="hint wiki-body">
                      {status.default_duration_rounds > 0
                        ? `lasts ${status.default_duration_rounds} rounds by default`
                        : 'lasts until removed'}
                    </p>
                    {status.effect_description && <p className="hint wiki-body">{status.effect_description}</p>}
                  </div>
                </li>
              ))}
          </ul>
        </div>
      )}
    </section>
  )
}

/** Actions live on a second endpoint, so they are only fetched when expanded. */
function MonsterDetail({ monster, onError }: { monster: Monster; onError: (error: unknown) => void }) {
  const [actions, setActions] = useState<{ id: number; name: string; attack_bonus: number; damage: string; damage_type: string; description: string }[] | null>(null)

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
      {tags.length > 0 && <p className="hint wiki-body">{tags.join(' · ')}</p>}

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

      {monster.damage_resistances && <p className="hint wiki-body">Resists: {monster.damage_resistances}</p>}
      {monster.damage_immunities && <p className="hint wiki-body">Immune: {monster.damage_immunities}</p>}
      {monster.damage_vulnerabilities && <p className="hint wiki-body">Vulnerable: {monster.damage_vulnerabilities}</p>}
      {monster.traits && <p className="hint wiki-body">{monster.traits}</p>}

      {actions === null && <p className="hint">Loading actions…</p>}
      {actions !== null && actions.length === 0 && <p className="hint">No actions recorded.</p>}
      {actions !== null &&
        actions.map((action) => (
          <p key={action.id} className="hint wiki-body">
            <strong>{action.name}</strong>
            {action.damage && ` — ${action.attack_bonus >= 0 ? '+' : ''}${action.attack_bonus} to hit, ${action.damage} ${action.damage_type}`}
            {action.description && ` — ${action.description}`}
          </p>
        ))}
    </div>
  )
}
