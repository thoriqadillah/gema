package example

import (
	"context"

	"github.com/thoriqadillah/gema"
)

type Store interface {
	Foo(ctx context.Context) error
	Hello(ctx context.Context) string
}

type store struct {
	db *gema.DB
}

func newStore(db *gema.DB) Store {
	return &store{db}
}

func (s *store) Hello(ctx context.Context) string {
	db := s.db.Tx(ctx)
	_ = db

	// do something with db
	return "Hello world"
}

func (s *store) Foo(ctx context.Context) error {
	db := s.db.Tx(ctx)
	_ = db

	// do something with db
	return nil
}
