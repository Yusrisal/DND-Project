import { useState } from 'react'
import type { Campaign, Character, GameMap, Token, User } from '../api/types'
import { AccountBar } from './AccountBar'
import { BestiaryPanel } from './BestiaryPanel'
import { CharacterSheetPanel } from './CharacterSheetPanel'
import { DMDashboardPanel } from './DMDashboardPanel'
import { HandoutsPanel } from './HandoutsPanel'
import { InvitePanel } from './InvitePanel'
import { SavepointsPanel } from './SavepointsPanel'
import { ShopPanel } from './ShopPanel'

type Tab = 'board' | 'party' | 'bestiary' | 'shop' | 'dashboard' | 'lore' | 'saves' | 'setup'

const TABS: { id: Tab; label: string }[] = [
  { id: 'board', label: 'board' },
  { id: 'party', label: 'party' },
  { id: 'bestiary', label: 'bestiary' },
  { id: 'shop', label: 'shop' },
  { id: 'dashboard', label: 'tools' },
  { id: 'lore', label: 'lore' },
  { id: 'saves', label: 'saves' },
  { id: 'setup', label: 'setup' },
]

type Props = {
  user: User
  onLogout: () => void

  campaigns: Campaign[]
  campaign: Campaign | null
  onSelectCampaign: (campaign: Campaign) => void
  onCreateCampaign: () => void
  search: string
  onSearch: (value: string) => void

  maps: GameMap[]
  map: GameMap | null
  onSelectMap: (map: GameMap) => void
  onCreateMap: () => void
  onEditMap: (map: GameMap) => void
  onDeleteMap: (map: GameMap) => void

  characters: Character[]
  characterId: number | null
  onSelectCharacter: (id: number) => void
  onCreateCharacter: () => void

  tokens: Token[]
  targetTokenId: number | null
  onSelectToken: (id: number) => void
  onCreateToken: () => void
  onEditToken: (token: Token) => void
  onToggleCondition: (token: Token) => void
  onDeleteToken: (id: number) => void
  ownerName: (token: Token) => string

  activeTokenId: number | null
  onRoll: (expression: string, label: string, characterId: number | null) => void
  onAttack: (request: {
    item_id?: number
    action_id?: number
    attacker_token_id?: number
    target_token_id: number
  }) => void
  onError: (error: unknown) => void
}

/**
 * The DM sidebar gives quick access to board maps/tokens, player sheets, bestiary,
 * merchant shops, and the comprehensive DM tools dashboard.
 */
