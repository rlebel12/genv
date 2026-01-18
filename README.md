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

### Basic Usage

The `Bind()`/`BindMany()` API provides a clean, ergonomic way to declare and parse environment variables with automatic type inference via generics:

```go
type Config struct {
    AppName     string
    Port        int
    Debug       bool
    Timeout     float64
    DatabaseURL url.URL
    Tags        []string
}

env := genv.New(
    genv.WithAllowDefault(func(*genv.Genv) (bool, error) { return true, nil }),
)

var config Config

// All variables registered and parsed in one call with type inference!
err := genv.Parse(env,
    genv.Bind("APP_NAME", &config.AppName).Default("MyApp"),
    genv.Bind("PORT", &config.Port).Default("8080"),
    genv.Bind("DEBUG", &config.Debug).Default("false"),
    genv.Bind("TIMEOUT", &config.Timeout).Default("30.5"),
    genv.Bind("DATABASE_URL", &config.DatabaseURL).Default("https://db.example.com"),
    genv.BindMany("TAGS", &config.Tags).Default("api,web,production"),
)
if err != nil {
    log.Fatal(err)
}
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

The package supports registering custom type parsers using functional options:

#### Registering Custom Types

Use `WithParser` to register custom types when creating a registry:

```go
// Define custom types
type UserID string
type Department string

// Create registry with custom parsers using functional options
registry := genv.NewDefaultRegistry(
    genv.WithParser(func(s string) (UserID, error) {
        if s == "" {
            return "", errors.New("UserID cannot be empty")
        }
        if !strings.HasPrefix(s, "user_") {
            return UserID("user_" + s), nil
        }
        return UserID(s), nil
    }),
    genv.WithParser(func(s string) (Department, error) {
        validDepts := map[string]bool{
            "engineering": true,
            "marketing":   true,
            "sales":       true,
        }
        dept := strings.ToLower(strings.TrimSpace(s))
        if !validDepts[dept] {
            return "", fmt.Errorf("invalid department: %s", s)
        }
        return Department(dept), nil
    }),
)

// Use custom registry with the simplified API
env := genv.New(
    genv.WithRegistry(registry),
    genv.WithAllowDefault(func(*genv.Genv) (bool, error) { return true, nil }),
)

var config struct {
    UserID     UserID
    Department Department
}

err := genv.Parse(env,
    genv.Bind("USER_ID", &config.UserID).Default("demo123"),
    genv.Bind("DEPARTMENT", &config.Department).Default("engineering"),
)
```

#### Custom Parser Registries

Create isolated parser registries for different contexts:

```go
// Create empty registry (no built-in parsers)
customRegistry := genv.NewRegistry(
    genv.WithParser(func(s string) (string, error) {
        return strings.ToUpper(s), nil // Custom string behavior
    }),
)

// Create registry with default parsers plus custom types
defaultRegistry := genv.NewDefaultRegistry(
    genv.WithParser(func(s string) (UserID, error) {
        return UserID("user_" + s), nil
    }),
)

// Use custom registry
env := genv.New(genv.WithRegistry(customRegistry))
```

#### Registry Isolation

Different `Genv` instances can have completely different parsing behavior:

```go
type UserID string

// Registry for production with strict validation
prodRegistry := genv.NewDefaultRegistry(
    genv.WithParser(func(s string) (UserID, error) {
        if !strings.HasPrefix(s, "user_") {
            return "", errors.New("invalid UserID format")
        }
        return UserID(s), nil
    }),
)

// Registry for testing with lenient validation
testRegistry := genv.NewDefaultRegistry(
    genv.WithParser(func(s string) (UserID, error) {
        return UserID("user_" + s), nil // Always add prefix
    }),
)

prodEnv := genv.New(genv.WithRegistry(prodRegistry))
testEnv := genv.New(genv.WithRegistry(testRegistry))
```

## Advanced Usage

### Working with Custom Types

The registry system with `WithParser` enables full end-to-end custom type parsing:

```go
type ServicePort int
type LogLevel string

// Create registry with custom parsers using functional options
registry := genv.NewDefaultRegistry(
    genv.WithParser(func(s string) (ServicePort, error) {
        port, err := strconv.Atoi(s)
        if err != nil {
            return ServicePort(0), err
        }
        if port < 1024 || port > 65535 {
            return ServicePort(0), errors.New("port out of range")
        }
        return ServicePort(port), nil
    }),
    genv.WithParser(func(s string) (LogLevel, error) {
        level := strings.ToUpper(s)
        if level == "DEBUG" || level == "INFO" || level == "WARN" || level == "ERROR" {
            return LogLevel(level), nil
        }
        return "", errors.New("invalid log level")
    }),
)

// Use custom types with simplified API
env := genv.New(
    genv.WithRegistry(registry),
    genv.WithAllowDefault(func(*genv.Genv) (bool, error) { return true, nil }),
)

var config struct {
    Port     ServicePort
    LogLevel LogLevel
}

err := genv.Parse(env,
    genv.Bind("SERVICE_PORT", &config.Port).Default("8080"),
    genv.Bind("LOG_LEVEL", &config.LogLevel).Default("INFO"),
)
if err != nil {
    log.Fatal(err)
}
```

### Custom Type Slices

Custom types work with slice parsing using `BindMany`:

```go
type Priority int

registry := genv.NewDefaultRegistry(
    genv.WithParser(func(s string) (Priority, error) {
        switch strings.ToLower(s) {
        case "low": return Priority(1), nil
        case "medium": return Priority(2), nil
        case "high": return Priority(3), nil
        default: return Priority(0), errors.New("invalid priority")
        }
    }),
)

env := genv.New(
    genv.WithRegistry(registry),
    genv.WithAllowDefault(func(*genv.Genv) (bool, error) { return true, nil }),
)

var priorities []Priority
err := genv.Parse(env,
    genv.BindMany("TASK_PRIORITIES", &priorities).Default("low,high,medium"),
)
```

### Optional and Default with Custom Types

Custom types fully support optional and default value patterns:

```go
type Environment string

registry := genv.NewDefaultRegistry(
    genv.WithParser(func(s string) (Environment, error) {
        return Environment(strings.ToLower(s)), nil
    }),
)

env := genv.New(genv.WithRegistry(registry))

var config struct {
    OptionalEnv Environment
    EnvWithDefault Environment
}

err := genv.Parse(env,
    genv.Bind("APP_ENV", &config.OptionalEnv).Optional(), // Won't error if missing
    genv.Bind("APP_ENV_DEFAULT", &config.EnvWithDefault).Default("development"),
)
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

### Core Functions

- `genv.Bind[T](key, &target)` - Create VarFunc that binds environment variable to target pointer
- `genv.BindMany[T](key, &target)` - Create VarFunc that binds comma-separated values to slice pointer
- `genv.Parse(env, ...VarFunc)` - Execute all VarFuncs and parse environment variables

### Registry Management

- `genv.NewRegistry(opts...)` - Creates empty registry with optional parsers
- `genv.NewDefaultRegistry(opts...)` - Creates registry with built-in parsers plus optional custom parsers
- `genv.WithRegistry(registry)` - Configure Genv instance with specific registry

### Parser Registration

- `genv.WithParser[T](parseFunc)` - Functional option to register type-safe parser for type T

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
