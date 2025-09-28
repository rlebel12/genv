# Generic Type Parsing Pattern

## Overview

The Generic Type Parsing Pattern combines Go generics with reflection to provide type-safe environment variable parsing with runtime type dispatch. This pattern enables compile-time type safety while maintaining flexibility for custom type registration and extension.

## Pattern Structure

### Core Components

1. **Generic Parsing Function**: Type-safe `parseOne[T any]()` function
2. **Reflection-Based Dispatch**: Runtime type lookup in parser registry
3. **Type Safety Bridge**: Compile-time generics with runtime parser functions
4. **Error Context**: Consistent error handling with type information

### Key Files and Locations

- `genv.go:487-513` - Generic parsing implementation
- `genv.go:370-445` - ParserRegistry type dispatch system
- `genv.go:170-380` - Type-specific API methods using generic backend

## Implementation Details

### Generic Parsing Core

```go
// Generic function provides compile-time type safety
func (v *Var) parseOne[T any]() (T, error) {
    var zero T

    // Use reflection to get runtime type information
    targetType := reflect.TypeOf((*T)(nil)).Elem()

    // Look up parser function in registry
    parser, exists := v.genv.registry.GetParseFunc(targetType)
    if !exists {
        return zero, fmt.Errorf("no parser registered for type %T", zero)
    }

    // Get environment value
    envValue := os.Getenv(v.key)

    // Handle default values
    if envValue == "" && v.fallback.hasDefault() {
        envValue = v.fallback.defaultValue
    }

    // Parse with type-specific function
    parsed, err := parser(envValue)
    if err != nil {
        return zero, fmt.Errorf(errFmtInvalidVar, v.key, envValue, err)
    }

    // Type assertion back to generic type
    result, ok := parsed.(T)
    if !ok {
        return zero, fmt.Errorf("parser returned wrong type for %s", v.key)
    }

    return result, nil
}
```

### Type-Specific API Integration

```go
// Each type-specific method uses the generic backend
func (v *Var) NewString() string {
    var result string
    v.genv.varFuncs = append(v.genv.varFuncs, func() error {
        parsed, err := v.parseOne[string]()  // Generic parsing
        if err != nil {
            return err
        }
        result = parsed
        return nil
    })
    return result
}

func (v *Var) NewInt() int {
    var result int
    v.genv.varFuncs = append(v.genv.varFuncs, func() error {
        parsed, err := v.parseOne[int]()     // Same generic function
        if err != nil {
            return err
        }
        result = parsed
        return nil
    })
    return result
}
```

### Parser Registry Integration

```go
// Registry maps types to parsing functions
type ParserRegistry struct {
    parsers map[reflect.Type]func(string) (any, error)
}

// Type-safe registration using generics
func RegisterTypedParserOn[T any](registry *ParserRegistry, parseFunc func(string) (T, error)) {
    targetType := reflect.TypeOf((*T)(nil)).Elem()

    // Wrap type-safe function for registry storage
    registry.parsers[targetType] = func(s string) (any, error) {
        return parseFunc(s)
    }
}
```

## Architectural Benefits

### 1. Compile-Time Type Safety
- Generic functions catch type errors at compile time
- IDE support for autocompletion and error detection
- No runtime type casting errors in user code

### 2. Runtime Flexibility
- Dynamic parser registration for custom types
- Extension without modifying core library
- Plugin-like architecture for domain-specific types

### 3. Consistent Error Handling
- Standardized error format across all types
- Context-rich error messages with variable names
- Type information included in error messages

### 4. Performance Optimization
- Single reflection lookup per type, not per variable
- Type-specific parsers optimized for each type
- No boxing/unboxing overhead in user code

## Usage Patterns

### Built-in Type Usage

```go
env := genv.New()

// Generics infer types automatically
port := env.Var("PORT").Default("8080").NewInt()           // int
timeout := env.Var("TIMEOUT").Default("30s").NewDuration() // time.Duration
url := env.Var("API_URL").NewURL()                         // *url.URL
```

### Custom Type Registration

```go
type UserID string

func parseUserID(s string) (UserID, error) {
    if s == "" {
        return "", errors.New("UserID cannot be empty")
    }
    if !strings.HasPrefix(s, "user_") {
        return "", errors.New("UserID must start with 'user_'")
    }
    return UserID(s), nil
}

// Register custom parser
registry := genv.NewRegistry()
genv.RegisterTypedParserOn(registry, parseUserID)

// Use with type safety
env := genv.New(genv.WithRegistry(registry))
userID := env.Var("USER_ID").parseOne[UserID]() // Type-safe parsing
```

