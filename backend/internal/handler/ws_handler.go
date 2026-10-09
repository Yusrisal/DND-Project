package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/coder/websocket"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/service"
	"dndproject/backend/internal/ws"
)

// writeTimeout bounds a single frame write so one stuck client cannot pin a goroutine.
const writeTimeout = 10 * time.Second

const EventMapPing = "map.ping"

// Rooms are keyed by campaign, not by map. Chat and rolls belong to the table
// rather than to whichever map happens to be open, and a client that is sent a
// token event for another map simply ignores it: token payloads carry map_id.
type WSHandler struct {
	tokens     *service.TokenService
	maps       *service.MapService
	campaigns  *service.CampaignService
	chat       *service.ChatService
	rolls      *service.RollService
	characters *service.CharacterService
	combat     *service.CombatService
	auth       *service.AuthService
	hub        *ws.Hub
	origins    []string
}

func NewWSHandler(
	tokens *service.TokenService,
	maps *service.MapService,
	campaigns *service.CampaignService,
	chat *service.ChatService,
	rolls *service.RollService,
	characters *service.CharacterService,
	combat *service.CombatService,
	auth *service.AuthService,
	hub *ws.Hub,
	origins []string,
) *WSHandler {
	return &WSHandler{
		tokens: tokens, maps: maps, campaigns: campaigns, chat: chat, rolls: rolls,
		characters: characters, combat: combat, auth: auth, hub: hub, origins: origins,
	}
}

// session is what one connection knows about itself.
type session struct {
	client     *ws.Client
	user       model.User
	campaignID int64
	mapID      int64
}

// Serve upgrades GET /ws?map_id=N and keeps the client in that campaign room
// until it disconnects.
func (h *WSHandler) Serve(w http.ResponseWriter, r *http.Request) {
	// Authenticate before upgrading: a rejection is a plain HTTP response the
	// browser can actually surface, where a close frame would be invisible.
	cookie, err := r.Cookie(SessionCookie)
	if err != nil {
		writeError(w, service.ErrUnauthorized)
		return
	}
	user, err := h.auth.UserFromToken(r.Context(), cookie.Value)
	if err != nil {
		writeError(w, err)
		return
	}

	mapID, err := strconv.ParseInt(r.URL.Query().Get("map_id"), 10, 64)
	if err != nil || mapID <= 0 {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "map_id is required", Code: "bad_request"})
		return
	}

	// The campaign comes from the stored map, never from the query string.
	gameMap, err := h.maps.Get(r.Context(), mapID)
	if err != nil {
		writeError(w, err)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: h.origins})
	if err != nil {
		slog.Warn("websocket accept", "error", err)
		return
	}
	defer conn.CloseNow()

	client := h.hub.Join(gameMap.CampaignID)
	defer h.hub.Leave(client)

	s := &session{client: client, user: user, campaignID: gameMap.CampaignID, mapID: mapID}

	ctx := r.Context()
	// Everything this connection sends goes through the client queue, so
	// writePump stays the only goroutine writing to the socket.
	go h.writePump(ctx, conn, client)

	if err := h.sendBacklog(ctx, s); err != nil {
		slog.Warn("websocket backlog", "error", err)
		return
	}

	h.readPump(ctx, conn, s)
}

// sendBacklog gives a joining client the board for its map plus the chat and
// roll history for the campaign, respecting the DM's retention window.
func (h *WSHandler) sendBacklog(ctx context.Context, s *session) error {
	tokens, err := h.tokens.Board(ctx, s.mapID)
	if err != nil {
		return err
	}

	retentionDays := 0
	if h.campaigns != nil {
		if c, err := h.campaigns.Get(ctx, s.campaignID); err == nil {
			retentionDays = c.ChatRetentionDays
		}
	}

	messages, err := h.chat.Recent(ctx, s.campaignID, retentionDays)
	if err != nil {
		return err
	}
	rolls, err := h.rolls.Recent(ctx, s.campaignID, retentionDays)
	if err != nil {
		return err
	}

	frames := []struct {
		event   string
		payload any
	}{
		{EventBoard, tokens},
		{EventChatHistory, messages},
		{EventRollHistory, rolls},
	}

	// A map with no fight underway simply has no encounter frame.
	if encounter, err := h.combat.Encounter(ctx, s.mapID); err == nil {
		frames = append(frames, struct {
			event   string
			payload any
		}{EventEncounter, encounter})
	}

	for _, frame := range frames {
		msg, err := ws.Encode(frame.event, frame.payload)
		if err != nil {
			return err
		}
		h.hub.SendTo(s.client, msg)
	}
	return nil
}

// writePump drains the client queue until the hub closes it or the request ends.
func (h *WSHandler) writePump(ctx context.Context, conn *websocket.Conn, client *ws.Client) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, open := <-client.Out():
			if !open {
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := conn.Write(writeCtx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				slog.Warn("websocket write", "error", err)
				return
			}
		}
	}
}

