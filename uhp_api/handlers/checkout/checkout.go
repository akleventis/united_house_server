package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/akleventis/united_house_server/lib"
	"github.com/stripe/stripe-go"
	stripev73 "github.com/stripe/stripe-go/v73"
	"github.com/stripe/stripe-go/v73/checkout/session"
	"github.com/stripe/stripe-go/v73/price"
	"github.com/stripe/stripe-go/v73/product"
)

type Handler struct {
	clientURL string
}

func NewHandler(clientURL string) *Handler {
	return &Handler{
		clientURL: clientURL,
	}
}

type ProductV2 struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Size     string `json:"size"`
	ImageURL string `json:"image_url"`
	Price    int64  `json:"price"`
	Quantity int    `json:"quantity"`
	PriceID  string
}

type lineItemsV2 struct {
	Products []*ProductV2 `json:"items"`
}

type checkoutResponse struct {
	URL     string     `json:"url"`
	Product *ProductV2 `json:"product"`
}

// GetProducts returns json array of all products
func (h *Handler) GetProducts() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var productInfo []*ProductV2

		params := &stripev73.PriceListParams{}

		// expand on products
		params.AddExpand("data.product")

		i := price.List(params)
		for i.Next() {
			p := i.Price()

			// do not append archived products
			if !p.Product.Active {
				continue
			}

			quantity, err := strconv.Atoi(p.Product.Metadata["quantity"])
			if err != nil {
				http.Error(w, lib.ErrQuantity.Error(), http.StatusInternalServerError)
				return
			}

			// grab only necessary values
			product := &ProductV2{
				ID:       p.Product.ID,
				Name:     p.Product.Name,
				Size:     p.Product.Description,
				ImageURL: p.Product.Images[0],
				Price:    p.UnitAmount,
				Quantity: quantity,
			}
			productInfo = append(productInfo, product)
		}

		// returns price with expanded product //
		lib.ApiResponse(w, http.StatusOK, productInfo)
	}
}

// fulfillOrder verifies items are in stock and returns the resulting products array
func (h *Handler) fulfillOrder(items lineItemsV2) ([]*ProductV2, error) {
	var products []*ProductV2

	params := &stripev73.PriceSearchParams{}

	for _, v := range items.Products {
		p := &ProductV2{}

		// get product from stripe
		product, err := product.Get(v.ID, nil)
		if err != nil {
			return nil, err
		}

		quantity, err := strconv.Atoi(product.Metadata["quantity"])
		if err != nil {
			return nil, err
		}

		if quantity < v.Quantity {
			return []*ProductV2{{Name: v.Name, Quantity: quantity}}, lib.ErrOutOfStock
		}

		// get price ID of product
		params.Query = *stripe.String(fmt.Sprintf("product:'%s'", product.ID))
		iter := price.Search(params)
		for iter.Next() {
			result := iter.Current()
			priceData, ok := result.(*stripev73.Price)
			if !ok {
				return nil, lib.ErrFetchingProduct
			}
			p.PriceID = priceData.ID
		}

		p.Quantity = v.Quantity

		products = append(products, p)
	}

	return products, nil
}

// createLineItems converts products array to stripe LineItems
func createLineItems(products []*ProductV2) []*stripev73.CheckoutSessionLineItemParams {
	var cli []*stripev73.CheckoutSessionLineItemParams
	for _, v := range products {
		item := &stripev73.CheckoutSessionLineItemParams{
			Price:    stripev73.String(v.PriceID),
			Quantity: stripev73.Int64(int64(v.Quantity)),
		}
		cli = append(cli, item)
	}
	return cli
}

// createCheckoutSession creates and returns a stripe checkout session object
func (h *Handler) createCheckoutSession(cli []*stripev73.CheckoutSessionLineItemParams) (*stripev73.CheckoutSession, error) {
	params := &stripev73.CheckoutSessionParams{
		Mode:               stripev73.String(string(stripe.CheckoutSessionModePayment)),
		PaymentMethodTypes: []*string{stripev73.String("card")},
		SuccessURL:         stripev73.String(h.clientURL),
		CancelURL:          stripev73.String(h.clientURL),
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
		var items lineItemsV2

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
		sesh, err := h.createCheckoutSession(checkoutLineItems)
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
