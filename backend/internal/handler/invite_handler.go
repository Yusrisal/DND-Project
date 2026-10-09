package handler

import (
	"net/http"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
	"dndproject/backend/internal/service"
)

type InviteHandler struct {
	svc *service.InviteService
}

func NewInviteHandler(svc *service.InviteService) *InviteHandler {
	return &InviteHandler{svc: svc}
}

func (h *InviteHandler) List(w http.ResponseWriter, r *http.Request) {
	q := model.ParseQuery(r, repository.InviteSortColumns(), "created_at")
	page, err := h.svc.List(r.Context(), q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *InviteHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in model.InviteInput
	if !decode(w, r, &in) {
		return
	}
	invite, err := h.svc.Create(r.Context(), userFrom(r.Context()).ID, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, invite)
}

func (h *InviteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid code", Code: "bad_request"})
		return
	}
	if err := h.svc.Delete(r.Context(), code); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
