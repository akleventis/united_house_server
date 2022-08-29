package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/akleventis/united_house_server/lib"
)

// Admin only access
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqToken := r.Header.Get("Authorization")
		splitToken := strings.Split(reqToken, "Bearer")
		if len(splitToken) != 2 {
			http.Error(w, lib.ErrInvalidTokenFormat.Error(), http.StatusBadRequest)
			return
		}

		reqToken = strings.TrimSpace(splitToken[1])
		auth := os.Getenv("BEARER")
		if reqToken != auth {
			http.Error(w, lib.ErrInvalidToken.Error(), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}
