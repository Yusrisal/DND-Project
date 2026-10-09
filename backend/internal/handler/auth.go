package handler

import (
	"context"
	"net/http"
	"time"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/service"
)

// SessionCookie holds the raw session token. It is HttpOnly, so no script can
// read it, and it rides along on the websocket upgrade for free because the app
// and the API share an origin.
const SessionCookie = "dnd_session"

type contextKey struct{}

var userContextKey contextKey

// guard turns a session cookie into a user and rejects requests that do not
// clear the bar for a route. Routes opt in explicitly in the router, so reading
// the route table tells you who may call what.
type guard struct {
	auth   *service.AuthService
	secure bool
}

// user requires any signed-in account.
func (g *guard) user(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := g.resolve(w, r)
		if !ok {
			return
		}
		next(w, r.WithContext(withUser(r.Context(), user)))
	}
}

// dm requires the DM role.
func (g *guard) dm(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := g.resolve(w, r)
		if !ok {
			return
		}
		if !user.IsDM() {
			writeError(w, service.ErrForbidden)
			return
		}
		next(w, r.WithContext(withUser(r.Context(), user)))
	}
}

func (g *guard) resolve(w http.ResponseWriter, r *http.Request) (model.User, bool) {
	cookie, err := r.Cookie(SessionCookie)
	if err != nil {
		writeError(w, service.ErrUnauthorized)
		return model.User{}, false
	}
	user, err := g.auth.UserFromToken(r.Context(), cookie.Value)
	if err != nil {
		// A stale cookie should not keep failing, so clear it on the way out.
		g.clearCookie(w)
		writeError(w, err)
		return model.User{}, false
	}
	return user, true
}

func (g *guard) setCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		Secure:   g.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (g *guard) clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   g.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func withUser(ctx context.Context, user model.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// userFrom is safe only inside a handler wrapped by guard.user or guard.dm.
func userFrom(ctx context.Context) model.User {
	user, _ := ctx.Value(userContextKey).(model.User)
	return user
}
