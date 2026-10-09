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
    <aside className="sidebar">
      <AccountBar user={props.user} onLogout={props.onLogout} />

      <div className="tabs">
        {TABS.map(({ id, label }) => (
          <button
            key={id}
            className={tab === id ? 'tab active' : 'tab'}
            onClick={() => setTab(id)}
          >
            {label}
          </button>
        ))}
      </div>

      {tab === 'board' && (
        <>
          <section>
            <header>
              <h2>Maps</h2>
              <button onClick={props.onCreateMap}>+</button>
            </header>
            <ul>
              {props.maps.map((item) => (
                <li key={item.id}>
                  <button
                    className={item.id === props.map?.id ? 'row active' : 'row'}
                    onClick={() => props.onSelectMap(item)}
                  >
                    {item.name}
                  </button>
                </li>
              ))}
            </ul>
            {props.maps.length === 0 && <p className="hint">No maps in this campaign.</p>}
          </section>

          <section>
            <header>
              <h2>Tokens</h2>
              <button onClick={props.onCreateToken}>+</button>
            </header>
            <ul>
              {props.tokens.map((token) => (
                <li
                  key={token.id}
                  className={token.id === props.targetTokenId ? 'token-row targeted' : 'token-row'}
                >
                  <span className="swatch" style={{ background: token.color }} />
                  <button className="token-name" onClick={() => props.onSelectToken(token.id)}>
                    {token.name}
                    <em className="owner">
                      {token.hp_max > 0 &&
                        `${token.hp_current}/${token.hp_max} hp · AC ${token.armor_class}`}
                      {token.conditions.length > 0 && ` · ${token.conditions.join(', ')}`}
                      {props.ownerName(token) && ` · ${props.ownerName(token)}`}
                    </em>
                  </button>
                  <button className="link" onClick={() => props.onToggleCondition(token)}>
                    cond
                  </button>
                  <button className="link" onClick={() => props.onDeleteToken(token.id)}>
                    remove
                  </button>
                </li>
              ))}
            </ul>
            {props.tokens.length === 0 && <p className="hint">Nothing on this map yet.</p>}
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
    </aside>
  )
}
