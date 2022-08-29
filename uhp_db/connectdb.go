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

// uhpdb=# \d merch_t
//  id       | character varying(50) |           | not null |
//  name     | character varying(50) |           | not null |
//  size     | character varying(50) |           | not null |
//  price    | numeric               |           | not null |
//  quantity | integer               |           | not null |
func (db *UhpDB) createMerchTable() error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS merch_t (
		id VARCHAR( 50 ) PRIMARY KEY NOT NULL,
		name VARCHAR( 50 ) NOT NULL,
		size VARCHAR( 50 ) NOT NULL,
		price NUMERIC NOT NULL,
		quantity INT NOT NULL
	)`)
	if err != nil {
		return err
	}

	return nil
}

// uhpdb=# \d events_t;
//  id            | integer               |           | not null | nextval('events_t_id_seq'::regclass)
//  headliner     | json                  |           | not null |
//  openers       | json                  |           |          |
//  image_url     | character varying(50) |           | not null |
//  location_name | character varying(50) |           | not null |
//  location_url  | character varying(50) |           | not null |
//  start_time    | character varying(50) |           |          |
//  end_time      | character varying(50) |           |          |
//  ticket_url    | character varying(50) |           |          |
func (db *UhpDB) createEventsTable() error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS events_t (
		id SERIAL PRIMARY KEY,
		headliner json NOT NULL,
		openers json,
		image_url VARCHAR( 50 ) NOT NULL,
		location_name VARCHAR( 50 ) NOT NULL,
		location_url VARCHAR( 50 ) NOT NULL,
		ticket_url VARCHAR( 50 ), 
		start_time TIMESTAMP without time zone,
		end_time TIMESTAMP without time zone
	)`)
	if err != nil {
		return err
	}
	return nil
}

// uhpdb=# \d featured_artists_t;
// Column         |         Type          | Collation | Nullable |                    Default
// -----------------------+-----------------------+-----------+----------+------------------------------------------------
//  id                    | integer               |           | not null | nextval('featured_artists_t_id_seq'::regclass)
//  name                  | character varying(50) |           | not null |
//  redirect_url          | character varying(50) |           |          |
//  soundcloud_iframe_url | character varying(50) |           | not null |
//  sequence              | integer               |           | not null |
func (db *UhpDB) createFeaturedSongsTable() error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS featured_artists_t (
		id SERIAL PRIMARY KEY,
		name VARCHAR( 50 ) NOT NULL,
		redirect_url VARCHAR( 50 ),
		soundcloud_iframe_url VARCHAR( 50 ) NOT NULL,
		sequence INT NOT NULL
	)`)
	if err != nil {
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

	pDB := &UhpDB{
		db,
	}

	err = pDB.createMerchTable()
	if err != nil {
		return nil, err
	}
	err = pDB.createEventsTable()
	if err != nil {
		return nil, err
	}
	err = pDB.createFeaturedSongsTable()
	if err != nil {
		return nil, err
	}

	return pDB, nil
}
