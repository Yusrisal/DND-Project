package handler

import (
	"net/http"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
	"dndproject/backend/internal/service"
)

type ShopHandler struct {
	svc *service.ShopService
}

func NewShopHandler(svc *service.ShopService) *ShopHandler {
	return &ShopHandler{svc: svc}
}

func (h *ShopHandler) List(w http.ResponseWriter, r *http.Request) {
	q := model.ParseQuery(r, repository.ShopSortColumns(), "name")
	page, err := h.svc.List(r.Context(), queryID(r, "campaign_id"), q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *ShopHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	details, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, details)
}

func (h *ShopHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in model.ShopInput
	if !decode(w, r, &in) {
		return
	}
	shop, err := h.svc.Create(r.Context(), userFrom(r.Context()), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, shop)
}

func (h *ShopHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.ShopInput
	if !decode(w, r, &in) {
		return
	}
	shop, err := h.svc.Update(r.Context(), userFrom(r.Context()), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, shop)
}

func (h *ShopHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

func (h *ShopHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.ShopItemInput
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

func (h *ShopHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.ShopItemInput
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

func (h *ShopHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
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

func (h *ShopHandler) BuyItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var req model.BuyItemRequest
	if !decode(w, r, &req) {
		return
	}
	item, err := h.svc.BuyItem(r.Context(), userFrom(r.Context()), id, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ShopHandler) SellItem(w http.ResponseWriter, r *http.Request) {
	var req model.SellItemRequest
	if !decode(w, r, &req) {
		return
	}
	result, err := h.svc.SellItem(r.Context(), userFrom(r.Context()), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
