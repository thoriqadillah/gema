package env

import "github.com/thoriqadillah/gema"

var (
	APP_ENV = gema.Env("APP_ENV").String()
	PORT    = gema.Env("APP_PORT").String(":8000")
	DB_URL  = gema.Env("DB_URL").String("postgres://postgres:@localhost:5432/postgres?sslmode=disable")
)
