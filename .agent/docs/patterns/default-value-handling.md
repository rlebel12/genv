# Default Value Handling with Conditional Logic

## Overview

The Default Value Handling Pattern provides a sophisticated fallback system with global and per-variable default value control. This pattern enables conditional default behavior based on environment configuration while maintaining explicit control over when defaults are applied.

## Pattern Structure

### Core Components

1. **Fallback System**: `fallback` struct managing default values and conditions
2. **Global Default Control**: `GENV_ALLOW_DEFAULT` environment variable
3. **Per-Variable Override**: `WithAllowDefault()` and `WithAllowDefaultAlways()` options
4. **Conditional Logic**: Complex decision tree for when to apply defaults

### Key Files and Locations

- `genv.go:50-80` - `fallback` struct implementation
- `genv.go:95-120` - Default value application logic
- `genv.go:515-530` - Default control options
- `genv_test.go:300-400` - Default value behavior testing

## Implementation Details

### Fallback System Architecture

```go
// Fallback manages default value logic
type fallback struct {
    defaultValue string
    allowFn      func() bool  // Conditional allow function
}

// Check if default should be used
func (f fallback) hasDefault() bool {
    if f.defaultValue == "" {
        return false
    }
    if f.allowFn == nil {
        return true
    }
    return f.allowFn()
}

// Get default value if allowed
func (f fallback) getDefault() string {
    if f.hasDefault() {
        return f.defaultValue
    }
    return ""
}
```

### Global Default Control

```go
// Global default control via environment variable
func checkGlobalDefaultAllowed() bool {
    allowDefault := os.Getenv("GENV_ALLOW_DEFAULT")
    if allowDefault == "" {
        return true  // Default behavior: allow defaults
    }

    // Parse boolean value
    allowed, err := strconv.ParseBool(allowDefault)
    if err != nil {
        return true  // Default to allowing on parse error
    }
    return allowed
}

// Global allow function
var globalAllowDefault = checkGlobalDefaultAllowed
```

### Per-Variable Default Options

```go
// Always allow defaults for this variable
func WithAllowDefaultAlways() Opt[*Var] {
    return func(v *Var) {
        v.fallback.allowFn = func() bool {
            return true
        }
    }
}

// Custom conditional logic for defaults
func WithAllowDefault(allowFn func() bool) Opt[*Var] {
    return func(v *Var) {
        v.fallback.allowFn = allowFn
    }
}

// Respect global default setting (default behavior)
func WithAllowDefaultGlobal() Opt[*Var] {
    return func(v *Var) {
        v.fallback.allowFn = globalAllowDefault
    }
}
```

### Default Application Logic

```go
// Complex default application in parsing
func (v *Var) parseWithDefaults() (string, error) {
    envValue := os.Getenv(v.key)

    // Apply default if environment variable is empty
    if envValue == "" {
        if v.fallback.hasDefault() {
            defaultValue := v.fallback.getDefault()
            return defaultValue, nil
        }

        // No default available and variable is required
        if !v.optional {
            return "", fmt.Errorf("required environment variable %s not set", v.key)
        }

        // Optional variable with no default
        return "", nil
    }

    // Use environment value
    return envValue, nil
}
```

## Architectural Benefits

### 1. Flexible Default Control
- **Global Control**: Disable all defaults via `GENV_ALLOW_DEFAULT=false`
- **Per-Variable Override**: Individual variables can override global setting
- **Conditional Logic**: Complex conditions for when defaults apply

### 2. Environment-Aware Behavior
- **Production Safety**: Can disable defaults in production environments
- **Development Convenience**: Defaults enabled for local development
- **Testing Isolation**: Different default behavior per test environment

### 3. Explicit Configuration
- **Clear Intent**: Default behavior is explicitly configured
- **Auditable**: Default usage can be tracked and controlled
- **Predictable**: Consistent behavior across different environments

