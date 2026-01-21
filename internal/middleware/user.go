package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
)

type contextKey string

const usernameKey contextKey = "username"

func UserSession(store *sessions.CookieStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := store.Get(r, "user-session")

			username, ok := session.Values["username"].(string)
			if ok && username != "" {
				ctx := context.WithValue(r.Context(), usernameKey, username)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetUsername(r *http.Request) string {
	username, _ := r.Context().Value(usernameKey).(string)
	return username
}