// readPump receives client frames and dispatches them to the right service.
func (h *WSHandler) readPump(ctx context.Context, conn *websocket.Conn, s *session) {
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) && websocket.CloseStatus(err) != websocket.StatusNormalClosure {
				slog.Debug("websocket read closed", "error", err)
			}
			return
		}

		var envelope ws.Envelope
		if err := json.Unmarshal(data, &envelope); err != nil {
			h.replyError(s.client, "malformed frame")
			continue
		}

		switch envelope.Type {
		case EventTokenMoved:
			h.handleMove(ctx, s, envelope.Payload)
		case EventChatMessage:
			h.handleChat(ctx, s, envelope.Payload)
		case EventRoll:
			h.handleRoll(ctx, s, envelope.Payload)
		case EventAttack:
			h.handleAttack(ctx, s, envelope.Payload)
		case EventMapPing:
			h.handlePing(ctx, s, envelope.Payload)
		default:
			h.replyError(s.client, "unknown message type "+envelope.Type)
		}
	}
}

func (h *WSHandler) handleMove(ctx context.Context, s *session, raw json.RawMessage) {
	var in struct {
		ID int64   `json:"id"`
		X  float64 `json:"x"`
		Y  float64 `json:"y"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		h.replyError(s.client, "invalid move payload")
		return
	}

	token, err := h.tokens.Move(ctx, s.user, model.TokenMove{ID: in.ID, X: in.X, Y: in.Y})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			h.replyError(s.client, "you cannot move that token")
		case errors.Is(err, service.ErrNotFound):
			h.replyError(s.client, "token not found")
		default:
			h.replyError(s.client, validationMessage(err, "invalid move"))
		}
		return
	}

	if token.MapID != s.mapID {
		h.replyError(s.client, "token is not on this map")
		return
	}

	msg, err := ws.Encode(EventTokenMoved, token)
	if err != nil {
		slog.Error("encode token move", "error", err)
		return
	}
	// The mover already rendered the drag locally, so skip them.
	h.hub.Broadcast(s.campaignID, s.client, msg)
}

func (h *WSHandler) handleChat(ctx context.Context, s *session, raw json.RawMessage) {
	var in struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		h.replyError(s.client, "invalid chat payload")
		return
	}

	msg, err := h.chat.Post(ctx, s.user, s.campaignID, in.Body)
	if err != nil {
		h.replyError(s.client, validationMessage(err, "could not post message"))
		return
	}

	frame, err := ws.Encode(EventChatMessage, msg)
	if err != nil {
		slog.Error("encode chat message", "error", err)
		return
	}
	// Broadcast to everyone including the author, who needs the server timestamp.
	h.hub.Broadcast(s.campaignID, nil, frame)
}

func (h *WSHandler) handleRoll(ctx context.Context, s *session, raw json.RawMessage) {
	var in model.RollRequest
	if err := json.Unmarshal(raw, &in); err != nil {
		h.replyError(s.client, "invalid roll payload")
		return
	}

	if in.CharacterID != nil {
		if _, err := h.characters.Get(ctx, s.user, *in.CharacterID); err != nil {
			h.replyError(s.client, "that is not your character to roll for")
			return
		}
	}

	roll, err := h.rolls.Roll(ctx, s.user, s.campaignID, in)
	if err != nil {
		h.replyError(s.client, validationMessage(err, "could not evaluate the roll"))
		return
	}

	frame, err := ws.Encode(EventRoll, roll)
	if err != nil {
		slog.Error("encode roll", "error", err)
		return
	}
	h.hub.Broadcast(s.campaignID, nil, frame)
}

func (h *WSHandler) handleAttack(ctx context.Context, s *session, raw json.RawMessage) {
	var in model.AttackRequest
	if err := json.Unmarshal(raw, &in); err != nil {
		h.replyError(s.client, "invalid attack payload")
		return
	}

	roll, err := h.combat.Attack(ctx, s.user, s.campaignID, in)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			h.replyError(s.client, "that is not yours to attack with")
		case errors.Is(err, service.ErrNotFound):
			h.replyError(s.client, "no such weapon or target")
		default:
			h.replyError(s.client, validationMessage(err, "could not resolve the attack"))
		}
		return
	}

	// 1. Broadcast the attack resolution to the feed.
	frame, err := ws.Encode(EventAttack, roll)
	if err != nil {
		slog.Error("encode attack roll", "error", err)
		return
	}
	h.hub.Broadcast(s.campaignID, nil, frame)

	// 2. Broadcast the target's updated token so its health bar redraws.
	targetToken, err := h.tokens.Get(ctx, in.TargetTokenID)
	if err == nil {
		if tokenFrame, err := ws.Encode(EventTokenUpdated, targetToken); err == nil {
			h.hub.Broadcast(s.campaignID, nil, tokenFrame)
		}
	}
}

func (h *WSHandler) handlePing(ctx context.Context, s *session, raw json.RawMessage) {
	var in struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return
	}
	ping := model.MapPing{
		MapID:      s.mapID,
		X:          in.X,
		Y:          in.Y,
		AuthorName: s.user.DisplayName,
		Color:      "#f1c40f",
	}
	frame, err := ws.Encode(EventMapPing, ping)
	if err != nil {
		return
	}
	h.hub.Broadcast(s.campaignID, nil, frame)
}

func (h *WSHandler) replyError(client *ws.Client, message string) {
	frame, err := ws.Encode(EventError, map[string]string{"message": message})
	if err != nil {
		return
	}
	h.hub.SendTo(client, frame)
}

func validationMessage(err error, fallback string) string {
	var invalid service.ValidationError
	if errors.As(err, &invalid) {
		return invalid.Message
	}
	return fallback
}