### 4. Security Benefits
- **Production Hardening**: Prevents accidental use of development defaults
- **Configuration Validation**: Forces explicit environment configuration
- **Fail-Fast**: Missing required variables detected early

## Usage Patterns

### Basic Default Values

```go
env := genv.New()

// Simple defaults (subject to global control)
port := env.Var("PORT").Default("8080").NewInt()
host := env.Var("HOST").Default("localhost").NewString()
debug := env.Var("DEBUG").Default("false").NewBool()

if err := env.Parse(); err != nil {
    log.Fatal(err)
}
```

### Global Default Control

```bash
# Allow defaults (default behavior)
export GENV_ALLOW_DEFAULT=true

# Disable all defaults (production)
export GENV_ALLOW_DEFAULT=false

# Run application
./app
```

### Per-Variable Default Override

```go
env := genv.New()

// Always use default regardless of global setting
apiKey := env.Var("API_KEY").
    Default("development-key").
    WithAllowDefaultAlways().
    NewString()

// Use default only in development environment
dbURL := env.Var("DATABASE_URL").
    Default("postgres://localhost/dev").
    WithAllowDefault(func() bool {
        return os.Getenv("ENV") == "development"
    }).
    NewString()

// Respect global default setting (default behavior)
timeout := env.Var("TIMEOUT").
    Default("30s").
    WithAllowDefaultGlobal().
    NewDuration()

if err := env.Parse(); err != nil {
    log.Fatal(err)
}
```

### Environment-Specific Configuration

```go
// Environment-aware default configuration
func createEnvConfig() *genv.Genv {
    env := genv.New()

    // Different defaults based on environment
    environment := os.Getenv("ENVIRONMENT")

    switch environment {
    case "production":
        // No defaults in production
        return env.WithGlobalDefaultPolicy(func() bool { return false })

    case "development":
        // Liberal defaults for development
        return env.WithGlobalDefaultPolicy(func() bool { return true })

    case "testing":
        // Conditional defaults for testing
        return env.WithGlobalDefaultPolicy(func() bool {
            return os.Getenv("USE_TEST_DEFAULTS") == "true"
        })

    default:
        // Conservative defaults for unknown environments
        return env.WithGlobalDefaultPolicy(func() bool { return false })
    }
}
```

### Complex Conditional Logic

```go
env := genv.New()

// Default only during business hours
apiEndpoint := env.Var("API_ENDPOINT").
    Default("http://staging-api.com").
    WithAllowDefault(func() bool {
        now := time.Now()
        hour := now.Hour()
        return hour >= 9 && hour <= 17  // 9 AM to 5 PM
    }).
    NewString()

// Default only if debugging enabled
verboseLogging := env.Var("VERBOSE_LOGGING").
    Default("true").
    WithAllowDefault(func() bool {
        return os.Getenv("DEBUG") == "true"
    }).
    NewBool()

// Default only for specific users
userSpecificConfig := env.Var("USER_CONFIG").
    Default("default-config.json").
    WithAllowDefault(func() bool {
        user := os.Getenv("USER")
        return user == "developer" || user == "admin"
    }).
    NewString()

if err := env.Parse(); err != nil {
    log.Fatal(err)
}
```

## Design Decisions

### Why Conditional Defaults?

1. **Security**: Prevent accidental use of development defaults in production
2. **Flexibility**: Different behavior based on runtime conditions
3. **Explicit Control**: Force explicit configuration when needed
4. **Testing**: Predictable behavior across different test scenarios

### Global vs Per-Variable Control

**Global Control Benefits**:
- **Simple Override**: Single environment variable controls all defaults
- **Production Safety**: Easy to disable all defaults in production
- **Consistent Behavior**: All variables follow same default policy

**Per-Variable Control Benefits**:
- **Fine-Grained Control**: Individual variables can have different policies
- **Complex Logic**: Variable-specific conditions for default usage
- **Gradual Migration**: Can migrate variables individually

