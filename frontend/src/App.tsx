import { useCallback, useEffect, useState } from 'react'
import { ApiError, api } from './api/client'
import { CONDITIONS } from './api/types'
import type { Campaign, Character, GameMap, Token, User } from './api/types'
import { ActionBar } from './components/ActionBar'
import { AuthForm } from './components/AuthForm'
import { ChatPanel } from './components/ChatPanel'
import { DMSidebar } from './components/DMSidebar'
import { InitiativePanel } from './components/InitiativePanel'
import { MapCanvas } from './components/MapCanvas'
import { PlayerSidebar } from './components/PlayerSidebar'
import { useAuth } from './hooks/useAuth'
import { useSession } from './hooks/useSession'

const TOKEN_COLORS = ['#c0392b', '#2980b9', '#27ae60', '#8e44ad', '#d35400', '#16a085']

export default function App() {
  const { user, loading, login, register, logout, onUnauthorized } = useAuth()

  if (loading) return <div className="splash">Loading…</div>
  if (!user) return <AuthForm onLogin={login} onRegister={register} />

  return <Table user={user} onLogout={logout} onUnauthorized={onUnauthorized} />
}

type TableProps = {
  user: User
  onLogout: () => Promise<void>
  onUnauthorized: (error: unknown) => boolean
}

/**
 * One layout, two arrangements: the left column is what you manage, the centre
 * is the board, and the right is the table itself. A player manages a character
 * sheet; the DM manages the world, behind tabs so only one takes space.
 */
