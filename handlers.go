package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	uhp_db "github.com/akleventis/united_house_server/db"
	log "github.com/sirupsen/logrus"
	stripe "github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	webhook "github.com/stripe/stripe-go/webhook"
)

type lineItems struct {
	Items []*uhp_db.Product `json:"items"`
}

// createLineItems verifies items are in stock and returns a product array
func (s *server) createLineItems(items lineItems) ([]*uhp_db.Product, error) {
	var products []*uhp_db.Product
	for _, v := range items.Items {
		product, err := s.db.GetProductByID(v.ID, v.Quantity)
		if err != nil {
			if err == uhp_db.ErrOutOfStock {
				var message string
				switch product.Quantity {
				case 0:
					message = fmt.Sprintf("%s %s is out of stock. Please update cart", product.Size, product.Name)
				default:
					message = fmt.Sprintf("Only %d %s %s(s), in stock. Please update cart", product.Quantity, product.Size, product.Name)
				}
				return nil, errors.New(message)
			}
			return nil, uhp_db.ErrDB
		}
		products = append(products, product)
	}
	return products, nil
}

// createLineItems converts product array to stripe LineItems
func createLineItems(products []*uhp_db.Product) []*stripe.CheckoutSessionLineItemParams {
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

// handleCheckout receives array of items from client, verifies items are in stock and grabs from database, creates checkout session and returns stripe redirect url
func (s *server) handleCheckout() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var items lineItems

		err := json.NewDecoder(req.Body).Decode(&items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		products, err := s.createLineItems(items)
		if err != nil {
			switch err {
			case uhp_db.ErrDB:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			default:
				http.Error(w, fmt.Sprint(err), http.StatusBadRequest)
			}
			return
		}

		checkoutLineItems := createLineItems(products)

		sesh, err := createCheckoutSession(checkoutLineItems)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send redirect URL back to client
		jsonResp, err := json.Marshal(map[string]string{"url": sesh.URL})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResp)
	}
}

// getProducts returns json array of all products
func (s *server) getProducts() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}
		products, err := s.db.GetProducts()
		if err != nil {
			log.Fatal(err)
		}

		jsonResp, err := json.Marshal(products)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(jsonResp)
	}
}

// updateInventory is a helper function for reducing inventory upon successful checkout session
func (s *server) updateInventory(productID string, quantity int) error {
	if err := s.db.UpdateQuantity(productID, quantity); err != nil {
		return err
	}
	return nil
}

// handleWebhook() listens for Checkout Session Confirmation then Update Inventory accordingly
func (s *server) handleWebhook() http.HandlerFunc {
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
		endpointSecret := "whsec_997481b842ea0e014921893e6a5767e23bd7c32b18e2a424db9046a99268adb3"
		// Pass the request body and Stripe-Signature header to ConstructEvent, along with the webhook signing key.
		event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"),
			endpointSecret)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest) // Return a 400 error on a bad signature
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

				err = s.updateInventory(id, quantity)
				if err != nil {
					fmt.Fprintf(os.Stderr, "UPDATE INVENTORY ERROR: %v\n", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}
