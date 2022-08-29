package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/akleventis/united_house_server/lib"
	"github.com/akleventis/united_house_server/uhp_db"
	log "github.com/sirupsen/logrus"
	stripe_product "github.com/stripe/stripe-go/product"
	stripe "github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"

	webhook "github.com/stripe/stripe-go/webhook"
)

type Handler struct {
	db *uhp_db.UhpDB
}

func NewHandler(db *uhp_db.UhpDB) *Handler {
	return &Handler{
		db: db,
	}
}

type lineItems struct {
	Items []*uhp_db.Product `json:"items"`
}

type checkoutResponse struct {
	URL     string          `json:"url"`
	Product *uhp_db.Product `json:"product"`
}

// fulfillOrder verifies items are in stock and returns the resulting products array
func (h *Handler) fulfillOrder(items lineItems) ([]*uhp_db.Product, error) {
	var products []*uhp_db.Product
	for _, v := range items.Items {
		p, err := h.db.GetOrder(v.ID, v.Quantity)
		if err != nil {
			if err == lib.ErrOutOfStock && p != nil {
				return []*uhp_db.Product{p}, lib.ErrOutOfStock
			}
			return nil, err
		}
		// get product image from Stripe
		pInfo, err := stripe_product.Get(p.ID, nil)
		if err != nil {
			return nil, lib.ErrStripeImage
		}
		if len(pInfo.Images) > 0 {
			p.ImageURL = pInfo.Images[0]
		}
		products = append(products, p)
	}
	return products, nil
}

// createLineItems converts products array to stripe LineItems
func createLineItems(products []*uhp_db.Product) []*stripe.CheckoutSessionLineItemParams {
	var cli []*stripe.CheckoutSessionLineItemParams
	for _, v := range products {
		item := &stripe.CheckoutSessionLineItemParams{
			Name:        stripe.String(v.Name + " " + v.Size),
			Description: stripe.String(string(v.ID)),
			Amount:      stripe.Int64(int64(v.Price * 100)),
			Currency:    stripe.String("usd"),
			Quantity:    stripe.Int64(int64(v.Quantity)),
			Images:      stripe.StringSlice([]string{v.ImageURL}),
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
	stripe.Key = os.Getenv("STRIPE_KEY")
	sesh, err := session.New(params)
	if err != nil {
		return nil, err
	}
	return sesh, nil
}

// HandleCheckout receives array of items from client and returns a stripe checkout redirect url
func (h *Handler) HandleCheckout() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var items lineItems

		err := json.NewDecoder(req.Body).Decode(&items)
		if err != nil {
			http.Error(w, lib.ErrInvalidArgJsonBody.Error(), http.StatusBadRequest)
			return
		}

		// grab/verify items are in stock
		products, err := h.fulfillOrder(items)
		if err != nil {
			if err == lib.ErrOutOfStock && len(products) > 0 {
				res := &checkoutResponse{
					Product: products[0],
				}
				// return out of stock product to client
				lib.ApiResponse(w, http.StatusAccepted, res)
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

		res := &checkoutResponse{
			URL: sesh.URL,
		}

		lib.ApiResponse(w, http.StatusOK, res)
	}
}

// WEBHOOK
// handleWebhook() listens for Checkout Session Confirmation then Update Inventory accordingly
func (h *Handler) HandleWebhook() http.HandlerFunc {
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
			log.Errorf("Error verifying webhook signature: %v\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Unmarshal the event data into an appropriate struct depending on its Type
		if event.Type == "checkout.session.completed" {
			// Grab Session Data
			var sesh stripe.CheckoutSession
			err = json.Unmarshal(event.Data.Raw, &sesh)
			if err != nil {
				log.Errorf("Error parsing webhook JSON: %v\n", err)
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

				if err := h.db.UpdateQuantity(id, quantity); err != nil {
					http.Error(w, lib.ErrDB.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
		log.Info("")
		w.WriteHeader(http.StatusOK)
	}
}
