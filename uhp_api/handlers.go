package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	e "github.com/akleventis/united_house_server/errors"
	"github.com/akleventis/united_house_server/merchdb"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	stripe "github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	webhook "github.com/stripe/stripe-go/webhook"
)

type lineItems struct {
	Items []*merchdb.Product `json:"items"`
}

func apiResponse(w http.ResponseWriter, code int, obj interface{}) {
	r, err := json.Marshal(obj)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
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
			if err == e.ErrOutOfStock && p != nil {
				return []*merchdb.Product{p}, e.ErrOutOfStock
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
		var items lineItems

		err := json.NewDecoder(req.Body).Decode(&items)
		if err != nil {
			http.Error(w, e.ErrInvalidArgJsonBody.Error(), http.StatusBadRequest)
			return
		}

		// grab/verify items are in stock
		products, err := s.fulfillOrder(items)
		// TODO: handle error message front end
		// 				message := fmt.Sprintf("Only %d %s %s(s), in stock. Please update cart", p.Quantity, p.Size, p.Name)
		// 				message = fmt.Sprintf("%s %s is out of stock. Please update cart", p.Size, p.Name)
		if err != nil {
			if err == e.ErrOutOfStock && len(products) > 0 {
				apiResponse(w, http.StatusAccepted, products[0]) // one product should exist, have front end deal with string logic
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
		products, err := s.db.GetProducts()
		if err != nil {
			http.Error(w, e.ErrDB.Error(), http.StatusInternalServerError)
			return
		}

		apiResponse(w, http.StatusOK, products)
	}
}

// ADMIN ONLY: GetProduct retrieves a product using and id
func (s *server) GetProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if vars == nil {
			http.Error(w, e.ErrInvalidID.Error(), http.StatusBadRequest)
		}
		id := vars["id"]

		p, err := s.db.Get(id)
		if err != nil {
			http.Error(w, e.ErrDB.Error(), http.StatusInternalServerError)
			return
		}
		if p == nil {
			http.Error(w, http.StatusText(404), http.StatusNotFound)
			return
		}
		apiResponse(w, http.StatusOK, p)
	}
}

// ADMIN ONLY: CreateProduct creates a product based on provided fields (id, name, size, price, quantity)
func (s *server) CreateProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p merchdb.Product
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, e.ErrInvalidArgJsonBody.Error(), http.StatusBadRequest)
			return
		}
		res, err := s.db.Create(p)
		if err != nil {
			http.Error(w, e.ErrDB.Error(), http.StatusInternalServerError)
			return
		}
		apiResponse(w, http.StatusCreated, res)
	}
}

// ADMIN ONLY: UpdateProduct updates an existing product based on provided fields (name, size, price, quantity)
func (s *server) UpdateProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// grab id from url /product/{id}
		vars := mux.Vars(r)
		if vars == nil {
			http.Error(w, e.ErrInvalidID.Error(), http.StatusBadRequest)
		}
		id := vars["id"]

		// Get product to update
		updateProduct, err := s.db.Get(id)
		if err != nil {
			http.Error(w, e.ErrDB.Error(), http.StatusInternalServerError)
			return
		}
		if updateProduct == nil {
			http.Error(w, http.StatusText(404), http.StatusNotFound)
			return
		}

		// Decode body, use updateProduct object and unmarshal over to replace any fields that found in req body
		if err := json.NewDecoder(r.Body).Decode(&updateProduct); err != nil {
			http.Error(w, e.ErrInvalidArgJsonBody.Error(), http.StatusBadRequest)
			return
		}
		// prevent client from modifying id
		updateProduct.ID = id

		// Update product
		p, err := s.db.Update(updateProduct)
		if err != nil {
			http.Error(w, e.ErrDB.Error(), http.StatusInternalServerError)
			return
		}
		apiResponse(w, http.StatusOK, p)
	}
}

// ADMIN ONLY: DeleteProduct deletes an existing product using id
func (s *server) DeleteProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if vars == nil {
			http.Error(w, e.ErrInvalidID.Error(), http.StatusBadRequest)
			return
		}
		id := vars["id"]

		if err := s.db.Delete(id); err != nil {
			http.Error(w, e.ErrDB.Error(), http.StatusInternalServerError)
			return
		}
		apiResponse(w, http.StatusGone, http.StatusText(410))
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
