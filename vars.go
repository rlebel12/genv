package goenvvars

import (
	"os"
	"strconv"
)

type envVarData struct {
	key      string
	value    string
	found    bool
	optional bool
}

func EnvVar(key string, opts ...envVarOpt) envVarData {
	ev := envVarData{key: key}
	ev.value, ev.found = os.LookupEnv(key)
	for _, opt := range opts {
		opt(&ev)
	}
	if !ev.optional && ev.value == "" {
		panic("Missing required environment variable: " + ev.key)
	}
	return ev
}

func (e envVarData) String() string {
	return e.value
}

func (e envVarData) Bool() bool {
	ret, err := strconv.ParseBool(e.value)
	if err != nil {
		panic("Invalid boolean environment variable: " + e.value)
	}
	return ret
}

func (e envVarData) Int() int {
	ret, err := strconv.Atoi(e.value)
	if err != nil {
		panic("Invalid integer environment variable: " + e.value)
	}
	return ret
}

func (e envVarData) Float64() float64 {
	ret, err := strconv.ParseFloat(e.value, 64)
	if err != nil {
		panic("Invalid float environment variable: " + e.value)
	}
	return ret
}

type envVarOpt func(*envVarData)

func Fallback(value string) envVarOpt {
	return func(e *envVarData) {
		if !e.found && allowFallbacks() {
			e.value = value
		}
	}
}

func Optional() envVarOpt {
	return func(e *envVarData) {
		e.optional = true
	}
}

// Returns true if the environment variable with the given key is set and non-empty
func Presence(key string) bool {
	val, ok := os.LookupEnv(key)
	return ok && val != ""
}

func allowFallbacks() bool {
	return IsDev() || IsTest()
}
