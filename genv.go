package goenvvars

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Genv struct {
	defaultAllowFallback func() bool
	environment          environment
	environmentKey       string
	defaultSplitKey      string
}

func NewGenv(opts ...genvOpt) (*Genv, error) {
	genv := new(Genv)
	genv.environmentKey = "ENV"
	genv.defaultAllowFallback = func() bool { return !genv.IsProd() }
	genv.defaultSplitKey = ","

	environment, err := newEnvironment(genv)
	if err != nil {
		return nil, fmt.Errorf("invalid environment: %w", err)
	}
	genv.environment = environment

	for _, opt := range opts {
		opt(genv)
	}
	return genv, nil
}

func WithDefaultSplitKey(splitKey string) genvOpt {
	return func(genv *Genv) {
		genv.defaultSplitKey = splitKey
	}
}

// Returns a new environment variable with the given key.
func (genv *Genv) New(key string, opts ...envVarOpt) *envVar {
	ev := new(envVar)
	ev.key = key
	ev.allowFallback = genv.defaultAllowFallback
	ev.splitKey = genv.defaultSplitKey
	ev.value, ev.found = os.LookupEnv(key)

	for _, opt := range opts {
		opt(ev)
	}

	return ev
}

// Returns a new environment variable with the given key. Alias for New.
func (genv *Genv) Env(key string, opts ...envVarOpt) *envVar {
	return genv.New(key, opts...)
}

func (genv *Genv) IsDev() bool {
	return genv.environment == Dev
}

func (genv *Genv) IsProd() bool {
	return genv.environment == Prod
}

func (genv *Genv) IsTest() bool {
	return genv.environment == Test
}

type envVar struct {
	key           string
	value         string
	found         bool
	optional      bool
	allowFallback func() bool
	splitKey      string
}

func (ev *envVar) Optional() *envVar {
	ev.optional = true
	return ev
}

type fallback struct {
	allow bool
}

type fallbackOpt func(*fallback)

// Sets the default value for the environment variable if not present. Alias for Fallback.
func (ev *envVar) Default(value string, opts ...fallbackOpt) *envVar {
	return ev.Fallback(value, opts...)
}

// Sets the default value for the environment variable if not present.
func (ev *envVar) Fallback(value string, opts ...fallbackOpt) *envVar {
	fb := new(fallback)
	if ev.allowFallback != nil {
		fb.allow = ev.allowFallback()
	}

	for _, opt := range opts {
		opt(fb)
	}

	if !ev.found && fb.allow {
		ev.value = value
	}
	return ev
}

type manyOpt func(*envVar)

func WithSplitKey(splitKey string) manyOpt {
	return func(mev *envVar) {
		mev.splitKey = splitKey
	}
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
	ret, err := ev.TryString()
	if err != nil {
		panic(err)
	}
	return ret
}

func (ev *envVar) TryString() (string, error) {
	if err := ev.validate(); err != nil {
		return "", fmt.Errorf("invalid string environment variable for %s ('%s'): %w", ev.key, ev.value, err)
	}
	return ev.value, nil
}

func (ev *envVar) TryManyString(opts ...manyOpt) ([]string, error) {
	return parseMany(ev, (*envVar).TryString, opts...)
}

func (ev *envVar) ManyString(opts ...manyOpt) []string {
	ret, err := ev.TryManyString(opts...)
	if err != nil {
		panic(err)
	}
	return ret
}

func (ev *envVar) Bool() bool {
	ret, err := ev.TryBool()
	if err != nil {
		panic(err)
	}
	return ret
}

