package goevars

import (
	"os"
	"strconv"
)

type eVar string

// Returns the value of the environment variable with the given key.
// If the variable is not set and fallbacks are not allowed, this function panics.
// If the variable is not set and fallbacks are allowed, this function returns the fallback.
func Required(key string, fallback string) eVar {
	if value, ok := os.LookupEnv(key); ok {
		return eVar(value)
	} else if allowFallbacks() {
		return eVar(fallback)
	} else {
		panic("Missing required environment variable: " + key)
	}
}

// Returns the value of the environment variable with the given key, or the fallback if it is not set
func Optional(key string, fallback string) eVar {
	if value, ok := os.LookupEnv(key); ok {
		return eVar(value)
	} else {
		return eVar(fallback)
	}
}

func (e eVar) String() string {
	return string(e)
}

func (e eVar) Bool() bool {
	ret, err := strconv.ParseBool(e.String())
	if err != nil {
		panic("Invalid boolean environment variable: " + e.String())
	}
	return ret
}

func (e eVar) Int() int {
	ret, err := strconv.Atoi(e.String())
	if err != nil {
		panic("Invalid integer environment variable: " + e.String())
	}
	return ret
}

func (e eVar) Float64() float64 {
	ret, err := strconv.ParseFloat(e.String(), 64)
	if err != nil {
		panic("Invalid float environment variable: " + e.String())
	}
	return ret
}

func allowFallbacks() bool {
	return IsDev() || IsTest()
}
