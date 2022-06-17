package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/akleventis/united_house_server/merchdb"
	log "github.com/sirupsen/logrus"
	stripe "github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	webhook "github.com/stripe/stripe-go/webhook"
)

type lineItems struct {
	Items []*merchdb.Product `json:"items"`
}

func apiResponse(w http.ResponseWriter, code int, obj interface{}) {
	r, _ := json.Marshal(obj)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(r)
}

// fulfillOrder verifies items are in stock and returns the resulting products array
func (s *server) fulfillOrder(items lineItems) ([]*merchdb.Product, error) {
	var products []*merchdb.Product
	for _, v := range items.Items {
		p, err := s.db.GetOrder(v.ID, v.Quantity)
		if err != nil {
			if err == merchdb.ErrOutOfStock && p != nil {
				return []*merchdb.Product{p}, merchdb.ErrOutOfStock
			}
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

// createLineItems converts products array to stripe LineItems
func createLineItems(products []*merchdb.Product) []*stripe.CheckoutSessionLineItemParams {
	var cli []*stripe.CheckoutSessionLineItemParams
	for _, v := range products {
		item := &stripe.CheckoutSessionLineItemParams{
			Name:        stripe.String(string(v.Name)),
			Description: stripe.String(string(v.ID)), // product_id for webhook update database inventory
			Amount:      stripe.Int64(int64(v.Price * 100)),
			Currency:    stripe.String("usd"),
			Quantity:    stripe.Int64(int64(v.Quantity)),
		}
		cli = append(cli, item)
	}
	return cli
}

// createCheckoutSession creates and returns a stripe checkout session object
func createCheckoutSession(cli []*stripe.CheckoutSessionLineItemParams) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
		PaymentMethodTypes: []*string{stripe.String("card")},
		SuccessURL:         stripe.String(os.Getenv("CLIENT_URL")),
		CancelURL:          stripe.String(os.Getenv("CLIENT_URL")),
		LineItems:          cli,
	}
	sesh, err := session.New(params)
	if err != nil {
		return nil, err
	}
	return sesh, nil
}

// handleCheckout receives array of items from client and returns a stripe checkout redirect url
func (s *server) HandleCheckout() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			http.Error(w, "INVALID_REQUEST_METHOD", http.StatusMethodNotAllowed)
			return
		}

		var items lineItems

		err := json.NewDecoder(req.Body).Decode(&items)
		if err != nil {
			http.Error(w, "INVALID_JSON", http.StatusBadRequest)
			return
		}

		// grab/verify items are in stock
		products, err := s.fulfillOrder(items)
		// TODO: handle error message front end
		// 				message := fmt.Sprintf("Only %d %s %s(s), in stock. Please update cart", p.Quantity, p.Size, p.Name)
		// 				message = fmt.Sprintf("%s %s is out of stock. Please update cart", p.Size, p.Name)
		if err != nil {
			if err == merchdb.ErrOutOfStock && len(products) > 0 {
				apiResponse(w, http.StatusAccepted, products[0]) // one product should exist, have front end deal with string logic
				return
			}
			http.Error(w, "INTERNAL_ERROR", http.StatusInternalServerError)
			return
		}

		checkoutLineItems := createLineItems(products)

		// create checkout session
		sesh, err := createCheckoutSession(checkoutLineItems)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sessionURL, err := json.Marshal(sesh.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		apiResponse(w, http.StatusOK, sessionURL)
	}
}

// getProducts returns json array of all products
func (s *server) GetProducts() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			http.Error(w, "INVALID_REQUEST_METHOD", http.StatusMethodNotAllowed)
			return
		}

		products, err := s.db.GetProducts()
		if err != nil {
			http.Error(w, "DB_ERROR", http.StatusInternalServerError)
			return
		}

		apiResponse(w, http.StatusOK, products)
	}
}

// ADMIN ONLY //

// PUT
func (s *server) UpdateProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			http.Error(w, "INVALID_REQUEST_METHOD", http.StatusMethodNotAllowed)
			return
		}

		// Validate auth token
		reqToken := r.Header.Get("Authorization")
		splitToken := strings.Split(reqToken, "Bearer")
		if len(splitToken) != 2 {
			http.Error(w, "INVALID_TOKEN_FORMAT", http.StatusBadRequest)
			return
		}

		reqToken = strings.TrimSpace(splitToken[1])
		auth := os.Getenv("BEARER")
		if reqToken != auth {
			http.Error(w, "INVALID_TOKEN", http.StatusForbidden)
			return
		}

		// Decode body
		var p merchdb.Product
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "JSON_ERROR", http.StatusBadRequest)
			return
		}

		if p.ID == "" {
			http.Error(w, "INVALID_ARG_ID", http.StatusBadRequest)
			return
		}

		// Get product to update
		updateProduct, err := s.db.GetProductById(p.ID)
		if err != nil {
			http.Error(w, "DB_ERROR", http.StatusInternalServerError)
			return
		}

		if updateProduct == nil {
			http.Error(w, "NOT_FOUND", http.StatusNotFound)
		}

		updateProduct.Name = p.Name
		updateProduct.Price = p.Price
		updateProduct.Size = p.Size
		updateProduct.Quantity = p.Quantity

		// Update product
		if err := s.db.Update(updateProduct); err != nil {
			http.Error(w, "DB_ERROR", http.StatusInternalServerError)
			return
		}
		apiResponse(w, http.StatusOK, updateProduct)
	}
}

// ==================================================================== //

// WEBHOOK
// handleWebhook() listens for Checkout Session Confirmation then Update Inventory accordingly
func (s *server) HandleWebhook() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const MaxBodyBytes = int64(65536)
		req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
		payload, err := ioutil.ReadAll(req.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		// Stripe CLI webhook secret for testing your endpoint locally.
		// create for prod https://dashboard.stripe.com/webhooks
		endpointSecret := os.Getenv("WHSEC")
		// Pass the request body and Stripe-Signature header to ConstructEvent, along with the webhook signing key.
		event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"),
			endpointSecret)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Unmarshal the event data into an appropriate struct depending on its Type
		if event.Type == "checkout.session.completed" {
			// Grab Session Data
			var sesh stripe.CheckoutSession
			err = json.Unmarshal(event.Data.Raw, &sesh)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// Grab each session line items Product ID and Quantity
			params := &stripe.CheckoutSessionListLineItemsParams{}
			i := session.ListLineItems(sesh.ID, params)
			for i.Next() {
				li := i.LineItem()

				id := li.Description // stripe product ID
				quantity := int(li.Quantity)

				if err := s.db.UpdateQuantity(id, quantity); err != nil {
					fmt.Fprintf(os.Stderr, "UPDATE INVENTORY ERROR: %v\n", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
		log.Info("")
		w.WriteHeader(http.StatusOK)
	}
}
