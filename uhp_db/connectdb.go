package uhp_db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type UhpDB struct {
	*sql.DB
}

// Column    | Type | Collation | Nullable | Default
// ----------+------+-----------+----------+---------
// username  | text |           | not null |
// password  | text |           |          |
// Indexes:
//   "auth_t_pkey" PRIMARY KEY, btree (username)
func (db *UhpDB) createAuthTable() error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS auth_t (
		username text PRIMARY KEY,
		password text
		)`); err != nil {
		return err
	}
	return nil
}

// Opens database connection, creates tables if not exists
func Open() (*UhpDB, error) {
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

	if err = db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(20)
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(time.Minute * 5)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	log.Printf("%s database connected", dbName)

	pDB := &UhpDB{
		db,
	}

	if err = pDB.createAuthTable(); err != nil {
		return nil, err
	}

	return pDB, nil
}
