package main

import (
	"net/http"
	"os"

	uhp_db "github.com/akleventis/united_house_server/db"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	stripe "github.com/stripe/stripe-go/v72"
)

// server struct (https://pace.dev/blog/2018/05/09/how-I-write-http-services-after-eight-years.html)
type server struct {
	db     *uhp_db.ProductDB
	router *http.ServeMux
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	stripe.Key = os.Getenv("STRIPE_KEY")

	pDB, err := uhp_db.OpenDBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer pDB.DB.Close()

	s := &server{
		db:     pDB,
		router: http.NewServeMux(),
	}

	s.router.HandleFunc("/checkout", s.handleCheckout())
	s.router.HandleFunc("/products", s.getProducts())
	s.router.HandleFunc("/webhook", s.handleWebhook())

	handler := cors.Default().Handler(s.router)
	http.ListenAndServe(":5001", handler)
}
