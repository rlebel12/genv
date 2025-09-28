# Genv Project Documentation Index

Project-specific documentation for the `genv` Go package - a type-safe environment variable parsing library with fluent API design and extensible type system.

## Quick Reference

### Architecture Patterns
- **[Registry Isolation](architectures/registry-isolation.md)** - Per-instance parser registries for environment-specific behavior
- **[Fluent API with Deferred Execution](architectures/fluent-deferred-execution.md)** - Method chaining with batch validation and parsing
- **[Dual API Pattern](architectures/dual-api-pattern.md)** - Pointer-based and value-based APIs with unified backend

### Implementation Patterns
- **[Generic Type Parsing](patterns/generic-type-parsing.md)** - Type-safe parsing with runtime type dispatch
- **[Array Parsing](patterns/array-parsing.md)** - Configurable delimiter support with automatic filtering
- **[Default Value Handling](patterns/default-value-handling.md)** - Conditional defaults with global and per-variable control

### Testing Strategies
- **[Table-Driven Tests](testing/table-driven-tests.md)** - Comprehensive test coverage with descriptive scenarios
- **[Behavioral Contracts](testing/behavioral-contracts.md)** - Cross-implementation consistency validation

## Project Overview

**Genv** is a Go package that provides type-safe environment variable parsing with a fluent API design. The library emphasizes:

- **Type Safety**: Compile-time type checking with runtime flexibility
- **Registry Isolation**: Independent parser configurations per instance
- **Fluent Interface**: Chainable method calls for readable configuration
- **Extensibility**: Custom type support without modifying core library
- **Validation**: Comprehensive error handling and default value management

## Architecture Overview

### Core Design Principles

1. **Registry-Per-Instance Isolation**: Each `Genv` instance contains its own `ParserRegistry`, enabling different parsing behaviors without global state conflicts
2. **Deferred Execution**: Configuration operations collected and executed in batch during `Parse()` call
3. **Generic Type System**: Compile-time type safety with runtime type dispatch through reflection
4. **Dual API Support**: Both pointer-based (`String(&variable)`) and value-based (`NewString()`) APIs

### Key Components

**Main Types**:
- `Genv` - Main configuration object managing defaults, parsing, and parser registry
- `Var` - Individual environment variable with parsing rules and options
- `ParserRegistry` - Type-to-parser mapping with isolation between instances
- `fallback` - Default value handling with conditional allow functions

**Supported Types**:
- Built-in: `string`, `bool`, `int`, `float64`, `url.URL`, `uuid.UUID`, `time.Time`
- Arrays: `[]string`, `[]bool`, `[]int`, `[]float64`, `[]url.URL`, `[]uuid.UUID`
- Custom: User-defined types via parser registration

## Architecture Patterns Deep Dive

### 1. Registry Isolation Pattern
**Location**: [architectures/registry-isolation.md](architectures/registry-isolation.md)

Enables different `Genv` instances to have completely different parsing behaviors:
- Environment-specific validation rules (prod vs dev vs test)
- Custom type parsing without global state pollution
- Thread-safe concurrent access without synchronization
- Plugin-like architecture for domain-specific types

**Key Innovation**: Per-instance registries eliminate action-at-a-distance effects and enable sophisticated testing strategies.

### 2. Fluent API with Deferred Execution
**Location**: [architectures/fluent-deferred-execution.md](architectures/fluent-deferred-execution.md)

Combines declarative configuration with atomic validation:
- Method chaining for readable configuration (`env.Var("PORT").Default("8080").NewInt()`)
- Operations collected in `varFuncs` slice for batch execution
- All-or-nothing parsing prevents partial configuration states
- Early validation detects configuration errors before application startup

**Key Innovation**: Deferred execution enables comprehensive validation while maintaining fluent interface ergonomics.

### 3. Dual API Pattern
**Location**: [architectures/dual-api-pattern.md](architectures/dual-api-pattern.md)

