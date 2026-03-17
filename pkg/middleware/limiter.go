package middleware

import (
	"encoding/json"
	"fmt"
	"link-generator/pkg/helpers"
	"link-generator/pkg/limiter"
	"net/http"
)

const (
	errRateLimitUnauthorized = "Unauthorized"
	errTooManyRequests       = "too many requests"
)

func RateLimit(svc *limiter.LimiterService, keyType limiter.KeyType) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			value, ok := resolveKey(r, keyType)
			if !ok {
				writeJSONError(w, errRateLimitUnauthorized, http.StatusUnauthorized)
				return
			}

			result, err := svc.Allow(r.Context(), value)
			if err != nil || !result.Allowed {
				writeJSONError(w, errTooManyRequests, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func resolveKey(r *http.Request, keyType limiter.KeyType) (string, bool) {
	switch keyType {
	case limiter.KeyByAccountID:
		accountID, ok := r.Context().Value(ContextAccountIDKey).(uint)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("%d", accountID), true
	case limiter.KeyByIP:
		return helpers.GetClientIP(r), true
	default:
		return "", false
	}
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
