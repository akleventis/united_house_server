package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/akleventis/united_house_server/lib"
	"github.com/mailjet/mailjet-apiv3-go"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	mc *mailjet.Client
}

func NewHandler(mc *mailjet.Client) *Handler {
	return &Handler{
		mc: mc,
	}
}

type Email struct {
	Name string `json:"name"`
	From string `json:"from"`
	Body string `json:"body"`
}

func (h *Handler) SendEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uhpEmailAddress := os.Getenv("UHP_EMAIL")
		var email Email
		if err := json.NewDecoder(r.Body).Decode(&email); err != nil {
			http.Error(w, lib.ErrInvalidArgJsonBody.Error(), http.StatusBadRequest)
			return
		}
		messagesInfo := []mailjet.InfoMessagesV31{
			{
				From: &mailjet.RecipientV31{
					Email: uhpEmailAddress,
					Name:  "Booking",
				},
				To: &mailjet.RecipientsV31{
					mailjet.RecipientV31{
						Email: uhpEmailAddress,
						Name:  "Paul",
					},
				},
				Subject: "Booking Inquiry",
				HTMLPart: fmt.Sprintf(`<big>Email from %s:</big>
										<br/><br/>
										<big><i>&emsp;%s</i></big>
										<br/><br/>
										<big>You can reach %s back at %s</big>`, email.Name, email.Body, email.Name, email.From),
			},
		}
		messages := mailjet.MessagesV31{Info: messagesInfo}
		res, err := h.mc.SendMailV31(&messages)
		if err != nil {
			http.Error(w, lib.ErrEmail.Error(), http.StatusInternalServerError)
			return
		}
		log.Infof("Data: %+v\n", res)
		lib.ApiResponse(w, http.StatusAccepted, "message sent")
	}
}
