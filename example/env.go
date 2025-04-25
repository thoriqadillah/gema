package main

import (
	"os"

	"github.com/thoriqadillah/gema"
)

func env(k string) gema.Parser {
	return gema.ParseString(os.Getenv(k))
}

var (
	APP_ENV = env("APP_ENV").String("development")
	PORT    = env("APP_PORT").String(":8001")
	DB_URL  = env("DB_URL").String("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
)
