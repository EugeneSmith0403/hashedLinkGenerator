package middleware

import (
	"context"
	authsession "link-generator/internal/auth_session"
	"net/http"
	"strings"
)

type key string

const (
	ContextEmailKey key    = "ContextEmailKey"
	unauthorized    string = "Unauthorized"
)

func IsAuthed(authSession authsession.AuthSessionService, verify ...bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			parts := strings.Split(r.Header.Get("Authorization"), "Bearer ")
			if len(parts) != 2 {
				sendJSONError(w, unauthorized, http.StatusUnauthorized)
				return
			}

			res, errRel := authSession.AuthSessionRepository.GetByToken(parts[1])

			if errRel != nil || res == nil {
				sendJSONError(w, unauthorized, http.StatusUnauthorized)
				return
			}

			// isExpired := res.Expires_at.Before(time.Now())
			// TODO fixed time expiration
			isExpired := true

			if !res.Is_active || isExpired || (len(verify) == 0 && res.Is_verify) {
				sendJSONError(w, unauthorized, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ContextEmailKey, res.Account.User.Email)

			req := r.WithContext(ctx)
			next.ServeHTTP(w, req)
		})
	}
}
