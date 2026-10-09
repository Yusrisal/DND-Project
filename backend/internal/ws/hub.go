// Package ws keeps the websocket fan-out. It is deliberately domain-free:
// rooms are opaque int64 keys and messages are opaque bytes, so the handler
// layer owns what a message means.
package ws

import (
	"encoding/json"
	"sync"
)

// sendBuffer is how many messages may queue for one slow client before it is
// dropped. A dropped client reconnects and reloads the board.
const sendBuffer = 32

type Client struct {
	room int64
	send chan []byte
}

func (c *Client) Out() <-chan []byte { return c.send }

type Hub struct {
	mu    sync.RWMutex
	rooms map[int64]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[int64]map[*Client]struct{})}
}

func (h *Hub) Join(room int64) *Client {
	c := &Client{room: room, send: make(chan []byte, sendBuffer)}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*Client]struct{})
	}
	h.rooms[room][c] = struct{}{}
	return c
}

func (h *Hub) Leave(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients, ok := h.rooms[c.room]; ok {
		if _, joined := clients[c]; joined {
			delete(clients, c)
			close(c.send)
		}
		if len(clients) == 0 {
			delete(h.rooms, c.room)
		}
	}
}

// SendTo queues a message for one client. It takes the same lock Broadcast
// does, so it can never race with Leave closing the channel.
func (h *Hub) SendTo(c *Client, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if _, joined := h.rooms[c.room][c]; !joined {
		return
	}
	select {
	case c.send <- msg:
	default:
	}
}

// Broadcast sends msg to every client in room except skip, which may be nil.
func (h *Hub) Broadcast(room int64, skip *Client, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.rooms[room] {
		if c == skip {
			continue
		}
		select {
		case c.send <- msg:
		default: // slow client: drop the update rather than stall the room
		}
	}
}

// Envelope is the wire format in both directions.
type Envelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func Encode(msgType string, payload any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Envelope{Type: msgType, Payload: raw})
}
