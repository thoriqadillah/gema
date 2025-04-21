package main

import "github.com/thoriqadillah/gema"

var (
	DB_URL = gema.Env("DB_URL").String("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
)
