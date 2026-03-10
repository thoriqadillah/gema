package db

import (
	"github.com/uptrace/bun"
)

func UnwrapTx(db bun.IDB) *bun.Tx {
	return db.(*bun.Tx)
}
