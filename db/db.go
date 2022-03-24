package db

import "database/sql"

type Product struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Size     string `json:"size"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
}

// returns array of all products
func GetProducts(db *sql.DB) ([]*Product, error) {
	products := make([]*Product, 0)

	query := "SELECT * from merch;"
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
