package gema

import (
	"strconv"
	"time"
)

type Parser interface {
	String(defaults ...string) string
	Int(defaults ...int) int
	Bool(defaults ...bool) bool
	Float64(defaults ...float64) float64
	Time(defaults ...time.Time) time.Time
	Duration(defaults ...time.Duration) time.Duration
}

type parser struct {
	value string
}

func ParseString(str string) Parser {
	return &parser{str}
}

func (p *parser) String(defaults ...string) string {
	if p.value == "" && len(defaults) > 0 {
		return defaults[0]
	}

	if p.value == "" && len(defaults) == 0 {
		return ""
	}

	return p.value
}

func (p *parser) Bool(defaults ...bool) bool {
	if p.value == "" && len(defaults) > 0 {
		return defaults[0]
	}

	if p.value == "" && len(defaults) == 0 {
		return false
	}

	return p.value == "true"
}

func (p *parser) Int(defaults ...int) int {
	if p.value == "" && len(defaults) > 0 {
		return defaults[0]
	}

	if p.value == "" && len(defaults) == 0 {
		return 0
	}

	num, err := strconv.Atoi(p.value)
	if err != nil && len(defaults) > 0 {
		return defaults[0]
	}

	if err != nil && len(defaults) == 0 {
		return 0
	}

	return num
}

func (p *parser) Float64(defaults ...float64) float64 {
	if p.value == "" && len(defaults) > 0 {
		return defaults[0]
	}

	if p.value == "" && len(defaults) == 0 {
		return 0
	}

	num, err := strconv.ParseFloat(p.value, 32)
	if err != nil && len(defaults) > 0 {
		return defaults[0]
	}

	if err != nil && len(defaults) == 0 {
		return 0
	}

	return num
}

func (p *parser) Time(defaults ...time.Time) time.Time {
	if p.value == "" && len(defaults) > 0 {
		return defaults[0]
	}

	if p.value == "" && len(defaults) == 0 {
		return time.Time{}
	}

	date, err := time.Parse("2006-01-02", p.value)
	if err != nil && len(defaults) > 0 {
		return defaults[0]
	}

	if err != nil && len(defaults) == 0 {
		return time.Time{}
	}

	return date
}

func (p *parser) Duration(defaults ...time.Duration) time.Duration {
	if p.value == "" && len(defaults) > 0 {
		return defaults[0]
	}

	if p.value == "" && len(defaults) == 0 {
		return 0
	}

	duration, err := time.ParseDuration(p.value)
	if err != nil && len(defaults) > 0 {
		return defaults[0]
	}

	if err != nil && len(defaults) == 0 {
		return 0
	}

	return duration
}
