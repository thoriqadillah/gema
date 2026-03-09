package env

import (
	"github.com/thoriqadillah/gema"
)

var (
	DB_URL string
)

func Load() {
	DB_URL = gema.Env("DB_URL").String("postgres://postgres:gema@localhost:5433/gema?sslmode=disable")
}
