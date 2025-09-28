# Fluent API with Deferred Execution Pattern

## Overview

The Fluent API with Deferred Execution Pattern combines method chaining for configuration with delayed execution for validation. This pattern enables clean, declarative environment variable configuration while providing comprehensive validation before any parsing occurs.

## Pattern Structure

### Core Components

1. **Fluent Interface**: Method chaining that returns `*Var` for configuration
2. **Deferred Collection**: Operations stored in `varFuncs` slice for batch execution
3. **Validation Gate**: All configuration validated before any parsing begins
4. **Execution Phase**: Batch execution of all collected operations

### Key Files and Locations

- `genv.go:145-162` - Fluent method chaining implementation
- `genv.go:20` - `varFuncs` slice for deferred execution
- `genv.go:456-462` - Parse() method executing collected functions

## Implementation Details

### Declaration Phase - Fluent Configuration

```go
env := genv.New()

// Method chaining builds configuration
userID := env.Var("USER_ID").
    Default("guest").
    Optional().
    NewString()

port := env.Var("PORT").
    Default("8080").
    NewInt()
```

### Collection Phase - Operation Storage

```go
// Each fluent call adds to varFuncs slice
type Genv struct {
    varFuncs []func() error  // Operations collected here
    // ... other fields
}

func (v *Var) NewString() string {
    // Store parsing operation for later execution
    v.genv.varFuncs = append(v.genv.varFuncs, func() error {
        return v.parseIntoTarget(/* parsing logic */)
    })
    return v.target // Return placeholder
}
```

### Execution Phase - Batch Processing

```go
// All operations executed together
func (g *Genv) Parse() error {
    for _, varFunc := range g.varFuncs {
        if err := varFunc(); err != nil {
            return err // Fail fast on any error
        }
    }
    return nil
}
```

## Architectural Benefits

### 1. Early Validation
- All environment variable declarations validated before parsing
- Configuration errors detected before application startup
- Prevents partial parsing states that could leave application in inconsistent state

### 2. Clean Separation of Concerns
- **Declaration Phase**: What variables are needed and their rules
- **Execution Phase**: Actually reading and parsing environment values
- **Validation Phase**: Ensuring all requirements are met

### 3. Atomic Operations
- Either all environment variables parse successfully or none do
- No partial configuration states
- Clear error reporting for missing or invalid variables

### 4. Readable Configuration
- Declarative syntax clearly expresses intent
- Method chaining reduces boilerplate
- Self-documenting variable requirements

## Usage Patterns

### Basic Fluent Configuration
```go
env := genv.New()

dbURL := env.Var("DATABASE_URL").NewString()
port := env.Var("PORT").Default("8080").NewInt()
debug := env.Var("DEBUG").Default("false").NewBool()

// Parse all at once
if err := env.Parse(); err != nil {
    log.Fatal("Configuration error:", err)
}
```

### Complex Configuration with Validation
```go
env := genv.New()

// Complex chaining with multiple options
apiKey := env.Var("API_KEY").
    WithAllowDefaultAlways().
    Default("development-key").
    NewString()

endpoints := env.Var("ENDPOINTS").
    WithSplitKey(";").
    Default("http://localhost:8080").
    NewStringSlice()

timeout := env.Var("TIMEOUT").
    Default("30s").
    NewDuration()

// Single validation and parsing step
if err := env.Parse(); err != nil {
    return fmt.Errorf("environment configuration failed: %w", err)
}
```

### Error Handling Strategy
```go
func configureFromEnv() (*Config, error) {
    env := genv.New()

    config := &Config{
        Port:    env.Var("PORT").Default("8080").NewInt(),
        Host:    env.Var("HOST").Default("localhost").NewString(),
        Timeout: env.Var("TIMEOUT").Default("30s").NewDuration(),
    }

    // All parsing happens here - fail fast if any issues
    if err := env.Parse(); err != nil {
        return nil, fmt.Errorf("configuration parsing failed: %w", err)
    }

    return config, nil
}
```

## Design Decisions

### Why Deferred Execution?

1. **Validation Before Side Effects**: Detect all configuration issues before starting application
2. **Atomic Configuration**: All-or-nothing parsing prevents inconsistent states
3. **Better Error Reporting**: Can provide comprehensive error messages about all missing variables
4. **Testability**: Can mock environment completely before parsing

### Method Chaining vs Builder Pattern

The fluent interface is preferred over separate builder because:
- **Immediate Declaration**: Variables are declared where they're used
- **Type Safety**: Return types enforce correct method sequences
- **Readability**: Configuration reads like natural language
- **IDE Support**: Better autocomplete and error detection

### Function Collection vs Immediate Execution

Deferred execution through function collection provides:
- **Consistency**: All variables parsed with same environment snapshot
- **Performance**: Single pass through environment variables
- **Error Handling**: Centralized error processing and reporting
- **Testing**: Easier to mock entire environment parsing process

## Testing Strategies

### Testing Deferred Execution
```go
func TestDeferredExecution(t *testing.T) {
    env := genv.New()

    // Configure multiple variables
    var1 := env.Var("VAR1").Default("default1").NewString()
    var2 := env.Var("VAR2").Default("default2").NewString()

    // Set environment after configuration
    t.Setenv("VAR1", "value1")
    t.Setenv("VAR2", "value2")

    // Parse should use current environment values
    err := env.Parse()
    assert.NoError(t, err)
    assert.Equal(t, "value1", var1)
    assert.Equal(t, "value2", var2)
}
```

### Testing Error Conditions
```go
func TestParsingFailure(t *testing.T) {
    env := genv.New()

    port := env.Var("PORT").NewInt()
    // Don't set PORT environment variable

    err := env.Parse()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "PORT")
}
```

## Anti-Patterns to Avoid

1. **Parsing Before All Variables Declared**: Always call Parse() after all variable declarations
2. **Multiple Parse() Calls**: Parse() should be called exactly once per Genv instance
3. **Accessing Variables Before Parse()**: Variable values are only valid after Parse() succeeds
4. **Ignoring Parse() Errors**: Always handle Parse() errors appropriately

## Performance Considerations

### Memory Usage
- Function closures in `varFuncs` hold references to variables
- Memory released after Parse() completes
- Consider scope lifetime for long-running applications

### Execution Efficiency
- Single pass through environment variables
- No redundant parsing of same environment variable
- Batch validation more efficient than individual checks

## Related Patterns

- **Command Pattern**: Each operation stored as executable command
- **Batch Processing**: Operations collected and executed together
- **Builder Pattern**: Fluent interface for object construction
- **Fail-Fast Pattern**: Early validation prevents runtime errors