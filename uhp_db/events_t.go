package uhp_db

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/akleventis/united_house_server/lib"
	log "github.com/sirupsen/logrus"
)

type Openers []*Artist

type Artist struct {
	Name     string `json:"name"`
	Url      string `json:"url"`
	Sequence int    `json:"sequence"`
}

type Event struct {
	ID           int       `json:"id,omitempty"`
	Headliner    Artist    `json:"headliner"`
	Openers      Openers   `json:"openers"`
	ImageURL     string    `json:"image_url"`
	LocationName string    `json:"location_name"`
	LocationURL  string    `json:"location_url"`
	TicketURL    string    `json:"ticket_url"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
}

// Desc events_t
// Column       |            Type             | Collation | Nullable |               Default
// --------------------+-----------------------------+-----------+----------+--------------------------------------
//  id                 | integer                     |           | not null | nextval('events_t_id_seq'::regclass)
//  headliner          | json                        |           | not null |
//  openers            | json                        |           |          |
//  image_url          | character varying(50)       |           | not null |
//  location_name      | character varying(50)       |           | not null |
//  location_url       | character varying(50)       |           | not null |
//  ticket_url         | character varying(50)       |           |          |
//  start_time         | timestamp without time zone |           |          |
//  end_time           | timestamp without time zone |           |          |
// Indexes:
//     "events_t_pkey" PRIMARY KEY, btree (id)

func (uDB *UhpDB) GetEvents() ([]Event, error) {
	events := make([]Event, 0)

	query := `SELECT * FROM events_t;`
	rows, err := uDB.Query(query)
	if err != nil {
		return nil, lib.ErrDB
	}
	defer rows.Close()
	for rows.Next() {
		event := Event{}
		err := rows.Scan(&event.ID, &event.Headliner, &event.Openers, &event.ImageURL, &event.LocationName, &event.LocationURL, &event.TicketURL, &event.StartTime, &event.EndTime)
		if err != nil {
			log.Info(err)
			return nil, lib.ErrDB
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, lib.ErrDB
	}

	return events, nil
}

func (uDB *UhpDB) GetEvent(id string) (*Event, error) {
	var event Event
	query := `SELECT * FROM events_t WHERE id=$1 LIMIT 1;`

	if err := uDB.QueryRow(query, id).Scan(&event.ID, &event.Headliner, &event.Openers, &event.ImageURL, &event.LocationName, &event.LocationURL, &event.TicketURL, &event.StartTime, &event.EndTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, lib.ErrDB
	}

	return &event, nil
}

func (uDB *UhpDB) CreateEvent(event Event) (*Event, error) {
	query := `INSERT INTO events_t (headliner, openers, image_url, location_name, location_url, ticket_url, start_time, end_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`
	if _, err := uDB.Exec(query, event.Headliner, event.Openers, event.ImageURL, event.LocationName, event.LocationURL, event.TicketURL, event.StartTime, event.EndTime); err != nil {
		return nil, lib.ErrDB
	}
	// Grab auto-generated id
	// query = `SELECT id FROM events_t WHERE image_url=$1 and location_name=$2 and location_url=$3 and ticket_url=$4 and start_time=$5 and end_time=$6 LIMIT 1;`
	// if err := uDB.QueryRow(query, event.ImageURL, event.LocationName, event.LocationURL, event.TicketURL, event.StartTime, event.EndTime).Scan(&event.ID); err != nil {
	// 	return nil, lib.ErrDB
	// }
	return &event, nil
}

func (uDB *UhpDB) UpdateEvent(event *Event) (*Event, error) {
	query := `UPDATE events_t SET headliner=$1, openers=$2, image_url=$3, location_name=$4, location_url=$5, ticket_url=$6, start_time=$7, end_time=$8 WHERE id=$9;`
	if _, err := uDB.Exec(query, event.Headliner, event.Openers, event.ImageURL, event.LocationName, event.LocationURL, event.TicketURL, event.StartTime, event.EndTime, event.ID); err != nil {
		return nil, lib.ErrDB
	}
	return event, nil
}

func (uDB *UhpDB) DeleteEvent(id string) error {
	query := `DELETE FROM events_t WHERE id=$1;`
	if _, err := uDB.Exec(query, id); err != nil {
		return lib.ErrDB
	}
	return nil
}

// Artist struct implement driver.Value interface (https://pkg.go.dev/database/sql/driver#Valuer)
func (a Artist) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *Artist) Scan(v interface{}) error {
	b, ok := v.([]byte)
	if len(b) == 0 {
		return nil
	}
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

// Openers struct implement driver.Value interface (https://pkg.go.dev/database/sql/driver#Valuer)
func (o Openers) Value() (driver.Value, error) {
	return json.Marshal(o)
}

func (o *Openers) Scan(v interface{}) error {
	b, ok := v.([]byte)
	if len(b) == 0 {
		return nil
	}
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &o)
}
