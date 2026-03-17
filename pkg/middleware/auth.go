package middleware

import (
	"context"
	authsession "link-generator/internal/auth_session"
	"net/http"
	"strings"
	"time"
)

type key string

const (
	ContextEmailKey     key    = "ContextEmailKey"
	ContextAccountIDKey key    = "ContextAccountIDKey"
	unauthorized        string = "Unauthorized"
)

func IsAuthed(authSession authsession.AuthSessionService, allowVerify ...bool) func(http.Handler) http.Handler {
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

			isExpired := res.Expires_at.Before(time.Now())

			isVerify := res.Is_verify
			if len(allowVerify) > 0 && allowVerify[0] && !isVerify {
				isVerify = true
			}

			if !res.Is_active || isExpired || !isVerify {
				sendJSONError(w, unauthorized, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ContextEmailKey, res.Account.User.Email)
			ctx = context.WithValue(ctx, ContextAccountIDKey, res.Account.ID)

			req := r.WithContext(ctx)
			next.ServeHTTP(w, req)
		})
	}
}
