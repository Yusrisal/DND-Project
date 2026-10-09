package handler

import (
	"net/http"
	"strconv"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
	"dndproject/backend/internal/service"
)

// CharacterHandler serves sheets. Ownership is decided in the service, so every
// route here is open to any signed-in account and the DM simply sees more.
type CharacterHandler struct {
	svc *service.CharacterService
}

func NewCharacterHandler(svc *service.CharacterService) *CharacterHandler {
	return &CharacterHandler{svc: svc}
}

func (h *CharacterHandler) List(w http.ResponseWriter, r *http.Request) {
	q := model.ParseQuery(r, repository.CharacterSortColumns(), "name")
	page, err := h.svc.List(r.Context(), userFrom(r.Context()), queryID(r, "campaign_id"), q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

// Sheet is the one call the character panel makes: stats, inventory, spells and
// slots together.
func (h *CharacterHandler) Sheet(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	sheet, err := h.svc.Sheet(r.Context(), userFrom(r.Context()), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sheet)
}

func (h *CharacterHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in model.CharacterInput
	if !decode(w, r, &in) {
		return
	}
	character, err := h.svc.Create(r.Context(), userFrom(r.Context()), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, character)
}

func (h *CharacterHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.CharacterInput
	if !decode(w, r, &in) {
		return
	}
	character, err := h.svc.Update(r.Context(), userFrom(r.Context()), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, character)
}

func (h *CharacterHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.svc.Delete(r.Context(), userFrom(r.Context()), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Rest & Death Saves

func (h *CharacterHandler) LongRest(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	sheet, err := h.svc.LongRest(r.Context(), userFrom(r.Context()), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sheet)
}

func (h *CharacterHandler) ShortRest(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in struct {
		HitDiceSpent int `json:"hit_dice_spent"`
		HPHealed     int `json:"hp_healed"`
	}
	if !decode(w, r, &in) {
		return
	}
	sheet, err := h.svc.ShortRest(r.Context(), userFrom(r.Context()), id, in.HitDiceSpent, in.HPHealed)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sheet)
}

func (h *CharacterHandler) DeathSave(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in struct {
		Roll int `json:"roll"`
	}
	if !decode(w, r, &in) {
		return
	}
	sheet, err := h.svc.DeathSave(r.Context(), userFrom(r.Context()), id, in.Roll)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sheet)
}

func (h *CharacterHandler) SplitLoot(w http.ResponseWriter, r *http.Request) {
	var in model.LootSplitRequest
	if !decode(w, r, &in) {
		return
	}
	if err := h.svc.SplitPartyLoot(r.Context(), userFrom(r.Context()), in); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Inventory

func (h *CharacterHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.CharacterItemInput
	if !decode(w, r, &in) {
		return
	}
	item, err := h.svc.AddItem(r.Context(), userFrom(r.Context()), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *CharacterHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.CharacterItemInput
	if !decode(w, r, &in) {
		return
	}
	item, err := h.svc.UpdateItem(r.Context(), userFrom(r.Context()), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CharacterHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteItem(r.Context(), userFrom(r.Context()), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Spells

func (h *CharacterHandler) AddSpell(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.CharacterSpellInput
	if !decode(w, r, &in) {
		return
	}
	spell, err := h.svc.AddSpell(r.Context(), userFrom(r.Context()), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, spell)
}

func (h *CharacterHandler) UpdateSpell(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.CharacterSpellInput
	if !decode(w, r, &in) {
		return
	}
	spell, err := h.svc.UpdateSpell(r.Context(), userFrom(r.Context()), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, spell)
}

func (h *CharacterHandler) DeleteSpell(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteSpell(r.Context(), userFrom(r.Context()), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// SetSlot writes one spell level. Spending a slot is this call with used raised
// by one, so the server stays the only place the count changes.
func (h *CharacterHandler) SetSlot(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	level, err := strconv.Atoi(r.PathValue("level"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid spell level", Code: "bad_request"})
		return
	}

	var in model.SpellSlotInput
	if !decode(w, r, &in) {
		return
	}
	slots, err := h.svc.SetSlot(r.Context(), userFrom(r.Context()), id, level, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, slots)
}

// SetProficiency marks a skill or save as trained, or clears it, and answers
// with the whole sheet so the derived modifiers come back with it.
func (h *CharacterHandler) SetProficiency(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in struct {
		model.Proficiency
		Trained bool `json:"trained"`
	}
	in.Trained = true
	if !decode(w, r, &in) {
		return
	}

	sheet, err := h.svc.SetProficiency(r.Context(), userFrom(r.Context()), id, in.Proficiency, in.Trained)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sheet)
}
