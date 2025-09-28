# Registry Isolation Pattern

## Overview

The Registry Isolation Pattern provides per-instance type parser registries, enabling different `Genv` instances to have completely different parsing behaviors without global state conflicts. This architectural pattern is fundamental to genv's design and enables flexible, testable, and environment-specific configurations.

## Pattern Structure

### Core Components

1. **ParserRegistry**: Maps Go types to parsing functions using reflection
2. **Genv Instance**: Contains its own isolated ParserRegistry
3. **Registry Factories**: Create registries with different parser sets

### Key Files and Locations

- `genv.go:370-445` - ParserRegistry implementation
- `genv.go:515-530` - Registry options pattern
- `genv_test.go:481-532` - Registry isolation testing

## Implementation Details

### Registry Creation

```go
// Empty registry for custom-only parsing
registry := genv.NewRegistry()

// Registry with all built-in parsers
registry := genv.NewDefaultRegistry()

// Configure Genv with specific registry
env := genv.New(genv.WithRegistry(registry))
```

### Type-Safe Registration

```go
// Register custom type parser
genv.RegisterTypedParserOn(registry, func(s string) (CustomType, error) {
    // Custom parsing logic
    return parseCustomType(s)
})
```

## Architectural Benefits

### 1. Environment Isolation
- Different validation rules per environment (prod vs dev vs test)
- Custom type parsing without affecting other instances
- Clean separation of concerns between application contexts

### 2. Testing Flexibility
- Mock parsers for testing without global state pollution
- Isolated test environments with predictable behavior
- Ability to test different parser combinations independently

### 3. Zero Global State
- No singleton registries or global parser maps
- Thread-safe concurrent access without locks
- Eliminates action-at-a-distance effects

### 4. Extensibility
- Custom type support without modifying core library
- Plugin-like architecture for domain-specific types
- Backward compatibility through default registry behavior

## Usage Patterns

### Production Configuration
```go
// Production environment with strict validation
prodRegistry := genv.NewDefaultRegistry()
genv.RegisterTypedParserOn(prodRegistry, strictUserIDParser)
prodEnv := genv.New(genv.WithRegistry(prodRegistry))
```

### Development/Testing Configuration
```go
// Development environment with lenient parsing
devRegistry := genv.NewDefaultRegistry()
genv.RegisterTypedParserOn(devRegistry, lenientUserIDParser)
devEnv := genv.New(genv.WithRegistry(devRegistry))
```

### Custom Type Domain
```go
// Microservice with domain-specific types
serviceRegistry := genv.NewRegistry()
genv.RegisterTypedParserOn(serviceRegistry, parseServiceID)
genv.RegisterTypedParserOn(serviceRegistry, parseAPIKey)
serviceEnv := genv.New(genv.WithRegistry(serviceRegistry))
```

## Design Decisions

### Why Per-Instance Over Global Registry?

1. **Testability**: Tests can run in parallel with different parser sets
2. **Modularity**: Different parts of application can have different type needs
3. **Safety**: No risk of accidentally affecting other code through registration
4. **Flexibility**: Runtime configuration of parsing behavior

### Reflection vs Interface-Based

The pattern uses reflection for type mapping rather than interface-based approaches because:
- Enables registration of external types (url.URL, uuid.UUID)
- Provides type safety through generics
- Allows parsing functions to return concrete types
- Maintains clean API without interface pollution

## Testing Strategies

### Registry Isolation Verification
```go
func TestRegistryIsolation(t *testing.T) {
    registry1 := genv.NewRegistry()
    registry2 := genv.NewRegistry()

    // Register different parsers in each registry
    genv.RegisterTypedParserOn(registry1, parser1)
    genv.RegisterTypedParserOn(registry2, parser2)

    // Verify independent behavior
    env1 := genv.New(genv.WithRegistry(registry1))
    env2 := genv.New(genv.WithRegistry(registry2))

    // Assertions verify different parsing behavior
}
```

## Anti-Patterns to Avoid

1. **Sharing Registries Between Unrelated Components**: Each logical grouping should have its own registry
2. **Modifying Default Registry**: Create custom registries instead of modifying the default
3. **Global Registry Variables**: Always pass registries through dependency injection
4. **Registry Mutation After Use**: Registries should be configured once during initialization

## Related Patterns

- **Dependency Injection**: Registries are injected into Genv instances
- **Factory Pattern**: Registry creation through factory functions
- **Strategy Pattern**: Different parsing strategies via different registries