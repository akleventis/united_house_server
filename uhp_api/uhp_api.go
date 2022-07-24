package main

import (
	"net/http"
	"os"

	m "github.com/akleventis/united_house_server/middleware"
	"github.com/akleventis/united_house_server/uhp_db"
	"github.com/gorilla/mux"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	stripe "github.com/stripe/stripe-go"
)

type server struct {
	// db     merchdb.Datastore => unit testing w/ testclient
	db     *uhp_db.UhpDB
	router *mux.Router
}

func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	stripe.Key = os.Getenv("STRIPE_KEY")

	db, err := uhp_db.Open()
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
	// admin only
	s.router.HandleFunc("/product/{id}", m.Limit(m.Auth(s.GetProduct()))).Methods("GET")
	s.router.HandleFunc("/product", m.Limit(m.Auth(s.CreateProduct()))).Methods("POST")
	s.router.HandleFunc("/product/{id}", m.Limit(m.Auth(s.UpdateProduct()))).Methods("PATCH")
	s.router.HandleFunc("/product/{id}", m.Limit(m.Auth(s.DeleteProduct()))).Methods("DELETE")

	// TODO
	// events
	// s.router.HandleFunc("/events", m.Limit(s.GetEvents())).Methods("GET")
	// admin only
	// s.router.HandleFunc("/event", m.Limit(m.Auth(s.CreateEvent()))).Methods("POST")
	// s.router.HandleFunc("/event/{id}", m.Limit(m.Auth(s.UpdateEvent()))).Methods("PATCH")
	// s.router.HandleFunc("/event/{id}", m.Limit(m.Auth(s.DeleteEvent()))).Methdos("DELETE")

	// featured artists (soundcloud)
	// s.router.HandleFunc("/artists", m.Limit(s.GetArtists())).Methods("GET")
	// admin only
	// s.router.HandleFunc("/artist", m.Limit(m.Auth(s.CreateArtist()))).Methods("POST")
	// s.router.HandleFunc("/artist/{id}", m.Limit(m.Auth(s.UpdateArtist()))).Methods("PATCH")
	// s.router.HandleFunc("/artist/{id}", m.Limit(m.Auth(s.DeleteArtist()))).Methods("DELETE")

	handler := cors.Default().Handler(s.router)
	http.ListenAndServe(":5001", handler)
}
