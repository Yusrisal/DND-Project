package handler

import (
	"net/http"
	"strconv"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
	"dndproject/backend/internal/service"
)

type CompendiumHandler struct {
	svc *service.CompendiumService
}

func NewCompendiumHandler(svc *service.CompendiumService) *CompendiumHandler {
	return &CompendiumHandler{svc: svc}
}

// Environments

func (h *CompendiumHandler) ListEnvironments(w http.ResponseWriter, r *http.Request) {
	q := model.ParseQuery(r, repository.CompendiumSortColumns(), "name")
	page, err := h.svc.ListEnvironments(r.Context(), queryID(r, "campaign_id"), queryID(r, "map_id"), q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *CompendiumHandler) GetEnvironment(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	env, err := h.svc.GetEnvironment(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, env)
}

func (h *CompendiumHandler) CreateEnvironment(w http.ResponseWriter, r *http.Request) {
	var in model.EnvironmentInput
	if !decode(w, r, &in) {
		return
	}
	env, err := h.svc.CreateEnvironment(r.Context(), userFrom(r.Context()), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, env)
}

func (h *CompendiumHandler) UpdateEnvironment(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.EnvironmentInput
	if !decode(w, r, &in) {
		return
	}
	env, err := h.svc.UpdateEnvironment(r.Context(), userFrom(r.Context()), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, env)
}

func (h *CompendiumHandler) DeleteEnvironment(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteEnvironment(r.Context(), userFrom(r.Context()), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Compendium Items

func (h *CompendiumHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	q := model.ParseQuery(r, repository.CompendiumSortColumns(), "name")
	category := r.URL.Query().Get("category")
	page, err := h.svc.ListItems(r.Context(), queryID(r, "campaign_id"), category, q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *CompendiumHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	item, err := h.svc.GetItem(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CompendiumHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var in model.CompendiumItemInput
	if !decode(w, r, &in) {
		return
	}
	item, err := h.svc.CreateItem(r.Context(), userFrom(r.Context()), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *CompendiumHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.CompendiumItemInput
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

func (h *CompendiumHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
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

// Monster Drops

func (h *CompendiumHandler) ListMonsterDrops(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	drops, err := h.svc.ListMonsterDrops(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, drops)
}

func (h *CompendiumHandler) CreateMonsterDrop(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.MonsterDropInput
	if !decode(w, r, &in) {
		return
	}
	in.MonsterID = id
	drop, err := h.svc.CreateMonsterDrop(r.Context(), userFrom(r.Context()), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, drop)
}

func (h *CompendiumHandler) DeleteMonsterDrop(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteMonsterDrop(r.Context(), userFrom(r.Context()), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CompendiumHandler) RollMonsterLoot(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	result, err := h.svc.RollMonsterLoot(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// Compendium Spells

func (h *CompendiumHandler) ListSpells(w http.ResponseWriter, r *http.Request) {
	q := model.ParseQuery(r, repository.CompendiumSortColumns(), "name")
	levelStr := r.URL.Query().Get("level")
	level := -1
	if levelStr != "" {
		if lvl, err := strconv.Atoi(levelStr); err == nil {
			level = lvl
		}
	}
	page, err := h.svc.ListSpells(r.Context(), queryID(r, "campaign_id"), level, q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *CompendiumHandler) GetSpell(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	spell, err := h.svc.GetSpell(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, spell)
}

func (h *CompendiumHandler) CreateSpell(w http.ResponseWriter, r *http.Request) {
	var in model.CompendiumSpellInput
	if !decode(w, r, &in) {
		return
	}
	spell, err := h.svc.CreateSpell(r.Context(), userFrom(r.Context()), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, spell)
}

func (h *CompendiumHandler) UpdateSpell(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.CompendiumSpellInput
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

func (h *CompendiumHandler) DeleteSpell(w http.ResponseWriter, r *http.Request) {
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

// Status Effects

func (h *CompendiumHandler) ListStatusEffects(w http.ResponseWriter, r *http.Request) {
	q := model.ParseQuery(r, repository.CompendiumSortColumns(), "name")
	kind := r.URL.Query().Get("kind")
	page, err := h.svc.ListStatusEffects(r.Context(), queryID(r, "campaign_id"), kind, q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *CompendiumHandler) GetStatusEffect(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	effect, err := h.svc.GetStatusEffect(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, effect)
}

func (h *CompendiumHandler) CreateStatusEffect(w http.ResponseWriter, r *http.Request) {
	var in model.StatusEffectInput
	if !decode(w, r, &in) {
		return
	}
	effect, err := h.svc.CreateStatusEffect(r.Context(), userFrom(r.Context()), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, effect)
}

func (h *CompendiumHandler) UpdateStatusEffect(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.StatusEffectInput
	if !decode(w, r, &in) {
		return
	}
	effect, err := h.svc.UpdateStatusEffect(r.Context(), userFrom(r.Context()), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, effect)
}

func (h *CompendiumHandler) DeleteStatusEffect(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteStatusEffect(r.Context(), userFrom(r.Context()), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
