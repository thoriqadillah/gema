package main

import (
	"github.com/thoriqadillah/gema"
)

var (
	DB_URL string
)

func init() {
	DB_URL = gema.Env("DB_URL").String("postgres://postgres:gema@localhost:5433/gema?sslmode=disable")
}
