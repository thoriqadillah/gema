package db

import (
	"database/sql"

	"github.com/uptrace/bun"
)

func UnwrapTx(db bun.IDB) *sql.Tx {
	return db.(*bun.Tx).Tx
}
