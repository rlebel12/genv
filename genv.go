package genv

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type (
	Genv struct {
		allowDefault allowFunc
		splitKey     string
		varFuncs     []func() error
	}

	allowFunc func(*Genv) (bool, error)
)

func New(opts ...genvOpt) *Genv {
	genv := &Genv{
		allowDefault: func(genv *Genv) (bool, error) {
			genv = genv.Clone()
			allow := genv.Var("GENV_ALLOW_DEFAULT").
				Default("false", genv.WithAllowDefaultAlways()).
				Bool()
			if err := genv.Parse(); err != nil {
				return false, fmt.Errorf("parse GENV_ALLOW_DEFAULT: %w", err)
			}
			return *allow, nil
		},
		splitKey: ",",
	}

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

func WithAllowDefault(allowFn allowFunc) genvOpt {
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

func (genv *Genv) Parse() (err error) {
	defer func() { genv.varFuncs = nil }()
	for _, f := range genv.varFuncs {
		if err = f(); err != nil {
			return
		}
	}
	return
}

func (genv *Genv) Clone() *Genv {
	clone := New(
		WithAllowDefault(genv.allowDefault),
		WithSplitKey(genv.splitKey),
	)
	return clone
}

func (genv *Genv) WithAllowDefault(allow func(genv *Genv) (bool, error)) fallbackOpt {
	return func(f *fallback) {
		f.allow = allow
	}
}

func (genv *Genv) WithAllowDefaultAlways() fallbackOpt {
	return genv.WithAllowDefault(func(*Genv) (bool, error) {
		return true, nil
	})
}

type Var struct {
	key          string
	value        string
	found        bool
	optional     bool
	allowDefault allowFunc
	splitKey     string
	genv         *Genv
	fb           *fallback
}

type fallback struct {
	allow allowFunc
	value string
}

type fallbackOpt func(*fallback)

func (ev *Var) Optional() *Var {
	ev.optional = true
	return ev
}

// Sets the default value for the environment variable if not present
func (ev *Var) Default(value string, opts ...fallbackOpt) *Var {
	fb := new(fallback)
	fb.allow = ev.allowDefault
	fb.value = value

	for _, opt := range opts {
		opt(fb)
	}

	ev.fb = fb
	return ev
}

type manyOpt func(*Var)

func (genv *Genv) WithSplitKey(splitKey string) manyOpt {
	return func(mev *Var) {
		mev.splitKey = splitKey
	}
}

func (v *Var) StringVar(s *string) {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error { return v.parseString(s) })
}

func (v *Var) String() *string {
	s := new(string)
	v.StringVar(s)
	return s
}

func (v *Var) ManyStringVar(s *[]string, opts ...manyOpt) {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error {
		return parseMany(v, s, func(ev *Var, result *string) error {
			return ev.parseString(result)
		}, opts...)
	})
}

func (v *Var) ManyString(opts ...manyOpt) *[]string {
	s := new([]string)
	v.ManyStringVar(s, opts...)
	return s
}

func (v *Var) parseString(s *string) (err error) {
	*s, err = parse(v, func(value string) (string, error) {
		return value, nil
	})
	if err != nil {
		err = fmt.Errorf("parse string: %w", err)
	}
	return
}

func (v *Var) BoolVar(b *bool) {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error { return v.parseBool(b) })
}

func (v *Var) Bool() *bool {
	b := new(bool)
	v.BoolVar(b)
	return b
}

func (v *Var) ManyBoolVar(b *[]bool, opts ...manyOpt) {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error {
		return parseMany(v, b, func(ev *Var, result *bool) error {
			return ev.parseBool(result)
		}, opts...)
	})
}

func (v *Var) ManyBool(opts ...manyOpt) *[]bool {
	b := new([]bool)
	v.ManyBoolVar(b, opts...)
	return b
}

func (v *Var) parseBool(b *bool) (err error) {
	*b, err = parse(v, strconv.ParseBool)
	if err != nil {
		err = fmt.Errorf("parse bool: %w", err)
	}
	return
}

func (v *Var) IntVar(i *int) {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error { return v.parseInt(i) })
}

func (v *Var) Int() *int {
	i := new(int)
	v.IntVar(i)
	return i
}

func (v *Var) ManyIntVar(i *[]int, opts ...manyOpt) {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error {
		return parseMany(v, i, func(ev *Var, result *int) error {
			return ev.parseInt(result)
		}, opts...)
	})
}

