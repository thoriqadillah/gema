package main

import "github.com/thoriqadillah/gema"

var (
	APP_ENV = gema.Env("APP_ENV").String("development")
	PORT    = gema.Env("APP_PORT").String(":8000")
	DB_URL  = gema.Env("DB_URL").String("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
)
