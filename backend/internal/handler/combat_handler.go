package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/service"
	"dndproject/backend/internal/ws"
)

// Combat event names. Initiative and conditions are broadcast so every screen
// shows the same turn and the same statuses without polling.
const (
	EventEncounter = "encounter"
	EventAttack    = "attack"
)

// CombatHandler owns initiative, conditions and attacks. Everything here that
// changes the fight is DM-gated in the router except Attack, which a player
// uses for their own weapon.
type CombatHandler struct {
	svc  *service.CombatService
	maps *service.MapService
	hub  *ws.Hub
}

func NewCombatHandler(svc *service.CombatService, maps *service.MapService, hub *ws.Hub) *CombatHandler {
	return &CombatHandler{svc: svc, maps: maps, hub: hub}
}

// broadcastEncounter pushes the new initiative state to the campaign room.
func (h *CombatHandler) broadcastEncounter(r *http.Request, mapID int64, encounter model.Encounter) {
	gameMap, err := h.maps.Get(r.Context(), mapID)
	if err != nil {
		slog.Error("resolve campaign for encounter broadcast", "map_id", mapID, "error", err)
		return
	}
	msg, err := ws.Encode(EventEncounter, encounter)
	if err != nil {
		slog.Error("encode encounter", "error", err)
		return
	}
	h.hub.Broadcast(gameMap.CampaignID, nil, msg)
}

// mapID reads {id} as a map id, which every encounter route is keyed by.
func (h *CombatHandler) respond(w http.ResponseWriter, r *http.Request, mapID int64, encounter model.Encounter, err error) {
	if err != nil {
		writeError(w, err)
		return
	}
	h.broadcastEncounter(r, mapID, encounter)
	writeJSON(w, http.StatusOK, encounter)
}

// Encounter answers 200 with the running encounter, or 404 when combat is not
// underway, which is the normal quiet state rather than an error.
func (h *CombatHandler) Encounter(w http.ResponseWriter, r *http.Request) {
	mapID, ok := pathID(w, r)
	if !ok {
		return
	}
	encounter, err := h.svc.Encounter(r.Context(), mapID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, encounter)
}

func (h *CombatHandler) Start(w http.ResponseWriter, r *http.Request) {
	mapID, ok := pathID(w, r)
	if !ok {
		return
	}
	encounter, err := h.svc.Start(r.Context(), mapID)
	h.respond(w, r, mapID, encounter, err)
}

func (h *CombatHandler) End(w http.ResponseWriter, r *http.Request) {
	mapID, ok := pathID(w, r)
	if !ok {
		return
	}
	encounter, err := h.svc.End(r.Context(), mapID)
	h.respond(w, r, mapID, encounter, err)
}

func (h *CombatHandler) NextTurn(w http.ResponseWriter, r *http.Request) {
	mapID, ok := pathID(w, r)
	if !ok {
		return
	}
	encounter, err := h.svc.NextTurn(r.Context(), mapID)
	h.respond(w, r, mapID, encounter, err)
}

func (h *CombatHandler) AddCombatant(w http.ResponseWriter, r *http.Request) {
	mapID, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.CombatantInput
	if !decode(w, r, &in) {
		return
	}
	encounter, err := h.svc.AddCombatant(r.Context(), mapID, in)
	h.respond(w, r, mapID, encounter, err)
}

func (h *CombatHandler) SetInitiative(w http.ResponseWriter, r *http.Request) {
	mapID, ok := pathID(w, r)
	if !ok {
		return
	}
	combatantID, err := strconv.ParseInt(r.PathValue("combatant"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid combatant id", Code: "bad_request"})
		return
	}

	var in struct {
		Initiative int `json:"initiative"`
	}
	if !decode(w, r, &in) {
		return
	}
	encounter, err := h.svc.SetInitiative(r.Context(), mapID, combatantID, in.Initiative)
	h.respond(w, r, mapID, encounter, err)
}

func (h *CombatHandler) RemoveCombatant(w http.ResponseWriter, r *http.Request) {
	mapID, ok := pathID(w, r)
	if !ok {
		return
	}
	combatantID, err := strconv.ParseInt(r.PathValue("combatant"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid combatant id", Code: "bad_request"})
		return
	}
	encounter, err := h.svc.RemoveCombatant(r.Context(), mapID, combatantID)
	h.respond(w, r, mapID, encounter, err)
}

// SetCondition toggles a status on a token and broadcasts the updated token, so
// the board redraws for everyone.
func (h *CombatHandler) SetCondition(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.ConditionInput
	if !decode(w, r, &in) {
		return
	}

	token, err := h.svc.SetCondition(r.Context(), id, in)
	if err != nil {
		writeError(w, err)
		return
	}

	gameMap, err := h.maps.Get(r.Context(), token.MapID)
	if err == nil {
		if msg, encodeErr := ws.Encode(EventTokenUpdated, token); encodeErr == nil {
			h.hub.Broadcast(gameMap.CampaignID, nil, msg)
		}
	}
	writeJSON(w, http.StatusOK, token)
}
