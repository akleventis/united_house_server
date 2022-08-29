package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/akleventis/united_house_server/lib"
	log "github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

type Handler struct {
	mc *gomail.Message
}

func NewHandler(mc *gomail.Message) *Handler {
	return &Handler{
		mc: mc,
	}
}

// https://www.courier.com/guides/golang-send-email/
type Email struct {
	From string `json:"name"`
	Body string `json:"body"`
}

func (h *Handler) SendEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var email Email
		if err := json.NewDecoder(r.Body).Decode(&email); err != nil {
			http.Error(w, lib.ErrInvalidArgJsonBody.Error(), http.StatusBadRequest)
			return
		}
		h.mc.SetHeader("From", email.From)
		h.mc.SetHeader("To", "unitedhouseproductions@gmail.com")
		h.mc.SetBody("text/html", email.Body)

		n := gomail.NewDialer("smtp.gmail.com", 587, os.Getenv("MAIL_KEY"), os.Getenv("MAIL_PW"))

		// Send the email
		if err := n.DialAndSend(h.mc); err != nil {
			log.Info(err)
			http.Error(w, lib.ErrEmail.Error(), http.StatusInternalServerError)
			return
		}
		lib.ApiResponse(w, http.StatusAccepted, "message sent")
	}
}
