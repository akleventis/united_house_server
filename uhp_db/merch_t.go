package uhp_db

import (
	"database/sql"
	"fmt"

	lib "github.com/akleventis/united_house_server/lib"
	log "github.com/sirupsen/logrus"
)

type Product struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Size     string  `json:"size"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	ImageURL string  `json:"image_url,omitempty"`
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
func (uDB *UhpDB) GetProduct(id string) (*Product, error) {
	var p Product
	query := `SELECT * from merch_t where id=$1 LIMIT 1;`
	if err := uDB.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Size, &p.Price, &p.Quantity); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error(err)
		return nil, lib.ErrDB
	}
	return &p, nil
}

// Update will update a product using product_id
func (uDB *UhpDB) UpdateProduct(p *Product) (*Product, error) {
	// string format price 2 decimal precision
	query := fmt.Sprintf(`UPDATE merch_t SET name=$1, size=$2, price=%.2f, quantity=$3 WHERE id=$4;`, p.Price)
	if _, err := uDB.Exec(query, p.Name, p.Size, p.Quantity, p.ID); err != nil {
		log.Error(err)
		return nil, lib.ErrDB
	}
	return p, nil
}

// Delete will remove a product using product_id
func (uDB *UhpDB) DeleteProduct(id string) error {
	query := `DELETE FROM merch_t WHERE id=$1;`
	if _, err := uDB.Exec(query, id); err != nil {
		log.Error(err)
		return lib.ErrDB
	}
	return nil
}

// Create will post a new product to merch_t
func (uDB *UhpDB) CreateProduct(p Product) (*Product, error) {
	query := `INSERT INTO merch_t (id, name, size, price, quantity) VALUES ($1, $2, $3, $4, $5, $6);`
	if _, err := uDB.Exec(query, p.ID, p.Name, p.Size, p.Price, p.Quantity); err != nil {
		log.Error(err)
		return nil, lib.ErrDB
	}
	return &p, nil
}

// GetProducts returns an array of all products
func (uDB *UhpDB) GetProducts() ([]Product, error) {
	products := make([]Product, 0)

	query := `SELECT * FROM merch_t;`
	rows, err := uDB.Query(query)
	if err != nil {
		log.Error(err)
		return nil, lib.ErrDB
	}
	defer rows.Close()
	for rows.Next() {
		p := Product{}
		err := rows.Scan(&p.ID, &p.Name, &p.Size, &p.Price, &p.Quantity)
		if err != nil {
			log.Error(err)
			return nil, lib.ErrDB
		}

		products = append(products, p)
	}
	err = rows.Err()
	if err != nil {
		log.Error(err)
		return nil, lib.ErrDB
	}

	return products, nil
}

// GetOrder retrieves a product by ID and verifies order can be fulfilled. Returns error if quantity cannot be fulfilled
func (uDB *UhpDB) GetOrder(id string, quantity int) (*Product, error) {
	var p Product
	query := `SELECT * FROM merch_t WHERE id=$1 LIMIT 1;`
	if err := uDB.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Size, &p.Price, &p.Quantity); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error(err)
		return nil, lib.ErrDB
	}
	// if req quantity > in stock return product and error so we can tell front end how much in-stock
	if p.Quantity < quantity {
		return &p, lib.ErrOutOfStock
	}
	p.Quantity = quantity

	return &p, nil
}

// UpdateQuantity reduces quantity in database using product_id
func (uDB *UhpDB) UpdateQuantity(id string, quantity int) error {
	query := `UPDATE merch_t SET quantity=quantity-$1 WHERE id=$2;`
	if _, err := uDB.Exec(query, quantity, id); err != nil {
		log.Error(err)
		return lib.ErrDB
	}
	return nil
}
