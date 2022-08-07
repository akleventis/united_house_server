package uhp_db

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"

	e "github.com/akleventis/united_house_server/errors"
	log "github.com/sirupsen/logrus"
)

type Headliner struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Openers []*DJ

type DJ struct {
	Order string
	Name  string
	Url   string
}

type Event struct {
	ID           int          `json:"id"`
	Headliner    Headliner    `json:"headliner"`
	Openers      Openers      `json:"openers"`
	ImageURL     string       `json:"image_url"`
	LocationName string       `json:"location_name"`
	LocationURL  string       `json:"location_url"`
	TicketURL    string       `json:"ticket_url"`
	StartTime    sql.NullTime `json:"start_time"`
	EndTime      sql.NullTime `json:"end_time"`
}

// headliner json
// {
// 	"name": "Beebo",
// 	"url": "https://instagram.com/BEEBO_MUSIC/"
// }

// openers json
// [{
//		"order": "1",
// 		"name": "Beebo",
// 		"url": "https://instagram.com/BEEBO_MUSIC/"
// 	},
// 	{
// 		etc..
// 	}]

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
		return nil, e.ErrDB
	}
	defer rows.Close()
	for rows.Next() {
		event := Event{}
		err := rows.Scan(&event.ID, &event.Headliner, &event.Openers, &event.ImageURL, &event.LocationName, &event.LocationURL, &event.TicketURL, &event.StartTime, &event.EndTime)
		if err != nil {
			log.Info(err)
			return nil, e.ErrDB
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (uDB *UhpDB) CreateEvent(event Event) (*Event, error) {
	query := `INSERT INTO events_t (headliner, openers, image_url, location_name, location_url, ticket_url, start_time, end_time, date) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`
	if _, err := uDB.Exec(query, event.Headliner, event.Openers, event.ImageURL, event.LocationName, event.LocationURL, event.TicketURL, event.StartTime, event.EndTime); err != nil {
		return nil, e.ErrDB
	}
	return &event, nil
}

func (uDB *UhpDB) UpdateEvent(event *Event) (*Event, error) {
	query := `UPDATE events_t SET headliner=$1, openers=$2, image_url=$3, location_name=$4, location_url=$5, ticket_url=$6, start_time=$7, end_time=$8;`
	if _, err := uDB.Exec(query, event.Headliner, event.Openers, event.ImageURL, event.LocationName, event.LocationURL, event.TicketURL, event.StartTime, event.EndTime); err != nil {
		return nil, e.ErrDB
	}
	return event, nil
}

func (uDB *UhpDB) DeleteEvent(id string) error {
	query := `DELETE FROM events_t WHERE id=$1;`
	if _, err := uDB.Exec(query, id); err != nil {
		return err
	}
	return nil
}

func (uDB *UhpDB) GetEvent(id string) (*Event, error) {
	var event Event
	query := `SELECT * FROM events_t WHERE id=$1 LIMIT 1;`

	if err := uDB.QueryRow(query, id).Scan(&event.ID, &event.Headliner, &event.Openers, &event.ImageURL, &event.LocationName, &event.LocationURL, &event.TicketURL, &event.StartTime, &event.EndTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Info(err)
		return nil, e.ErrDB
	}

	return &event, nil
}

// Headliner struct implements the driver.Value interface (https://pkg.go.dev/database/sql/driver#Valuer)
func (h Headliner) Value() (driver.Value, error) {
	return json.Marshal(h)
}

func (a *Headliner) Scan(v interface{}) error {
	b, ok := v.([]byte)
	if len(b) == 0 {
		return nil
	}
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

// Openers struct implements the driver.Value interface (https://pkg.go.dev/database/sql/driver#Valuer)
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
