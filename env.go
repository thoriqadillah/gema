package gema

import "os"

func Env(key string) Parser {
	return parseString(os.Getenv(key))
}