func (ev *envVar) TryBool() (bool, error) {
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

func (ev *envVar) TryManyBool(opts ...manyOpt) ([]bool, error) {
	return parseMany(ev, (*envVar).TryBool, opts...)
}

func (ev *envVar) ManyBool(opts ...manyOpt) []bool {
	ret, err := ev.TryManyBool(opts...)
	if err != nil {
		panic(err)
	}
	return ret
}

func (ev *envVar) Int() int {
	ret, err := ev.TryInt()
	if err != nil {
		panic(err)
	}
	return ret
}

func (ev *envVar) TryInt() (int, error) {
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

func (ev *envVar) TryManyInt(opts ...manyOpt) ([]int, error) {
	return parseMany(ev, (*envVar).TryInt, opts...)
}

func (ev *envVar) ManyInt(opts ...manyOpt) []int {
	ret, err := ev.TryManyInt(opts...)
	if err != nil {
		panic(err)
	}
	return ret
}

func (ev *envVar) Float64() float64 {
	ret, err := ev.TryFloat64()
	if err != nil {
		panic(err)
	}
	return ret
}

func (ev *envVar) TryFloat64() (float64, error) {
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

func (ev *envVar) TryManyFloat64(opts ...manyOpt) ([]float64, error) {
	return parseMany(ev, (*envVar).TryFloat64, opts...)
}

func (ev *envVar) ManyFloat64(opts ...manyOpt) []float64 {
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
func (ev *envVar) URL() *url.URL {
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
func (ev *envVar) TryURL() (*url.URL, error) {
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

func (ev *envVar) TryManyURL(opts ...manyOpt) ([]*url.URL, error) {
	return parseMany(ev, (*envVar).TryURL, opts...)
}

func (ev *envVar) ManyURL(opts ...manyOpt) []*url.URL {
	ret, err := ev.TryManyURL(opts...)
	if err != nil {
		panic(err)
	}
	return ret
}

func parseMany[T any](ev *envVar, parser func(*envVar) (T, error), opts ...manyOpt) ([]T, error) {
	for _, opt := range opts {
		opt(ev)
	}
	split := strings.Split(ev.value, ev.splitKey)
	vars := make([]envVar, 0, len(split))
	for _, val := range split {
		if val == "" {
			continue
		}
		vars = append(vars, envVar{
			key:           ev.key,
			value:         val,
			found:         ev.found,
			optional:      ev.optional,
			allowFallback: ev.allowFallback,
		})
	}
	if !ev.optional && len(vars) == 0 {
		return nil, fmt.Errorf("missing required environment variable: %s", ev.key)
	}

	result := make([]T, len(vars))
	for i, ev := range vars {
		val, err := parser(&ev)
		if err != nil {
			return nil, err
		}
		result[i] = val
	}
	return result, nil
}

// Returns true if the environment variable with the given key is set and non-empty
func Presence(key string) bool {
	val, ok := os.LookupEnv(key)
	return ok && val != ""
}

type envVarOpt func(*envVar)

func (ev *envVar) validate() error {
	if !ev.optional && ev.value == "" {
		return fmt.Errorf("missing required environment variable: %s", ev.key)
	}
	return nil
}

type genvOpt func(*Genv)

type environment uint8

const (
	Dev environment = iota
	Prod
	Test
)

func newEnvironment(genv *Genv) (environment, error) {
	envStr, err := genv.New(genv.environmentKey).
		Fallback("DEVELOPMENT", OverrideAllow(func() bool { return true })).
		TryString()
	if err != nil {
		return 0, fmt.Errorf("invalid environment - must be a string: %w", err)
	}
	env, ok := environments()[envStr]
	if !ok {
		return 0, fmt.Errorf("invalid environment value: %w", err)
	}
	return env, nil
}

func environments() map[string]environment {
	return map[string]environment{
		"DEVELOPMENT": Dev,
		"DEV":         Dev,
		"PRODUCTION":  Prod,
		"PROD":        Prod,
		"TEST":        Test,
	}
}

func DefaultAllowFallback(allow func() bool) genvOpt {
	return func(genv *Genv) {
		genv.defaultAllowFallback = allow
	}
}

func EnvironmentKey(key string) genvOpt {
	return func(genv *Genv) {
		genv.environmentKey = key
	}
}
