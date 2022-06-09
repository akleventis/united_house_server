package main

import (
	"net/http"
	"os"

	"github.com/akleventis/united_house_server/merchdb"
	rl "github.com/akleventis/united_house_server/ratelimit"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	stripe "github.com/stripe/stripe-go/v72"
)

type server struct {
	db     merchdb.Datastore
	router *http.ServeMux
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	stripe.Key = os.Getenv("STRIPE_KEY")

	db, err := merchdb.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()

	s := &server{
		db:     db,
		router: http.NewServeMux(),
	}

	s.router.HandleFunc("/checkout", rl.Limit(s.HandleCheckout()))
	s.router.HandleFunc("/products", rl.Limit(s.GetProducts()))
	s.router.HandleFunc("/webhook", rl.Limit(s.HandleWebhook()))

	handler := cors.Default().Handler(s.router)
	http.ListenAndServe(":5001", handler)
}
