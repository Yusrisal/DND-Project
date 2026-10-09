package handler

import (
	"net/http"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
	"dndproject/backend/internal/service"
)

type HandoutHandler struct {
	svc *service.HandoutService
}

func NewHandoutHandler(svc *service.HandoutService) *HandoutHandler {
	return &HandoutHandler{svc: svc}
}

func (h *HandoutHandler) List(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r.Context())
	isDM := u.Role == model.RoleDM
	q := model.ParseQuery(r, repository.HandoutSortColumns(), "created_at")
	page, err := h.svc.List(r.Context(), queryID(r, "campaign_id"), isDM, q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *HandoutHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	u := userFrom(r.Context())
	isDM := u.Role == model.RoleDM
	handout, err := h.svc.Get(r.Context(), id, isDM)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, handout)
}

func (h *HandoutHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in model.HandoutInput
	if !decode(w, r, &in) {
		return
	}
	handout, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, handout)
}

func (h *HandoutHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.HandoutInput
	if !decode(w, r, &in) {
		return
	}
	handout, err := h.svc.Update(r.Context(), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, handout)
}

func (h *HandoutHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
