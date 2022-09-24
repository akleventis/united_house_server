package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/akleventis/united_house_server/lib"
	auth "github.com/akleventis/united_house_server/uhp_api/handlers/auth"
)

// Admin only access
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse authorization header
		token := r.Header.Get("Authorization")
		splitToken := strings.Split(token, "Bearer ")

		// grab token
		token = splitToken[1]

		// verify token is still valid
		UserSession, exists := auth.UserSessions[token]
		if !exists {
			http.Error(w, lib.ErrInvalidToken.Error(), http.StatusForbidden)
			return
		}

		// remove expired token from map
		if UserSession.Expires.Before(time.Now()) {
			delete(auth.UserSessions, token)
			http.Error(w, lib.ErrTokenExpired.Error(), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}
