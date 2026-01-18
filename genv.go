package genv

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type (
	Genv struct {
		allowDefault allowFunc
		splitKey     string
		varFuncs     []func() error
		registry     *ParserRegistry
	}

	allowFunc func(*Genv) (bool, error)
)

func New(opts ...Opt[Genv]) *Genv {
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
		registry: NewDefaultRegistry(),
	}

	for _, opt := range opts {
		opt(genv)
	}
	return genv
}

func WithSplitKey(splitKey string) Opt[Genv] {
	return func(genv *Genv) {
		genv.splitKey = splitKey
	}
}

func WithAllowDefault(allowFn allowFunc) Opt[Genv] {
	return func(genv *Genv) {
		genv.allowDefault = allowFn
	}
}

func WithRegistry(registry *ParserRegistry) Opt[Genv] {
	return func(genv *Genv) {
		genv.registry = registry
	}
}

// Var Returns a new environment variable with the given key.
func (genv *Genv) Var(key string, opts ...Opt[Var]) *Var {
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
		WithRegistry(genv.registry),
	)
	return clone
}

func (genv *Genv) WithAllowDefault(allow func(genv *Genv) (bool, error)) Opt[fallback] {
	return func(f *fallback) {
		f.allow = allow
	}
}

func (genv *Genv) WithAllowDefaultAlways() Opt[fallback] {
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

func (fb *fallback) resolve(genv *Genv) (string, error) {
	if fb == nil || fb.allow == nil {
		return "", nil
	}

	isAllowed, err := fb.allow(genv)
	if err != nil {
		return "", fmt.Errorf("resolve fallback allow: %w", err)
	} else if !isAllowed {
		return "", nil
	}

	return fb.value, nil
}

func (v *Var) Optional() *Var {
	v.optional = true
	return v
}

// Default Sets the default value for the environment variable if not present
func (v *Var) Default(value string, opts ...Opt[fallback]) *Var {
	fb := new(fallback)
	fb.allow = v.allowDefault
	fb.value = value

	for _, opt := range opts {
		opt(fb)
	}

	v.fb = fb
	return v
}

func (genv *Genv) WithSplitKey(splitKey string) Opt[Var] {
	return func(mev *Var) {
		mev.splitKey = splitKey
	}
}

func (v *Var) String(s *string) *Var {
	return Type(v, s)
}

func (v *Var) NewString() *string {
	return NewType[string](v)
}

func (v *Var) Strings(s *[]string, opts ...Opt[Var]) *Var {
	return Types(v, s, opts...)
}

func (v *Var) NewStrings(opts ...Opt[Var]) *[]string {
	return NewTypes[string](v, opts...)
}

func (v *Var) Bool(b *bool) *Var {
	return Type(v, b)
}

func (v *Var) NewBool() *bool {
	return NewType[bool](v)
}

func (v *Var) Bools(b *[]bool, opts ...Opt[Var]) *Var {
	return Types(v, b, opts...)
}

func (v *Var) NewBools(opts ...Opt[Var]) *[]bool {
	return NewTypes[bool](v, opts...)
}

func (v *Var) Int(i *int) *Var {
	return Type(v, i)
}

func (v *Var) NewInt() *int {
	return NewType[int](v)
}

func (v *Var) Ints(i *[]int, opts ...Opt[Var]) *Var {
	return Types(v, i, opts...)
}

func (v *Var) NewInts(opts ...Opt[Var]) *[]int {
	return NewTypes[int](v, opts...)
}

func (v *Var) Float64(f *float64) *Var {
	return Type(v, f)
}

func (v *Var) NewFloat64() *float64 {
	return NewType[float64](v)
}

func (v *Var) Float64s(f *[]float64, opts ...Opt[Var]) *Var {
	return Types(v, f, opts...)
}

func (v *Var) NewFloat64s(opts ...Opt[Var]) *[]float64 {
	return NewTypes[float64](v, opts...)
}

func (v *Var) URL(u *url.URL) *Var {
	return Type(v, u)
}

func (v *Var) NewURL() *url.URL {
	return NewType[url.URL](v)
}

func (v *Var) URLs(u *[]url.URL, opts ...Opt[Var]) *Var {
	return Types(v, u, opts...)
}

func (v *Var) NewURLs(opts ...Opt[Var]) *[]url.URL {
	return NewTypes[url.URL](v, opts...)
}

func (v *Var) UUID(id *uuid.UUID) *Var {
	return Type(v, id)
}

func (v *Var) NewUUID() *uuid.UUID {
	return NewType[uuid.UUID](v)
}

func (v *Var) UUIDs(id *[]uuid.UUID, opts ...Opt[Var]) *Var {
	return Types(v, id, opts...)
}

func (v *Var) NewUUIDs(opts ...Opt[Var]) *[]uuid.UUID {
	return NewTypes[uuid.UUID](v, opts...)
}

func (v *Var) resolveValue() (string, error) {
	if v.found {
		return v.value, nil
	}

	fb, err := v.fb.resolve(v.genv)
	if err != nil {
		return "", fmt.Errorf("resolve fallback: %w", err)
	}
	return fb, nil
}

const errFmtInvalidVar = "%s is invalid: %w"

var ErrRequiredEnvironmentVariable = errors.New("environment variable is empty or unset")

func parseMany[T any](ev *Var, result *[]T, opts ...Opt[Var]) error {
	for _, opt := range opts {
		opt(ev)
	}

	if ev.splitKey == "" {
		return errors.New("split key cannot be empty")
	}

	value, err := ev.resolveValue()
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
			found:        true,
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
		err := assignValue(&ev, v)
		if err != nil {
			return fmt.Errorf(errFmtInvalidVar, ev.key, err)
		}
		*result = append(*result, *v)
	}
	return nil
}

