package lib

import "errors"

var (
	// db
	ErrOutOfStock = errors.New("OUT_OF_STOCK")
	ErrDB         = errors.New("DB_ERROR")

	// 400s
	ErrInvalidTokenFormat = errors.New("INVALID_TOKEN_FORMAT")
	ErrInvalidToken       = errors.New("INVALID_TOKEN")
	ErrInvalidID          = errors.New("INVALID_ID")

	// 500s
	ErrInvalidArgJsonBody = errors.New("INVALID_JSON")

	// STRIPE
	ErrStripeImage = errors.New("STRIPE_IMAGE_ERROR")

	// Email
	ErrEmail = errors.New("EMAIL_ERROR")

	// IMAGES
	ErrImageFile     = errors.New("IMAGE_FILE_ERROR")
	ErrImageTooLarge = errors.New("IMAGE_MUST_BE_LESS_THAN_3_MEGABYTES")
	ErrFormValue     = errors.New("INVALID_FORM_VALUE")
	ErrFileType      = errors.New("FILE_TYPE_NOT_ALLOWED")
	ErrNoImage       = errors.New("IMAGE_NOT_FOUND")
)
