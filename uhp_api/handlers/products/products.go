package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/akleventis/united_house_server/lib"
	"github.com/akleventis/united_house_server/uhp_db"
	"github.com/gorilla/mux"
	stripe_product "github.com/stripe/stripe-go/product"
)

type Handler struct {
	db *uhp_db.UhpDB
}

func NewHandler(db *uhp_db.UhpDB) *Handler {
	return &Handler{
		db: db,
	}
}

// GetProducts returns json array of all products
func (h *Handler) GetProducts() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		products, err := h.db.GetProducts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// get product images from Stripe
		for i, p := range products {
			pInfo, err := stripe_product.Get(p.ID, nil)
			if err != nil {
				http.Error(w, lib.ErrStripeImage.Error(), http.StatusInternalServerError)
				return
			}
			if len(pInfo.Images) > 0 {
				products[i].ImageURL = pInfo.Images[0]
			}
		}
		lib.ApiResponse(w, http.StatusOK, products)
	}
}

// GetProduct retrieves a product using and id
func (h *Handler) GetProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		if vars == nil {
			http.Error(w, lib.ErrInvalidID.Error(), http.StatusBadRequest)
		}
		id := vars["id"]

		p, err := h.db.GetProduct(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if p == nil {
			http.Error(w, http.StatusText(404), http.StatusNotFound)
			return
		}

		// get product image from Stripe
		pInfo, err := stripe_product.Get(p.ID, nil)
		if err != nil {
			http.Error(w, lib.ErrStripeImage.Error(), http.StatusInternalServerError)
			return
		}
		if len(pInfo.Images) > 0 {
			p.ImageURL = pInfo.Images[0]
		}

		lib.ApiResponse(w, http.StatusOK, p)
	}
}

// CreateProduct creates a product based on provided fields (id, name, size, price, quantity)
func (h *Handler) CreateProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p uhp_db.Product
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, lib.ErrInvalidArgJsonBody.Error(), http.StatusBadRequest)
			return
		}
		res, err := h.db.CreateProduct(p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lib.ApiResponse(w, http.StatusCreated, res)
	}
}

// UpdateProduct updates an existing product based on provided fields (name, size, price, quantity)
func (h *Handler) UpdateProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// grab id from url
		vars := mux.Vars(r)
		if vars == nil {
			http.Error(w, lib.ErrInvalidID.Error(), http.StatusBadRequest)
		}
		id := vars["id"]

		// Get product to update
		updateProduct, err := h.db.GetProduct(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if updateProduct == nil {
			http.Error(w, http.StatusText(404), http.StatusNotFound)
			return
		}

		// Decode body, use updateProduct object and unmarshal over to replace any fields that found in req body
		if err := json.NewDecoder(r.Body).Decode(&updateProduct); err != nil {
			http.Error(w, lib.ErrInvalidArgJsonBody.Error(), http.StatusBadRequest)
			return
		}
		// prevent client from modifying id
		updateProduct.ID = id

		// Update product
		p, err := h.db.UpdateProduct(updateProduct)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lib.ApiResponse(w, http.StatusOK, p)
	}
}

// DeleteProduct deletes an existing product using id
func (h *Handler) DeleteProduct() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if vars == nil {
			http.Error(w, lib.ErrInvalidID.Error(), http.StatusBadRequest)
			return
		}
		id := vars["id"]

		if err := h.db.DeleteProduct(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lib.ApiResponse(w, http.StatusGone, "Gone")
	}
}
