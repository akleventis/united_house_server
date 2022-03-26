package db

import (
	"database/sql"
	"errors"
)

type Product struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Size     string `json:"size"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
}

var (
	ErrOutOfStock  = errors.New("OUT_OF_STOCK")
	ErrUpdateStock = errors.New("UPDATE_STOCK_ERROR")
)

// returns array of all products
func GetProducts(db *sql.DB) ([]*Product, error) {
	products := make([]*Product, 0)

	query := `SELECT * from merch;`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		product := Product{}
		err := rows.Scan(&product.ID, &product.Name, &product.Size, &product.Price, &product.Quantity)
		if err != nil {
			return nil, err
		}
		products = append(products, &product)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return products, nil
}

func GetProductById(db *sql.DB, id int, quantity int) (*Product, error) {
	var p Product

	query := `SELECT * FROM merch WHERE id=$1 LIMIT 1;`
	err := db.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Size, &p.Price, &p.Quantity)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return nil, nil
	}

	// if req quantity > in stock return product and error so we can tell front end how much in-stock
	if p.Quantity < quantity {
		return &p, ErrOutOfStock
	}

	p.Quantity = quantity

	return &p, nil
}

func UpdateQuantity(db *sql.DB, id int, quantity int) error {
	query := `UPDATE merch SET quantity = quantity - $1 WHERE id=$2;`
	if _, err := db.Exec(query, quantity, id); err != nil {
		return ErrUpdateStock
	}
	return nil
}
