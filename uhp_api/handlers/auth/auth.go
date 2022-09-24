package handlers

import (
	"net/http"
	"time"

	"github.com/akleventis/united_house_server/lib"
	"github.com/akleventis/united_house_server/uhp_db"
	"github.com/google/uuid"

	// log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var (
	UserSessions = map[string]UserSession{}
)

type ReqToken struct {
	Token string `json:"token"`
}

type UserSession struct {
	Username string
	Expires  time.Time
}

type Handler struct {
	db *uhp_db.UhpDB
}

func NewHandler(db *uhp_db.UhpDB) *Handler {
	return &Handler{
		db: db,
	}
}

func (h *Handler) SignIn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse secure basic authorization header
		username, password, ok := r.BasicAuth()
		bPass := []byte(password)

		// set a password
		// hashPass, err := bcrypt.GenerateFromPassword(bPass, bcrypt.DefaultCost)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		// log.Info(string(hashPass))

		if ok {
			// validate user exists and grab associated password
			expectedPassword, err := h.db.GetAdminUserPass(username)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// compare req password & db password. If successful => issue new token that expires in an hour
			if err = bcrypt.CompareHashAndPassword([]byte(expectedPassword), bPass); err == nil {
				var requestToken = &ReqToken{}
				token := uuid.NewString()
				expiresAt := time.Now().Add(1 * time.Hour)

				// create user session map entry
				UserSessions[token] = UserSession{
					Username: username,
					Expires:  expiresAt,
				}

				// return request token json obj
				requestToken.Token = token
				lib.ApiResponse(w, http.StatusAccepted, requestToken)
				return
			}
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
	}
}
