import { useEffect, useMemo, useRef, useState } from 'react'
import { api } from '../api/client'
import type { AttackDetail, Campaign, DiceRoll, FeedEntry } from '../api/types'

type Props = {
  feed: FeedEntry[]
  connected: boolean
  campaign?: Campaign | null
  isDM?: boolean
  characterId: number | null
  onSend: (body: string) => void
  onRoll: (expression: string, label?: string, characterId?: number | null) => void
  onUpdateCampaign?: (campaign: Campaign) => void
}

const QUICK_ROLLS = ['1d20', '2d20kh1', '2d20kl1', '1d4', '1d6', '1d8', '1d10', '1d12', '4d6kh3']

const time = (iso: string) =>
  new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })

function asAttack(roll: DiceRoll): AttackDetail | null {
  const detail = roll.detail as unknown as { kind?: string }
  return detail?.kind === 'attack' ? (detail as unknown as AttackDetail) : null
}

function breakdown(roll: DiceRoll): string {
  const parts = roll.detail.groups.map((group) => {
    const kept = group.rolls.filter((value) => !group.dropped.includes(value))
    const shown = group.dropped.length > 0 ? `${kept.join(' + ')} (dropped ${group.dropped.join(', ')})` : group.rolls.join(' + ')
    return `${group.expression}: ${shown}`
  })
  if (roll.detail.modifier !== 0) {
    parts.push(`${roll.detail.modifier > 0 ? '+' : ''}${roll.detail.modifier}`)
  }
  return parts.join('  ')
}

export function ChatPanel({
  feed,
  connected,
  campaign,
  isDM = false,
  characterId,
  onSend,
  onRoll,
  onUpdateCampaign,
}: Props) {
  const [draft, setDraft] = useState('')
  const [showFullArchive, setShowFullArchive] = useState(false)
  const endRef = useRef<HTMLDivElement | null>(null)

  const retentionDays = campaign?.chat_retention_days ?? 0

  const handleRetentionChange = async (days: number) => {
    if (!campaign) return
    try {
      const updated = await api.setCampaignRetention(campaign.id, { chat_retention_days: days })
      onUpdateCampaign?.(updated)
    } catch {
      // Ignored
    }
  }

  // Filter feed based on lingering retention window
  const visibleFeed = useMemo(() => {
    if (showFullArchive || retentionDays <= 0) return feed
    const cutoff = Date.now() - retentionDays * 24 * 60 * 60 * 1000
    return feed.filter((entry) => new Date(entry.at).getTime() >= cutoff)
  }, [feed, retentionDays, showFullArchive])

  useEffect(() => {
    endRef.current?.scrollIntoView({ block: 'end' })
  }, [visibleFeed.length])

  const submit = (event: React.FormEvent) => {
    event.preventDefault()
    const body = draft.trim()
    if (!body) return

    const match = body.match(/^\/(?:roll|r)\s+(\S+)\s*(.*)$/i)
    if (match) {
      onRoll(match[1], match[2], characterId)
    } else {
      onSend(body)
    }
    setDraft('')
  }

  return (
    <aside className="chat">
      <header className="chat-header flex items-center justify-between pb-2 border-b border-border">
        <div className="flex items-center gap-2">
          <h2 className="m-0 text-sm font-semibold">Table Log</h2>
          <span className={connected ? 'dot online' : 'dot offline'} />
        </div>

        {/* DM Lingering / Retention Window Control */}
        {isDM && campaign ? (
          <div className="flex items-center gap-1.5 text-xs text-muted">
            <span>Show:</span>
            <select
              className="bg-bg border border-border rounded px-1.5 py-0.5 text-[11px] text-text"
              value={retentionDays}
              onChange={(e) => handleRetentionChange(parseInt(e.target.value, 10))}
              title="Select how long chat and story history lingers for the table"
            >
              <option value={0}>All Time (Permanent)</option>
              <option value={30}>Last 1 Month</option>
              <option value={14}>Last 2 Weeks</option>
              <option value={7}>Last 1 Week</option>
              <option value={1}>Last 24 Hours</option>
            </select>
          </div>
        ) : retentionDays > 0 ? (
          <span className="text-[10px] text-muted">
            Showing past {retentionDays === 1 ? '24 hours' : `${retentionDays} days`}
          </span>
        ) : null}
      </header>

      {isDM && retentionDays > 0 && (
        <div className="flex items-center justify-between px-2 py-1 bg-panel/50 text-[11px] text-muted border-b border-border">
          <span>{visibleFeed.length} entries visible to table</span>
          <button
            className="link text-accent text-[11px]"
            onClick={() => setShowFullArchive(!showFullArchive)}
          >
            {showFullArchive ? 'Apply Table Filter' : '👁️ View Full Archive'}
          </button>
        </div>
      )}

      <div className="chat-feed">
        {visibleFeed.length === 0 && (
          <p className="hint text-xs text-muted">No messages in this visibility window.</p>
        )}

        {visibleFeed.map((entry) =>
          entry.kind === 'message' ? (
            <div key={`m${entry.message.id}`} className="feed-message">
              <span className="feed-meta text-xs text-muted">
                {entry.message.author_name} · {time(entry.at)}
              </span>
              <p className="m-0 text-sm">{entry.message.body}</p>
            </div>
          ) : (
            <AttackOrRoll key={`r${entry.roll.id}`} roll={entry.roll} at={entry.at} />
          ),
        )}
        <div ref={endRef} />
      </div>

      <div className="quick-rolls">
        {QUICK_ROLLS.map((expression) => (
          <button key={expression} onClick={() => onRoll(expression, '', characterId)}>
            {expression}
          </button>
        ))}
      </div>

      <form className="chat-input" onSubmit={submit}>
        <input
          value={draft}
          onChange={(event) => setDraft(event.target.value)}
          placeholder="Message, or /roll 1d20+5"
          maxLength={2000}
        />
        <button type="submit" className="btn-primary">
          Send
        </button>
      </form>
    </aside>
  )
}

