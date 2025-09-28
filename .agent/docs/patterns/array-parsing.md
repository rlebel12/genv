# Array Parsing with Configurable Delimiters

## Overview

The Array Parsing Pattern enables parsing of environment variables into Go slices with configurable delimiter support and automatic empty value filtering. This pattern provides flexible array parsing while maintaining type safety and consistent error handling for individual elements.

## Pattern Structure

### Core Components

1. **Configurable Delimiter**: Runtime delimiter configuration via `WithSplitKey()`
2. **Generic Array Parsing**: Type-safe slice parsing using `parseSlice[T any]()`
3. **Element-Level Validation**: Individual parsing of each array element
4. **Empty Value Filtering**: Automatic removal of empty strings after splitting

### Key Files and Locations

- `genv.go:282-324` - Array parsing implementation
- `genv.go:515-530` - Split delimiter configuration options
- `genv_test.go:200-250` - Array parsing tests with custom delimiters

## Implementation Details

### Generic Slice Parsing

```go
// Generic slice parsing with configurable delimiter
func (v *Var) parseSlice[T any]() ([]T, error) {
    envValue := os.Getenv(v.key)

    // Handle default values
    if envValue == "" && v.fallback.hasDefault() {
        envValue = v.fallback.defaultValue
    }

    // Split using configured delimiter (default: ",")
    delimiter := v.genv.splitKey
    parts := strings.Split(envValue, delimiter)

    // Filter out empty strings
    var nonEmptyParts []string
    for _, part := range parts {
        trimmed := strings.TrimSpace(part)
        if trimmed != "" {
            nonEmptyParts = append(nonEmptyParts, trimmed)
        }
    }

    // Parse each element using generic parsing
    result := make([]T, 0, len(nonEmptyParts))
    for i, part := range nonEmptyParts {
        // Use single-value parser for each element
        parsed, err := v.parseElementAs[T](part)
        if err != nil {
            return nil, fmt.Errorf("element %d: %w", i, err)
        }
        result = append(result, parsed)
    }

    return result, nil
}
```

### Delimiter Configuration

```go
// Configurable delimiter support
type Genv struct {
    splitKey  string // Default: ","
    // ... other fields
}

// Option pattern for delimiter configuration
func WithSplitKey(delimiter string) Opt[*Genv] {
    return func(g *Genv) {
        g.splitKey = delimiter
    }
}

// Fluent API for per-variable delimiter
func (v *Var) WithSplitKey(delimiter string) *Var {
    // Create variable-specific configuration
    newVar := *v
    newVar.genv = &Genv{
        splitKey: delimiter,
        registry: v.genv.registry, // Share registry
    }
    return &newVar
}
```

### Type-Specific Array APIs

```go
// String slice parsing
func (v *Var) NewStringSlice() []string {
    var result []string
    v.genv.varFuncs = append(v.genv.varFuncs, func() error {
        parsed, err := v.parseSlice[string]()
        if err != nil {
            return err
        }
        result = parsed
        return nil
    })
    return result
}

// Integer slice parsing
func (v *Var) NewIntSlice() []int {
    var result []int
    v.genv.varFuncs = append(v.genv.varFuncs, func() error {
        parsed, err := v.parseSlice[int]()
        if err != nil {
            return err
        }
        result = parsed
        return nil
    })
    return result
}
```

## Architectural Benefits

### 1. Flexible Delimiter Support
- Configurable delimiters for different data formats
- Global and per-variable delimiter configuration
- Common delimiters: comma `,`, semicolon `;`, pipe `|`, space ` `

### 2. Type Safety for Elements
- Each array element parsed with appropriate type parser
- Compile-time type checking for slice elements
- Consistent error handling for malformed elements

### 3. Robust Parsing
- Automatic whitespace trimming around elements
- Empty value filtering prevents empty slice elements
- Element-level error reporting with index information

### 4. Consistent API
- Same fluent API patterns as single-value parsing
- Default value support for entire arrays
- Optional array support with standard patterns

## Usage Patterns

### Basic Array Parsing

```go
env := genv.New()

// Comma-separated (default)
ports := env.Var("PORTS").Default("8080,8081,8082").NewIntSlice()
// Input: "8080,8081,8082" → Output: []int{8080, 8081, 8082}

hosts := env.Var("HOSTS").Default("localhost,api.example.com").NewStringSlice()
// Input: "localhost,api.example.com" → Output: []string{"localhost", "api.example.com"}

if err := env.Parse(); err != nil {
    log.Fatal(err)
}
```

### Custom Delimiter Configuration

```go
env := genv.New()

// Semicolon-separated URLs
urls := env.Var("API_ENDPOINTS").
    WithSplitKey(";").
    Default("http://api1.com;http://api2.com").
    NewURLSlice()

// Pipe-separated feature flags
features := env.Var("FEATURES").
    WithSplitKey("|").
    Default("feature1|feature2|feature3").
    NewStringSlice()

// Space-separated numbers
numbers := env.Var("NUMBERS").
    WithSplitKey(" ").
    Default("1 2 3 4 5").
    NewIntSlice()

if err := env.Parse(); err != nil {
    log.Fatal(err)
}
```

### Complex Type Arrays

