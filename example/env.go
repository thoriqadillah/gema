package main

import (
	"github.com/thoriqadillah/gema"
)

var (
	APP_ENV      string
	PORT         string
	DB_URL       string
	DB_QUEUE_URL string
	MAILER_HOST  string
	MAILER_PORT  int
	MAILER_USER  string
	MAILER_PASS  string
	MAILER_FROM  string
	MAILER_NAME  string
)

func init() {
	APP_ENV = gema.Env("APP_ENV").String("development")
	PORT = gema.Env("APP_PORT").String(":8001")
	DB_URL = gema.Env("DB_URL").String("postgres://postgres:gema@localhost:5433/gema?sslmode=disable")
	MAILER_HOST = gema.Env("MAILER_HOST").String("smtp.gmail.com")
	MAILER_PORT = gema.Env("MAILER_PORT").Int(587)
	MAILER_USER = gema.Env("MAILER_USER").String()
	MAILER_PASS = gema.Env("MAILER_PASS").String()
	MAILER_FROM = gema.Env("MAILER_FROM").String()
	MAILER_NAME = gema.Env("MAILER_NAME").String("Gema")
}
