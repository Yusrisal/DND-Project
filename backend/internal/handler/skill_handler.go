package handler

import (
	"net/http"

	"dndproject/backend/internal/service"
)

type SkillHandler struct {
	svc  *service.SkillService
	char *service.CharacterService
}

func NewSkillHandler(svc *service.SkillService, char *service.CharacterService) *SkillHandler {
	return &SkillHandler{svc: svc, char: char}
}

func (h *SkillHandler) ListTrees(w http.ResponseWriter, r *http.Request) {
	className := r.URL.Query().Get("class_name")
	skills, err := h.svc.ListClassSkills(r.Context(), className)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, skills)
}

func (h *SkillHandler) Learn(w http.ResponseWriter, r *http.Request) {
	charID, ok := pathID(w, r)
	if !ok {
		return
	}
	var in struct {
		SkillID int64 `json:"skill_id"`
	}
	if !decode(w, r, &in) {
		return
	}
	u := userFrom(r.Context())
	if err := h.svc.LearnSkill(r.Context(), u, charID, in.SkillID); err != nil {
		writeError(w, err)
		return
	}
	sheet, err := h.char.Sheet(r.Context(), u, charID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sheet)
}

func (h *SkillHandler) ToggleActive(w http.ResponseWriter, r *http.Request) {
	charID, ok := pathID(w, r)
	if !ok {
		return
	}
	var in struct {
		SkillID int64 `json:"skill_id"`
		Active  bool  `json:"active"`
	}
	if !decode(w, r, &in) {
		return
	}
	u := userFrom(r.Context())
	if err := h.svc.ToggleSkillActive(r.Context(), u, charID, in.SkillID, in.Active); err != nil {
		writeError(w, err)
		return
	}
	sheet, err := h.char.Sheet(r.Context(), u, charID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sheet)
}

func (h *SkillHandler) UseScroll(w http.ResponseWriter, r *http.Request) {
	charID, ok := pathID(w, r)
	if !ok {
		return
	}
	var in struct {
		ItemID int64 `json:"item_id"`
	}
	if !decode(w, r, &in) {
		return
	}
	u := userFrom(r.Context())
	if err := h.svc.UseSkillScroll(r.Context(), u, charID, in.ItemID); err != nil {
		writeError(w, err)
		return
	}
	sheet, err := h.char.Sheet(r.Context(), u, charID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sheet)
}

func (h *SkillHandler) AllocateStat(w http.ResponseWriter, r *http.Request) {
	charID, ok := pathID(w, r)
	if !ok {
		return
	}
	var in struct {
		Ability string `json:"ability"`
	}
	if !decode(w, r, &in) {
		return
	}
	u := userFrom(r.Context())
	sheet, err := h.char.AllocateStat(r.Context(), u, charID, in.Ability)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sheet)
}

func (h *SkillHandler) TransferItem(w http.ResponseWriter, r *http.Request) {
	charID, ok := pathID(w, r)
	if !ok {
		return
	}
	var in struct {
		ToCharacterID int64 `json:"to_character_id"`
		ItemID        int64 `json:"item_id"`
		Quantity      int   `json:"quantity"`
	}
	if !decode(w, r, &in) {
		return
	}
	u := userFrom(r.Context())
	if err := h.char.TransferItem(r.Context(), u, charID, in.ToCharacterID, in.ItemID, in.Quantity); err != nil {
		writeError(w, err)
		return
	}
	sheet, err := h.char.Sheet(r.Context(), u, charID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sheet)
}

// DM Gating & Grants
func (h *SkillHandler) AddXP(w http.ResponseWriter, r *http.Request) {
	charID, ok := pathID(w, r)
	if !ok {
		return
	}
	var in struct {
		Amount int64 `json:"amount"`
	}
	if !decode(w, r, &in) {
		return
	}
	char, err := h.char.AddXP(r.Context(), charID, in.Amount)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, char)
}

func (h *SkillHandler) SetLevelBlock(w http.ResponseWriter, r *http.Request) {
	charID, ok := pathID(w, r)
	if !ok {
		return
	}
	var in struct {
		Blocked bool   `json:"blocked"`
		Reason  string `json:"reason"`
	}
	if !decode(w, r, &in) {
		return
	}
	char, err := h.char.SetLevelBlock(r.Context(), charID, in.Blocked, in.Reason)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, char)
}

func (h *SkillHandler) GrantPoints(w http.ResponseWriter, r *http.Request) {
	charID, ok := pathID(w, r)
	if !ok {
		return
	}
	var in struct {
		StatPoints  int `json:"stat_points"`
		SkillPoints int `json:"skill_points"`
	}
	if !decode(w, r, &in) {
		return
	}
	char, err := h.char.GrantPoints(r.Context(), charID, in.StatPoints, in.SkillPoints)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, char)
}

// Pending Loot Drops
func (h *SkillHandler) ListPendingLoot(w http.ResponseWriter, r *http.Request) {
	campaignID := queryID(r, "campaign_id")
	drops, err := h.svc.ListPendingLoot(r.Context(), campaignID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, drops)
}

func (h *SkillHandler) ResolvePendingLoot(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in struct {
		Approve bool   `json:"approve"`
		Reason  string `json:"reason"`
	}
	if !decode(w, r, &in) {
		return
	}
	if err := h.svc.ResolvePendingLoot(r.Context(), id, in.Approve, in.Reason); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
