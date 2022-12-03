package main

import (
	"net/http"
	"os"

	m "github.com/akleventis/united_house_server/middleware"
	checkout "github.com/akleventis/united_house_server/uhp_api/handlers/checkout"
	email "github.com/akleventis/united_house_server/uhp_api/handlers/email"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/mailjet/mailjet-apiv3-go"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	stripev73 "github.com/stripe/stripe-go/v73"
)

func main() {
	port := ":5001"
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	stripev73.Key = os.Getenv("STRIPE_KEY")

	router := mux.NewRouter()

	// stripe
	checkout := checkout.NewHandler()
	router.HandleFunc("/checkout", m.Limit(checkout.HandleCheckout(), m.RL10)).Methods("POST")
	router.HandleFunc("/products", m.Limit(checkout.GetProducts(), m.RL50)).Methods("GET")

	// email
	mailjetClient := mailjet.NewMailjetClient(os.Getenv("MAILJET_KEY"), os.Getenv("MAILJET_SECRET"))
	email := email.NewHandler(mailjetClient)
	router.HandleFunc("/mail", m.Limit(email.SendEmail(), m.RL5)).Methods("POST")

	handler := cors.Default().Handler(router)

	log.Infof("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(port, handler))
}
