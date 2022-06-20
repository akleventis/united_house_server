package merchdb

import (
	"database/sql"
	"fmt"

	e "github.com/akleventis/united_house_server/errors"
	log "github.com/sirupsen/logrus"
)

type Product struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Size     string  `json:"size"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

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
func (pDB *ProductDB) Get(id string) (*Product, error) {
	var p *Product
	query := `SELECT * from merch_t where id=$1 LIMIT 1;`
	if err := pDB.QueryRow(query).Scan(&p.ID, &p.Name, &p.Size, &p.Price, &p.Quantity); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, e.ErrDB
	}
	return p, nil
}

// Update will update a product using product_id
func (pDB *ProductDB) Update(p *Product) (*Product, error) {
	// string format price 2 decimal precision
	query := fmt.Sprintf(`UPDATE merch_t SET name=$1, size=$2, price=%.2f, quantity=$3 WHERE id=$4;`, p.Price)
	if _, err := pDB.Exec(query, p.Name, p.Size, p.Quantity, p.ID); err != nil {
		return nil, e.ErrDB
	}
	return p, nil
}

// Delete will remove a product using product_id
func (pDB *ProductDB) Delete(id string) error {
	query := `DELETE FROM merch_t WHERE id=$1;`
	if _, err := pDB.Exec(query, id); err != nil {
		return e.ErrDB
	}
	return nil
}

// Create will post a new product to merch_t
func (pDB *ProductDB) Create(p Product) (*Product, error) {
	query := `INSERT INTO merch_t (id, name, size, price, quantity) VALUES ($1, $2, $3, $4, $5, $6);`
	log.Info(query)
	if _, err := pDB.Exec(query, p.ID, p.Name, p.Size, p.Price, p.Quantity); err != nil {
		return nil, e.ErrDB
	}
	return &p, nil
}

// GetProducts returns an array of all products
func (pDB *ProductDB) GetProducts() ([]Product, error) {
	products := make([]Product, 0)

	query := `SELECT * from merch_t;`
	rows, err := pDB.Query(query)
	if err != nil {
		return nil, e.ErrDB
	}
	defer rows.Close()
	for rows.Next() {
		p := Product{}
		err := rows.Scan(&p.ID, &p.Name, &p.Size, &p.Price, &p.Quantity)
		if err != nil {
			return nil, e.ErrDB
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
func (pDB *ProductDB) GetOrder(id string, quantity int) (*Product, error) {
	var p Product

	query := `SELECT * FROM merch_t WHERE id=$1 LIMIT 1;`
	if err := pDB.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Size, &p.Price, &p.Quantity); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, e.ErrDB
	}
	// if req quantity > in stock return product and error so we can tell front end how much in-stock
	if p.Quantity < quantity {
		return &p, e.ErrOutOfStock
	}
	p.Quantity = quantity

	return &p, nil
}

// UpdateQuantity reduces quantity in database using product_id
func (pDB *ProductDB) UpdateQuantity(id string, quantity int) error {
	query := `UPDATE merch_t SET quantity=quantity-$1 WHERE id=$2;`
	if _, err := pDB.Exec(query, quantity, id); err != nil {
		return e.ErrDB
	}
	return nil
}
