package uhp_db

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/akleventis/united_house_server/lib"
)

type FeaturedArtist struct {
	ID            int    `json:"id,omitempty"`
	Artist        Artist `json:"artist"`
	SoundcloudURL string `json:"soundcloud_iframe_url"`
	Sequence      int    `json:"sequence"`
}

// Column         |         Type          | Collation | Nullable |                    Default
// -----------------------+-----------------------+-----------+----------+------------------------------------------------
//  id                    | integer               |           | not null | nextval('featured_artists_t_id_seq'::regclass)
//  artist                | json                  |           | not null |
//  soundcloud_iframe_url | character varying(50) |           | not null |
//  sequence              | integer               |           | not null |
// Indexes:
//     "featured_artists_t_pkey" PRIMARY KEY, btree (id)

func (uDB *UhpDB) GetFeaturedArtists() ([]FeaturedArtist, error) {
	songs := make([]FeaturedArtist, 0)

	query := `SELECT * FROM featured_artists_t;`
	rows, err := uDB.Query(query)
	if err != nil {
		return nil, lib.ErrDB
	}
	defer rows.Close()
	for rows.Next() {
		song := FeaturedArtist{}
		if err := rows.Scan(&song.ID, &song.Artist, &song.SoundcloudURL, &song.Sequence); err != nil {
			return nil, lib.ErrDB
		}
		songs = append(songs, song)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return songs, nil
}

func (uDB *UhpDB) GetFeaturedArtist(id string) (*FeaturedArtist, error) {
	var song FeaturedArtist
	query := `SELECT * FROM featured_artists_t WHERE id=$1 LIMIT 1;`

	if err := uDB.QueryRow(query, id).Scan(&song.ID, &song.Artist, &song.SoundcloudURL, &song.Sequence); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, lib.ErrDB
	}

	return &song, nil
}

func (uDB *UhpDB) CreateFeaturedArtist(song FeaturedArtist) (*FeaturedArtist, error) {
	query := `INSERT INTO featured_artists_t (name, redirect_url, soundcloud_iframe_url, sequence) VALUES ($1, $2, $3, $4);`
	if _, err := uDB.Exec(query, song.Artist, song.SoundcloudURL, song.Sequence); err != nil {
		return nil, lib.ErrDB
	}
	// Grab auto-generated id
	// query = `SELECT id FROM featured_artists_t WHERE name=$1 and redirect_url=$2 and soundcloud_iframe_url=$3 and sequence=$4 LIMIT 1`
	// if err := uDB.QueryRow(query, song.Artist, song.SoundcloudURL, song.Sequence).Scan(&song.ID); err != nil {
	// 	return nil, lib.ErrDB
	// }
	return &song, nil
}

func (uDB *UhpDB) UpdateFeaturedArtist(song *FeaturedArtist) (*FeaturedArtist, error) {
	query := `UPDATE featured_artists_t SET name=$1, redirect_url=$2, soundcloud_iframe_url=$3, sequence=$4 WHERE id=$5;`
	if _, err := uDB.Exec(query, song.Artist, song.SoundcloudURL, song.Sequence, song.ID); err != nil {
		return nil, lib.ErrDB
	}
	return song, nil
}

func (uDB *UhpDB) DeleteFeaturedArtist(id string) error {
	query := `DELETE FROM featured_artists_t WHERE id=$1;`
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
