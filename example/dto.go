package main

import "github.com/thoriqadillah/gema"

type foo struct {
	gema.Validate
	Bar string `json:"bar" validate:"required"`
}
