package goenvvars

import (
	"os"
	"strconv"
)

type EnvVar struct {
	Key           string
	Value         string
	Found         bool
	Optional      bool
	allowFallback func() bool
}

type envVarOpt func(*EnvVar)

type fallbackOpt func(*EnvVar)

var DefaultAllowFallback = defaultAllowFallback

func New(key string, opts ...envVarOpt) *EnvVar {
	ev := new(EnvVar)
	ev.Key = key
	ev.allowFallback = DefaultAllowFallback
	ev.Value, ev.Found = os.LookupEnv(key)

	for _, opt := range opts {
		opt(ev)
	}

	if !ev.Optional && ev.Value == "" {
		panic("Missing required environment variable: " + ev.Key)
	}

	return ev
}

func Optional() envVarOpt {
	return func(e *EnvVar) {
		e.Optional = true
	}
}

func Fallback(value string, opts ...fallbackOpt) envVarOpt {
	return func(e *EnvVar) {
		for _, opt := range opts {
			opt(e)
		}

		if !e.Found && e.allowFallback() {
			e.Value = value
		}
	}
}

func OverrideAllowFallback(af func() bool) fallbackOpt {
	return func(e *EnvVar) {
		e.allowFallback = af
	}
}

func (e EnvVar) String() string {
	return e.Value
}

func (e EnvVar) Bool() bool {
	ret, err := strconv.ParseBool(e.Value)
	if err != nil {
		panic("Invalid boolean environment variable: " + e.Value)
	}
	return ret
}

func (e EnvVar) Int() int {
	ret, err := strconv.Atoi(e.Value)
	if err != nil {
		panic("Invalid integer environment variable: " + e.Value)
	}
	return ret
}

func (e EnvVar) Float64() float64 {
	ret, err := strconv.ParseFloat(e.Value, 64)
	if err != nil {
		panic("Invalid float environment variable: " + e.Value)
	}
	return ret
}

// Returns true if the environment variable with the given key is set and non-empty
func Presence(key string) bool {
	val, ok := os.LookupEnv(key)
	return ok && val != ""
}

func defaultAllowFallback() bool {
	return !IsProd()
}
