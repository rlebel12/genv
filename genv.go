package genv

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type (
	Genv struct {
		allowDefault func(*Genv) bool
		splitKey     string
	}
)

func New(opts ...genvOpt) *Genv {
	genv := new(Genv)
	genv.allowDefault = func(genv *Genv) bool {
		return genv.
			Var("GENV_ALLOW_DEFAULT").
			Default("false", genv.WithAllowDefaultAlways()).
			Bool()
	}
	genv.splitKey = ","

	for _, opt := range opts {
		opt(genv)
	}
	return genv
}

func WithSplitKey(splitKey string) genvOpt {
	return func(genv *Genv) {
		genv.splitKey = splitKey
	}
}

func WithAllowDefault(allowFn func(*Genv) bool) genvOpt {
	return func(genv *Genv) {
		genv.allowDefault = allowFn
	}
}

// Returns a new environment variable with the given key.
func (genv *Genv) Var(key string, opts ...envVarOpt) *Var {
	ev := new(Var)
	ev.key = key
	ev.allowDefault = genv.allowDefault
	ev.splitKey = genv.splitKey
	ev.value, ev.found = os.LookupEnv(key)
	ev.genv = genv

	for _, opt := range opts {
		opt(ev)
	}

	return ev
}

func (genv *Genv) WithAllowDefault(allow func(genv *Genv) bool) defaultOpt {
	return func(f *fallback) {
		f.allow = allow
	}
}

func (genv *Genv) WithAllowDefaultAlways() defaultOpt {
	return genv.WithAllowDefault(func(*Genv) bool {
		return true
	})
}

type Var struct {
	key          string
	value        string
	found        bool
	optional     bool
	allowDefault func(*Genv) bool
	splitKey     string
	genv         *Genv
}

type fallback struct {
	allow func(*Genv) bool
}

type defaultOpt func(*fallback)

func (ev *Var) Optional() *Var {
	ev.optional = true
	return ev
}

// Sets the default value for the environment variable if not present
func (ev *Var) Default(value string, opts ...defaultOpt) *Var {
	fb := new(fallback)
	fb.allow = ev.allowDefault

	for _, opt := range opts {
		opt(fb)
	}

	if !ev.found && fb.allow != nil && fb.allow(ev.genv) {
		ev.value = value
	}
	return ev
}

type manyOpt func(*Var)

func (genv *Genv) WithSplitKey(splitKey string) manyOpt {
	return func(mev *Var) {
		mev.splitKey = splitKey
	}
}

func (ev *Var) String() string {
	ret, err := ev.TryString()
	if err != nil {
		panic(err)
	}
	return ret
}

func (ev *Var) TryString() (string, error) {
	if err := ev.validate(); err != nil {
		return "", fmt.Errorf("invalid string environment variable for %s ('%s'): %w", ev.key, ev.value, err)
	}
	return ev.value, nil
}

func (ev *Var) TryManyString(opts ...manyOpt) ([]string, error) {
	return parseMany(ev, (*Var).TryString, opts...)
}

func (ev *Var) ManyString(opts ...manyOpt) []string {
	ret, err := ev.TryManyString(opts...)
	if err != nil {
		panic(err)
	}
	return ret
}

func (ev *Var) Bool() bool {
	ret, err := ev.TryBool()
	if err != nil {
		panic(err)
	}
	return ret
}

func (ev *Var) TryBool() (bool, error) {
	if err := ev.validate(); err != nil {
		return false, err
	}
	if ev.value == "" {
		return false, nil
	}
	ret, err := strconv.ParseBool(ev.value)
	if err != nil {
		return false, fmt.Errorf("invalid boolean environment variable for %s ('%s'): %w", ev.key, ev.value, err)
	}
	return ret, nil
}

func (ev *Var) TryManyBool(opts ...manyOpt) ([]bool, error) {
	return parseMany(ev, (*Var).TryBool, opts...)
}

func (ev *Var) ManyBool(opts ...manyOpt) []bool {
	ret, err := ev.TryManyBool(opts...)
	if err != nil {
		panic(err)
	}
	return ret
}

func (ev *Var) Int() int {
	ret, err := ev.TryInt()
	if err != nil {
		panic(err)
	}
	return ret
}

