package main

import (
	"net/http"
	"os"

	"github.com/akleventis/united_house_server/merchdb"
	m "github.com/akleventis/united_house_server/middleware"
	"github.com/gorilla/mux"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	stripe "github.com/stripe/stripe-go"
)

type server struct {
	// db     merchdb.Datastore
	db     *merchdb.ProductDB
	router *mux.Router
}

func main() {
	if err := godotenv.Load("../.env"); err != nil {
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
		router: mux.NewRouter(),
	}

	// stripe
	s.router.HandleFunc("/checkout", m.Limit(s.HandleCheckout())).Methods("POST")
	s.router.HandleFunc("/webhook", m.Limit(s.HandleWebhook())).Methods("POST")

	// products
	s.router.HandleFunc("/products", m.Limit(s.GetProducts())).Methods("GET")

	// product
	s.router.HandleFunc("/product", m.Limit(m.Auth(s.CreateProduct()))).Methods("POST")
	s.router.HandleFunc("/product/{id}", m.Limit(m.Auth(s.GetProduct()))).Methods("GET")
	s.router.HandleFunc("/product/{id}", m.Limit(m.Auth(s.DeleteProduct()))).Methods("DELETE")
	s.router.HandleFunc("/product/{id}", m.Limit(m.Auth(s.UpdateProduct()))).Methods("PATCH")

	handler := cors.Default().Handler(s.router)
	http.ListenAndServe(":5001", handler)
}
