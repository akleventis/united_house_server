package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
)

// create struct
type ProductDB struct {
	*sql.DB
}

func OpenDBConnection() (*ProductDB, error) {
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
	pDB := &ProductDB{
		db,
	}

	return pDB, nil
}

// CREATE TABLE IF NOT EXISTS merch_t (
// 	id VARCHAR( 50 ) PRIMARY KEY NOT NULL,
// 	name VARCHAR( 50 ) NOT NULL,
// 	size VARCHAR( 50 ) NOT NULL,
// 	price INT NOT NULL,
// 	quantity INT NOT NULL
//  )
