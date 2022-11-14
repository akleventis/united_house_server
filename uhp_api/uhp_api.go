package main

import (
	"net/http"
	"os"

	m "github.com/akleventis/united_house_server/middleware"

	// TODO: figure out why imports only work with named variable

	auth "github.com/akleventis/united_house_server/uhp_api/handlers/auth"
	checkout "github.com/akleventis/united_house_server/uhp_api/handlers/checkout"
	email "github.com/akleventis/united_house_server/uhp_api/handlers/email"
	"github.com/akleventis/united_house_server/uhp_db"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mailjet/mailjet-apiv3-go"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"

	stripev73 "github.com/stripe/stripe-go/v73"
)

func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := uhp_db.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()

	stripev73.Key = os.Getenv("STRIPE_KEY")

	router := mux.NewRouter()

	// stripe
	checkout := checkout.NewHandler(db)
	router.HandleFunc("/checkout", m.Limit(checkout.HandleCheckout(), m.RL10)).Methods("POST")
	// router.HandleFunc("/webhook", m.Limit(checkout.HandleWebhook(), m.RL10)).Methods("POST")
	router.HandleFunc("/products", m.Limit(checkout.GetProducts(), m.RL50)).Methods("GET")

	// email
	mailjetClient := mailjet.NewMailjetClient(os.Getenv("MAILJET_KEY"), os.Getenv("MAILJET_SECRET"))
	email := email.NewHandler(mailjetClient)
	router.HandleFunc("/mail", m.Limit(email.SendEmail(), m.RL5)).Methods("POST")

	// admin sign-in
	auth := auth.NewHandler(db)
	router.HandleFunc("/auth", auth.SignIn()).Methods("POST")

	handler := cors.AllowAll().Handler(router)
	http.ListenAndServe(":5001", handler)
}
