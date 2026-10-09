import { useState } from 'react'
import type { Character, User } from '../api/types'
import { AccountBar } from './AccountBar'
import { CharacterSheetPanel } from './CharacterSheetPanel'
import { HandoutsPanel } from './HandoutsPanel'
import { ShopPanel } from './ShopPanel'

type Props = {
  campaignId: number | null
  user: User
  onLogout: () => void
  characters: Character[]
  characterId: number | null
  onSelectCharacter: (id: number) => void
  onCreateCharacter: () => void
  targetTokenId: number | null
  onRoll: (expression: string, label: string, characterId: number | null) => void
  onAttack: (request: { item_id: number; target_token_id: number }) => void
  onError: (error: unknown) => void
}

type Tab = 'sheet' | 'shop' | 'lore'

export function PlayerSidebar({
  campaignId,
  user,
  onLogout,
  characters,
  characterId,
  onSelectCharacter,
  onCreateCharacter,
  targetTokenId,
  onRoll,
  onAttack,
  onError,
}: Props) {
  const [tab, setTab] = useState<Tab>('sheet')

  return (
    <aside className="sidebar">
      <AccountBar user={user} onLogout={onLogout} />

      <div className="tabs" style={{ marginBottom: '6px' }}>
        <button
          className={tab === 'sheet' ? 'tab active' : 'tab'}
          onClick={() => setTab('sheet')}
        >
          sheet
        </button>
        <button
          className={tab === 'shop' ? 'tab active' : 'tab'}
          onClick={() => setTab('shop')}
        >
          shop
        </button>
        <button
          className={tab === 'lore' ? 'tab active' : 'tab'}
          onClick={() => setTab('lore')}
        >
          lore
        </button>
      </div>

      {tab === 'sheet' && (
        <>
          {characters.length === 0 && (
            <section>
              <p className="hint">No character yet.</p>
              <button className="primary" onClick={onCreateCharacter}>
                Create your character
              </button>
            </section>
          )}

          {/* Only worth a picker when there is something to pick between. */}
          {characters.length > 1 && (
            <div className="tabs wrap">
              {characters.map((character) => (
                <button
                  key={character.id}
                  className={character.id === characterId ? 'tab active' : 'tab'}
                  onClick={() => onSelectCharacter(character.id)}
                >
                  {character.name}
                </button>
              ))}
            </div>
          )}

          {characterId !== null && (
            <CharacterSheetPanel
              key={characterId}
              characterId={characterId}
              onRoll={onRoll}
              targetTokenId={targetTokenId}
              onAttack={onAttack}
              onError={onError}
            />
          )}
        </>
      )}

      {tab === 'shop' && (
        <ShopPanel
          campaignId={campaignId}
          user={user}
          characters={characters}
          activeCharacterId={characterId}
          onError={onError}
        />
      )}

      {tab === 'lore' && (
        <HandoutsPanel
          campaignId={campaignId}
          isDM={false}
          onError={onError}
        />
      )}
    </aside>
  )
}
