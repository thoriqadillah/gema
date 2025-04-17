package gema

import (
	"os"
)

func Env(key string) Parser {
	v := os.Getenv(key)
	return parseString(v)
}
