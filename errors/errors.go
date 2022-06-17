package errors

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
)
