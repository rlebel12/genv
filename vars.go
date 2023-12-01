package goenvvars

import (
	"net/url"
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

type fallback struct {
	allow bool
}

type fallbackOpt func(*fallback)

func (ev *envVar) Fallback(value string, opts ...fallbackOpt) *envVar {
	fb := &fallback{
		allow: ev.allowFallback(),
	}

	for _, opt := range opts {
		opt(fb)
	}

	if !ev.found && fb.allow {
		ev.value = value
	}
	return ev
}

func OverrideAllow(allow func() bool) fallbackOpt {
	return func(f *fallback) {
		f.allow = allow()
	}
}

func AllowAlways() fallbackOpt {
	return OverrideAllow(func() bool {
		return true
	})
}

func (ev *envVar) String() string {
	ev.validate()
	return ev.value
}

func (ev *envVar) Bool() bool {
	ev.validate()
	if ev.value == "" {
		return false
	}
	ret, err := strconv.ParseBool(ev.value)
	if err != nil {
		panic("Invalid boolean environment variable: " + ev.value)
	}
	return ret
}

func (ev *envVar) Int() int {
	ev.validate()
	if ev.value == "" {
		return 0
	}
	ret, err := strconv.Atoi(ev.value)
	if err != nil {
		panic("Invalid integer environment variable: " + ev.value)
	}
	return ret
}

func (ev *envVar) Float64() float64 {
	ev.validate()
	if ev.value == "" {
		return 0
	}
	ret, err := strconv.ParseFloat(ev.value, 64)
	if err != nil {
		panic("Invalid float environment variable: " + ev.value)
	}
	return ret
}

// Returns the value of the environment variable as a URL.
// Panics if the value is not a valid URL, but this may happen
// if a scheme is not specified. See the documentation for
// url.Parse for more information.
func (ev *envVar) URL() *url.URL {
	ev.validate()
	if ev.value == "" {
		return &url.URL{}
	}
	ret, err := url.Parse(ev.value)
	if err != nil {
		panic("Invalid URL environment variable: " + ev.value)
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
