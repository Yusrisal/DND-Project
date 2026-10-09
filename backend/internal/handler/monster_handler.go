package handler

import (
	"log/slog"
	"net/http"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
	"dndproject/backend/internal/service"
	"dndproject/backend/internal/ws"
)

// MonsterHandler serves the bestiary. Reading is open to any signed-in account
// so a player can look up a creature they have met; everything that writes, and
// spawning, is DM only.
type MonsterHandler struct {
	svc  *service.MonsterService
	maps *service.MapService
	hub  *ws.Hub
}

func NewMonsterHandler(svc *service.MonsterService, maps *service.MapService, hub *ws.Hub) *MonsterHandler {
	return &MonsterHandler{svc: svc, maps: maps, hub: hub}
}

func (h *MonsterHandler) List(w http.ResponseWriter, r *http.Request) {
	q := model.ParseQuery(r, repository.MonsterSortColumns(), "name")
	page, err := h.svc.List(r.Context(), q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

// StatBlock is the whole entry: vitals, skills and saves, and actions.
func (h *MonsterHandler) StatBlock(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	block, err := h.svc.StatBlock(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, block)
}

func (h *MonsterHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in model.MonsterInput
	if !decode(w, r, &in) {
		return
	}
	monster, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, monster)
}

func (h *MonsterHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.MonsterInput
	if !decode(w, r, &in) {
		return
	}
	monster, err := h.svc.Update(r.Context(), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, monster)
}

func (h *MonsterHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MonsterHandler) SetProficiency(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in struct {
		model.MonsterProficiency
		Keep bool `json:"keep"`
	}
	in.Keep = true
	if !decode(w, r, &in) {
		return
	}

	list, err := h.svc.SetProficiency(r.Context(), id, in.MonsterProficiency, in.Keep)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *MonsterHandler) AddAction(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.MonsterActionInput
	if !decode(w, r, &in) {
		return
	}
	action, err := h.svc.AddAction(r.Context(), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, action)
}

func (h *MonsterHandler) UpdateAction(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.MonsterActionInput
	if !decode(w, r, &in) {
		return
	}
	action, err := h.svc.UpdateAction(r.Context(), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, action)
}

func (h *MonsterHandler) DeleteAction(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteAction(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Spawn places copies of a stat block on a map and broadcasts each one, so the
// party watches the goblins appear.
func (h *MonsterHandler) Spawn(w http.ResponseWriter, r *http.Request) {
	var in model.SpawnInput
	if !decode(w, r, &in) {
		return
	}

	tokens, err := h.svc.Spawn(r.Context(), in)
	if err != nil {
		writeError(w, err)
		return
	}

	gameMap, err := h.maps.Get(r.Context(), in.MapID)
	if err != nil {
		slog.Error("resolve campaign for spawn broadcast", "map_id", in.MapID, "error", err)
	} else {
		for _, token := range tokens {
			msg, encodeErr := ws.Encode(EventTokenCreated, token)
			if encodeErr != nil {
				slog.Error("encode spawned token", "error", encodeErr)
				continue
			}
			h.hub.Broadcast(gameMap.CampaignID, nil, msg)
		}
	}
	writeJSON(w, http.StatusCreated, tokens)
}