func (v *Var) ManyInt(opts ...manyOpt) *[]int {
	i := new([]int)
	v.ManyIntVar(i, opts...)
	return i
}

func (v *Var) parseInt(i *int) (err error) {
	*i, err = parse(v, strconv.Atoi)
	if err != nil {
		err = fmt.Errorf("parse int: %w", err)
	}
	return
}

func (v *Var) Float64Var(f *float64) {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error { return v.parseFloat(f) })
}

func (v *Var) Float64() *float64 {
	f := new(float64)
	v.Float64Var(f)
	return f
}

func (v *Var) ManyFloat64Var(f *[]float64, opts ...manyOpt) {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error {
		return parseMany(v, f, func(ev *Var, result *float64) error {
			return ev.parseFloat(result)
		}, opts...)
	})
}

func (v *Var) ManyFloat64(opts ...manyOpt) *[]float64 {
	f := new([]float64)
	v.ManyFloat64Var(f, opts...)
	return f
}

func (v *Var) parseFloat(f *float64) (err error) {
	*f, err = parse(v, func(s string) (float64, error) {
		return strconv.ParseFloat(s, 64)
	})
	if err != nil {
		err = fmt.Errorf("parse float64: %w", err)
	}
	return
}

func (v *Var) URLVar(u *url.URL) {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error { return v.parseURL(u) })
}

func (v *Var) URL() *url.URL {
	u := new(url.URL)
	v.URLVar(u)
	return u
}

func (v *Var) ManyURLVar(u *[]url.URL, opts ...manyOpt) {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error {
		return parseMany(v, u, func(ev *Var, result *url.URL) error {
			return ev.parseURL(result)
		}, opts...)
	})
}

func (v *Var) ManyURL(opts ...manyOpt) *[]url.URL {
	u := new([]url.URL)
	v.ManyURLVar(u, opts...)
	return u
}

func (v *Var) parseURL(u *url.URL) (err error) {
	*u, err = parse(v, func(s string) (url.URL, error) {
		result, err := url.Parse(s)
		if err != nil {
			return url.URL{}, err
		}
		return *result, nil
	})
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	return nil
}

func (v *Var) parseValue() (string, error) {
	if v.value == "" && v.fb != nil && v.fb.allow != nil {
		allow, err := v.fb.allow(v.genv)
		if err != nil {
			return "", fmt.Errorf(errFmtInvalidVar, v.key, err)
		}
		if allow {
			return v.fb.value, nil
		}
	}
	return v.value, nil
}

const errFmtInvalidVar = "%s is invalid: %w"

func parse[T any](ev *Var, fn func(string) (T, error)) (T, error) {
	var (
		result T
		err    error
	)

	value, err := ev.parseValue()
	if err != nil {
		return result, fmt.Errorf(errFmtInvalidVar, ev.key, err)
	}

	if !ev.optional && value == "" {
		return result, fmt.Errorf(errFmtInvalidVar, ev.key, ErrRequiredEnvironmentVariable)
	}

	if value == "" {
		// If validation succeeded, then the value being empty means it was
		// optional (or just an empty string is the desired output).
		// In that case, use the zero value.
		return result, nil
	}

	result, err = fn(value)
	if err != nil {
		return result, fmt.Errorf(errFmtInvalidVar, ev.key, err)
	}
	return result, nil
}

var ErrRequiredEnvironmentVariable = errors.New("environment variable is empty or unset")

func parseMany[T any](ev *Var, result *[]T, fn func(*Var, *T) error, opts ...manyOpt) error {
	for _, opt := range opts {
		opt(ev)
	}

	if ev.splitKey == "" {
		return errors.New("split key cannot be empty")
	}

	value, err := ev.parseValue()
	if err != nil {
		return fmt.Errorf(errFmtInvalidVar, ev.key, err)
	}

	split := strings.Split(value, ev.splitKey)
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
		return fmt.Errorf(errFmtInvalidVar, ev.key, ErrRequiredEnvironmentVariable)
	}

	for _, ev := range vars {
		v := new(T)
		err := fn(&ev, v)
		if err != nil {
			return fmt.Errorf(errFmtInvalidVar, ev.key, err)
		}
		*result = append(*result, *v)
	}
	return nil
}

type envVarOpt func(*Var)

type genvOpt func(*Genv)
