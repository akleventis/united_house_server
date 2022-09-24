package uhp_db

import (
	"database/sql"

	"github.com/akleventis/united_house_server/lib"
	log "github.com/sirupsen/logrus"
)

// Column    | Type | Collation | Nullable | Default
// ----------+------+-----------+----------+---------
// username  | text |           | not null |
// password  | text |           |          |

func (uDB *UhpDB) GetAdminUserPass(username string) (string, error) {
	var password string
	query := `SELECT password FROM auth_t WHERE username=$1 LIMIT 1;`

	if err := uDB.QueryRow(query, username).Scan(&password); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		log.Error(err)
		return "", lib.ErrDB
	}
	return password, nil
}
