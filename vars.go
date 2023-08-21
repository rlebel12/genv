package goevars

import (
	"os"
	"strconv"
)

type eVar struct {
	key      string
	value    string
	found    bool
	optional bool
}

type eVarOpt func(*eVar)

func Fallback(value string) eVarOpt {
	return func(e *eVar) {
		if !e.found && allowFallbacks() {
			e.value = value
		}
	}
}

func Optional() eVarOpt {
	return func(e *eVar) {
		e.optional = true
	}
}

func New(key string, opts ...eVarOpt) eVar {
	ev := eVar{key: key}
	ev.value, ev.found = os.LookupEnv(key)
	for _, opt := range opts {
		opt(&ev)
	}
	if !ev.optional && ev.value == "" {
		panic("Missing required environment variable: " + ev.key)
	}
	return ev
}

// Returns true if the environment variable with the given key is set and non-empty
func Presence(key string) bool {
	val, ok := os.LookupEnv(key)
	return ok && val != ""
}

func (e eVar) String() string {
	return e.convert()
}

func (e eVar) Bool() bool {
	ret, err := strconv.ParseBool(e.convert())
	if err != nil {
		panic("Invalid boolean environment variable: " + e.convert())
	}
	return ret
}

func (e eVar) Int() int {
	ret, err := strconv.Atoi(e.convert())
	if err != nil {
		panic("Invalid integer environment variable: " + e.convert())
	}
	return ret
}

func (e eVar) Float64() float64 {
	ret, err := strconv.ParseFloat(e.convert(), 64)
	if err != nil {
		panic("Invalid float environment variable: " + e.convert())
	}
	return ret
}

func (ev eVar) convert() string {
	return ev.value
}

func allowFallbacks() bool {
	return IsDev() || IsTest()
}
