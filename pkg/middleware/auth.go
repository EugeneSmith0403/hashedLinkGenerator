package middleware

import (
	"adv/go-http/internal/auth"
	internalJWT "adv/go-http/internal/jwt"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type key string

const (
	ContextEmailKey key = "ContextEmailKey"
)

func sendJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func IsAuthed(jwtService *internalJWT.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			parts := strings.Split(r.Header.Get("Authorization"), "Bearer ")
			if len(parts) != 2 {
				sendJSONError(w, auth.Unauthorized, http.StatusUnauthorized)
				return
			}

			claims := &jwt.MapClaims{}
			checkedClaims, err := jwtService.CheckToken(strings.TrimSpace(parts[1]), claims)
			if err != nil {
				sendJSONError(w, auth.Unauthorized, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ContextEmailKey, (*checkedClaims)["email"])

			req := r.WithContext(ctx)
			next.ServeHTTP(w, req)
		})
	}
}
