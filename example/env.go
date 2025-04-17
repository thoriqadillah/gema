package main

import "github.com/thoriqadillah/gema"

var (
	APP_ENV = gema.Env("APP_ENV").String()
	PORT    = gema.Env("APP_PORT").String(":8000")
	DB_URL  = gema.Env("DB_URL").String("postgres://packform:packform@localhost:5432/packform?sslmode=disable")
)
