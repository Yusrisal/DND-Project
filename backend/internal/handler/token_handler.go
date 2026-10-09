package handler

import (
	"context"
	"log/slog"
	"net/http"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
	"dndproject/backend/internal/service"
	"dndproject/backend/internal/ws"
)

// Event names shared by the REST handlers and the websocket handler, so a
// change made over HTTP looks the same to clients as one made over the socket.
const (
	EventBoard        = "board"
	EventTokenCreated = "token.created"
	EventTokenUpdated = "token.updated"
	EventTokenMoved   = "token.moved"
	EventTokenDeleted = "token.deleted"
	EventChatHistory  = "chat.history"
	EventChatMessage  = "chat.message"
	EventRollHistory  = "roll.history"
	EventRoll         = "roll"
	EventError        = "error"
)

type TokenHandler struct {
	svc  *service.TokenService
	maps *service.MapService
	hub  *ws.Hub
}

func NewTokenHandler(svc *service.TokenService, maps *service.MapService, hub *ws.Hub) *TokenHandler {
	return &TokenHandler{svc: svc, maps: maps, hub: hub}
}

// broadcast fans an event out to everyone at the table, skipping nobody: REST
// callers get the authoritative row in their own response anyway. Rooms are
// keyed by campaign, so the map has to be resolved to its campaign first.
func (h *TokenHandler) broadcast(ctx context.Context, mapID int64, event string, payload any) {
	gameMap, err := h.maps.Get(ctx, mapID)
	if err != nil {
		slog.Error("resolve campaign for broadcast", "map_id", mapID, "error", err)
		return
	}
	msg, err := ws.Encode(event, payload)
	if err != nil {
		slog.Error("encode broadcast", "event", event, "error", err)
		return
	}
	h.hub.Broadcast(gameMap.CampaignID, nil, msg)
}

func (h *TokenHandler) List(w http.ResponseWriter, r *http.Request) {
	q := model.ParseQuery(r, repository.TokenSortColumns(), "id")
	page, err := h.svc.List(r.Context(), queryID(r, "map_id"), q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

// Board serves GET /api/maps/{id}/tokens: the whole board in one call.
func (h *TokenHandler) Board(w http.ResponseWriter, r *http.Request) {
	mapID, ok := pathID(w, r)
	if !ok {
		return
	}
	tokens, err := h.svc.Board(r.Context(), mapID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

func (h *TokenHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	token, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, token)
}

func (h *TokenHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in model.TokenInput
	if !decode(w, r, &in) {
		return
	}
	token, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeError(w, err)
		return
	}
	h.broadcast(r.Context(), token.MapID, EventTokenCreated, token)
	writeJSON(w, http.StatusCreated, token)
}

func (h *TokenHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.TokenInput
	if !decode(w, r, &in) {
		return
	}
	token, err := h.svc.Update(r.Context(), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	h.broadcast(r.Context(), token.MapID, EventTokenUpdated, token)
	writeJSON(w, http.StatusOK, token)
}

func (h *TokenHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	// Read the token first: once it is gone there is no map id left to broadcast to.
	token, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	h.broadcast(r.Context(), token.MapID, EventTokenDeleted, model.TokenMove{ID: id})
	w.WriteHeader(http.StatusNoContent)
}
