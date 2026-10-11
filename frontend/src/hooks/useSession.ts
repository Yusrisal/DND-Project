import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { api } from '../api/client'
import type { AttackDetail, ChatMessage, DiceRoll, Encounter, FeedEntry, MapPing, Token } from '../api/types'
import { playCritSound, playDiceSound, playHitSound, playMissSound, playPingSound } from '../utils/audio'

type Envelope = { type: string; payload: unknown }

const RECONNECT_DELAY = 2000

function socketURL(mapId: number): string {
  const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
  return `${protocol}://${location.host}/ws?map_id=${mapId}`
}

/** Replaces an entry by id, or appends it when it is new. */
function upsert<T extends { id: number }>(list: T[], entry: T): T[] {
  return list.some((item) => item.id === entry.id)
    ? list.map((item) => (item.id === entry.id ? entry : item))
    : [...list, entry]
}

/**
 * useSession owns the one websocket this app opens. The room is the campaign,
 * so chat and rolls keep flowing when the DM switches maps, while token events
 * are filtered down to the map on screen.
 */
export function useSession(mapId: number | null) {
  const [tokens, setTokens] = useState<Token[]>([])
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [rolls, setRolls] = useState<DiceRoll[]>([])
  const [encounter, setEncounter] = useState<Encounter | null>(null)
  const [pings, setPings] = useState<MapPing[]>([])
  const [connected, setConnected] = useState(false)
  const [rejected, setRejected] = useState<string | null>(null)
  const socketRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    if (mapId === null) {
      setTokens([])
      setEncounter(null)
      return
    }
    // A fresh map starts with no fight until the server says otherwise.
    setEncounter(null)

    let closedByUs = false
    let retry: number | undefined

    const connect = () => {
      const socket = new WebSocket(socketURL(mapId))
      socketRef.current = socket

      socket.onopen = () => setConnected(true)

      socket.onmessage = (event) => {
        const { type, payload } = JSON.parse(event.data) as Envelope
        switch (type) {
          case 'board':
            setTokens(Array.isArray(payload) ? (payload as Token[]) : [])
            break
          case 'token.created': {
            const token = payload as Token
            // The room is the campaign, so events for other maps arrive here too.
            if (token.map_id === mapId) setTokens((current) => upsert(current, token))
            break
          }
          case 'token.updated':
          case 'token.moved': {
            const token = payload as Token
            if (token.map_id === mapId) setTokens((current) => upsert(current, token))
            break
          }
          case 'token.deleted': {
            const { id } = payload as { id: number }
            setTokens((current) => current.filter((t) => t.id !== id))
            break
          }
          case 'chat.history':
            setMessages(Array.isArray(payload) ? (payload as ChatMessage[]) : [])
            break
          case 'chat.message':
            setMessages((current) => upsert(current, payload as ChatMessage))
            break
          case 'roll.history':
            setRolls(Array.isArray(payload) ? (payload as DiceRoll[]) : [])
            break
          case 'roll': {
            playDiceSound()
            setRolls((current) => upsert(current, payload as DiceRoll))
            break
          }
          case 'attack': {
            const roll = payload as DiceRoll
            const detail = roll.detail as unknown as AttackDetail
            if (detail && detail.kind === 'attack') {
              if (detail.critical) playCritSound()
              else if (detail.hit) playHitSound()
              else playMissSound()
            } else {
              playDiceSound()
            }
            setRolls((current) => upsert(current, roll))
            break
          }
          case 'encounter': {
            const next = payload as Encounter | null
            setEncounter(next && next.active ? next : null)
            break
          }
          case 'map.ping': {
            const ping = payload as MapPing
            if (ping.map_id === mapId) {
              playPingSound()
              setPings((current) => [...current, ping])
              setTimeout(() => {
                setPings((current) => current.filter((p) => p !== ping))
              }, 4000)
            }
            break
          }
          case 'error': {
            // A rejected move means the optimistic update was wrong, so take the
            // authoritative board back from the server rather than guessing.
            const { message } = payload as { message: string }
            setRejected(message)
            api
              .board(mapId)
              .then((data) => setTokens(Array.isArray(data) ? data : []))
              .catch(() => undefined)
            break
          }
        }
      }

      socket.onclose = () => {
        setConnected(false)
        if (!closedByUs) retry = setTimeout(connect, RECONNECT_DELAY)
      }
    }

    connect()

    return () => {
      closedByUs = true
      clearTimeout(retry)
      socketRef.current?.close()
      socketRef.current = null
    }
  }, [mapId])

  const send = useCallback((type: string, payload: unknown) => {
    socketRef.current?.send(JSON.stringify({ type, payload }))
  }, [])

  const moveToken = useCallback(
    (id: number, x: number, y: number) => {
      setRejected(null)
      setTokens((current) => current.map((t) => (t.id === id ? { ...t, x, y } : t)))
      send('token.moved', { id, x, y })
    },
    [send],
  )

  const sendMessage = useCallback(
    (body: string) => {
      setRejected(null)
      send('chat.message', { body })
    },
    [send],
  )

  const roll = useCallback(
    (expression: string, label = '', characterId: number | null = null) => {
      setRejected(null)
      send('roll', { expression, label, character_id: characterId })
    },
    [send],
  )

  // The server decides hit, damage and the new hit points; this only names the
  // weapon or action and the target.
  const attack = useCallback(
    (request: {
      item_id?: number
      action_id?: number
      attacker_token_id?: number
      target_token_id: number
      advantage?: boolean
      disadvantage?: boolean
    }) => {
      setRejected(null)
      send('attack', request)
    },
    [send],
  )

  const sendPing = useCallback(
    (x: number, y: number) => {
      send('map.ping', { x, y })
    },
    [send],
  )

  // Chat and rolls share one feed, ordered by when the server recorded them.
  const feed = useMemo<FeedEntry[]>(() => {
    const validMessages = Array.isArray(messages) ? messages : []
    const validRolls = Array.isArray(rolls) ? rolls : []
    const entries: FeedEntry[] = [
      ...validMessages.map((message) => ({ kind: 'message' as const, at: message.created_at, message })),
      ...validRolls.map((entry) => ({ kind: 'roll' as const, at: entry.created_at, roll: entry })),
    ]
    return entries.sort((a, b) => a.at.localeCompare(b.at))
  }, [messages, rolls])

  return { tokens, feed, encounter, pings, connected, rejected, moveToken, sendMessage, roll, attack, sendPing }
}
