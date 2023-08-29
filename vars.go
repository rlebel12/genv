package goenvvars

import (
	"os"
	"strconv"
)

var DefaultAllowFallback = defaultAllowFallback

func New(key string, opts ...envVarOpt) *envVar {
	ev := new(envVar)
	ev.key = key
	ev.allowFallback = DefaultAllowFallback
	ev.value, ev.found = os.LookupEnv(key)

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

func (ev *envVar) String() string {
	ev.validate()
	return ev.value
}

func (ev *envVar) Bool() bool {
	ev.validate()
	ret, err := strconv.ParseBool(ev.value)
	if err != nil {
		panic("Invalid boolean environment variable: " + ev.value)
	}
	return ret
}

func (ev *envVar) Int() int {
	ev.validate()
	ret, err := strconv.Atoi(ev.value)
	if err != nil {
		panic("Invalid integer environment variable: " + ev.value)
	}
	return ret
}

func (ev *envVar) Float64() float64 {
	ev.validate()
	ret, err := strconv.ParseFloat(ev.value, 64)
	if err != nil {
		panic("Invalid float environment variable: " + ev.value)
	}
	return ret
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

func (ev *envVar) validate() {
	if !ev.optional && ev.value == "" {
		panic("Missing required environment variable: " + ev.key)
	}
}

func defaultAllowFallback() bool {
	return !IsProd()
}
