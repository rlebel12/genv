# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is `genv`, a Go package for parsing environment variables with type safety and validation. It provides a fluent API for declaring environment variables with support for defaults, optional values, and various data types (string, int, bool, float64, URL, UUID, and slice variants).

## Development Commands

### Testing
- `go test` - Run all tests
- `go test -v` - Run tests with verbose output
- `go test ./...` - Run tests in all packages including example

### Coverage
- `make cover` - Generate test coverage report and open HTML coverage report in browser

### Building
- `go build` - Build the package
- `go get .` - Install dependencies

### Running Examples
- `cd example && go run .` - Run the example program (requires setting environment variables)

## Architecture

### Core Components

**Main Package (`genv.go`)**:
- `Genv` struct: Main configuration object that manages defaults, parsing, and parser registry
- `Var` struct: Represents a single environment variable with its parsing rules
- `fallback` struct: Handles default value logic with conditional allow functions
- `ParserRegistry` struct: Manages type-to-parser mappings with isolation between instances

**Key Design Patterns**:
- Fluent API: Methods return `*Var` to allow chaining (e.g., `env.Var("KEY").Default("value").Optional()`)
- Deferred parsing: Variable parsing functions are collected in `varFuncs` slice and executed during `Parse()`
- Pointer-based and value-based APIs: Both `String(&variable)` and `NewString()` patterns
- Generic parsing: `parse[T any]()` function handles type conversion with consistent error handling
- Registry-per-instance: Each `Genv` instance has its own `ParserRegistry` for type parsing isolation

### Type Support
The package supports these types with both single and slice variants:
- `string`, `[]string`
- `bool`, `[]bool` 
- `int`, `[]int`
- `float64`, `[]float64`
- `url.URL`, `[]url.URL`
- `uuid.UUID`, `[]uuid.UUID`

### Default Value System
- Global default control via `GENV_ALLOW_DEFAULT` environment variable
- Per-variable default overrides via `WithAllowDefault()` options
- `WithAllowDefaultAlways()` convenience method for always allowing defaults

### Array Parsing
- Configurable split delimiter (default: `,`, can be changed via `WithSplitKey()`)
- Empty values after splitting are automatically filtered out
- Example: `"1,2,3"` becomes `[]int{1, 2, 3}`

### Parser Registry System
The package uses a registry-based architecture for type parsing that provides isolation and extensibility:

**Registry Architecture**:
- Each `Genv` instance contains its own `ParserRegistry`
- Registries map Go types to parsing functions using reflection
- Default behavior: `New()` creates instances with all built-in parsers pre-registered
- Custom behavior: `New(WithRegistry(customRegistry))` allows using custom parser sets

**Registry Creation**:
- `NewRegistry()` - Creates empty registry for custom-only parsing
- `NewDefaultRegistry()` - Creates registry with all built-in parsers (string, bool, int, float64, url.URL, uuid.UUID, time.Time)
- `WithRegistry(registry)` - Option to configure `Genv` with specific registry


### Custom Type Support
The package supports registering custom type parsers for domain-specific types:

**Registration Methods**:
- `RegisterTypedParserOn[T](registry, parseFunc)` - Type-safe registration on specific registry
- `registry.RegisterParseFunc(reflect.Type, parseFunc)` - Reflection-based registration on specific registry

**Custom Type Example**:
```go
type UserID string

// Register custom parser on specific registry
registry := genv.NewRegistry()
genv.RegisterTypedParserOn(registry, func(s string) (UserID, error) {
    if s == "" {
        return "", errors.New("UserID cannot be empty")
    }
    return UserID("user_" + s), nil
})

// Use registry with Genv instance
env := genv.New(genv.WithRegistry(registry))
userID := env.Var("USER_ID").NewUserID() // This would require additional type methods
```

**Registry Isolation Benefits**:
- Different `Genv` instances can have completely different parser sets
- Test environments can use mock parsers without affecting production
- Microservices can have domain-specific type parsing
- No global state conflicts between different parts of application

## Testing Notes

The test suite uses testify/assert and covers:
- Basic type parsing and validation
- Default value behavior and conditional logic
- Optional variable handling
- Array parsing with custom delimiters
- Registry creation and isolation
- Custom type parser registration
- Error cases and edge conditions

**Registry Testing Patterns**:
- Test registry isolation by creating separate instances with different parsers
- Verify custom type registration using `RegisterTypedParserOn[T]()`
- Test empty registries fail appropriately when built-in types not available

When adding new features, follow the existing test patterns and ensure both success and failure cases are covered.