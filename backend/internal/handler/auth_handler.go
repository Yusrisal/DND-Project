package handler

import (
	"net/http"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
	"dndproject/backend/internal/service"
)

type AuthHandler struct {
	svc   *service.AuthService
	guard *guard
}

func NewAuthHandler(svc *service.AuthService, g *guard) *AuthHandler {
	return &AuthHandler{svc: svc, guard: g}
}

// Register creates the account and signs it in, so an invited player lands on
// the board instead of on a login form.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var in model.RegisterInput
	if !decode(w, r, &in) {
		return
	}
	if _, err := h.svc.Register(r.Context(), in); err != nil {
		writeError(w, err)
		return
	}
	h.issueSession(w, r, model.Credentials{Email: in.Email, Password: in.Password}, http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var in model.Credentials
	if !decode(w, r, &in) {
		return
	}
	h.issueSession(w, r, in, http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(SessionCookie); err == nil {
		if err := h.svc.Logout(r.Context(), cookie.Value); err != nil {
			writeError(w, err)
			return
		}
	}
	h.guard.clearCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// Me is how the frontend learns on load whether it has a session and what role
// it holds.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, userFrom(r.Context()))
}

// Users backs the owner picker: the DM needs names to assign tokens to.
func (h *AuthHandler) Users(w http.ResponseWriter, r *http.Request) {
	q := model.ParseQuery(r, repository.UserSortColumns(), "display_name")
	page, err := h.svc.ListUsers(r.Context(), q)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *AuthHandler) issueSession(w http.ResponseWriter, r *http.Request, in model.Credentials, status int) {
	user, token, expiresAt, err := h.svc.Login(r.Context(), in)
	if err != nil {
		writeError(w, err)
		return
	}
	h.guard.setCookie(w, token, expiresAt)
	writeJSON(w, status, user)
}
