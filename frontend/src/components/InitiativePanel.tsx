import { api } from '../api/client'
import type { Encounter, Token } from '../api/types'

type Props = {
  mapId: number
  encounter: Encounter | null
  tokens: Token[]
  isDM: boolean
  selectedTokenId: number | null
  onSelectToken: (tokenId: number | null) => void
  onError: (error: unknown) => void
}

/**
 * The initiative order. Every change goes through the API and comes back on the
 * websocket, so each screen shows the same turn without polling.
 */
export function InitiativePanel({
  mapId,
  encounter,
  tokens,
  isDM,
  selectedTokenId,
  onSelectToken,
  onError,
}: Props) {
  if (!encounter) {
    if (!isDM) return null
    return (
      <section className="initiative">
        <header>
          <h2>Initiative</h2>
          <button onClick={() => api.startEncounter(mapId).catch(onError)}>start combat</button>
        </header>
      </section>
    )
  }

  // Tokens already in the order are not offered again.
  const listed = new Set(encounter.combatants.map((c) => c.token_id).filter(Boolean))
  const available = tokens.filter((token) => !listed.has(token.id))

  const addEveryone = async () => {
    try {
      // Sequential on purpose: each call answers with the whole order, and the
      // initiative rolls read better when they do not race.
      for (const token of available) {
        await api.addCombatant(mapId, { token_id: token.id })
      }
    } catch (err) {
      onError(err)
    }
  }

  return (
    <section className="initiative">
      <header>
        <h2>
          Initiative <em>round {encounter.round}</em>
        </h2>
        {isDM && (
          <span className="actions">
            <button onClick={() => api.nextTurn(mapId).catch(onError)}>next</button>
            <button onClick={() => api.endEncounter(mapId).catch(onError)}>end</button>
          </span>
        )}
      </header>

      <ol className="order">
        {encounter.combatants.map((combatant, index) => {
          const token = tokens.find((t) => t.id === combatant.token_id)
          const classes = ['order-row']
          if (index === encounter.turn_index) classes.push('current')
          if (token && token.id === selectedTokenId) classes.push('targeted')

          return (
            <li key={combatant.id} className={classes.join(' ')}>
              <span className="order-init">{combatant.initiative}</span>
              <button
                className="order-name"
                title={token ? 'Target this creature' : undefined}
                onClick={() => onSelectToken(token ? token.id : null)}
              >
                {combatant.name}
                {token && token.hp_max > 0 && (
                  <em>
                    {token.hp_current}/{token.hp_max}
                  </em>
                )}
              </button>
              {isDM && (
                <button
                  className="link"
                  onClick={() => api.removeCombatant(mapId, combatant.id).catch(onError)}
                >
                  drop
                </button>
              )}
            </li>
          )
        })}
      </ol>

      {encounter.combatants.length === 0 && <p className="hint">Nobody has rolled yet.</p>}

      {isDM && available.length > 0 && (
        <div className="quick-rolls">
          <button onClick={addEveryone}>add all ({available.length})</button>
          {available.slice(0, 6).map((token) => (
            <button
              key={token.id}
              onClick={() => api.addCombatant(mapId, { token_id: token.id }).catch(onError)}
            >
              + {token.name}
            </button>
          ))}
        </div>
      )}
    </section>
  )
}