### Complex Type Parsing

```go
type Config struct {
    Host string
    Port int
}

func parseConfig(s string) (Config, error) {
    parts := strings.Split(s, ":")
    if len(parts) != 2 {
        return Config{}, errors.New("config must be in format host:port")
    }

    port, err := strconv.Atoi(parts[1])
    if err != nil {
        return Config{}, fmt.Errorf("invalid port: %w", err)
    }

    return Config{Host: parts[0], Port: port}, nil
}

// Register and use complex type
registry := genv.NewRegistry()
genv.RegisterTypedParserOn(registry, parseConfig)

env := genv.New(genv.WithRegistry(registry))
config := env.Var("SERVER_CONFIG").parseOne[Config]()
```

## Design Decisions

### Why Generics + Reflection?

1. **Best of Both Worlds**: Compile-time safety with runtime flexibility
2. **Extension Path**: Can register parsers for any type, including external types
3. **Performance**: Type-specific parsers can be optimized individually
4. **Compatibility**: Works with existing Go types without modification

### Type System Design Choices

**Generic Constraints**: Uses `any` constraint rather than specific interfaces
- Enables parsing of any Go type
- No interface implementation required
- Compatible with primitive types and external types

**Return Type Strategy**: Parser functions return concrete types, not interfaces
- Eliminates type assertions in user code
- Better error messages with specific type information
- Enables type-specific validation and formatting

### Error Handling Integration

```go
const errFmtInvalidVar = "invalid value for environment variable %s: %q: %w"

// Consistent error formatting across all types
func (v *Var) parseOne[T any]() (T, error) {
    // ... parsing logic ...

    if err != nil {
        var zero T
        return zero, fmt.Errorf(errFmtInvalidVar, v.key, envValue, err)
    }

    return result, nil
}
```

## Testing Strategies

### Generic Function Testing

```go
func TestGenericParsing(t *testing.T) {
    env := genv.New()

    tests := []struct {
        name     string
        envVar   string
        envValue string
        parser   func() (any, error)
        expected any
    }{
        {
            name: "string parsing",
            envVar: "TEST_STRING",
            envValue: "hello",
            parser: func() (any, error) {
                v := env.Var("TEST_STRING")
                return v.parseOne[string]()
            },
            expected: "hello",
        },
        {
            name: "int parsing",
            envVar: "TEST_INT",
            envValue: "42",
            parser: func() (any, error) {
                v := env.Var("TEST_INT")
                return v.parseOne[int]()
            },
            expected: 42,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Setenv(tt.envVar, tt.envValue)

            result, err := tt.parser()
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Custom Type Testing

```go
func TestCustomTypeRegistration(t *testing.T) {
    type CustomID string

    parseCustomID := func(s string) (CustomID, error) {
        if s == "" {
            return "", errors.New("CustomID cannot be empty")
        }
        return CustomID("custom_" + s), nil
    }

    registry := genv.NewRegistry()
    genv.RegisterTypedParserOn(registry, parseCustomID)

    env := genv.New(genv.WithRegistry(registry))
    v := env.Var("CUSTOM_ID")

    t.Setenv("CUSTOM_ID", "test")

    result, err := v.parseOne[CustomID]()
    assert.NoError(t, err)
    assert.Equal(t, CustomID("custom_test"), result)
}
```

## Performance Considerations

### Type Lookup Optimization
- Registry uses `reflect.Type` as map key for O(1) lookup
- Type reflection performed once per variable type, not per variable
- Parser functions cached in registry for reuse

### Memory Efficiency
- Generic functions compiled separately for each type
- No interface{} boxing in hot paths
- Type-specific optimizations possible in parser functions

### Runtime Overhead
- Single reflection call per parsing operation
- Type assertion overhead eliminated through generics
- Parser functions can be optimized for specific types

## Anti-Patterns to Avoid

1. **Type Assertion in User Code**: Use generics instead of interface{} returns
2. **Multiple Parser Registration**: Don't register multiple parsers for same type
3. **Generic Type Parameters in Structs**: Keep generics at function level
4. **Runtime Type Switching**: Use registry dispatch instead of type switches

## Related Patterns

- **Registry Pattern**: Type-to-parser mapping
- **Strategy Pattern**: Different parsing strategies per type
- **Template Method**: Generic parsing template with type-specific implementations
- **Type-Safe Builder**: Generic builder methods for fluent APIs