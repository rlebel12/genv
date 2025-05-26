package genv

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
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
				NewBool()
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

func (genv *Genv) Parse(vars ...*Var[any, any]) (err error) {
	for _, ev := range vars {
		value := os.Getenv(ev.key)
		if ev.many {
			split := strings.Split(value, ev.splitKey)
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
		}
		if err = v.parse(value); err != nil {
			return
		}
		// if err = v.parse() {

		// }
	}
	return nil
}

// Returns a new environment variable with the given key.
// func (genv *Genv) Var(key string, opts ...envVarOpt) *Var[T, U] {
// 	ev := new(Var)
// 	ev.key = key
// 	ev.allowDefault = genv.allowDefault
// 	ev.splitKey = genv.splitKey
// 	ev.value, ev.found = os.LookupEnv(key)
// 	ev.genv = genv

// 	for _, opt := range opts {
// 		opt(ev)
// 	}

// 	return ev
// }

// func (genv *Genv) Parse() (err error) {
// 	defer func() { genv.varFuncs = nil }()
// 	for _, f := range genv.varFuncs {
// 		if err = f(); err != nil {
// 			return
// 		}
// 	}
// 	return
// }

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

type oneOrMany[U any] interface {
	U | []U
}

type Var[T any, U any] struct {
	key    string
	target *T
	parse  func(string) (U, error)
	many   bool
	// value        string
	// found        bool
	optional     bool
	allowDefault allowFunc
	splitKey     string
	fb           *fallback
}

func (ev *Var[T, U]) parseSingle(value string) (err error) {
	return nil
}

func (ev *Var[T, U]) parseMany(value string) (err error) {
	results := []U{}
	values := strings.Split(value, ev.splitKey)
	for _, value := range values {
		if value == "" {
			continue
		}
		parsed, err := ev.parse(value)
		if err != nil {
			return err
		}
		results = append(results, parsed)
	}
	return nil
}

type fallback struct {
	allow allowFunc
	value string
}

type fallbackOpt func(*fallback)

func (ev *Var[T, U]) Optional() *Var[T, U] {
	ev.optional = true
	return ev
}

// Sets the default value for the environment variable if not present
func (ev *Var[T, U]) Default(value string, opts ...fallbackOpt) *Var[T, U] {
	fb := new(fallback)
	fb.allow = ev.allowDefault
	fb.value = value

	for _, opt := range opts {
		opt(fb)
	}

	ev.fb = fb
	return ev
}

type manyOpt func(*Var[T, U])

func (genv *Genv) WithSplitKey(splitKey string) manyOpt {
	return func(mev *Var[T, U]) {
		mev.splitKey = splitKey
	}
}

// String parses an environment variable into any type with string as underlying type
func String[T ~string](key string, target *T) *Var[T, T] {
	return newVar(key, target, false, parseString[T])
}

func StringSlice[S []T, T ~string](key string, target *S) *Var[S, T] {
	return newVar(key, target, true, parseString[T])
}

func parseString[T ~string](value string) (T, error) {
	return T(value), nil
}

func parseManyy[S []T, T any](parse func(string) (T, error), values ...string) (S, error) {
	var err error
	result := make(S, len(values))
	for i, value := range values {
		result[i], err = parse(value)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func newVar[T any, U any](key string, target *T, many bool, parse func(string) (U, error)) *Var[T, U] {
	return &Var[T, U]{
		key:    key,
		target: target,
		parse:  parse,
		many:   many,
	}
}

func NewString[T ~string](v *Var[T, U]) *T {
	s := new(T)
	String(s, v)
	return s
}

// Strings parses an environment variable into a slice of any type with string as underlying type
// func Strings[T ~string](s *[]T, v *Var[T, U], opts ...manyOpt) {
// 	prepareManyVars(v, s, parseString, opts...)
// }

func NewStrings[T ~string](v *Var[T, U], opts ...manyOpt) *[]T {
	s := new([]T)
	Strings(s, v, opts...)
	return s
}

func (v *Var[T, U]) parseString(s *string) (err error) {
	*s, err = parse(v, func(value string) (string, error) {
		return value, nil
	})
	if err != nil {
		err = fmt.Errorf("parse string: %w", err)
	}
	return
}

func (v *Var[T, U]) Bool(b *bool) *Var[T, U] {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error { return v.parseBool(b) })
	return v
}

func (v *Var[T, U]) NewBool() *bool {
	b := new(bool)
	v.Bool(b)
	return b
}

func (v *Var[T, U]) Bools(b *[]bool, opts ...manyOpt) *Var[T, U] {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error {
		return parseMany(v, b, func(ev *Var[T, U], result *bool) error {
			return ev.parseBool(result)
		}, opts...)
	})
	return v
}