```go
// Custom type arrays
type ServiceConfig struct {
    Name string
    Port int
}

func parseServiceConfig(s string) (ServiceConfig, error) {
    parts := strings.Split(s, ":")
    if len(parts) != 2 {
        return ServiceConfig{}, errors.New("format must be name:port")
    }

    port, err := strconv.Atoi(parts[1])
    if err != nil {
        return ServiceConfig{}, fmt.Errorf("invalid port: %w", err)
    }

    return ServiceConfig{Name: parts[0], Port: port}, nil
}

// Register parser and use with arrays
registry := genv.NewRegistry()
genv.RegisterTypedParserOn(registry, parseServiceConfig)

env := genv.New(genv.WithRegistry(registry))
services := env.Var("SERVICES").
    WithSplitKey(";").
    Default("api:8080;web:3000;db:5432").
    parseSlice[ServiceConfig]() // Would need additional method implementation
```

### Empty Value Handling

```go
env := genv.New()

// Empty values are filtered out
values := env.Var("VALUES").Default("a,,b, ,c").NewStringSlice()
// Input: "a,,b, ,c" → Output: []string{"a", "b", "c"}

// Leading/trailing whitespace trimmed
trimmed := env.Var("TRIMMED").Default(" one , two , three ").NewStringSlice()
// Input: " one , two , three " → Output: []string{"one", "two", "three"}

if err := env.Parse(); err != nil {
    log.Fatal(err)
}
```

## Design Decisions

### Why Configurable Delimiters?

1. **Data Format Flexibility**: Different systems use different separators
2. **CSV Compatibility**: Standard comma separation
3. **Shell Compatibility**: PATH-style colon separation
4. **URL Safety**: Semicolon separation when commas appear in values

### Empty Value Filtering Strategy

**Automatic Filtering Rationale**:
- **User Intent**: Empty values rarely intended in environment variables
- **Common Pattern**: Environment variables often have trailing delimiters
- **Clean API**: No need for explicit filtering in user code
- **Predictable Behavior**: Consistent empty value handling

### Element-Level Error Handling

```go
// Detailed error context for array parsing
func (v *Var) parseSlice[T any]() ([]T, error) {
    for i, part := range nonEmptyParts {
        parsed, err := v.parseElementAs[T](part)
        if err != nil {
            // Include element index and value in error
            return nil, fmt.Errorf("element %d (%q): %w", i, part, err)
        }
        result = append(result, parsed)
    }
    return result, nil
}
```

## Testing Strategies

### Delimiter Testing

```go
func TestCustomDelimiters(t *testing.T) {
    tests := []struct {
        name      string
        delimiter string
        input     string
        expected  []string
    }{
        {
            name:      "comma separated",
            delimiter: ",",
            input:     "a,b,c",
            expected:  []string{"a", "b", "c"},
        },
        {
            name:      "semicolon separated",
            delimiter: ";",
            input:     "a;b;c",
            expected:  []string{"a", "b", "c"},
        },
        {
            name:      "pipe separated",
            delimiter: "|",
            input:     "a|b|c",
            expected:  []string{"a", "b", "c"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            env := genv.New()
            result := env.Var("TEST_VAR").WithSplitKey(tt.delimiter).NewStringSlice()

            t.Setenv("TEST_VAR", tt.input)

            err := env.Parse()
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Edge Case Testing

```go
func TestArrayEdgeCases(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected []string
    }{
        {
            name:     "empty string",
            input:    "",
            expected: []string{},
        },
        {
            name:     "single value",
            input:    "single",
            expected: []string{"single"},
        },
        {
            name:     "empty elements filtered",
            input:    "a,,b",
            expected: []string{"a", "b"},
        },
        {
            name:     "whitespace trimmed",
            input:    " a , b , c ",
            expected: []string{"a", "b", "c"},
        },
        {
            name:     "only delimiters",
            input:    ",,,",
            expected: []string{},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            env := genv.New()
            result := env.Var("TEST_VAR").NewStringSlice()

            t.Setenv("TEST_VAR", tt.input)

            err := env.Parse()
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Type Safety Testing

```go
func TestTypedArrayParsing(t *testing.T) {
    env := genv.New()

    intResult := env.Var("INT_ARRAY").NewIntSlice()
    boolResult := env.Var("BOOL_ARRAY").NewBoolSlice()

    t.Setenv("INT_ARRAY", "1,2,3")
    t.Setenv("BOOL_ARRAY", "true,false,true")

    err := env.Parse()
    assert.NoError(t, err)
    assert.Equal(t, []int{1, 2, 3}, intResult)
    assert.Equal(t, []bool{true, false, true}, boolResult)
}
```

## Performance Considerations

### Memory Allocation
- Pre-allocates result slice with capacity hint
- Single allocation for split operation
- Minimal string copying with trim operations

### Parsing Efficiency
- Reuses single-value parsers for each element
- No reflection overhead per element (cached in registry)
- Fail-fast on first parsing error

### Large Array Handling
```go
// For very large arrays, consider streaming approach
func (v *Var) parseSliceStream[T any]() ([]T, error) {
    // Stream parsing for memory efficiency
    // Process elements one at a time without storing split results
}
```

## Anti-Patterns to Avoid

1. **Nested Delimiters**: Don't use same delimiter for different nesting levels
2. **Delimiter in Values**: Ensure delimiter doesn't appear in actual values
3. **Manual Empty Filtering**: Let the pattern handle empty value filtering
4. **Type Mixing**: Don't mix different types in same array environment variable

## Related Patterns

- **Generic Type Parsing**: Element parsing uses same generic parsing infrastructure
- **Configuration Pattern**: Arrays often used for configuration lists
- **CSV Parsing**: Similar to CSV field parsing with configurable delimiters
- **Builder Pattern**: Fluent API for delimiter configuration