function AttackOrRoll({ roll, at }: { roll: DiceRoll; at: string }) {
  const attack = asAttack(roll)

  if (!attack) {
    return (
      <div className="feed-roll">
        <span className="feed-meta text-xs text-muted">
          {roll.author_name} · {time(at)}
        </span>
        <div className="roll-line">
          {roll.label && <span className="roll-label">{roll.label}</span>}
          <code>{roll.expression}</code>
          <strong className="roll-total">{roll.total}</strong>
        </div>
        <div className="roll-breakdown">{breakdown(roll)}</div>
      </div>
    )
  }

  const { hit, critical, fumble, target_ac, attack_roll, damage, damage_type, target, weapon_name, target_name } = attack
  const verdict = critical ? 'Critical!' : fumble ? 'Fumble' : hit ? 'Hit' : 'Miss'
  const verdictClass = critical ? 'crit' : hit ? 'hit' : 'miss'

  return (
    <div className={`attack-card ${verdictClass}`}>
      <span className="feed-meta text-xs text-muted">
        {attack.attacker_name} · {time(at)}
      </span>
      <div className="attack-headline">
        <span>
          attacks <strong>{target_name}</strong> with {weapon_name}
        </span>
        <strong className={`attack-verdict ${verdictClass}`}>{verdict}</strong>
      </div>
      <div className="roll-line">
        <span className="roll-label">Attack vs AC {target_ac}</span>
        <code>{attack_roll.expression}</code>
        <strong className="roll-total">{attack_roll.total}</strong>
      </div>
      <div className="roll-breakdown">{breakdown({ ...roll, detail: attack_roll.detail })}</div>

      {hit && damage && (
        <div className="mt-1 pt-1 border-t border-border/50">
          <div className="roll-line">
            <span className="roll-label">Damage ({damage_type || 'untyped'})</span>
            <code>{damage.expression}</code>
            <strong className="roll-total text-danger">-{damage.total}</strong>
          </div>
          <div className="roll-breakdown">{breakdown({ ...roll, detail: damage.detail })}</div>
          <div className="text-[11px] text-muted mt-0.5">
            Target HP: {target.hp_current} / {target.hp_max}
            {target.hp_temp > 0 && ` (+${target.hp_temp} temp)`}
          </div>
        </div>
      )}
    </div>
  )
}
