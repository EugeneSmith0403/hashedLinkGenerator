package middleware

import (
	"adv/go-http/internal/auth"
	"adv/go-http/internal/user"
	"net/http"
)

const (
	errUserNotFound             = "user not found"
	errSubscriptionCheckFailed  = "subscription check failed"
	errActiveSubscriptionRequired = "active subscription required"
)

type SubChecker interface {
	HasActiveSubscription(userID uint) (bool, error)
}

func HasActiveSubscription(userRepo *user.UserRepository, subChecker SubChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			email, ok := r.Context().Value(ContextEmailKey).(string)
			if !ok || email == "" {
				sendJSONError(w, auth.Unauthorized, http.StatusUnauthorized)
				return
			}

			currentUser, err := userRepo.FindByEmail(email)
			if err != nil || currentUser == nil {
				sendJSONError(w, errUserNotFound, http.StatusUnauthorized)
				return
			}

			hasActive, err := subChecker.HasActiveSubscription(currentUser.ID)
			if err != nil {
				sendJSONError(w, errSubscriptionCheckFailed, http.StatusInternalServerError)
				return
			}
			if !hasActive {
				sendJSONError(w, errActiveSubscriptionRequired, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
