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
	// db     merchdb.Datastore
	db     merchdb.ProductDB
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
		db:     *db,
		router: http.NewServeMux(),
	}

	// stripe
	s.router.HandleFunc("/checkout", rl.Limit(s.HandleCheckout())) // POST
	s.router.HandleFunc("/webhook", rl.Limit(s.HandleWebhook()))   // POST => from stripe

	// products
	s.router.HandleFunc("/get_products", rl.Limit(s.GetProducts()))     // GET
	s.router.HandleFunc("/update_product", rl.Limit(s.UpdateProduct())) // PUT (include all fields)

	handler := cors.Default().Handler(s.router)
	http.ListenAndServe(":5001", handler)
}
