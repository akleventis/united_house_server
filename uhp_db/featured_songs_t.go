package uhp_db

type FeaturedSong struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	SoundcloudURL string `json:"soundcloud_iframe_url"`
}

// Desc featured_songs_t
// Column         |         Type          | Collation | Nullable |                    Default
// -----------------------+-----------------------+-----------+----------+------------------------------------------------
//  id                    | integer               |           | not null | nextval('featured_artists_t_id_seq'::regclass)
//  name                  | character varying(50) |           | not null |
//  soundcloud_iframe_url | character varying(50) |           |          |
// Indexes:
//     "featured_artists_t_pkey" PRIMARY KEY, btree (id)
