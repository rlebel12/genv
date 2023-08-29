package goenvvars

import (
	"errors"
	"os"
	"strconv"
)

var DefaultAllowFallback = defaultAllowFallback

func New(key string, opts ...envVarOpt) *envVar {
	ev := new(envVar)
	ev.key = key
	ev.allowFallback = DefaultAllowFallback
	ev.value, ev.found = os.LookupEnv(key)

	for _, opt := range opts {
		opt(ev)
	}

	return ev
}

func (ev *envVar) Optional() *envVar {
	ev.optional = true
	return ev
}

func (ev *envVar) Fallback(value string, opts ...envVarOpt) *envVar {
	for _, opt := range opts {
		opt(ev)
	}

	if !ev.found && ev.allowFallback() {
		ev.value = value
	}
	return ev
}

func OverrideAllow(af func() bool) envVarOpt {
	return func(e *envVar) {
		e.allowFallback = af
	}
}

func (ev *envVar) String() (string, error) {
	if err := ev.validate(); err != nil {
		return "", err
	}
	return ev.value, nil
}

func (ev *envVar) Bool() (bool, error) {
	if err := ev.validate(); err != nil {
		return false, err
	}
	return strconv.ParseBool(ev.value)
}

func (ev *envVar) Int() (int, error) {
	if err := ev.validate(); err != nil {
		return 0, err
	}
	return strconv.Atoi(ev.value)
}

func (ev *envVar) Float64() (float64, error) {
	if err := ev.validate(); err != nil {
		return 0, err
	}
	return strconv.ParseFloat(ev.value, 64)
}

// Returns true if the environment variable with the given key is set and non-empty
func Presence(key string) bool {
	val, ok := os.LookupEnv(key)
	return ok && val != ""
}

type envVar struct {
	key           string
	value         string
	found         bool
	optional      bool
	allowFallback func() bool
}

type envVarOpt func(*envVar)

func (ev *envVar) validate() error {
	required := !ev.optional
	empty := ev.value == ""
	if required && empty {
		return errors.New("Missing required environment variable: " + ev.key)
	}
	return nil
}

func defaultAllowFallback() bool {
	return !IsProd()
}
