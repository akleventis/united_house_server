package uhp_db

import (
	"database/sql"

	"github.com/akleventis/united_house_server/lib"
	log "github.com/sirupsen/logrus"
)

type Artist struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

// Column |         Type          | Collation | Nullable |                Default
// --------+-----------------------+-----------+----------+---------------------------------------
//  id     | integer               |           | not null | nextval('artists_t_id_seq'::regclass)
//  name   | character varying(50) |           | not null |
//  url    | character varying(50) |           |          |
// Indexes:
//     "artists_t_pkey" PRIMARY KEY, btree (id)

func (uDB *UhpDB) GetArtists() ([]Artist, error) {
	artists := make([]Artist, 0)

	query := `SELECT * FROM artists_t;`
	rows, err := uDB.Query(query)
	if err != nil {
		return nil, lib.ErrDB
	}
	defer rows.Close()
	for rows.Next() {
		artist := Artist{}
		if err := rows.Scan(&artist.ID, &artist.Name, &artist.Url); err != nil {
			log.Error(err)
			return nil, lib.ErrDB
		}
		artists = append(artists, artist)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return artists, nil
}

func (uDB *UhpDB) GetArtist(id string) (*Artist, error) {
	var artist Artist
	query := `SELECT * FROM artists_t WHERE id=$1 LIMIT 1;`

	if err := uDB.QueryRow(query, id).Scan(&artist.ID, &artist.Name, &artist.Url); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, lib.ErrDB
	}
	return &artist, nil
}

func (uDB *UhpDB) CreateArtist(artist Artist) (*Artist, error) {
	query := `INSERT INTO artists_t (name, url) VALUES ($1, $2);`
	if _, err := uDB.Exec(query, artist.Name, artist.Url); err != nil {
		return nil, lib.ErrDB
	}
	// Grab auto-generated id
	// query = `SELECT id FROM artists_t WHERE name=$1 and url=$2 LIMIT 1`
	// if err := uDB.QueryRow(query, artist.Name, artist.Url).Scan(&artist.ID); err != nil {
	// 	return nil, lib.ErrDB
	// }
	return &artist, nil
}

func (uDB *UhpDB) UpdateArtist(artist *Artist) (*Artist, error) {
	query := `UPDATE artists_t SET name=$1, url=$2 WHERE id=$3`
	if _, err := uDB.Exec(query, artist.Name, artist.Url, artist.ID); err != nil {
		return nil, lib.ErrDB
	}
	return artist, nil
}

func (uDB *UhpDB) DeleteArtist(id string) error {
	query := `DELETE FROM artists_t WHERE id=$1;`
	if _, err := uDB.Exec(query, id); err != nil {
		return lib.ErrDB
	}
	return nil
}
