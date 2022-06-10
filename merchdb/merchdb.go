package merchdb

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type Datastore interface {
	GetProducts() ([]*Product, error)
	GetProductOrder(id string, quantity int) (*Product, error)
	UpdateQuantity(id string, quantity int) error
	GetProductById(id string) (*Product, error)
	Update(p *Product) error
}

// create struct
type ProductDB struct {
	*sql.DB
}

func (db *ProductDB) createMerchTable() error {
	// create table
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS merch_t (
		id VARCHAR( 50 ) PRIMARY KEY NOT NULL,
		name VARCHAR( 50 ) NOT NULL,
		size VARCHAR( 50 ) NOT NULL,
		price INT NOT NULL,
		quantity INT NOT NULL)`)
	if err != nil {
		return err
	}

	return nil
}

func Open() (*ProductDB, error) {
	var (
		dbHost = os.Getenv("DB_HOST")
		dbPort = os.Getenv("DB_PORT")
		dbUser = os.Getenv("DB_USER")
		dbName = os.Getenv("DB_NAME")
	)

	dbPortInt, err := strconv.Atoi(dbPort)
	if err != nil {
		return nil, err
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"dbname=%s sslmode=disable", dbHost, dbPortInt, dbUser, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(20)
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(time.Minute * 5)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	log.Printf("%s database connected", dbName)

	pDB := &ProductDB{
		db,
	}

	err = pDB.createMerchTable()
	if err != nil {
		return nil, err
	}

	return pDB, nil
}
