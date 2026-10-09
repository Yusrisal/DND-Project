package handler

import (
	"net/http"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
	"dndproject/backend/internal/service"
)

type ChatHandler struct {
	svc *service.ChatService
}

func NewChatHandler(svc *service.ChatService) *ChatHandler {
	return &ChatHandler{svc: svc}
}

// List is the scrollback behind the live feed. Live messages arrive on the
// websocket; this is for reading further back.
func (h *ChatHandler) List(w http.ResponseWriter, r *http.Request) {
	campaignID := queryID(r, "campaign_id")
	if campaignID == 0 {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "campaign_id is required", Code: "bad_request"})
		return
	}

	q := model.ParseQuery(r, repository.ChatSortColumns(), "id")
	page, err := h.svc.List(r.Context(), campaignID, q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}
