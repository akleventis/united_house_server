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
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS merch_t (
		id VARCHAR( 50 ) PRIMARY KEY NOT NULL,
		name VARCHAR( 50 ) NOT NULL,
		size VARCHAR( 50 ) NOT NULL,
		price NUMERIC NOT NULL,
		quantity INT NOT NULL
		)`); err != nil {
		return err
	}
	return nil
}

// uhpdb=# \d artists_t;
// Column |         Type          | Collation | Nullable |                Default
// --------+-----------------------+-----------+----------+---------------------------------------
//  id     | integer               |           | not null | nextval('artists_t_id_seq'::regclass)
//  name   | character varying(50) |           | not null |
//  url    | character varying(50) |           |          |
// Indexes:
//     "artists_t_pkey" PRIMARY KEY, btree (id)
// Referenced by:
//     TABLE "events_t" CONSTRAINT "events_t_headliner_id_fkey" FOREIGN KEY (headliner_id) REFERENCES artists_t(id)
//     TABLE "featured_artists_t" CONSTRAINT "featured_artists_t_artist_id_fkey" FOREIGN KEY (artist_id) REFERENCES artists_t(id)
func (db *UhpDB) createArtistsTable() error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS artists_t (
			id SERIAL PRIMARY KEY,
			name VARCHAR( 50 ) NOT NULL,
			url VARCHAR( 50 )
		)`); err != nil {
		return err
	}
	return nil
}

// uhpdb=# \d featured_artists_t;
// Column         |         Type          | Collation | Nullable |                    Default
// -----------------------+-----------------------+-----------+----------+------------------------------------------------
//  id                    | integer               |           | not null | nextval('featured_artists_t_id_seq'::regclass)
//  artist                | json                  |           | not null |
//  soundcloud_iframe_url | character varying(50) |           | not null |
//  sequence              | integer               |           | not null |
// Indexes:
//     "featured_artists_t_pkey" PRIMARY KEY, btree (id)
func (db *UhpDB) createFeaturedArtistsTable() error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS featured_artists_t (
		id SERIAL PRIMARY KEY,
		artist json NOT NULL,
		soundcloud_iframe_url VARCHAR( 50 ) NOT NULL,
		sequence INT NOT NULL
		)`); err != nil {
		return err
	}
	return nil
}

// uhpdb=# \d events_t;
// Column     |            Type             | Collation | Nullable |               Default
// ---------------+-----------------------------+-----------+----------+--------------------------------------
//  id            | integer                     |           | not null | nextval('events_t_id_seq'::regclass)
//  headliner_id  | json                        |           | not null |
//  openers       | json                        |           |          |
//  image_url     | character varying(50)       |           | not null |
//  location_name | character varying(50)       |           | not null |
//  location_url  | character varying(50)       |           | not null |
//  ticket_url    | character varying(50)       |           |          |
//  start_time    | timestamp without time zone |           |          |
//  end_time      | timestamp without time zone |           |          |
// Indexes:
//     "events_t_pkey" PRIMARY KEY, btree (id)
func (db *UhpDB) createEventsTable() error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS events_t (
		id SERIAL PRIMARY KEY,
		headliner_id json NOT NULL,
		openers json,
		image_url VARCHAR( 50 ) NOT NULL,
		location_name VARCHAR( 50 ) NOT NULL,
		location_url VARCHAR( 50 ) NOT NULL,
		ticket_url VARCHAR( 50 ), 
		start_time TIMESTAMP without time zone,
		end_time TIMESTAMP without time zone
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

	if err = pDB.createMerchTable(); err != nil {
		return nil, err
	}
	if err = pDB.createArtistsTable(); err != nil {
		return nil, err
	}
	if err = pDB.createFeaturedArtistsTable(); err != nil {
		return nil, err
	}
	if err = pDB.createEventsTable(); err != nil {
		return nil, err
	}

	return pDB, nil
}
