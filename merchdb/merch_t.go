package merchdb

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

var ErrOutOfStock = errors.New("OUT_OF_STOCK")
var ErrDB = errors.New("DB_ERROR")

// Desc merch_t
// Column  |         Type          | Collation | Nullable | Default
// ----------+-----------------------+-----------+----------+---------
// id       | character varying(50) |           | not null |
// name     | character varying(50) |           | not null |
// size     | character varying(50) |           | not null |
// price    | integer               |           | not null |
// quantity | integer               |           | not null |
// Indexes:
//   "merch_pkey" PRIMARY KEY, btree (id)

// GetProducts returns an array of all products
func (productDB *ProductDB) GetProducts() ([]*Product, error) {
	products := make([]*Product, 0)

	query := `SELECT * from merch_t;`
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

// GetProductByID retrieves a product by ID and requested quanitity. Returns error if quantity can not be fulfilled
func (productDB *ProductDB) GetProductByID(id string, quantity int) (*Product, error) {
	var p Product

	query := `SELECT * FROM merch_t WHERE id=$1 LIMIT 1;`
	err := productDB.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Size, &p.Price, &p.Quantity)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, ErrDB
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
	query := fmt.Sprintf(`UPDATE merch_t SET quantity=quantity-%d WHERE id='%s';`, quantity, id)
	if _, err := productDB.Exec(query); err != nil {
		return err
	}
	return nil
}
