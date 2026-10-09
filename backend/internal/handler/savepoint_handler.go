package handler

import (
	"net/http"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
	"dndproject/backend/internal/service"
)

type SavepointHandler struct {
	svc *service.SavepointService
}

func NewSavepointHandler(svc *service.SavepointService) *SavepointHandler {
	return &SavepointHandler{svc: svc}
}

func (h *SavepointHandler) List(w http.ResponseWriter, r *http.Request) {
	campaignID := queryID(r, "campaign_id")
	if campaignID <= 0 {
		var ok bool
		campaignID, ok = pathID(w, r)
		if !ok {
			return
		}
	}
	q := model.ParseQuery(r, repository.SavepointSortColumns(), "created_at")
	page, err := h.svc.List(r.Context(), campaignID, q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *SavepointHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in model.SavepointInput
	if !decode(w, r, &in) {
		return
	}
	sp, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sp)
}

func (h *SavepointHandler) Restore(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	sp, err := h.svc.Restore(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sp)
}

func (h *SavepointHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
