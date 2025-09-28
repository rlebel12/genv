# Dual API Pattern

## Overview

The Dual API Pattern provides two complementary interfaces for environment variable parsing: a pointer-based API for direct assignment and a value-based API for functional programming styles. This pattern maximizes flexibility while maintaining type safety and consistent error handling.

## Pattern Structure

### Core Components

1. **Pointer-Based API**: Direct assignment to existing variables using pointers
2. **Value-Based API**: Returns parsed values for assignment or chaining
3. **Unified Backend**: Both APIs use the same underlying parsing infrastructure
4. **Type Safety**: Generic implementation ensures compile-time type checking

### Key Files and Locations

- `genv.go:170-270` - Pointer-based API implementations (String, Int, Bool, etc.)
- `genv.go:280-380` - Value-based API implementations (NewString, NewInt, NewBool, etc.)
- `genv.go:487-513` - Shared generic parsing backend

## Implementation Details

### Pointer-Based API

```go
// Direct assignment to existing variables
var (
    dbURL    string
    port     int
    enabled  bool
)

env := genv.New()
env.Var("DATABASE_URL").String(&dbURL)
env.Var("PORT").Default("8080").Int(&port)
env.Var("FEATURE_ENABLED").Default("false").Bool(&enabled)

if err := env.Parse(); err != nil {
    log.Fatal(err)
}
// Variables are now populated
```

### Value-Based API

```go
// Functional assignment with return values
env := genv.New()

dbURL := env.Var("DATABASE_URL").NewString()
port := env.Var("PORT").Default("8080").NewInt()
enabled := env.Var("FEATURE_ENABLED").Default("false").NewBool()

if err := env.Parse(); err != nil {
    log.Fatal(err)
}
// Variables are now populated
```

### Unified Generic Backend

```go
// Both APIs converge on shared parsing logic
func (v *Var) parseOne[T any]() (T, error) {
    parser, exists := v.genv.registry.GetParseFunc(reflect.TypeOf((*T)(nil)).Elem())
    if !exists {
        var zero T
        return zero, fmt.Errorf("no parser registered for type %T", zero)
    }

    // Unified parsing with consistent error handling
    return parseWithType[T](parser, envValue)
}
```

## Architectural Benefits

### 1. API Flexibility
- **Pointer API**: Integrates with existing variable declarations
- **Value API**: Supports functional programming and configuration structs
- **Choose Appropriate Style**: Different patterns for different use cases

### 2. Migration Support
- **Legacy Code**: Pointer API works with existing variable declarations
- **Modern Code**: Value API enables more functional approaches
- **Incremental Adoption**: Can mix both styles within same application

### 3. Type Safety Consistency
- Both APIs use same generic parsing infrastructure
- Compile-time type checking for both approaches
- Consistent error handling and validation

### 4. Performance Equivalence
- Same underlying parsing performance
- No overhead from API choice
- Single implementation reduces maintenance burden

## Usage Patterns

### Struct Configuration with Pointer API

```go
type Config struct {
    DatabaseURL string
    Port        int
    Debug       bool
    Endpoints   []string
}

func LoadConfig() (*Config, error) {
    config := &Config{}
    env := genv.New()

    // Direct assignment to struct fields
    env.Var("DATABASE_URL").String(&config.DatabaseURL)
    env.Var("PORT").Default("8080").Int(&config.Port)
    env.Var("DEBUG").Default("false").Bool(&config.Debug)
    env.Var("ENDPOINTS").WithSplitKey(";").StringSlice(&config.Endpoints)

    if err := env.Parse(); err != nil {
        return nil, err
    }

    return config, nil
}
```

### Functional Configuration with Value API

```go
func LoadConfig() (*Config, error) {
    env := genv.New()

    return &Config{
        DatabaseURL: env.Var("DATABASE_URL").NewString(),
        Port:        env.Var("PORT").Default("8080").NewInt(),
        Debug:       env.Var("DEBUG").Default("false").NewBool(),
        Endpoints:   env.Var("ENDPOINTS").WithSplitKey(";").NewStringSlice(),
    }, env.Parse()
}
```

### Mixed API Usage

```go
func LoadConfig() (*Config, error) {
    env := genv.New()
    config := &Config{}

    // Use pointer API for complex types
    env.Var("DATABASE_URL").URL(&config.DatabaseURL)
    env.Var("USER_ID").UUID(&config.UserID)

    // Use value API for simple assignments
    config.Port = env.Var("PORT").Default("8080").NewInt()
    config.Debug = env.Var("DEBUG").Default("false").NewBool()

    return config, env.Parse()
}
```

