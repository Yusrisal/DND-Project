import { useCallback, useEffect, useState } from 'react'
import { ApiError, api } from './api/client'
import { CONDITIONS } from './api/types'
import type { Campaign, Character, GameMap, Token, User } from './api/types'
import { ActionBar } from './components/ActionBar'
import { AuthForm } from './components/AuthForm'
import { ChatPanel } from './components/ChatPanel'
import { DMSidebar } from './components/DMSidebar'
import { InitiativePanel } from './components/InitiativePanel'
import { InputModal } from './components/InputModal'
import { MapCanvas } from './components/MapCanvas'
import { MapFormModal, type MapFormValues } from './components/MapFormModal'
import { PlayerSidebar } from './components/PlayerSidebar'
import { TokenFormModal, type TokenFormValues } from './components/TokenFormModal'
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

  // Modal dialog state (replaces window.prompt)
  const [modalConfig, setModalConfig] = useState<{
    isOpen: boolean
    title: string
    label?: string
    placeholder?: string
    onSubmit: (val: string) => void
  }>({
    isOpen: false,
    title: '',
    onSubmit: () => {},
  })

  // null closed, 'new' creating, a map when editing.
  const [mapForm, setMapForm] = useState<GameMap | 'new' | null>(null)
  const [tokenForm, setTokenForm] = useState<Token | 'new' | null>(null)

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

  const addCampaign = () => {
    setModalConfig({
      isOpen: true,
      title: 'Create Campaign',
      label: 'Campaign Name',
      placeholder: 'e.g. Curse of the Dragon King',
      onSubmit: async (name) => {
        try {
          const created = await api.createCampaign({ name, description: '' })
          setCampaigns((current) => [...current, created])
          setCampaign(created)
        } catch (err) {
          report(err)
        }
      },
    })
  }

  const addMap = () => {
    if (!campaign) {
      alert('Please select or create a Campaign first in the [SETUP] tab!')
      return
    }
    setMapForm('new')
  }

  const submitMapForm = async (values: MapFormValues) => {
    try {
      if (mapForm === 'new') {
        if (!campaign) return
        const created = await api.createMap({ campaign_id: campaign.id, ...values })
        setMaps((current) => [...current, created])
        setMap(created)
        return
      }
      if (!mapForm) return
      const updated = await api.updateMap(mapForm.id, values)
      setMaps((current) => current.map((m) => (m.id === updated.id ? updated : m)))
      setMap((current) => (current?.id === updated.id ? updated : current))
    } catch (err) {
      report(err)
    }
  }

  const removeMap = async (target: GameMap) => {
    if (!confirm(`Delete map "${target.name}"? Its tokens and fog go with it.`)) return
    try {
      await api.deleteMap(target.id)
      const remaining = maps.filter((m) => m.id !== target.id)
      setMaps(remaining)
      setMap((current) => (current?.id === target.id ? remaining[0] ?? null : current))
    } catch (err) {
      report(err)
    }
  }

  const addToken = () => {
    if (!map) {
      alert('Please create or select a Map first before adding tokens!')
      return
    }
    setTokenForm('new')
  }

  const submitTokenForm = async (values: TokenFormValues) => {
    try {
      if (tokenForm === 'new') {
        if (!map) return
        await api.createToken({ map_id: map.id, ...values })
        return
      }
      if (!tokenForm) return
      await api.updateToken(tokenForm.id, { ...tokenForm, ...values })
    } catch (err) {
      report(err)
    }
    // Either way the server broadcasts the change, so the board updates itself.
  }

  const addCharacter = () => {
    if (!campaign) return
    setModalConfig({
      isOpen: true,
      title: 'Create Character Sheet',
      label: 'Character Name',
      placeholder: 'e.g. Thorin Oakenshield',
      onSubmit: async (name) => {
        try {
          const created = await api.createCharacter({ campaign_id: campaign.id, name })
          setCharacters((current) => [...current, created])
          setCharacterId(created.id)
        } catch (err) {
          report(err)
        }
      },
    })
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
  const activeTokenId = encounter?.combatants?.[encounter.turn_index]?.token_id ?? null

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
          onEditMap={setMapForm}
          onDeleteMap={removeMap}
          characters={characters}
          characterId={characterId}
          onSelectCharacter={setCharacterId}
          onCreateCharacter={addCharacter}
          tokens={tokens}
          targetTokenId={targetTokenId}
          onSelectToken={setTargetTokenId}
          onCreateToken={addToken}
          onEditToken={setTokenForm}
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

      <main className="flex-1 flex flex-col min-w-0 bg-[#07080a] relative overflow-hidden">
        <div className="flex items-center gap-3 px-3 py-1.5 bg-[#0e1017] border-b-2 border-[#2b3145] text-xs">
          <span className={`w-2.5 h-2.5 border border-black ${connected ? 'bg-[#2ed573] shadow-[0_0_8px_#2ed573]' : 'bg-[#ff4757]'}`} />
          <span className="font-bold text-[#e0e2ec] tracking-wider text-[11px]">
            {map ? `MAP: ${map.name.toUpperCase()} [${connected ? 'ONLINE' : 'RECONNECTING'}]` : 'MAP: [NONE SELECTED]'}
          </span>
          {targetTokenId !== null && (
            <span className="flex items-center gap-1.5 px-2 py-0.5 bg-[#1e272e] border-2 border-[#ffa502] text-[#ffa502] text-[10px] font-bold">
              TARGET: {tokens.find((t) => t.id === targetTokenId)?.name.toUpperCase() ?? 'TOKEN'}
              <button
                className="text-[#ffa502] hover:text-white ml-1 bg-transparent border-none cursor-pointer"
                onClick={() => setTargetTokenId(null)}
              >
                [X]
              </button>
            </span>
          )}
          {isDM && campaign && (
            <button
              className={`ml-auto px-2.5 py-1 text-[10px] font-bold uppercase tracking-wider border-2 transition-all ${
                campaign.is_paused
                  ? 'bg-[#e17055] text-white border-[#fab1a0] shadow-[2px_2px_0px_#000]'
                  : 'bg-[#1e272e] text-[#d2dae2] border-[#485460] hover:border-[#66fcf1]'
              }`}
              onClick={handleTogglePause}
              title={campaign.is_paused ? 'Resume live session' : 'Pause live session'}
            >
              {campaign.is_paused ? '▶ RESUME' : '⏸ PAUSE'}
            </button>
          )}
        </div>

        {campaign?.is_paused && (
          <div className="bg-[#2d1406] border-b-2 border-[#e17055] text-[#fab1a0] px-4 py-2 text-xs flex items-center justify-between z-20 font-bold">
            <span>
              [PAUSED] {campaign.pause_message || 'SESSION SUSPENDED BY DUNGEON MASTER'}
            </span>
            {isDM && (
              <button
                className="px-2 py-0.5 bg-[#e17055] hover:bg-[#d63031] text-white border border-white text-[10px] font-bold"
                onClick={handleTogglePause}
              >
                UNPAUSE
              </button>
            )}
          </div>
        )}

        {(error || rejected) && (
          <div
            className="bg-[#380e0e] border-b-2 border-[#ff4757] text-[#ff6b81] px-4 py-2 text-xs flex items-center justify-between cursor-pointer font-bold"
            onClick={() => setError(null)}
          >
            <span>ERROR: {error ?? rejected}</span>
            <span className="text-[10px]">[DISMISS]</span>
          </div>
        )}

        <ActionBar
          character={characters.find((c) => c.id === characterId) ?? null}
          onSendChat={sendMessage}
          onRoll={roll}
        />

        <div className="flex-1 relative overflow-hidden flex items-center justify-center bg-[radial-gradient(#151822_1px,transparent_1px)] bg-[size:16px_16px]">
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
            <div className="text-center p-8 text-[#57606f] space-y-3">
              <div className="text-4xl">⚔️</div>
              <p className="text-xs font-bold uppercase tracking-widest text-[#747d8c]">
                {isDM ? 'NO MAP ON THE TABLE YET' : 'WAITING FOR THE DM TO LOAD A MAP...'}
              </p>
              {isDM && (
                <button
                  onClick={addMap}
                  className="px-4 py-2 bg-[#1f5156] border-2 border-[#66fcf1] text-[#66fcf1] hover:bg-[#45a29e] hover:text-black font-bold text-xs uppercase shadow-[2px_2px_0px_#000]"
                >
                  + New Map
                </button>
              )}
            </div>
          )}
        </div>
      </main>

      <div className="w-[340px] flex-shrink-0 bg-[#0f1118] border-l-4 border-[#2b3145] flex flex-col h-screen overflow-hidden text-[#c5c6c7]">
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

      {/* Modern In-Game Input Modal */}
      <InputModal
        isOpen={modalConfig.isOpen}
        title={modalConfig.title}
        label={modalConfig.label}
        placeholder={modalConfig.placeholder}
        onClose={() => setModalConfig((prev) => ({ ...prev, isOpen: false }))}
        onSubmit={modalConfig.onSubmit}
      />

      <MapFormModal
        key={mapForm === 'new' ? 'new' : (mapForm?.id ?? 'closed')}
        isOpen={mapForm !== null}
        map={mapForm === 'new' ? null : mapForm}
        onClose={() => setMapForm(null)}
        onSubmit={submitMapForm}
      />

      <TokenFormModal
        key={tokenForm === 'new' ? 'new-token' : (tokenForm?.id ?? 'closed-token')}
        isOpen={tokenForm !== null}
        token={tokenForm === 'new' ? null : tokenForm}
        players={players}
        characters={characters}
        defaultColor={TOKEN_COLORS[tokens.length % TOKEN_COLORS.length]}
        onClose={() => setTokenForm(null)}
        onSubmit={submitTokenForm}
      />
    </div>
  )
}
