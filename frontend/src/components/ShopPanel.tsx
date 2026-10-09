import { useCallback, useEffect, useState } from 'react'
import { api } from '../api/client'
import type { Character, Shop, ShopDetails, User } from '../api/types'

type Props = {
  campaignId: number | null
  user: User
  characters: Character[]
  activeCharacterId: number | null
  onError: (error: unknown) => void
  onSuccess?: (msg: string) => void
}

export function ShopPanel({ campaignId, user, characters, activeCharacterId, onError }: Props) {
  const [shops, setShops] = useState<Shop[]>([])
  const [activeShopId, setActiveShopId] = useState<number | null>(null)
  const [shopDetails, setShopDetails] = useState<ShopDetails | null>(null)
  const [buyerCharId, setBuyerCharId] = useState<number | null>(activeCharacterId)

  const isDM = user.role === 'dm'

  const refreshShops = useCallback(() => {
    if (!campaignId) {
      setShops([])
      setActiveShopId(null)
      return
    }
    api
      .listShops(campaignId)
      .then((page) => {
        setShops(page.items)
        setActiveShopId((curr) => curr ?? page.items[0]?.id ?? null)
      })
      .catch(onError)
  }, [campaignId, onError])

  useEffect(refreshShops, [refreshShops])

  useEffect(() => {
    if (!activeShopId) {
      setShopDetails(null)
      return
    }
    api.getShop(activeShopId).then(setShopDetails).catch(onError)
  }, [activeShopId, onError])

  useEffect(() => {
    if (activeCharacterId && (!buyerCharId || !characters.some((c) => c.id === buyerCharId))) {
      setBuyerCharId(activeCharacterId)
    }
  }, [activeCharacterId, characters, buyerCharId])

  const createShop = async () => {
    if (!campaignId) return
    const name = prompt('Shop / Merchant name')
    if (!name) return
    const description = prompt('Description or merchant flavor', '') ?? ''
    try {
      const created = await api.createShop({ campaign_id: campaignId, name, description })
      setShops((curr) => [...curr, created])
      setActiveShopId(created.id)
    } catch (err) {
      onError(err)
    }
  }

  const deleteShop = async (id: number) => {
    if (!confirm('Delete this shop?')) return
    try {
      await api.deleteShop(id)
      setActiveShopId(null)
      refreshShops()
    } catch (err) {
      onError(err)
    }
  }

  const addShopItem = async () => {
    if (!activeShopId) return
    const name = prompt('Item name')
    if (!name) return
    const costGP = Number(prompt('Cost in Gold (GP)', '10') ?? '0') || 0
    const quantity = Number(prompt('Stock quantity (-1 for unlimited)', '-1') ?? '-1')
    const weight = Number(prompt('Weight (lbs)', '0') ?? '0') || 0
    const damage = (prompt('Damage expression if weapon (e.g. 1d8+2). Leave empty for gear.', '') ?? '').trim()
    const attackBonus = damage ? Number(prompt('Attack bonus', '0') ?? '0') || 0 : 0
    const damageType = damage ? (prompt('Damage type (slashing, piercing, etc.)', 'slashing') ?? '').trim() : ''
    const description = prompt('Description', '') ?? ''

    try {
      await api.addShopItem(activeShopId, {
        name,
        cost_cp: Math.round(costGP * 100),
        quantity,
        weight,
        attack_bonus: attackBonus,
        damage,
        damage_type: damageType,
        description,
      })
      api.getShop(activeShopId).then(setShopDetails).catch(onError)
    } catch (err) {
      onError(err)
    }
  }

  const deleteShopItem = async (id: number) => {
    try {
      await api.deleteShopItem(id)
      if (activeShopId) {
        api.getShop(activeShopId).then(setShopDetails).catch(onError)
      }
    } catch (err) {
      onError(err)
    }
  }

  const buyItem = async (shopItemId: number, itemName: string, costCP: number) => {
    if (!buyerCharId) {
      alert('Select which character is buying!')
      return
    }
    const costText = costCP >= 100 ? `${costCP / 100} GP` : `${costCP} CP`
    if (!confirm(`Buy ${itemName} for ${costText}?`)) return

    try {
      await api.buyShopItem(shopItemId, { character_id: buyerCharId, quantity: 1 })
      alert(`Purchased ${itemName}! Added to character inventory.`)
      if (activeShopId) {
        api.getShop(activeShopId).then(setShopDetails).catch(onError)
      }
    } catch (err) {
      onError(err)
    }
  }

  const splitLoot = async () => {
    if (!campaignId) return
    const gp = parseInt(prompt('Party Gold (GP) to split equally:', '0') ?? '0', 10) || 0
    const sp = parseInt(prompt('Party Silver (SP) to split equally:', '0') ?? '0', 10) || 0
    const cp = parseInt(prompt('Party Copper (CP) to split equally:', '0') ?? '0', 10) || 0
    const ep = parseInt(prompt('Party Electrum (EP) to split equally:', '0') ?? '0', 10) || 0
    const pp = parseInt(prompt('Party Platinum (PP) to split equally:', '0') ?? '0', 10) || 0

    if (gp + sp + cp + ep + pp === 0) return
    try {
      await api.splitLoot({ campaign_id: campaignId, gp, sp, cp, ep, pp })
      alert('Loot split equally among active party characters!')
    } catch (err) {
      onError(err)
    }
  }

  return (
    <div className="shop-panel">
      <section>
        <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h2>Shops & Merchants</h2>
          <div style={{ display: 'flex', gap: '6px' }}>
            {isDM && (
              <>
                <button onClick={splitLoot} title="Split coin loot among all active characters">
                  Split Loot
                </button>
                <button onClick={createShop} title="Create a new shop">+</button>
              </>
            )}
          </div>
        </header>

        {shops.length > 0 && (
          <div className="tabs wrap" style={{ marginBottom: '8px' }}>
            {shops.map((s) => (
              <button
                key={s.id}
                className={s.id === activeShopId ? 'tab active' : 'tab'}
                onClick={() => setActiveShopId(s.id)}
              >
                {s.name}
              </button>
            ))}
          </div>
        )}

        {shops.length === 0 && <p className="hint">No merchants available.</p>}
      </section>

      {shopDetails && (
        <section className="shop-wares">
          <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div>
              <h3>{shopDetails.shop.name}</h3>
              {shopDetails.shop.description && <span className="hint">{shopDetails.shop.description}</span>}
            </div>
            {isDM && (
              <div style={{ display: 'flex', gap: '6px' }}>
                <button onClick={addShopItem}>+ item</button>
                <button className="link" onClick={() => deleteShop(shopDetails.shop.id)}>
                  delete shop
                </button>
              </div>
            )}
          </header>

          <div style={{ margin: '8px 0', fontSize: '0.9em', display: 'flex', alignItems: 'center', gap: '8px' }}>
            <span>Buyer:</span>
            <select
              value={buyerCharId ?? ''}
              onChange={(e) => setBuyerCharId(Number(e.target.value) || null)}
            >
              {characters.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name} ({c.gp} GP, {c.sp} SP)
                </option>
              ))}
            </select>
          </div>

          <div className="list">
            {shopDetails.items.length === 0 && <p className="hint">Shop has no items for sale.</p>}
            <ul>
              {shopDetails.items.map((item) => {
                const costDisplay =
                  item.cost_cp >= 100 ? `${(item.cost_cp / 100).toFixed(2).replace(/\.00$/, '')} GP` : `${item.cost_cp} CP`
                return (
                  <li key={item.id} className="list-row" style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <div style={{ display: 'flex', flexDirection: 'column' }}>
                      <span className="list-name">
                        <strong>{item.name}</strong> · <em>{costDisplay}</em>
                        {item.quantity >= 0 && <span> ({item.quantity} in stock)</span>}
                        {item.damage && <span> · {item.damage} {item.damage_type}</span>}
                      </span>
                      {item.description && <span className="hint" style={{ fontSize: '0.8em' }}>{item.description}</span>}
                    </div>

                    <div style={{ display: 'flex', gap: '6px', alignItems: 'center' }}>
                      <button
                        className="pill on"
                        disabled={item.quantity === 0 || !buyerCharId}
                        onClick={() => buyItem(item.id, item.name, item.cost_cp)}
                      >
                        Buy
                      </button>
                      {isDM && (
                        <button className="link" onClick={() => deleteShopItem(item.id)}>
                          remove
                        </button>
                      )}
                    </div>
                  </li>
                )
              })}
            </ul>
          </div>
        </section>
      )}
    </div>
  )
}
