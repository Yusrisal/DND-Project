package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"dndproject/backend/internal/service"
)

type errorBody struct {
	Error string `json:"error"`
	Code  string `json:"code"`
	Field string `json:"field,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("encode response", "error", err)
	}
}

// writeError maps the service layer's two sentinel failures onto status codes
// and hides anything else behind a 500.
func writeError(w http.ResponseWriter, err error) {
	var invalid service.ValidationError
	switch {
	case errors.Is(err, service.ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, errorBody{Error: "sign in required", Code: "unauthorized"})
	case errors.Is(err, service.ErrForbidden):
		writeJSON(w, http.StatusForbidden, errorBody{Error: "not allowed", Code: "forbidden"})
	case errors.Is(err, service.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorBody{Error: "not found", Code: "not_found"})
	case errors.As(err, &invalid):
		writeJSON(w, http.StatusUnprocessableEntity, errorBody{Error: invalid.Message, Code: "validation_failed", Field: invalid.Field})
	default:
		slog.Error("request failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorBody{Error: "internal server error", Code: "internal"})
	}
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid JSON body", Code: "bad_request"})
		return false
	}
	return true
}

// pathID reads {id} from the route pattern, answering 400 itself when it is not a positive number.
func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid id", Code: "bad_request"})
		return 0, false
	}
	return id, true
}

// queryID reads an optional numeric filter such as ?campaign_id=3, returning 0 when absent.
func queryID(r *http.Request, key string) int64 {
	id, err := strconv.ParseInt(r.URL.Query().Get(key), 10, 64)
	if err != nil || id < 0 {
		return 0
	}
	return id
}