// assignValue parses an environment variable into the target. The assignment only occurs
// if no errors occurred during parsing.
func assignValue[T any](ev *Var, target *T) error {
	value, err := parseOne[T](ev)
	if err != nil {
		return fmt.Errorf("parseOne: %w", err)
	}

	typedResult, ok := value.(T)
	if !ok {
		return fmt.Errorf("parser returned incorrect type: expected %T, got %T", *target, value)
	}
	*target = typedResult
	return nil
}

// Parser defines the strategy for parsing a string into any type using reflection
type Parser struct {
	ParseFn      func(string) (any, error)
	TargetTypeFn func() reflect.Type
	TypeNameFn   func() string
}

func (p Parser) Parse(value string) (any, error) {
	return p.ParseFn(value)
}

func (p Parser) TargetType() reflect.Type {
	return p.TargetTypeFn()
}

func (p Parser) TypeName() string {
	return p.TypeNameFn()
}

// newParser creates a new parser from a type and parsing function
func newParser(targetType reflect.Type, parseFn func(string) (any, error)) Parser {
	return Parser{
		ParseFn:      parseFn,
		TargetTypeFn: func() reflect.Type { return targetType },
		TypeNameFn:   func() string { return targetType.String() },
	}
}

// ParserRegistry manages parsers for different types using reflection
type ParserRegistry struct {
	parsers map[reflect.Type]Parser
}

func newParserRegistry() *ParserRegistry {
	return &ParserRegistry{
		parsers: make(map[reflect.Type]Parser),
	}
}

func (r *ParserRegistry) register(targetType reflect.Type, parser Parser) {
	if _, exists := r.parsers[targetType]; exists {
		panic(fmt.Sprintf("parser for type %s already registered", targetType))
	}
	r.parsers[targetType] = parser
}

func (r *ParserRegistry) get(targetType reflect.Type) (Parser, bool) {
	parser, exists := r.parsers[targetType]
	return parser, exists
}

// RegisterTypedParser registers a parser with type safety on a specific registry
func RegisterTypedParserOn[T any](registry *ParserRegistry, parseFn func(string) (T, error)) {
	var zero T
	targetType := reflect.TypeOf(zero)

	wrappedFn := func(s string) (any, error) {
		result, err := parseFn(s)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	registry.RegisterParseFunc(targetType, wrappedFn)
}

// RegisterParseFunc registers a parser using pure reflection (method API)
func (r *ParserRegistry) RegisterParseFunc(targetType reflect.Type, parseFn func(string) (any, error)) {
	parser := newParser(targetType, parseFn)
	r.register(targetType, parser)
}

// NewRegistry creates an empty parser registry
func NewRegistry() *ParserRegistry {
	return newParserRegistry()
}

// NewDefaultRegistry creates a parser registry with all built-in parsers
func NewDefaultRegistry() *ParserRegistry {
	registry := newParserRegistry()

	RegisterTypedParserOn(registry, func(s string) (string, error) {
		return s, nil
	})
	RegisterTypedParserOn(registry, strconv.ParseBool)
	RegisterTypedParserOn(registry, strconv.Atoi)
	RegisterTypedParserOn(registry, func(s string) (float64, error) {
		return strconv.ParseFloat(s, 64)
	})
	RegisterTypedParserOn(registry, func(s string) (url.URL, error) {
		result, err := url.Parse(s)
		if err != nil {
			return url.URL{}, err
		}
		return *result, nil
	})
	RegisterTypedParserOn(registry, uuid.Parse)
	RegisterTypedParserOn(registry, func(s string) (time.Time, error) {
		return time.Parse(time.RFC3339, s)
	})

	return registry
}

// getParser retrieves a parser for a specific type from the given registry
func getParser[T any](registry *ParserRegistry) (Parser, bool) {
	var zero T
	targetType := reflect.TypeOf(zero)
	return registry.get(targetType)
}

// Type sets the value of a variable of type T using the registered parser
func Type[T any](v *Var, target *T) *Var {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error {
		if err := assignValue(v, target); err != nil {
			return fmt.Errorf("assignValue: %w", err)
		}
		return nil
	})
	return v
}

