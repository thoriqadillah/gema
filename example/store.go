package main

import "github.com/uptrace/bun"

type Store interface {
	Hello() string
}

type store struct {
	db *bun.DB
}

func newStore(db *bun.DB) Store {
	return &store{db: db}
}

func (s *store) Hello() string {
	return s.db.String()
}
