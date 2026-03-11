package api

import "github.com/thoriqadillah/gema"

type Foo struct {
	gema.Validate
	Bar string `json:"bar" validate:"required"`
}