func (ev *Var) TryInt() (int, error) {
	if err := ev.validate(); err != nil {
		return 0, fmt.Errorf("invalid integer environment variable for %s ('%s'): %w", ev.key, ev.value, err)
	}
	if ev.value == "" {
		return 0, nil
	}
	ret, err := strconv.Atoi(ev.value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer environment variable for %s ('%s'): %w", ev.key, ev.value, err)
	}
	return ret, nil
}

func (ev *Var) TryManyInt(opts ...manyOpt) ([]int, error) {
	return parseMany(ev, (*Var).TryInt, opts...)
}

func (ev *Var) ManyInt(opts ...manyOpt) []int {
	ret, err := ev.TryManyInt(opts...)
	if err != nil {
		panic(err)
	}
	return ret
}

func (ev *Var) Float64() float64 {
	ret, err := ev.TryFloat64()
	if err != nil {
		panic(err)
	}
	return ret
}

func (ev *Var) TryFloat64() (float64, error) {
	if err := ev.validate(); err != nil {
		return 0, fmt.Errorf("invalid float environment variable for %s ('%s'): %w", ev.key, ev.value, err)
	}
	if ev.value == "" {
		return 0, nil
	}
	ret, err := strconv.ParseFloat(ev.value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid float environment variable for %s ('%s'): %w", ev.key, ev.value, err)
	}
	return ret, nil
}

func (ev *Var) TryManyFloat64(opts ...manyOpt) ([]float64, error) {
	return parseMany(ev, (*Var).TryFloat64, opts...)
}

func (ev *Var) ManyFloat64(opts ...manyOpt) []float64 {
	ret, err := ev.TryManyFloat64(opts...)
	if err != nil {
		panic(err)
	}
	return ret
}

// Returns the value of the environment variable as a URL.
// Panics if the value is not a valid URL, but this may happen
// if a scheme is not specified. See the documentation for
// url.Parse for more information.
func (ev *Var) URL() *url.URL {
	ret, err := ev.TryURL()
	if err != nil {
		panic(err)
	}
	return ret
}

// Returns the value of the environment variable as a URL.
// Fails if the value is not a valid URL, but this may happen
// if a scheme is not specified. See the documentation for
// url.Parse for more information.
func (ev *Var) TryURL() (*url.URL, error) {
	if err := ev.validate(); err != nil {
		return &url.URL{}, fmt.Errorf("invalid URL environment variable for %s ('%s'): %w", ev.key, ev.value, err)
	}
	if ev.value == "" {
		return &url.URL{}, nil
	}
	ret, err := url.Parse(ev.value)
	if err != nil {
		return &url.URL{}, fmt.Errorf("invalid URL environment variable for %s ('%s'): %w", ev.key, ev.value, err)
	}
	return ret, nil
}

func (ev *Var) TryManyURL(opts ...manyOpt) ([]*url.URL, error) {
	return parseMany(ev, (*Var).TryURL, opts...)
}

func (ev *Var) ManyURL(opts ...manyOpt) []*url.URL {
	ret, err := ev.TryManyURL(opts...)
	if err != nil {
		panic(err)
	}
	return ret
}

// Returns true if the environment variable with the given key is set and non-empty
func (genv *Genv) Present(key string) bool {
	result := genv.Var(key).Optional().String()
	return result != ""
}

func parseMany[T any](ev *Var, parser func(*Var) (T, error), opts ...manyOpt) ([]T, error) {
	for _, opt := range opts {
		opt(ev)
	}
	split := strings.Split(ev.value, ev.splitKey)
	vars := make([]Var, 0, len(split))
	for _, val := range split {
		if val == "" {
			continue
		}
		vars = append(vars, Var{
			key:          ev.key,
			value:        val,
			found:        ev.found,
			optional:     ev.optional,
			allowDefault: ev.allowDefault,
			genv:         ev.genv,
		})
	}
	if !ev.optional && len(vars) == 0 {
		return nil, fmt.Errorf("missing required environment variable: %s", ev.key)
	}

	result := make([]T, len(vars))
	for i, ev := range vars {
		val, err := parser(&ev)
		if err != nil {
			return nil, fmt.Errorf("invalid environment variable for %s ('%s'): %w", ev.key, ev.value, err)
		}
		result[i] = val
	}
	return result, nil
}

type envVarOpt func(*Var)

func (ev *Var) validate() error {
	if !ev.optional && ev.value == "" {
		return fmt.Errorf("missing required environment variable: %s", ev.key)
	}
	return nil
}

type genvOpt func(*Genv)
