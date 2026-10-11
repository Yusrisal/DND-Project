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

  characters: Character[]
  characterId: number | null
  onSelectCharacter: (id: number) => void
  onCreateCharacter: () => void

  tokens: Token[]
  targetTokenId: number | null
  onSelectToken: (id: number) => void
  onCreateToken: () => void
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
    <aside className="w-[340px] flex-shrink-0 bg-[#161820] border-r border-[#232834] flex flex-col h-screen overflow-hidden text-slate-200">
      <AccountBar user={props.user} onLogout={props.onLogout} />

      <div className="grid grid-cols-4 gap-1 p-2 bg-[#12141a] border-b border-[#232834]">
        {TABS.map(({ id, label }) => (
          <button
            key={id}
            className={`py-1.5 px-2 text-[11px] font-medium rounded capitalize transition-all border ${
              tab === id
                ? 'bg-indigo-600 text-white border-indigo-500 shadow-sm shadow-indigo-600/30'
                : 'bg-[#181b24] text-slate-400 border-[#262b3a] hover:bg-[#202430] hover:text-slate-200'
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
            <section className="space-y-2">
              <header className="flex items-center justify-between pb-1 border-b border-[#232834]">
                <h2 className="text-xs font-bold uppercase tracking-wider text-slate-400">Maps</h2>
                <button
                  onClick={props.onCreateMap}
                  className="px-2 py-0.5 text-xs bg-indigo-600 hover:bg-indigo-500 text-white rounded font-bold"
                >
                  +
                </button>
              </header>
              <ul className="space-y-1">
                {props.maps.map((item) => (
                  <li key={item.id}>
                    <button
                      className={`w-full text-left px-2.5 py-1.5 rounded text-xs transition-colors border ${
                        item.id === props.map?.id
                          ? 'bg-indigo-950/40 border-indigo-500/60 text-indigo-200 font-semibold'
                          : 'bg-[#1a1d26] border-[#262b38] text-slate-300 hover:bg-[#202430]'
                      }`}
                      onClick={() => props.onSelectMap(item)}
                    >
                      {item.name}
                    </button>
                  </li>
                ))}
              </ul>
              {props.maps.length === 0 && (
                <p className="text-xs text-slate-500 italic">No maps in this campaign.</p>
              )}
            </section>

            <section className="space-y-2">
              <header className="flex items-center justify-between pb-1 border-b border-[#232834]">
                <h2 className="text-xs font-bold uppercase tracking-wider text-slate-400">Tokens</h2>
                <button
                  onClick={props.onCreateToken}
                  className="px-2 py-0.5 text-xs bg-indigo-600 hover:bg-indigo-500 text-white rounded font-bold"
                >
                  +
                </button>
              </header>
              <ul className="space-y-1.5">
                {props.tokens.map((token) => (
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
                        {token.conditions.length > 0 && ` · ${token.conditions.join(', ')}`}
                        {props.ownerName(token) && ` · ${props.ownerName(token)}`}
                      </div>
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
              {props.tokens.length === 0 && (
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
              <button onClick={props.onCreateCharacter}>+</button>
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
        <>
          <section>
            <header>
              <h2>Campaigns</h2>
              <button onClick={props.onCreateCampaign}>+</button>
            </header>
            <input
              value={props.search}
              onChange={(event) => props.onSearch(event.target.value)}
              placeholder="Search campaigns"
            />
            <ul>
              {props.campaigns.map((item) => (
                <li key={item.id}>
                  <button
                    className={item.id === props.campaign?.id ? 'row active' : 'row'}
                    onClick={() => props.onSelectCampaign(item)}
                  >
                    {item.name}
                  </button>
                </li>
              ))}
            </ul>
          </section>

          <InvitePanel onError={props.onError} />
        </>
      )}
      </div>
    </aside>
  )
}
