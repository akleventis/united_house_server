package db

import (
	"database/sql"
	"errors"
	"fmt"
)

type Product struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Size     string `json:"size"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
}

// TODO: Global error variables?
var ErrOutOfStock = errors.New("OUT_OF_STOCK")

// TODO: An interface to call these methods on??

// Returns an array of all products
func (productDB *ProductDB) GetProducts() ([]*Product, error) {
	products := make([]*Product, 0)

	query := `SELECT * from merch;`
	rows, err := productDB.Query(query)
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

// GetProductByID retrieves a product by ID and requested quanitity
// Returns product & error if requested quantity > amount in stock
func (productDB *ProductDB) GetProductByID(id string, quantity int) (*Product, error) {
	var p Product

	query := `SELECT * FROM merch WHERE id=$1 LIMIT 1;`
	err := productDB.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Size, &p.Price, &p.Quantity)
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

// UpdateQuantity reduces quantity in database using productID (primary key)
func (productDB *ProductDB) UpdateQuantity(id string, quantity int) error {
	query := fmt.Sprintf(`UPDATE merch SET quantity=quantity-%d WHERE id='%s';`, quantity, id)
	if _, err := productDB.Exec(query); err != nil {
		return err
	}
	return nil
}
