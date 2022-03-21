package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
)

var (
	dbHost = os.Getenv("DB_HOST")
	dbPort = os.Getenv("DB_PORT")
	dbUser = os.Getenv("DB_USER")
	dbName = os.Getenv("DB_NAME")
)

func OpenDBConnection() (*sql.DB, error) {
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
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
