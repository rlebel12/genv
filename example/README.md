# GENV Examples

This directory demonstrates comprehensive usage of the `genv` package, including the new registry functionality for custom types and parser isolation.

## Quick Start

Run all examples:
```bash
go run .
```

Run with basic environment variables set:
```bash
STRING_VAR=hello INT_VAR=42 BOOL_VAR=true go run .
```

Run with custom registry environment variables:
```bash
STRING_VAR=hello INT_VAR=42 BOOL_VAR=true \
CUSTOM_USER_ID=alice CUSTOM_DEPARTMENT=engineering CUSTOM_EMAIL=alice@example.com \
SERVICE_NAME=auth-service TASK_PRIORITIES="high,medium" LOG_LEVEL=DEBUG \
go run .
```

## Examples Overview

### 1. Basic Usage Example (`NewSettings`)

Demonstrates traditional `genv` usage patterns:

**Required Environment Variables:**
- `STRING_VAR` - Any non-empty string
- `INT_VAR` - Integer value (e.g., `42`)
- `BOOL_VAR` - Boolean value (`true` or `false`)

**Optional Variables:**
- `ALWAYS_DEFAULT_STRING_VAR` - Falls back to "default value"
- `OPTIONAL_FLOAT_VAR` - Optional float, defaults to 0
- `ADVANCED_URL_VAR` - URL with conditional default logic
- `MANY_INT_VAR` - Semicolon-separated integers (e.g., `"123;456"`)

```bash
STRING_VAR="hello world" INT_VAR=8080 BOOL_VAR=true go run .
```

### 2. Custom Registry Example (`NewCustomRegistrySettings`)

Shows how to register and use custom types with validation:

**Custom Types:**
- `UserID` - Automatically adds "user_" prefix if missing
- `Department` - Validates against allowed departments (engineering, marketing, sales, hr)
- `ValidatedEmail` - Basic email format validation and normalization

**Environment Variables:**
- `CUSTOM_USER_ID` - User identifier (e.g., `"alice"` → `"user_alice"`)
- `CUSTOM_DEPARTMENT` - Department name (must be valid department)
- `CUSTOM_EMAIL` - Email address (basic validation)

```bash
CUSTOM_USER_ID=bob CUSTOM_DEPARTMENT=sales CUSTOM_EMAIL=Bob@Example.COM go run .
```

### 3. Registry Isolation Example (`DemonstrateRegistryIsolation`)

Demonstrates how different `genv` instances can have completely different parsing behavior using isolated registries.

**Features:**
- Production registry with strict validation
- Development registry with lenient defaults
- Complete isolation between registry instances

This example runs automatically and shows conceptual isolation.

### 4. Advanced Custom Type Features (`NewAdvancedCustomTypeSettings`)

Shows advanced patterns including slices, optional types, and defaults:

**Custom Types:**
- `Priority` - Enum-like parsing (low=1, medium=2, high=3, critical=4)
- `ServiceName` - Adds "svc-" prefix, minimum length validation
- `ExampleLogLevel` - Strict log level validation

**Environment Variables:**
- `TASK_PRIORITIES` - Pipe-separated priorities (e.g., `"medium|high|low"`)
- `SERVICE_NAME` - Optional service name (adds prefix if needed)
- `LOG_LEVEL` - Log level with default "INFO"

```bash
TASK_PRIORITIES="high|critical|medium" SERVICE_NAME=api LOG_LEVEL=DEBUG go run .
```

### 5. Production Configuration Example (`DemonstrateProductionConfig`)

Demonstrates real-world microservice configuration patterns with environment-specific validation.

**Features:**
- Environment-specific registries (development vs production)
- Production-grade validation
- Microservice configuration patterns

This example shows how to structure configuration for different deployment environments.

## Custom Type Examples

### Simple Custom Type

```go
type UserID string

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
)

env := genv.New(genv.WithRegistry(registry))
var userID UserID
err := genv.Parse(env, genv.Bind("USER_ID", &userID))
```

### Enum-like Custom Type

```go
type Priority int

registry := genv.NewDefaultRegistry(
    genv.WithParser(func(s string) (Priority, error) {
        switch strings.ToLower(s) {
        case "low": return Priority(1), nil
        case "medium": return Priority(2), nil
        case "high": return Priority(3), nil
        case "critical": return Priority(4), nil
        default: return Priority(0), fmt.Errorf("invalid priority: %s", s)
        }
    }),
)

env := genv.New(genv.WithRegistry(registry))
var priorities []Priority
err := genv.Parse(env, genv.BindMany("PRIORITIES", &priorities)) // "low,high,medium"
```

### Custom Type with Complex Validation

```go
type EmailAddress string

genv.RegisterTypedParserOn(registry, func(s string) (EmailAddress, error) {
    if !strings.Contains(s, "@") {
        return "", errors.New("invalid email format")
    }
    if !strings.Contains(s, ".") {
        return "", errors.New("email missing domain")
    }
    return EmailAddress(strings.ToLower(s)), nil
})
```

## Registry Patterns

### Development vs Production Registries

```go
func NewProductionRegistry() *genv.ParserRegistry {
    registry := genv.NewDefaultRegistry()
    
    // Strict validation for production
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
    
    return registry
}

func NewDevelopmentRegistry() *genv.ParserRegistry {
    registry := genv.NewDefaultRegistry()
    
    // Lenient validation for development
    genv.RegisterTypedParserOn(registry, func(s string) (ServicePort, error) {
        port, err := strconv.Atoi(s)
        if err != nil {
            return ServicePort(8080), nil // Default port
        }
        return ServicePort(port), nil
    })
    
    return registry
}
```

### Testing with Custom Registries

```go
func TestConfig(t *testing.T) {
    // Create test registry with mock behavior
    testRegistry := genv.NewDefaultRegistry()
    
    genv.RegisterTypedParserOn(testRegistry, func(s string) (UserID, error) {
        return UserID("test_" + s), nil // Always add test prefix
    })
    
    env := genv.New(genv.WithRegistry(testRegistry))
    // ... test with isolated parsing behavior
}
```

## Running Individual Examples

Each example function can be called independently:

```go
// Just run the basic example
settings, err := NewSettings()
if err != nil {
    log.Fatal(err)
}

// Just run custom registry example  
customSettings, err := NewCustomRegistrySettings()
if err != nil {
    log.Fatal(err)
}

// Just run advanced features
advancedSettings, err := NewAdvancedCustomTypeSettings()
if err != nil {
    log.Fatal(err)
}
```

## Error Handling

The examples demonstrate various error scenarios:

1. **Missing required variables** - Basic example fails without required env vars
2. **Invalid custom types** - Custom parsers validate input and return descriptive errors
3. **Type validation** - Custom types can enforce business rules
4. **Environment-specific validation** - Different validation rules per environment

## Best Practices Demonstrated

1. **Registry Isolation** - Use separate registries for different environments
2. **Type Safety** - Custom types prevent invalid values at parse time
3. **Default Values** - Graceful fallbacks for non-critical configuration
4. **Validation** - Input validation at the parser level
5. **Error Messages** - Clear, actionable error messages
6. **Testing** - Isolated registries enable better unit testing

## Integration with Other Tools

The microservice configuration example (`microservice_config.go`) shows how to integrate `genv` with:

- Environment-specific configuration
- Validation patterns
- Structured configuration types
- Production vs development settings

This provides a template for real-world applications.