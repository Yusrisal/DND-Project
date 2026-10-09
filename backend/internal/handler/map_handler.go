package handler

import (
	"net/http"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
	"dndproject/backend/internal/service"
)

type MapHandler struct {
	svc *service.MapService
}

func NewMapHandler(svc *service.MapService) *MapHandler {
	return &MapHandler{svc: svc}
}

func (h *MapHandler) List(w http.ResponseWriter, r *http.Request) {
	q := model.ParseQuery(r, repository.MapSortColumns(), "created_at")
	page, err := h.svc.List(r.Context(), queryID(r, "campaign_id"), q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *MapHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	m, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *MapHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in model.MapInput
	if !decode(w, r, &in) {
		return
	}
	m, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (h *MapHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.MapInput
	if !decode(w, r, &in) {
		return
	}
	m, err := h.svc.Update(r.Context(), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *MapHandler) UpdateFog(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.FogInput
	if !decode(w, r, &in) {
		return
	}
	m, err := h.svc.UpdateFog(r.Context(), id, in.FogEnabled, in.FogRevealed)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *MapHandler) SetWeather(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.WeatherInput
	if !decode(w, r, &in) {
		return
	}
	m, err := h.svc.SetWeather(r.Context(), id, in.Weather)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *MapHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

func (h *MapHandler) ListAoETemplates(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	list, err := h.svc.ListAoETemplates(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *MapHandler) CreateAoETemplate(w http.ResponseWriter, r *http.Request) {
	mapID, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.AoETemplateInput
	if !decode(w, r, &in) {
		return
	}
	in.MapID = mapID
	item, err := h.svc.CreateAoETemplate(r.Context(), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *MapHandler) DeleteAoETemplate(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteAoETemplate(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Map Objects (Doors, Chests, Traps, Markers)

func (h *MapHandler) ListObjects(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	list, err := h.svc.ListObjects(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *MapHandler) CreateObject(w http.ResponseWriter, r *http.Request) {
	mapID, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.MapObjectInput
	if !decode(w, r, &in) {
		return
	}
	in.MapID = mapID
	item, err := h.svc.CreateObject(r.Context(), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *MapHandler) UpdateObject(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in model.MapObjectInput
	if !decode(w, r, &in) {
		return
	}
	item, err := h.svc.UpdateObject(r.Context(), id, in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *MapHandler) DeleteObject(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteObject(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
