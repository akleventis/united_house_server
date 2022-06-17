package merchdb

import (
	"database/sql"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type Product struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Size     string  `json:"size"`
	Price    float64 `json:"price,omitempty"`
	Quantity int     `json:"quantity"`
}

var ErrOutOfStock = errors.New("OUT_OF_STOCK")
var ErrDB = errors.New("DB_ERROR")

// TODO: UPDATE ON OTHER DEVICES
// Desc merch_t
// Column  |         Type          | Collation | Nullable | Default
// ----------+-----------------------+-----------+----------+---------
// id       | character varying(50) |           | not null |
// name     | character varying(50) |           | not null |
// size     | character varying(50) |           | not null |
// price    | numeric	            |           | not null |
// quantity | integer               |           | not null |
// Indexes:
//   "merch_pkey" PRIMARY KEY, btree (id)

// Get returns a product using product_id
func (productDB *ProductDB) Get(id string) (*Product, error) {
	var p *Product

	query := `SELECT * from merch_t where id=$1 LIMIT 1;`
	err := productDB.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Size, &p.Price, &p.Quantity)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, ErrDB
		}
		return nil, nil
	}
	return p, nil
}

// Update will update a product using product_id
func (productDB *ProductDB) Update(p *Product) error {
	query := fmt.Sprintf(`UPDATE merch_t SET name='%s', size='%s', price=%.2f, quantity=%d WHERE id='%s';`, p.Name, p.Size, p.Price, p.Quantity, p.ID)
	log.Info(query)
	if _, err := productDB.Exec(query); err != nil {
		return err
	}
	return nil
}

// GetProducts returns an array of all products
func (productDB *ProductDB) GetProducts() ([]Product, error) {
	products := make([]Product, 0)

	query := `SELECT * from merch_t;`
	rows, err := productDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		p := Product{}
		err := rows.Scan(&p.ID, &p.Name, &p.Size, &p.Price, &p.Quantity)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return products, nil
}

// GetOrder retrieves a product by ID and verifies order can be fulfilled. Returns error if quantity cannot be fulfilled
func (productDB *ProductDB) GetOrder(id string, quantity int) (*Product, error) {
	var p Product

	query := `SELECT * FROM merch_t WHERE id=$1 LIMIT 1;`
	err := productDB.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Size, &p.Price, &p.Quantity)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, ErrDB
	}

	// if req quantity > in stock return product and error so we can tell front end how much in-stock
	if p.Quantity < quantity {
		return &p, ErrOutOfStock
	}
	p.Quantity = quantity

	return &p, nil
}

// UpdateQuantity reduces quantity in database using product_id
func (productDB *ProductDB) UpdateQuantity(id string, quantity int) error {
	query := fmt.Sprintf(`UPDATE merch_t SET quantity=quantity-%d WHERE id='%s';`, quantity, id)
	if _, err := productDB.Exec(query); err != nil {
		return err
	}
	return nil
}

func (productDB *ProductDB) GetProductById(id string) (*Product, error) {
	var p Product

	query := `SELECT * FROM merch_t WHERE id=$1 LIMIT 1;`
	err := productDB.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Size, &p.Price, &p.Quantity)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, ErrDB
		}
		return nil, nil
	}
	return &p, nil
}
