package main

import (
	"github.com/thoriqadillah/gema"
)

var (
	DB_URL = gema.Env("DB_URL").String("postgres://postgres:gema@localhost:5432/gema?sslmode=disable")
)