function Table({ user, onLogout, onUnauthorized }: TableProps) {
  const [campaigns, setCampaigns] = useState<Campaign[]>([])
  const [campaign, setCampaign] = useState<Campaign | null>(null)
  const [maps, setMaps] = useState<GameMap[]>([])
  const [map, setMap] = useState<GameMap | null>(null)
  const [players, setPlayers] = useState<User[]>([])
  const [characters, setCharacters] = useState<Character[]>([])
  const [characterId, setCharacterId] = useState<number | null>(null)
  const [search, setSearch] = useState('')
  const [error, setError] = useState<string | null>(null)
  // The token an attack lands on. Clicking the board picks it.
  const [targetTokenId, setTargetTokenId] = useState<number | null>(null)

  const { tokens, feed, encounter, connected, rejected, moveToken, sendMessage, roll, attack } =
    useSession(map?.id ?? null)
  const isDM = user.role === 'dm'

  const report = useCallback(
    (err: unknown) => {
      // An expired session drops straight back to the login form.
      if (onUnauthorized(err)) return
      setError(err instanceof ApiError ? err.message : 'Something went wrong')
    },
    [onUnauthorized],
  )

  useEffect(() => {
    const timer = setTimeout(() => {
      api
        .listCampaigns({ search, page_size: 50, sort_by: 'name', sort_dir: 'asc' })
        .then((page) => {
          setCampaigns(page.items)
          // One campaign runs at a time, so open it rather than making it a click.
          setCampaign((current) => current ?? page.items[0] ?? null)
        })
        .catch(report)
    }, 250)
    return () => clearTimeout(timer)
  }, [search, report])

  useEffect(() => {
    if (!campaign) {
      setMaps([])
      setMap(null)
      return
    }
    api
      .listMaps(campaign.id, { page_size: 50, sort_by: 'name', sort_dir: 'asc' })
      .then((page) => {
        setMaps(page.items)
        setMap(page.items[0] ?? null)
      })
      .catch(report)
  }, [campaign, report])

  // The server already filters this to the sheets the caller may see, so a
  // player gets their own and the DM gets everyone.
  useEffect(() => {
    if (!campaign) {
      setCharacters([])
      setCharacterId(null)
      return
    }
    api
      .listCharacters(campaign.id)
      .then((page) => {
        setCharacters(page.items)
        const mine = page.items.find((c) => c.user_id === user.id)
        setCharacterId(mine?.id ?? page.items[0]?.id ?? null)
      })
      .catch(report)
  }, [campaign, user.id, report])

  // Only the DM assigns tokens, so only the DM needs the roster.
  useEffect(() => {
    if (!isDM) return
    api
      .listUsers()
      .then((page) => setPlayers(page.items))
      .catch(report)
  }, [isDM, report])

  // Switching maps invalidates the target, which belongs to the old board.
  useEffect(() => {
    setTargetTokenId(null)
  }, [map?.id])

  const reloadCharacters = useCallback(() => {
    if (!campaign) return
    api
      .listCharacters(campaign.id)
      .then((page) => setCharacters(page.items))
      .catch(report)
  }, [campaign, report])

  const addCampaign = async () => {
    const name = prompt('Campaign name')
    if (!name) return
    try {
      const created = await api.createCampaign({ name, description: '' })
      setCampaigns((current) => [...current, created])
      setCampaign(created)
    } catch (err) {
      report(err)
    }
  }

  const addMap = async () => {
    if (!campaign) return
    const name = prompt('Map name')
    if (!name) return
    try {
      const created = await api.createMap({ campaign_id: campaign.id, name })
      setMaps((current) => [...current, created])
      setMap(created)
    } catch (err) {
      report(err)
    }
  }

  const addToken = async () => {
    if (!map) return
    const name = prompt('Token name')
    if (!name) return

    const roster = players.map((player, index) => `${index + 1}. ${player.display_name}`).join('\n')
    const choice = prompt(`Assign to which player? Empty for none.\n${roster}`)
    const owner = choice ? (players[Number(choice) - 1] ?? null) : null

    try {
      // The socket echo puts it on the board, so there is nothing to set here.
      await api.createToken({
        map_id: map.id,
        name,
        color: TOKEN_COLORS[tokens.length % TOKEN_COLORS.length],
        owner_user_id: owner?.id ?? null,
      })
    } catch (err) {
      report(err)
    }
  }

  const addCharacter = async () => {
    if (!campaign) return
    const name = prompt('Character name')
    if (!name) return
    try {
      const created = await api.createCharacter({ campaign_id: campaign.id, name })
      setCharacters((current) => [...current, created])
      setCharacterId(created.id)
    } catch (err) {
      report(err)
    }
  }

  const toggleCondition = async (token: Token) => {
    const current = token.conditions.join(', ') || 'none'
    const name = prompt(
      `Condition to toggle on ${token.name}.\nNow: ${current}\n\n${CONDITIONS.join(', ')}`,
    )
    if (!name) return
    const trimmed = name.trim().toLowerCase()
    try {
      await api.setCondition(token.id, trimmed, !token.conditions.includes(trimmed))
    } catch (err) {
      report(err)
    }
  }

  const canMove = useCallback(
    (token: Token) => {
      if (campaign?.is_paused && !isDM) return false
      return isDM || token.owner_user_id === user.id
    },
    [isDM, user.id, campaign?.is_paused],
  )

  const handleTogglePause = async () => {
    if (!campaign) return
    try {
      const isPaused = !campaign.is_paused
      let msg = campaign.pause_message || 'Session Paused by DM (Break time / Saved)'
      if (isPaused) {
        const custom = prompt('Pause reason or message to players:', msg)
        if (custom !== null && custom.trim() !== '') msg = custom.trim()
      }
      const updated = await api.setCampaignPause(campaign.id, { is_paused: isPaused, pause_message: msg })
      setCampaign(updated)
      setCampaigns((curr) => curr.map((c) => (c.id === updated.id ? updated : c)))
    } catch (err) {
      report(err)
    }
  }

  const ownerName = useCallback(
    (token: Token) => {
      if (token.owner_user_id === null) return isDM ? 'unassigned' : ''
      if (token.owner_user_id === user.id) return 'you'
      return players.find((player) => player.id === token.owner_user_id)?.display_name ?? ''
    },
    [isDM, players, user.id],
  )

  // The creature whose turn it is, used to name the attacker on a monster action.
  const activeTokenId = encounter?.combatants[encounter.turn_index]?.token_id ?? null

  const sheetError = (err: unknown) => {
    report(err)
    reloadCharacters()
  }

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-[#0c0d12] text-[#e7e8ea]">
      {isDM ? (
        <DMSidebar
          user={user}
          onLogout={() => void onLogout().catch(report)}
          campaigns={campaigns}
          campaign={campaign}
          onSelectCampaign={setCampaign}
          onCreateCampaign={addCampaign}
          search={search}
          onSearch={setSearch}
          maps={maps}
          map={map}
          onSelectMap={setMap}
          onCreateMap={addMap}
          characters={characters}
          characterId={characterId}
          onSelectCharacter={setCharacterId}
          onCreateCharacter={addCharacter}
          tokens={tokens}
          targetTokenId={targetTokenId}
          onSelectToken={setTargetTokenId}
          onCreateToken={addToken}
          onToggleCondition={toggleCondition}
          onDeleteToken={(id) => void api.deleteToken(id).catch(report)}
          ownerName={ownerName}
          activeTokenId={activeTokenId}
          onRoll={roll}
          onAttack={attack}
          onError={sheetError}
        />
      ) : (
        <PlayerSidebar
          campaignId={campaign?.id ?? null}
          user={user}
          onLogout={() => void onLogout().catch(report)}
          characters={characters}
          characterId={characterId}
          onSelectCharacter={setCharacterId}
          onCreateCharacter={addCharacter}
          targetTokenId={targetTokenId}
          onRoll={roll}
          onAttack={attack}
          onError={sheetError}
        />
      )}

      <main className="flex-1 flex flex-col min-w-0 bg-[#0a0b0f] relative overflow-hidden">
        <div className="flex items-center gap-3 px-4 py-2 bg-[#14161f] border-b border-[#232834] text-xs text-slate-300">
          <span className={`w-2 h-2 rounded-full ${connected ? 'bg-emerald-500 shadow-sm shadow-emerald-500/50' : 'bg-red-500'}`} />
          <span className="font-medium text-slate-200">
            {map ? `${map.name} — ${connected ? 'Live' : 'Reconnecting'}` : 'No active map'}
          </span>
          {targetTokenId !== null && (
            <span className="flex items-center gap-1.5 px-2 py-0.5 rounded bg-indigo-950/50 border border-indigo-700/50 text-indigo-300 text-[11px]">
              targeting {tokens.find((t) => t.id === targetTokenId)?.name ?? 'Token'}
              <button
                className="text-slate-400 hover:text-white ml-1 bg-transparent border-none cursor-pointer"
                onClick={() => setTargetTokenId(null)}
              >
                ✕
              </button>
            </span>
          )}
          {isDM && campaign && (
            <button
              className={`ml-auto px-2.5 py-1 text-xs rounded border transition-colors ${
                campaign.is_paused
                  ? 'bg-amber-600 hover:bg-amber-500 text-white border-amber-500 font-semibold'
                  : 'bg-[#1e222e] hover:bg-[#282d3d] text-slate-300 border-[#2d3446]'
              }`}
              onClick={handleTogglePause}
              title={campaign.is_paused ? 'Resume live session' : 'Pause live session'}
            >
              {campaign.is_paused ? '▶ Resume Session' : '⏸ Pause Session'}
            </button>
          )}
        </div>

        {campaign?.is_paused && (
          <div className="bg-amber-950/80 border-b border-amber-600 text-amber-200 px-4 py-2 text-xs flex items-center justify-between z-20">
            <span>
              ☕ <strong>Session Paused</strong> — {campaign.pause_message || 'Game saved at checkpoint. Break time!'}
            </span>
            {isDM && (
              <button
                className="px-2.5 py-1 bg-amber-600 hover:bg-amber-500 text-white rounded text-xs font-semibold"
                onClick={handleTogglePause}
              >
                Resume Live Session
              </button>
            )}
          </div>
        )}

        {(error || rejected) && (
          <div
            className="bg-red-500/10 border-b border-red-500/30 text-red-300 px-4 py-2 text-xs flex items-center justify-between cursor-pointer"
            onClick={() => setError(null)}
          >
            <span>{error ?? rejected}</span>
            <span className="text-[10px] text-red-400">Click to dismiss</span>
          </div>
        )}

        <ActionBar
          character={characters.find((c) => c.id === characterId) ?? null}
          onSendChat={sendMessage}
          onRoll={roll}
        />

        <div className="flex-1 relative overflow-hidden flex items-center justify-center">
          {map ? (
            <MapCanvas
              map={map}
              tokens={tokens}
              canMove={canMove}
              selectedTokenId={targetTokenId}
              onSelect={setTargetTokenId}
              onMove={moveToken}
              isDM={isDM}
              onUpdateMap={(updated) => {
                setMap(updated)
                setMaps((curr) => curr.map((m) => (m.id === updated.id ? updated : m)))
              }}
            />
          ) : (
            <div className="text-center p-8 text-slate-500 space-y-2">
              <div className="text-3xl">🗺️</div>
              <p className="text-sm">
                {isDM ? 'Select or create a map from the board tab to start playing.' : 'Waiting for the DM to open a map.'}
              </p>
            </div>
          )}
        </div>
      </main>

      <div className="w-[320px] flex-shrink-0 bg-[#161820] border-l border-[#232834] flex flex-col h-screen overflow-hidden">
        {map && (
          <InitiativePanel
            mapId={map.id}
            encounter={encounter}
            tokens={tokens}
            isDM={isDM}
            selectedTokenId={targetTokenId}
            onSelectToken={setTargetTokenId}
            onError={report}
          />
        )}
        <ChatPanel
          feed={feed}
          connected={connected}
          campaign={campaign}
          isDM={isDM}
          characterId={characterId}
          onSend={sendMessage}
          onRoll={roll}
          onUpdateCampaign={(updated) => {
            setCampaign(updated)
            setCampaigns((curr) => curr.map((c) => (c.id === updated.id ? updated : c)))
          }}
        />
      </div>
    </div>
  )
}
