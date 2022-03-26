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

const stripeBaseURL = "https://api.stripe.com"

// handle get post request
func (s *server) handleCheckout() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var items lineItems
		err := json.NewDecoder(req.Body).Decode(&items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var products []*uhp_db.Product
		for _, v := range items.Items {
			product, err := uhp_db.GetProductById(s.db, v.ID, v.Quantity)
			if err != nil {
				if err == uhp_db.ErrOutOfStock {
					// TODO: how to send error message to front end?
					http.Error(w, fmt.Sprintf("Oops, look like we only have %d %s(s) in stock, please update cart", product.Quantity, product.Name), http.StatusBadRequest)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			products = append(products, product)
		}

		var checkoutLineItems []*stripe.CheckoutSessionLineItemParams
		for _, v := range products {
			item := &stripe.CheckoutSessionLineItemParams{
				Name:     stripe.String(v.Name),
				Amount:   stripe.Int64(int64(v.Price * 100)),
				Currency: stripe.String("usd"),
				Quantity: stripe.Int64(int64(v.Quantity)),
			}
			checkoutLineItems = append(checkoutLineItems, item)
		}
		params := &stripe.CheckoutSessionParams{
			Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
			PaymentMethodTypes: []*string{stripe.String("card")},
			SuccessURL:         stripe.String("http://localhost:3000"),
			CancelURL:          stripe.String("http://localhost:3000"),

			// loop through and fill line items with product objects
			LineItems: checkoutLineItems,
		}
		sesh, err := session.New(params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		resp := map[string]string{
			"url": sesh.URL,
		}
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// confirm payment went through, then subtract quantity from database
		// https://stripe.com/docs/payments/payment-intents/verifying-status#handling-webhook-events
		// seshID := sesh.ID

		// reduce quanitity in database
		for _, v := range products {
			log.Info("PRODUCT: ", v)
			if err := uhp_db.UpdateQuantity(s.db, v.ID, v.Quantity); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
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
