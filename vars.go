package goenvvars

import (
	"os"
	"strconv"
)

type EnvVar struct {
	key      string
	value    string
	found    bool
	optional bool
}

func NewEnvVar(key string, opts ...envVarOpt) EnvVar {
	ev := EnvVar{key: key}
	ev.value, ev.found = os.LookupEnv(key)
	for _, opt := range opts {
		opt(&ev)
	}
	if !ev.optional && ev.value == "" {
		panic("Missing required environment variable: " + ev.key)
	}
	return ev
}

func (e EnvVar) String() string {
	return e.value
}

func (e EnvVar) Bool() bool {
	ret, err := strconv.ParseBool(e.value)
	if err != nil {
		panic("Invalid boolean environment variable: " + e.value)
	}
	return ret
}

func (e EnvVar) Int() int {
	ret, err := strconv.Atoi(e.value)
	if err != nil {
		panic("Invalid integer environment variable: " + e.value)
	}
	return ret
}

func (e EnvVar) Float64() float64 {
	ret, err := strconv.ParseFloat(e.value, 64)
	if err != nil {
		panic("Invalid float environment variable: " + e.value)
	}
	return ret
}

type envVarOpt func(*EnvVar)

func Fallback(value string) envVarOpt {
	return func(e *EnvVar) {
		if !e.found && AllowFallbacks() {
			e.value = value
		}
	}
}

func Optional() envVarOpt {
	return func(e *EnvVar) {
		e.optional = true
	}
}

// Returns true if the environment variable with the given key is set and non-empty
func Presence(key string) bool {
	val, ok := os.LookupEnv(key)
	return ok && val != ""
}

func AllowFallbacks() bool {
	return IsDev() || IsTest()
}
