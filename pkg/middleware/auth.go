package middleware

import (
	internalJWT "link-generator/internal/jwt"
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type key string

const (
	ContextEmailKey key    = "ContextEmailKey"
	unauthorized    string = "Unauthorized"
)

func IsAuthed(jwtService *internalJWT.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			parts := strings.Split(r.Header.Get("Authorization"), "Bearer ")
			if len(parts) != 2 {
				sendJSONError(w, unauthorized, http.StatusUnauthorized)
				return
			}

			claims := &jwt.MapClaims{}
			checkedClaims, err := jwtService.CheckToken(strings.TrimSpace(parts[1]), claims)
			if err != nil {
				sendJSONError(w, unauthorized, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ContextEmailKey, (*checkedClaims)["email"])

			req := r.WithContext(ctx)
			next.ServeHTTP(w, req)
		})
	}
}
