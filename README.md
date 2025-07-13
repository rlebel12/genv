# genv

[![Go Reference](https://pkg.go.dev/badge/github.com/rlebel12/genv.svg)](https://pkg.go.dev/github.com/rlebel12/genv)
[![Test](https://github.com/rlebel12/genv/actions/workflows/test.yml/badge.svg)](https://github.com/rlebel12/genv/actions/workflows/test.yml)

A small package to help work with environment variables in Go.

## Installation

```console
go get github.com/rlebel12/genv
```

## Usage

First, ensure that the package is imported:

```go
import "github.com/rlebel12/genv"
```

In its most basic form, the package can be used to retrieve environment variables and then parse them into specified types:

```go
env := genv.New() // Creates instance with default parsers

StringVar := env.Var("STRING_VAR").NewString()
BoolVar := env.Var("BOOL_VAR").NewBool()
IntVar := env.Var("INT_VAR").NewInt()
FloatVar := env.Var("FLOAT_VAR").NewFloat64()
URLVar := env.Var("URL_VAR").NewURL()

if err := env.Parse(); err != nil {
    slog.Error("env parse", "error", err.Error())
}

// Or instead, using pointers:

type Example struct {
    StringVar string
    IntVar    int
}

var example Example
env.Var("STRING_VAR").String(&example.StringVar)
env.Var("INT_VAR").Int(&example.IntVar)

if err := env.Parse(); err != nil {
    slog.Error("env parse", "error", err.Error())
}

// Advanced: Using custom registries (optional)
customRegistry := genv.NewDefaultRegistry() // or genv.NewRegistry() for empty
env := genv.New(genv.WithRegistry(customRegistry))
```

### Optional Variables

By default, parsing will fail if an environment variable is absent (either because the environment variable is not defined, or because it was set to an empty string). However, it is possible to specify that a variable is optional to prevent the failure and default to the zero value:

```go
var OptionalVar = genv.Var("OPTIONAL_VAR").Optional()
```

### Defaults

You can specify a default value to use if the environment variable is absent:

```go
var DefaultVar = genv.Var("DEFAULT_VAR").Default("default value")
```

This is intended to be used to allow speed and ease of development while ensuring that all environment variables are defined before deploying to production. Thus, the default behavior is to forbid defaults unless the `GENV_ALLOW_DEFAULT` environment variable evaluates to `true`. This behavior can be overridden in two ways.

#### Allow Defaults: Global Override

Override the behavior for all subsequent invocations of `genv.Var`:

```go
var genv := genv.New(
    genv.WithAllowDefault(func() bool { return true }),
)
```

#### Allow Defaults: Individual Override

Override the behavior for individual environment variables by passing an `WithAllowDefault` option to `Default`:

```go
var DefaultVar = genv.Var("DEFAULT_VAR").Default(
    "default value",
    genv.WithAllowDefault(func() bool { return true }),
)
```

This approach takes priority over the global override.

### Combining Options

Options can be chained together. For example, it is possible to declare that an environment variable is both
optional and has a default value. This means that the default value will be used if allowed and necessary, and the program
will not panic if the final value is still absent.

```go
var OptionalDefaultVar = genv.Var("OPTIONAL_DEFAULT_VAR").
    Default("default value").
    Optional()
```

### Custom Types & Parser Registry

The package supports registering custom type parsers and using isolated parser registries:

#### Registering Custom Types

You can register parsers for your own types on specific registries:

```go
// Define custom type
type UserID string

// Create registry and register custom type
registry := genv.NewDefaultRegistry()
genv.RegisterTypedParserOn(registry, func(s string) (UserID, error) {
    if s == "" {
        return "", errors.New("UserID cannot be empty")
    }
    return UserID("user_" + s), nil
})

// Use custom registry
env := genv.New(genv.WithRegistry(registry))
// Note: Direct custom type support would require additional API methods
```

#### Custom Parser Registries

Create isolated parser registries for different contexts:

```go
// Create empty registry (no built-in parsers)
customRegistry := genv.NewRegistry()

// Register only the types you need
genv.RegisterTypedParserOn(customRegistry, func(s string) (string, error) {
    return strings.ToUpper(s), nil // Custom string behavior
})

// Create registry with default parsers
defaultRegistry := genv.NewDefaultRegistry()

// Use custom registry
env := genv.New(genv.WithRegistry(customRegistry))
```

#### Registry Isolation

Different `Genv` instances can have completely different parsing behavior:

```go
// Registry for production with strict validation
prodRegistry := genv.NewDefaultRegistry()
genv.RegisterTypedParserOn(prodRegistry, func(s string) (UserID, error) {
    if !strings.HasPrefix(s, "user_") {
        return "", errors.New("invalid UserID format")
    }
    return UserID(s), nil
})

// Registry for testing with lenient validation
testRegistry := genv.NewDefaultRegistry()
genv.RegisterTypedParserOn(testRegistry, func(s string) (UserID, error) {
    return UserID("user_" + s), nil // Always add prefix
})

prodEnv := genv.New(genv.WithRegistry(prodRegistry))
testEnv := genv.New(genv.WithRegistry(testRegistry))
```

## Advanced Usage

### Working with Custom Types

The new registry system enables full end-to-end custom type parsing:

```go
type ServicePort int
type LogLevel string

// Create registry with custom parsers
registry := genv.NewDefaultRegistry()

genv.RegisterTypedParserOn(registry, func(s string) (ServicePort, error) {
    port, err := strconv.Atoi(s)
    if err != nil {
        return ServicePort(0), err
    }
    if port < 1024 || port > 65535 {
        return ServicePort(0), errors.New("port out of range")
    }
    return ServicePort(port), nil
})

genv.RegisterTypedParserOn(registry, func(s string) (LogLevel, error) {
    level := strings.ToUpper(s)
    if level == "DEBUG" || level == "INFO" || level == "WARN" || level == "ERROR" {
        return LogLevel(level), nil
    }
    return "", errors.New("invalid log level")
})

// Use custom types
env := genv.New(genv.WithRegistry(registry))
var config struct {
    Port     ServicePort
    LogLevel LogLevel
}

genv.Type(env.Var("SERVICE_PORT"), &config.Port)
genv.Type(env.Var("LOG_LEVEL"), &config.LogLevel)

if err := env.Parse(); err != nil {
    log.Fatal(err)
}
```

### Custom Type Slices

Custom types work with slice parsing using the same delimiter system:

```go
type Priority int

genv.RegisterTypedParserOn(registry, func(s string) (Priority, error) {
    switch strings.ToLower(s) {
    case "low": return Priority(1), nil
    case "medium": return Priority(2), nil
    case "high": return Priority(3), nil
    default: return Priority(0), errors.New("invalid priority")
    }
})

var priorities []Priority
genv.Types(env.Var("TASK_PRIORITIES"), &priorities) // e.g., "low,high,medium"
```

### Optional and Default with Custom Types

Custom types fully support optional and default value patterns:

```go
type Environment string

// Optional custom type
var env Environment
genv.Type(genv.Var("APP_ENV").Optional(), &env) // Won't error if missing

// Custom type with default
var envWithDefault Environment  
genv.Type(genv.Var("APP_ENV").Default("development"), &envWithDefault)
```

## Registry Patterns

### When to Use Different Registry Types

**`genv.New()` (Default Registry)**
- Use for simple applications
- All built-in types available  
- Global parser registration affects all instances
- Backward compatible

**`genv.New(genv.WithRegistry(genv.NewDefaultRegistry()))`**
- Use when you want instance-specific custom types
- Isolated from global registration
- All built-in types included
- Good for libraries

**`genv.New(genv.WithRegistry(genv.NewRegistry()))`**
- Use for complete control over available types
- No built-in parsers (you must register everything)
- Maximum isolation
- Good for specialized use cases


### Environment-Specific Registries

```go
func NewConfigForEnvironment(environment string) (*genv.Genv, error) {
    var registry *genv.ParserRegistry
    
    switch environment {
    case "production":
        registry = NewProductionRegistry() // Strict validation
    case "development":
        registry = NewDevelopmentRegistry() // Lenient with defaults
    case "testing":
        registry = NewTestingRegistry() // Mock-friendly
    default:
        return nil, fmt.Errorf("unknown environment: %s", environment)
    }
    
    return genv.New(genv.WithRegistry(registry)), nil
}
```

## API Reference

### Registry Management

- `genv.NewRegistry()` - Creates empty registry
- `genv.NewDefaultRegistry()` - Creates registry with built-in parsers  
- `genv.WithRegistry(registry)` - Configure Genv instance with specific registry

### Parser Registration

- `genv.RegisterTypedParserOn[T](registry, parseFunc)` - Type-safe registration on specific registry
- `registry.RegisterParseFunc(reflect.Type, parseFunc)` - Low-level registration using reflection

### Type Parsing

- `genv.Type[T](var, &target)` - Parse single value using generic type inference
- `genv.NewType[T](var)` - Create and return new parsed value
- `genv.Types[T](var, &target)` - Parse slice of values
- `genv.NewTypes[T](var)` - Create and return new parsed slice

### Built-in Types

Default registry includes parsers for:
- `string`, `[]string`
- `bool`, `[]bool`
- `int`, `[]int`
- `float64`, `[]float64`
- `url.URL`, `[]url.URL`
- `uuid.UUID`, `[]uuid.UUID`
- `time.Time`, `[]time.Time` (RFC3339 format)

### Example

See the `example` package for a more complete demonstration of how this package can be used.
