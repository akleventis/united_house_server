package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	uhp_db "github.com/akleventis/united_house_server/db"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
)

// server struct (https://pace.dev/blog/2018/05/09/how-I-write-http-services-after-eight-years.html)
type server struct {
	db     *sql.DB
	router *http.ServeMux
}

type lineItems struct {
	Items []*uhp_db.Product `json:"items"`
}

// handle get post request
func (s *server) handleCheckout() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var products lineItems
		err := json.NewDecoder(req.Body).Decode(&products)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// go into db grab item by id
		// if not enough quantity, return struct with quantity and bad request?
		// subtract quanitty from db
		// Grab item price and name
		// add to line items array
		// maybe build up the array before entering the checkout block below
		// make sure to subtract quantity AFTER successful checkout!
		params := &stripe.CheckoutSessionParams{
			Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
			PaymentMethodTypes: []*string{stripe.String("card")},
			SuccessURL:         stripe.String("http://localhost:3000"),
			CancelURL:          stripe.String("http://localhost:3000"),

			// loop through and fill line items with product objects
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{
					Name:     stripe.String("T-Shirt"),
					Amount:   stripe.Int64(1 * 100),
					Currency: stripe.String("usd"),
					Quantity: stripe.Int64(2),
				},
			},
		}
		s, err := session.New(params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		resp := map[string]string{
			"url": s.URL,
		}
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(jsonResp)
		// http.Redirect(w, req, s.URL, http.StatusSeeOther)
	}
}

func (s *server) getProducts() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
		// get products
		products, err := uhp_db.GetProducts(s.db)
		if err != nil {
			log.Fatal(err)
		}
		for _, v := range products {
			fmt.Println(v)
		}
		fmt.Println(products)

	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	stripe.Key = os.Getenv("STRIPE_KEY")

	db, err := uhp_db.OpenDBConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	s := &server{
		db:     db,
		router: http.NewServeMux(),
	}

	s.router.HandleFunc("/checkout", s.handleCheckout())
	s.router.HandleFunc("/products", s.getProducts())

	handler := cors.Default().Handler(s.router)
	http.ListenAndServe(":5001", handler)
}
