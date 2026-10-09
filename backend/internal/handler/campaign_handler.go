package handler

import (
	"net/http"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
	"dndproject/backend/internal/service"
)

type CampaignHandler struct {
	svc *service.CampaignService
}

func NewCampaignHandler(svc *service.CampaignService) *CampaignHandler {
	return &CampaignHandler{svc: svc}
}

func (h *CampaignHandler) List(w http.ResponseWriter, r *http.Request) {
	q := model.ParseQuery(r, repository.CampaignSortColumns(), "created_at")
	page, err := h.svc.List(r.Context(), q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *CampaignHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	campaign, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, campaign)
}

func (h *CampaignHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in model.CampaignInput
	if !decode(w, r, &in) {
		return
	}
	campaign, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, campaign)
}

func (h *CampaignHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.CampaignInput
	if !decode(w, r, &in) {
		return
	}
	campaign, err := h.svc.Update(r.Context(), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, campaign)
}

func (h *CampaignHandler) SetPause(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.PauseInput
	if !decode(w, r, &in) {
		return
	}
	campaign, err := h.svc.SetPause(r.Context(), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, campaign)
}

func (h *CampaignHandler) SetRetention(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.RetentionInput
	if !decode(w, r, &in) {
		return
	}
	campaign, err := h.svc.SetRetention(r.Context(), id, in.ChatRetentionDays)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, campaign)
}

func (h *CampaignHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