### Variable Declaration Patterns

```go
// Pattern 1: Pre-declared variables (Pointer API)
var (
    dbURL  string
    port   int
    debug  bool
)

func init() {
    env := genv.New()
    env.Var("DB_URL").String(&dbURL)
    env.Var("PORT").Default("8080").Int(&port)
    env.Var("DEBUG").Default("false").Bool(&debug)

    if err := env.Parse(); err != nil {
        log.Fatal(err)
    }
}

// Pattern 2: Inline declaration (Value API)
func getConfig() Config {
    env := genv.New()
    defer func() {
        if err := env.Parse(); err != nil {
            log.Fatal(err)
        }
    }()

    return Config{
        DBURL: env.Var("DB_URL").NewString(),
        Port:  env.Var("PORT").Default("8080").NewInt(),
        Debug: env.Var("DEBUG").Default("false").NewBool(),
    }
}
```

## Design Decisions

### Why Both APIs?

1. **Different Use Cases**: Pointer API for existing code, Value API for new code
2. **Team Preferences**: Some teams prefer functional style, others imperative
3. **Context Suitability**: Struct initialization vs variable assignment
4. **Migration Path**: Easier adoption when multiple patterns supported

### Implementation Strategy

Rather than duplicate logic, both APIs share:
- **Generic Parsing Functions**: `parseOne[T]()` used by both
- **Validation Logic**: Same rules apply regardless of API
- **Error Handling**: Consistent error types and messages
- **Registry Access**: Same parser registry for both approaches

### Type System Integration

```go
// Both APIs leverage same generic type system
func (v *Var) String(target *string) {
    v.genv.varFuncs = append(v.genv.varFuncs, func() error {
        value, err := v.parseOne[string]()  // Shared parsing
        if err != nil {
            return err
        }
        *target = value  // Pointer assignment
        return nil
    })
}

func (v *Var) NewString() string {
    var result string
    v.genv.varFuncs = append(v.genv.varFuncs, func() error {
        value, err := v.parseOne[string]()  // Same shared parsing
        if err != nil {
            return err
        }
        result = value  // Value assignment
        return nil
    })
    return result
}
```

## Testing Strategies

### API Equivalence Testing

```go
func TestAPIEquivalence(t *testing.T) {
    t.Setenv("TEST_VAR", "42")

    // Test pointer API
    var ptrResult int
    env1 := genv.New()
    env1.Var("TEST_VAR").Int(&ptrResult)
    err1 := env1.Parse()

    // Test value API
    env2 := genv.New()
    valueResult := env2.Var("TEST_VAR").NewInt()
    err2 := env2.Parse()

    // Verify equivalent behavior
    assert.NoError(t, err1)
    assert.NoError(t, err2)
    assert.Equal(t, ptrResult, valueResult)
}
```

### Error Handling Consistency

```go
func TestErrorConsistency(t *testing.T) {
    t.Setenv("INVALID_INT", "not-a-number")

    var ptrResult int
    env1 := genv.New()
    env1.Var("INVALID_INT").Int(&ptrResult)
    err1 := env1.Parse()

    env2 := genv.New()
    valueResult := env2.Var("INVALID_INT").NewInt()
    err2 := env2.Parse()

    // Errors should be equivalent
    assert.Error(t, err1)
    assert.Error(t, err2)
    assert.Contains(t, err1.Error(), "INVALID_INT")
    assert.Contains(t, err2.Error(), "INVALID_INT")
}
```

## Performance Considerations

### Memory Usage
- **Pointer API**: No additional memory allocation for result storage
- **Value API**: Temporary variables created during parsing
- **Minimal Overhead**: Difference negligible for typical use cases

### Execution Performance
- **Identical Backend**: Both APIs use same parsing functions
- **No Performance Penalty**: API choice doesn't affect runtime performance
- **Compiler Optimization**: Generic functions optimized identically

## Anti-Patterns to Avoid

1. **Mixing APIs for Same Variable**: Choose one API style per variable for consistency
2. **Assuming API Differences**: Both APIs have identical parsing behavior
3. **Performance Micro-Optimization**: API choice should be based on readability, not performance
4. **Inconsistent Error Handling**: Handle errors the same way regardless of API choice

## Related Patterns

- **Builder Pattern**: Both APIs support fluent configuration
- **Strategy Pattern**: Choose API strategy based on context
- **Adapter Pattern**: Pointer API adapts to existing variable patterns
- **Template Method**: Shared parsing template with different assignment strategies