func (v *Var[T, U]) NewBools(opts ...manyOpt) *[]bool {
	b := new([]bool)
	v.Bools(b, opts...)
	return b
}

func (v *Var[T, U]) parseBool(b *bool) (err error) {
	*b, err = parse(v, strconv.ParseBool)
	if err != nil {
		err = fmt.Errorf("parse bool: %w", err)
	}
	return
}

func (v *Var[T, U]) Int(i *int) *Var[T, U] {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error { return v.parseInt(i) })
	return v
}

func (v *Var[T, U]) NewInt() *int {
	i := new(int)
	v.Int(i)
	return i
}

func (v *Var[T, U]) Ints(i *[]int, opts ...manyOpt) *Var[T, U] {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error {
		return parseMany(v, i, func(ev *Var[T, U], result *int) error {
			return ev.parseInt(result)
		}, opts...)
	})
	return v
}

func (v *Var[T, U]) NewInts(opts ...manyOpt) *[]int {
	i := new([]int)
	v.Ints(i, opts...)
	return i
}

func (v *Var[T, U]) parseInt(i *int) (err error) {
	*i, err = parse(v, strconv.Atoi)
	if err != nil {
		err = fmt.Errorf("parse int: %w", err)
	}
	return
}

func (v *Var[T, U]) Float64(f *float64) *Var[T, U] {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error { return v.parseFloat(f) })
	return v
}

func (v *Var[T, U]) NewFloat64() *float64 {
	f := new(float64)
	v.Float64(f)
	return f
}

func (v *Var[T, U]) Float64s(f *[]float64, opts ...manyOpt) *Var[T, U] {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error {
		return parseMany(v, f, func(ev *Var[T, U], result *float64) error {
			return ev.parseFloat(result)
		}, opts...)
	})
	return v
}

func (v *Var[T, U]) NewFloat64s(opts ...manyOpt) *[]float64 {
	f := new([]float64)
	v.Float64s(f, opts...)
	return f
}

func (v *Var[T, U]) parseFloat(f *float64) (err error) {
	*f, err = parse(v, func(s string) (float64, error) {
		return strconv.ParseFloat(s, 64)
	})
	if err != nil {
		err = fmt.Errorf("parse float64: %w", err)
	}
	return
}

func (v *Var[T, U]) URL(u *url.URL) *Var[T, U] {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error { return v.parseURL(u) })
	return v
}

func (v *Var[T, U]) NewURL() *url.URL {
	u := new(url.URL)
	v.URL(u)
	return u
}

func (v *Var[T, U]) URLs(u *[]url.URL, opts ...manyOpt) *Var[T, U] {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error {
		return parseMany(v, u, func(ev *Var[T, U], result *url.URL) error {
			return ev.parseURL(result)
		}, opts...)
	})
	return v
}

func (v *Var[T, U]) NewURLs(opts ...manyOpt) *[]url.URL {
	u := new([]url.URL)
	v.URLs(u, opts...)
	return u
}

func (v *Var[T, U]) parseURL(u *url.URL) (err error) {
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

func (v *Var[T, U]) UUID(id *uuid.UUID) *Var[T, U] {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error { return v.parseUUID(id) })
	return v
}

func (v *Var[T, U]) NewUUID() *uuid.UUID {
	id := new(uuid.UUID)
	v.UUID(id)
	return id
}

func (v *Var[T, U]) UUIDs(id *[]uuid.UUID, opts ...manyOpt) *Var[T, U] {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error {
		return parseMany(v, id, func(ev *Var[T, U], result *uuid.UUID) error {
			return ev.parseUUID(result)
		}, opts...)
	})
	return v
}

func (v *Var[T, U]) NewUUIDs(opts ...manyOpt) *[]uuid.UUID {
	id := new([]uuid.UUID)
	v.UUIDs(id, opts...)
	return id
}

func (v *Var[T, U]) parseUUID(id *uuid.UUID) (err error) {
	*id, err = parse(v, uuid.Parse)
	if err != nil {
		return fmt.Errorf("parse uuid: %w", err)
	}
	return nil
}

