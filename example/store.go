package main

import (
	"context"

	"github.com/thoriqadillah/gema"
)

type Store interface {
	Foo(ctx context.Context) error
	Hello(ctx context.Context) string
}

type store struct {
	txHost *gema.TransactionalHost
}

func newStore(txHost *gema.TransactionalHost) Store {
	return &store{txHost}
}

func (s *store) Hello(ctx context.Context) string {
	db := s.txHost.Tx(ctx)
	_ = db

	// do something with db
	return "Hello world"
}

func (s *store) Foo(ctx context.Context) error {
	db := s.txHost.Tx(ctx)
	_ = db

	// do something with db
	return nil
}