export function DMSidebar(props: Props) {
  const [tab, setTab] = useState<Tab>('board')

  return (
    <aside className="w-[380px] flex-shrink-0 bg-[#0f1118] border-r-4 border-[#2b3145] flex flex-col h-screen overflow-hidden text-[#c5c6c7] select-none">
      <AccountBar user={props.user} onLogout={props.onLogout} />

      {/* 8-Bit Pixel Navigation Tabs */}
      <div className="grid grid-cols-4 gap-1.5 p-2 bg-[#08090d] border-b-2 border-[#2b3145]">
        {TABS.map(({ id, label }) => (
          <button
            key={id}
            className={`py-2 px-1 text-[10px] font-bold uppercase transition-all tracking-wide ${
              tab === id
                ? 'bg-[#1f5156] text-[#66fcf1] border-2 border-[#66fcf1] shadow-[2px_2px_0px_#051a1c]'
                : 'bg-[#151822] text-[#8e98af] border-2 border-[#2c3345] hover:border-[#45a29e] hover:text-[#e0e2ec]'
            }`}
            onClick={() => setTab(id)}
          >
            {label}
          </button>
        ))}
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-4">
        {tab === 'board' && (
          <>
            {!props.campaign && (
              <div className="p-3 bg-[#2d1214] border-2 border-[#ff4757] text-[#ff6b81] text-[11px] space-y-2">
                <p className="font-bold m-0 leading-tight">NO CAMPAIGN ACTIVE!</p>
                <p className="text-[10px] m-0 text-slate-300">
                  You must create or select a campaign before creating maps.
                </p>
                <button
                  onClick={() => setTab('setup')}
                  className="w-full py-1 bg-[#1f5156] border-2 border-[#66fcf1] text-[#66fcf1] hover:bg-[#45a29e] hover:text-black font-bold uppercase text-[10px]"
                >
                  GO TO SETUP →
                </button>
              </div>
            )}

            <section className="space-y-2">
              <header className="flex items-center justify-between pb-1 border-b border-[#2b3145]">
                <h2 className="text-[10px] font-bold uppercase tracking-wider text-[#e0e2ec]">Maps</h2>
                <button
                  onClick={props.onCreateMap}
                  className="px-2 py-0.5 text-[9px] font-bold bg-[#1f5156] border-2 border-[#66fcf1] text-[#66fcf1] hover:bg-[#45a29e] hover:text-black uppercase shadow-[1px_1px_0px_#000]"
                >
                  + Add Map
                </button>
              </header>
              <ul className="space-y-1">
                {props.maps.map((item) => (
                  <li
                    key={item.id}
                    className={`flex items-stretch gap-2 p-2 border-2 transition-all ${
                      item.id === props.map?.id
                        ? 'bg-[#1f5156] border-[#66fcf1] shadow-[2px_2px_0px_#051a1c]'
                        : 'bg-[#12141d] border-[#2b3145] hover:border-[#45a29e]'
                    }`}
                  >
                    <button
                      className="flex flex-1 items-center gap-2 min-w-0 text-left"
                      onClick={() => props.onSelectMap(item)}
                      title={`Open ${item.name}`}
                    >
                      {/* A thumbnail makes the list scannable at a glance. */}
                      <span className="w-11 h-8 flex-shrink-0 border border-[#2b3145] bg-[#08090d] overflow-hidden flex items-center justify-center">
                        {item.image_url ? (
                          <img
                            src={item.image_url}
                            alt=""
                            className="w-full h-full object-cover"
                          />
                        ) : (
                          <span className="text-[12px]">🗺️</span>
                        )}
                      </span>
                      <span className="min-w-0">
                        <span
                          className={`block truncate text-[11px] font-bold ${
                            item.id === props.map?.id ? 'text-[#66fcf1]' : 'text-[#8e98af]'
                          }`}
                        >
                          {item.name.toUpperCase()}
                        </span>
                        <span className="block text-[9px] text-[#57606f]">
                          {Math.floor(item.width / (item.grid_size || 50))} ×{' '}
                          {Math.floor(item.height / (item.grid_size || 50))} sq
                        </span>
                      </span>
                    </button>
                    <span className="flex flex-col gap-1 flex-shrink-0">
                      <button
                        onClick={() => props.onEditMap(item)}
                        className="px-1.5 py-0.5 text-[9px] font-bold uppercase bg-[#1e2333] border border-[#32394d] text-slate-300 hover:border-[#66fcf1] hover:text-white"
                        title="Edit map image, grid and size"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => props.onDeleteMap(item)}
                        className="px-1.5 py-0.5 text-[9px] font-bold uppercase bg-[#2d1214] border border-[#ff4757] text-[#ff6b81] hover:bg-[#ff4757] hover:text-white"
                        title="Delete this map"
                      >
                        Del
                      </button>
                    </span>
                  </li>
                ))}
              </ul>
              {props.maps.length === 0 && (
                <p className="text-[10px] text-[#57606f] italic text-center py-2">NO MAPS IN THIS CAMPAIGN.</p>
              )}
            </section>

            <section className="space-y-2">
              <header className="flex items-center justify-between pb-1 border-b border-[#2b3145]">
                <h2 className="text-[10px] font-bold uppercase tracking-wider text-[#e0e2ec]">Tokens</h2>
                <button
                  onClick={props.onCreateToken}
                  className="px-2 py-0.5 text-[9px] font-bold bg-[#1f5156] border-2 border-[#66fcf1] text-[#66fcf1] hover:bg-[#45a29e] hover:text-black uppercase shadow-[1px_1px_0px_#000]"
                >
                  + Add Token
                </button>
              </header>
              <ul className="space-y-1.5">
                {(props.tokens || []).map((token) => (
                  <li
                    key={token.id}
                    className={`flex items-center gap-2 p-2 rounded text-xs border transition-colors ${
                      token.id === props.targetTokenId
                        ? 'bg-amber-950/30 border-amber-500/50'
                        : 'bg-[#1a1d26] border-[#262b38]'
                    }`}
                  >
                    <span
                      className="w-3.5 h-3.5 rounded-full flex-shrink-0"
                      style={{ background: token.color }}
                    />
                    <button
                      className="flex-1 text-left truncate text-slate-200 hover:text-white"
                      onClick={() => props.onSelectToken(token.id)}
                    >
                      <div className="font-medium truncate">{token.name}</div>
                      <div className="text-[10px] text-slate-400 truncate">
                        {token.hp_max > 0 &&
                          `${token.hp_current}/${token.hp_max} hp · AC ${token.armor_class}`}
                        {(token.conditions || []).length > 0 && ` · ${(token.conditions || []).join(', ')}`}
                        {props.ownerName(token) && ` · ${props.ownerName(token)}`}
                        {token.character_id !== null && ' · sheet'}
                      </div>
                    </button>
                    <button
                      className="text-[11px] text-slate-400 hover:text-indigo-300 px-1"
                      onClick={() => props.onEditToken(token)}
                      title="Owner, sheet, portrait, size"
                    >
                      edit
                    </button>
                    <button
                      className="text-[11px] text-slate-400 hover:text-indigo-300 px-1"
                      onClick={() => props.onToggleCondition(token)}
                    >
                      cond
                    </button>
                    <button
                      className="text-[11px] text-slate-400 hover:text-red-400 px-1"
                      onClick={() => props.onDeleteToken(token.id)}
                    >
                      ×
                    </button>
                  </li>
                ))}
              </ul>
              {(!props.tokens || props.tokens.length === 0) && (
                <p className="text-xs text-slate-500 italic">Nothing on this map yet.</p>
              )}
            </section>
          </>
        )}

      {tab === 'party' && (
        <>
          <section>
            <header>
              <h2>Characters</h2>
              <button className="btn-add" onClick={props.onCreateCharacter}>
                + Add Character
              </button>
            </header>
            <ul>
              {props.characters.map((item) => (
                <li key={item.id}>
                  <button
                    className={item.id === props.characterId ? 'row active' : 'row'}
                    onClick={() => props.onSelectCharacter(item.id)}
                  >
                    <span>{item.name}</span>
                    {item.status && item.status !== 'active' && (
                      <em className="hint" style={{ fontSize: '0.8em', marginLeft: '6px' }}>
                        ({item.status})
                      </em>
                    )}
                  </button>
                </li>
              ))}
            </ul>
            {props.characters.length === 0 && <p className="hint">No sheets yet.</p>}
          </section>

          {props.characterId !== null && (
            <CharacterSheetPanel
              key={props.characterId}
              characterId={props.characterId}
              onRoll={props.onRoll}
              targetTokenId={props.targetTokenId}
              onAttack={props.onAttack}
              onError={props.onError}
            />
          )}
        </>
      )}

      {tab === 'bestiary' && (
        <BestiaryPanel
          mapId={props.map?.id ?? null}
          targetTokenId={props.targetTokenId}
          attackerTokenId={props.activeTokenId}
          onAttack={props.onAttack}
          onError={props.onError}
        />
      )}

      {tab === 'shop' && (
        <ShopPanel
          campaignId={props.campaign?.id ?? null}
          user={props.user}
          characters={props.characters}
          activeCharacterId={props.characterId}
          onError={props.onError}
        />
      )}

      {tab === 'dashboard' && (
        <DMDashboardPanel
          campaignId={props.campaign?.id ?? null}
          currentMapId={props.map?.id ?? null}
          characters={props.characters}
          onRoll={props.onRoll}
          onError={props.onError}
        />
      )}

      {tab === 'lore' && (
        <HandoutsPanel
          campaignId={props.campaign?.id ?? null}
          isDM={true}
          onError={props.onError}
        />
      )}

      {tab === 'saves' && (
        <SavepointsPanel
          campaignId={props.campaign?.id ?? null}
          currentMapId={props.map?.id ?? null}
          onError={props.onError}
        />
      )}

      {tab === 'setup' && (
        <div className="space-y-4">
          <section className="space-y-2">
            <header className="flex items-center justify-between pb-1 border-b border-[#2b3145]">
              <h2 className="text-[10px] font-bold uppercase tracking-wider text-[#e0e2ec]">Campaigns</h2>
              <button
                onClick={props.onCreateCampaign}
                className="px-2.5 py-1 text-[9px] font-bold bg-[#1f5156] border-2 border-[#66fcf1] text-[#66fcf1] hover:bg-[#45a29e] hover:text-black uppercase shadow-[1px_1px_0px_#000]"
              >
                + New Campaign
              </button>
            </header>
            <input
              value={props.search}
              onChange={(event) => props.onSearch(event.target.value)}
              placeholder="SEARCH CAMPAIGNS..."
              className="w-full bg-[#090a0f] border-2 border-[#2b3145] p-1.5 text-[10px] text-[#e0e2ec] placeholder-[#57606f] focus:outline-none focus:border-[#66fcf1]"
            />
            <div className="space-y-1">
              {props.campaigns.map((item) => (
                <button
                  key={item.id}
                  className={`w-full text-left p-2 border-2 text-[11px] font-bold transition-all ${
                    item.id === props.campaign?.id
                      ? 'bg-[#1f5156] border-[#66fcf1] text-[#66fcf1] shadow-[2px_2px_0px_#051a1c]'
                      : 'bg-[#12141d] border-[#2b3145] text-[#8e98af] hover:border-[#45a29e] hover:text-white'
                  }`}
                  onClick={() => props.onSelectCampaign(item)}
                >
                  ⚔️ {item.name.toUpperCase()}
                </button>
              ))}
              {props.campaigns.length === 0 && (
                <p className="text-[10px] text-[#57606f] italic text-center py-2">NO CAMPAIGNS CREATED YET.</p>
              )}
            </div>
          </section>

          <InvitePanel onError={props.onError} />
        </div>
      )}
      </div>
    </aside>
  )
}
