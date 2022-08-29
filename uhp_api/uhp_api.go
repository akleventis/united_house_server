package main

import (
	"net/http"
	"os"

	m "github.com/akleventis/united_house_server/middleware"
	checkout "github.com/akleventis/united_house_server/uhp_api/handlers/checkout"
	email "github.com/akleventis/united_house_server/uhp_api/handlers/email"
	events "github.com/akleventis/united_house_server/uhp_api/handlers/events"
	featured_artists "github.com/akleventis/united_house_server/uhp_api/handlers/featured_artists"
	products "github.com/akleventis/united_house_server/uhp_api/handlers/products"
	"github.com/akleventis/united_house_server/uhp_db"
	"github.com/gorilla/mux"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	stripe "github.com/stripe/stripe-go"
	gomail "gopkg.in/gomail.v2"
)

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

	router := mux.NewRouter()

	// stripe
	checkout := checkout.NewHandler(db)
	router.HandleFunc("/checkout", m.Limit(checkout.HandleCheckout(), m.RL10)).Methods("POST")
	router.HandleFunc("/webhook", m.Limit(checkout.HandleWebhook(), m.RL10)).Methods("POST")

	// products
	products := products.NewHandler(db)
	router.HandleFunc("/products", m.Limit(products.GetProducts(), m.RL50)).Methods("GET")
	router.HandleFunc("/product/{id}", m.Limit(m.Auth(products.GetProduct()), m.RL30)).Methods("GET")       // admin
	router.HandleFunc("/product", m.Limit(m.Auth(products.CreateProduct()), m.RL30)).Methods("POST")        // admin
	router.HandleFunc("/product/{id}", m.Limit(m.Auth(products.UpdateProduct()), m.RL30)).Methods("PATCH")  // admin
	router.HandleFunc("/product/{id}", m.Limit(m.Auth(products.DeleteProduct()), m.RL30)).Methods("DELETE") // admin

	// events
	events := events.NewHandler(db)
	router.HandleFunc("/events", m.Limit(events.GetEvents(), m.RL50)).Methods("GET")
	router.HandleFunc("/event/{id}", m.Limit(events.GetEvent(), m.RL50)).Methods("GET")
	router.HandleFunc("/event", m.Limit(m.Auth(events.CreateEvent()), m.RL30)).Methods("POST")        // admin
	router.HandleFunc("/event/{id}", m.Limit(m.Auth(events.UpdateEvent()), m.RL30)).Methods("PATCH")  // admin
	router.HandleFunc("/event/{id}", m.Limit(m.Auth(events.DeleteEvent()), m.RL30)).Methods("DELETE") // admin

	// featured artist soundcloud iframe
	fa := featured_artists.NewHandler(db)
	router.HandleFunc("/featured_artists", m.Limit(fa.GetFeaturedArtists(), m.RL50)).Methods("GET")
	router.HandleFunc("/featured_artist", m.Limit(m.Auth(fa.CreateFeaturedArtist()), m.RL30)).Methods("POST")        // admin
	router.HandleFunc("/featured_artist/{id}", m.Limit(m.Auth(fa.UpdateFeaturedArtist()), m.RL30)).Methods("PATCH")  // admin
	router.HandleFunc("/featured_artist/{id}", m.Limit(m.Auth(fa.DeleteFeaturedArtist()), m.RL30)).Methods("DELETE") // admin

	// email
	mc := gomail.NewMessage()
	email := email.NewHandler(mc)
	router.HandleFunc("/mail", m.Limit(email.SendEmail(), m.RL5)).Methods("POST")

	handler := cors.Default().Handler(router)
	http.ListenAndServe(":5001", handler)
}