func (v *Var[T, U]) getValue() (string, error) {
	value, _ := os.LookupEnv(v.key)
	if value == "" && v.fb != nil && v.fb.allow != nil {
		allow, err := v.fb.allow(genv)
		if err != nil {
			return "", fmt.Errorf(errFmtInvalidVar, v.key, err)
		}
		if allow {
			return v.fb.value, nil
		}
	}
	return value, nil
}

const errFmtInvalidVar = "%s is invalid: %w"

func parse[T any, U any](ev *Var[T, U], fn func(string) (T, error)) (T, error) {
	var (
		result T
		err    error
	)

	value, err := ev.getValue()
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

func parseMany[T any, U any](ev *Var[T, U], result *[]T, fn func(*Var[T, U], *T) error, opts ...manyOpt) error {
	for _, opt := range opts {
		opt(ev)
	}

	if ev.splitKey == "" {
		return errors.New("split key cannot be empty")
	}

	value, err := ev.getValue()
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

type envVarOpt func(*Var[T, U])

type genvOpt func(*Genv)

// Int parses an environment variable into any type with int as underlying type
func Int[T ~int](v *Var[T, U], i *T) *Var[T, U] {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error {
		var temp int
		err := v.parseInt(&temp)
		if err != nil {
			return err
		}
		*i = T(temp)
		return nil
	})
	return v
}

// Ints parses an environment variable into a slice of any type with int as underlying type
func Ints[T ~int](v *Var[T, U], i *[]T, opts ...manyOpt) *Var[T, U] {
	var temp []int
	result := v.Ints(&temp, opts...)
	v.genv.varFuncs = append(v.genv.varFuncs[:len(v.genv.varFuncs)-1], func() error {
		err := v.Ints(&temp, opts...).genv.varFuncs[len(v.genv.varFuncs)-1]()
		if err != nil {
			return err
		}
		*i = make([]T, len(temp))
		for j, val := range temp {
			(*i)[j] = T(val)
		}
		return nil
	})
	return result
}

// Bool parses an environment variable into any type with bool as underlying type
func Bool[T ~bool](v *Var[T, U], b *T) *Var[T, U] {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error {
		var temp bool
		err := v.parseBool(&temp)
		if err != nil {
			return err
		}
		*b = T(temp)
		return nil
	})
	return v
}

// Bools parses an environment variable into a slice of any type with bool as underlying type
func Bools[T ~bool](v *Var[T, U], b *[]T, opts ...manyOpt) *Var[T, U] {
	var temp []bool
	result := v.Bools(&temp, opts...)
	v.genv.varFuncs = append(v.genv.varFuncs[:len(v.genv.varFuncs)-1], func() error {
		err := v.Bools(&temp, opts...).genv.varFuncs[len(v.genv.varFuncs)-1]()
		if err != nil {
			return err
		}
		*b = make([]T, len(temp))
		for j, val := range temp {
			(*b)[j] = T(val)
		}
		return nil
	})
	return result
}

// Float64 parses an environment variable into any type with float64 as underlying type
func Float64[T ~float64](v *Var[T, U], f *T) *Var[T, U] {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error {
		var temp float64
		err := v.parseFloat(&temp)
		if err != nil {
			return err
		}
		*f = T(temp)
		return nil
	})
	return v
}

func prepareSingleVar[T any](v *Var[T, U], result *T, p func(value string) (T, error)) {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error {
		value, err := parse(v, p)
		if err != nil {
			return err
		}
		*result = value
		return nil
	})
}

func prepareManyVars[T any](ev *Var[T, U], result *[]T, p func(value string) (T, error), opts ...manyOpt) {
	ev.genv.varFuncs = append(ev.genv.varFuncs, func() error {
		for _, opt := range opts {
			opt(ev)
		}

		if ev.splitKey == "" {
			return errors.New("split key cannot be empty")
		}

		value, err := ev.getValue()
		if err != nil {
			return fmt.Errorf(errFmtInvalidVar, ev.key, err)
		}

		split := strings.Split(value, ev.splitKey)
		vals := make([]T, 0, len(split))
		for _, val := range split {
			if val == "" {
				continue
			}
			v, err := p(val)
			if err != nil {
				return fmt.Errorf(errFmtInvalidVar, ev.key, err)
			}
			vals = append(vals, T(v))
		}
		if !ev.optional && len(vals) == 0 {
			return fmt.Errorf(errFmtInvalidVar, ev.key, ErrRequiredEnvironmentVariable)
		}

		*result = vals
		return nil
	})
}