Provides two complementary interfaces for maximum flexibility:
- **Pointer API**: Direct assignment to existing variables (`env.Var("KEY").String(&variable)`)
- **Value API**: Functional style with return values (`variable := env.Var("KEY").NewString()`)
- **Unified Backend**: Both APIs use same parsing infrastructure for consistency
- **Migration Support**: Enables gradual adoption and different coding styles

## Implementation Patterns Deep Dive

### 1. Generic Type Parsing
**Location**: [patterns/generic-type-parsing.md](patterns/generic-type-parsing.md)

Combines Go generics with reflection for type-safe extensibility:
- `parseOne[T any]()` function provides compile-time type safety
- Runtime type dispatch through reflection-based registry lookup
- Custom type registration via `RegisterTypedParserOn[T](registry, parseFunc)`
- Consistent error handling with type information

**Key Innovation**: Bridges compile-time type safety with runtime parser registration flexibility.

### 2. Array Parsing with Configurable Delimiters
**Location**: [patterns/array-parsing.md](patterns/array-parsing.md)

Flexible array parsing with robust handling:
- Configurable delimiters (comma, semicolon, pipe, space, custom)
- Automatic empty value filtering and whitespace trimming
- Type-safe element parsing using generic infrastructure
- Element-level error reporting with index information

### 3. Default Value Handling
**Location**: [patterns/default-value-handling.md](patterns/default-value-handling.md)

Sophisticated fallback system with conditional logic:
- Global default control via `GENV_ALLOW_DEFAULT` environment variable
- Per-variable override through `WithAllowDefault()` and `WithAllowDefaultAlways()`
- Complex conditional logic for environment-aware defaults
- Security benefits through production default hardening

## Testing Strategy Overview

### 1. Table-Driven Testing Pattern
**Location**: [testing/table-driven-tests.md](testing/table-driven-tests.md)

Comprehensive test coverage through structured scenarios:
- Map-based test structure with descriptive scenario names
- `give*`/`want*` naming convention for inputs and expected outputs
- Three-phase execution (Arrange/Act/Assert) within table iterations
- Edge case and error condition testing

### 2. Behavioral Contract Testing
**Location**: [testing/behavioral-contracts.md](testing/behavioral-contracts.md)

Cross-implementation consistency validation:
- Registry isolation verification ensuring independent behavior
- API consistency testing between pointer and value APIs
- Default value contract validation across different configurations
- Parser extension contract testing for custom type integration

## Development Workflow Integration

### Spec-Driven Development Support

This documentation supports spec-driven development workflows through:

1. **Pattern Library**: Reusable patterns for similar Go libraries
2. **Architecture Decisions**: Documented rationale for design choices
3. **Testing Strategies**: Proven approaches for complex library testing
4. **Implementation Guidance**: Detailed patterns for feature development

### Usage in Planning Phase

During planning (`/spec` commands), reference:
- **Architecture patterns** for system design decisions
- **Implementation patterns** for specific feature development
- **Testing patterns** for validation strategy planning

### Usage in Implementation Phase

During development:
- Follow established patterns for consistency
- Apply testing strategies for comprehensive coverage
- Use architectural decisions as implementation guidance

## Contributing Guidelines

### Pattern Discovery
When implementing new features, document emerging patterns:
- Extract reusable patterns for future use
- Document architectural decisions and their rationale
- Add testing strategies for new complexity areas

### Documentation Maintenance
- Update patterns when better approaches are discovered
- Keep examples current with latest implementation
- Cross-reference related patterns for discoverability

## References

### External Pattern Sources
- **Global Patterns**: Reference patterns from `@~/.agent/docs/` when applicable
- **Go Language Patterns**: Contribute discoveries to global Go pattern library
- **Testing Patterns**: Align with established testing pattern library

### Project-Specific Adaptations
- **Registry System**: Unique to genv, potential contribution to global patterns
- **Deferred Configuration**: Novel approach worth documenting for reuse
- **Dual API Design**: Pattern applicable to other library designs

---

**Note**: This documentation serves as both implementation guidance and pattern library for similar Go library development. Patterns documented here may be contributed back to global pattern repositories for broader reuse.