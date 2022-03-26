package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	uhp_db "github.com/akleventis/united_house_server/db"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	stripe "github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	webhook "github.com/stripe/stripe-go/webhook"
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
		if req.Method != "POST" {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
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
					log.Info("OUT OF STOCK")
					http.Error(w, fmt.Sprintf("Oops, look like we only have %d %s(s) in stock, please update cart", product.Quantity, product.Name), http.StatusBadRequest)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			products = append(products, product)
		}

		// Creates "LineItems" from products array
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

		// Create checkout session
		params := &stripe.CheckoutSessionParams{
			Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
			PaymentMethodTypes: []*string{stripe.String("card")},
			SuccessURL:         stripe.String("http://localhost:3000"),
			CancelURL:          stripe.String("http://localhost:3000"),
			LineItems:          checkoutLineItems,
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
		w.Write(jsonResp)
	}
}

func (s *server) getProducts() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

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

func updateInventory(items *stripe.LineItemList) {
	for _, v := range items.Data {
		log.Info(v)
	}
	// var p uhp_db.Product
	// for _, v := range items.Data {
	// 	p.ID =
	// 	p.Quantity = int(v.Quantity)

	// }
	// for _, v := range products {
	// 	log.Info("PRODUCT: ", v)
	// 	if err := uhp_db.UpdateQuantity(s.db, v.ID, v.Quantity); err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	// }
}

func (s *server) handleWebhook() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const MaxBodyBytes = int64(65536)
		req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
		payload, err := ioutil.ReadAll(req.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		// This is your Stripe CLI webhook secret for testing your endpoint locally.
		endpointSecret := "whsec_997481b842ea0e014921893e6a5767e23bd7c32b18e2a424db9046a99268adb3"
		// Pass the request body and Stripe-Signature header to ConstructEvent, along
		// with the webhook signing key.
		event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"),
			endpointSecret)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
			w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
			return
		}

		// Unmarshal the event data into an appropriate struct depending on its Type
		switch event.Type {
		case "checkout.session.completed":
			log.Info("PAYMENT SUCCEEDED")
			var session stripe.CheckoutSession
			err = json.Unmarshal(event.Data.Raw, &session)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// TODO: Need to get data from lineitems in order for inventory management
			// updateInventory(session.LineItems)
			// log.Info(session.LineItems)
		default:
			fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
		}

		w.WriteHeader(http.StatusOK)

	}
}

// func (s *server) ping() http.HandlerFunc {
// 	return func(w http.ResponseWriter, req *http.Request) {
// 		log.Info("PING")
// 		w.WriteHeader(http.StatusOK)
// 	}
// }

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

	// s.router.HandleFunc("/", s.ping())
	s.router.HandleFunc("/checkout", s.handleCheckout())
	s.router.HandleFunc("/products", s.getProducts())
	s.router.HandleFunc("/webhook", s.handleWebhook())

	handler := cors.Default().Handler(s.router)
	http.ListenAndServe(":5001", handler)
}