// NewType creates a new variable of type T using the registered parser
func NewType[T any](v *Var) *T {
	target := new(T)
	Type(v, target)
	return target
}

// Types sets the value of a slice of type T using the registered parser
func Types[T any](v *Var, target *[]T, opts ...Opt[Var]) *Var {
	v.genv.varFuncs = append(v.genv.varFuncs, func() error {
		return parseMany(v, target, opts...)
	})
	return v
}

// NewTypes creates a new slice of type T using the registered parser
func NewTypes[T any](v *Var, opts ...Opt[Var]) *[]T {
	target := new([]T)
	Types(v, target, opts...)
	return target
}

// parseOne parses using the reflection-based parser interface
func parseOne[T any](ev *Var) (any, error) {
	parser, exists := getParser[T](ev.genv.registry)
	if !exists {
		var zero T
		return nil, fmt.Errorf("no parser registered for type %T", zero)
	}

	value, err := ev.resolveValue()
	if err != nil {
		return nil, fmt.Errorf(errFmtInvalidVar, ev.key, err)
	}

	if value == "" {
		if !ev.optional {
			return nil, fmt.Errorf(errFmtInvalidVar, ev.key, ErrRequiredEnvironmentVariable)
		}
		targetType := parser.TargetType()
		return reflect.Zero(targetType).Interface(), nil
	}

	result, err := parser.Parse(value)
	if err != nil {
		return nil, fmt.Errorf(errFmtInvalidVar, ev.key, err)
	}
	return result, nil
}

type Opt[T any] func(*T)

// VarFunc is a function that registers a variable on a Genv instance.
// It's returned by the package-level Bind() and BindMany() functions and used with Parse().
type VarFunc func(*Genv) *Var

// Bind creates a VarFunc that will register and parse a single value.
// The type is inferred from the pointer using generics (no reflection).
//
// Example:
//   err := genv.Parse(env,
//       genv.Bind("PORT", &port),          // T inferred as int
//       genv.Bind("NAME", &name).Default("unnamed"),
//       genv.Bind("DEBUG", &debug).Optional(),
//   )
func Bind[T any](key string, target *T) VarFunc {
	return func(env *Genv) *Var {
		v := env.Var(key)
		Type(v, target)
		return v
	}
}

// BindMany creates a VarFunc that will register and parse a slice of values.
// The type is inferred from the pointer using generics (no reflection).
//
// Example:
//   err := genv.Parse(env,
//       genv.BindMany("TAGS", &tags),
//       genv.BindMany("PORTS", &ports),
//   )
func BindMany[T any](key string, target *[]T, opts ...Opt[Var]) VarFunc {
	return func(env *Genv) *Var {
		v := env.Var(key)
		Types(v, target, opts...)
		return v
	}
}

// Parse registers all variables and parses them in one call.
// This is the recommended way to use the simplified API.
func Parse(env *Genv, vars ...VarFunc) error {
	for _, vf := range vars {
		vf(env)
	}
	return env.Parse()
}

// wrapVarFunc is a helper to create VarFunc wrapper methods
func wrapVarFunc(vf VarFunc, fn func(*Var) *Var) VarFunc {
	return func(env *Genv) *Var {
		v := vf(env)
		return fn(v)
	}
}

// Default sets the default value for this variable
func (vf VarFunc) Default(value string, opts ...Opt[fallback]) VarFunc {
	return wrapVarFunc(vf, func(v *Var) *Var {
		return v.Default(value, opts...)
	})
}

// Optional marks this variable as optional
func (vf VarFunc) Optional() VarFunc {
	return wrapVarFunc(vf, func(v *Var) *Var {
		return v.Optional()
	})
}
