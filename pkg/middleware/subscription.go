package middleware

import "net/http"

const (
	ErrUserNotFound               = "user not found"
	ErrSubscriptionCheckFailed    = "subscription check failed"
	ErrActiveSubscriptionRequired = "active subscription required"
)

type SubChecker interface {
	HasActiveSubscription(userID uint) (bool, error)
}

func HasActiveSubscription(getUserID func(email string) (uint, error), subChecker SubChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			email, ok := r.Context().Value(ContextEmailKey).(string)
			if !ok || email == "" {
				sendJSONError(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := getUserID(email)
			if err != nil {
				sendJSONError(w, ErrUserNotFound, http.StatusUnauthorized)
				return
			}

			hasActive, err := subChecker.HasActiveSubscription(userID)
			if err != nil {
				sendJSONError(w, ErrSubscriptionCheckFailed, http.StatusInternalServerError)
				return
			}
			if !hasActive {
				sendJSONError(w, ErrActiveSubscriptionRequired, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
