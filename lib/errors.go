package lib

import "errors"

var (
	// db
	ErrOutOfStock = errors.New("OUT_OF_STOCK")
	ErrDB         = errors.New("DB_ERROR")

	// products
	ErrQuantity        = errors.New("QUANTITY_ERR")
	ErrFetchingProduct = errors.New("STRIPE_PRODUCT_ERR")

	// 400s
	ErrInvalidToken = errors.New("INVALID_TOKEN")
	ErrTokenExpired = errors.New("EXPIRED_TOKEN")
	ErrInvalidID    = errors.New("INVALID_ID")

	// 500s
	ErrInvalidArgJsonBody = errors.New("INVALID_JSON")

	// STRIPE
	ErrStripeImage = errors.New("STRIPE_IMAGE_ERROR")

	// Email
	ErrEmail = errors.New("EMAIL_ERROR")
)