### Default Value Storage

```go
// String storage for all default values
type fallback struct {
    defaultValue string    // Always stored as string
    allowFn      func() bool
}
```

**Why String Storage?**:
- **Consistent Interface**: Same API regardless of target type
- **Type-Safe Parsing**: Default values parsed with same logic as environment values
- **Validation**: Default values subject to same validation as environment values

## Testing Strategies

### Global Default Testing

```go
func TestGlobalDefaultControl(t *testing.T) {
    tests := []struct {
        name                 string
        globalDefaultSetting string
        expectDefault        bool
    }{
        {
            name:                 "defaults allowed",
            globalDefaultSetting: "true",
            expectDefault:        true,
        },
        {
            name:                 "defaults disabled",
            globalDefaultSetting: "false",
            expectDefault:        false,
        },
        {
            name:                 "unset uses defaults",
            globalDefaultSetting: "",
            expectDefault:        true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.globalDefaultSetting != "" {
                t.Setenv("GENV_ALLOW_DEFAULT", tt.globalDefaultSetting)
            }

            env := genv.New()
            result := env.Var("TEST_VAR").Default("default-value").NewString()

            err := env.Parse()
            assert.NoError(t, err)

            if tt.expectDefault {
                assert.Equal(t, "default-value", result)
            } else {
                assert.Equal(t, "", result)
            }
        })
    }
}
```

### Per-Variable Override Testing

```go
func TestPerVariableDefaultOverride(t *testing.T) {
    t.Setenv("GENV_ALLOW_DEFAULT", "false")  // Globally disable defaults

    env := genv.New()

    // This variable should use default despite global setting
    alwaysDefault := env.Var("ALWAYS_DEFAULT").
        Default("always").
        WithAllowDefaultAlways().
        NewString()

    // This variable should not use default (respects global setting)
    respectsGlobal := env.Var("RESPECTS_GLOBAL").
        Default("never").
        NewString()

    err := env.Parse()
    assert.NoError(t, err)

    assert.Equal(t, "always", alwaysDefault)
    assert.Equal(t, "", respectsGlobal)
}
```

### Conditional Logic Testing

```go
func TestConditionalDefaultLogic(t *testing.T) {
    env := genv.New()

    callCount := 0
    conditionalVar := env.Var("CONDITIONAL").
        Default("conditional-value").
        WithAllowDefault(func() bool {
            callCount++
            return callCount <= 1  // Only allow default on first call
        }).
        NewString()

    // First parse should use default
    err := env.Parse()
    assert.NoError(t, err)
    assert.Equal(t, "conditional-value", conditionalVar)

    // Second parse should not use default
    env2 := genv.New()
    conditionalVar2 := env2.Var("CONDITIONAL").
        Default("conditional-value").
        WithAllowDefault(func() bool {
            callCount++
            return callCount <= 1
        }).
        NewString()

    err = env2.Parse()
    assert.NoError(t, err)
    assert.Equal(t, "", conditionalVar2)
}
```

## Performance Considerations

### Function Call Overhead
- Allow functions called only when defaults needed
- Functions should be lightweight and fast
- Avoid complex computations in allow functions

### Memory Usage
- Default values stored as strings (minimal memory impact)
- Allow functions stored as closures (consider capture scope)
- Global default function shared across all variables

## Anti-Patterns to Avoid

1. **Complex Allow Functions**: Keep conditional logic simple and fast
2. **Side Effects in Allow Functions**: Allow functions should be pure
3. **Ignoring Global Setting**: Always consider global default policy
4. **Default Value Validation**: Don't validate defaults differently than environment values

## Related Patterns

- **Strategy Pattern**: Different default strategies via allow functions
- **Chain of Responsibility**: Global → Per-variable → Default value resolution
- **Configuration Pattern**: Environment-aware configuration management
- **Fail-Fast Pattern**: Early detection of missing required variables