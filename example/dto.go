package main

import "github.com/thoriqadillah/gema"

type foo struct {
	Bar string `json:"bar" validate:"required"`
}

func (f *foo) Validate() error {
	return gema.ValidateStruct(f)
}
