# Table-Driven Testing Pattern

## Overview

The Table-Driven Testing Pattern uses structured test data with descriptive scenarios to provide comprehensive test coverage while maintaining readability and maintainability. This pattern is extensively used in genv for testing parsing behavior, error conditions, and edge cases across different types and configurations.

## Pattern Structure

### Core Components

1. **Test Data Structure**: Maps with descriptive field names and test scenarios
2. **Descriptive Naming**: `give*`/`want*` naming convention for inputs and expected outputs
3. **Three-Phase Execution**: Arrange/Act/Assert pattern within table iterations
4. **Comprehensive Coverage**: Multiple scenarios covering normal cases, edge cases, and error conditions

### Key Files and Locations

- `genv_test.go:38-67` - Basic table-driven test structure
- `genv_test.go:200-280` - Type-specific parsing tests
- `genv_test.go:617-851` - Advanced custom type testing
- `genv_test.go:855-1151` - Complex scenario testing

## Implementation Details

### Basic Table-Driven Structure

```go
func TestBasicParsing(t *testing.T) {
    tests := map[string]struct {
        giveEnvVar   string
        giveEnvValue string
        giveDefault  string
        wantResult   string
        wantError    bool
    }{
        "environment value used": {
            giveEnvVar:   "TEST_VAR",
            giveEnvValue: "env-value",
            giveDefault:  "default-value",
            wantResult:   "env-value",
            wantError:    false,
        },
        "default value used when env empty": {
            giveEnvVar:   "TEST_VAR",
            giveEnvValue: "",
            giveDefault:  "default-value",
            wantResult:   "default-value",
            wantError:    false,
        },
        "error when no env and no default": {
            giveEnvVar:   "TEST_VAR",
            giveEnvValue: "",
            giveDefault:  "",
            wantResult:   "",
            wantError:    true,
        },
    }

    for name, tc := range tests {
        t.Run(name, func(t *testing.T) {
            // Arrange
            if tc.giveEnvValue != "" {
                t.Setenv(tc.giveEnvVar, tc.giveEnvValue)
            }

            env := genv.New()
            var result string

            if tc.giveDefault != "" {
                env.Var(tc.giveEnvVar).Default(tc.giveDefault).String(&result)
            } else {
                env.Var(tc.giveEnvVar).String(&result)
            }

            // Act
            err := env.Parse()

            // Assert
            if tc.wantError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tc.wantResult, result)
            }
        })
    }
}
```

### Type-Specific Testing Pattern

```go
func TestTypeSpecificParsing(t *testing.T) {
    tests := map[string]struct {
        giveType     string
        giveValue    string
        wantResult   interface{}
        wantError    bool
        wantErrorMsg string
    }{
        "valid integer": {
            giveType:   "int",
            giveValue:  "42",
            wantResult: 42,
            wantError:  false,
        },
        "invalid integer": {
            giveType:     "int",
            giveValue:    "not-a-number",
            wantResult:   0,
            wantError:    true,
            wantErrorMsg: "invalid syntax",
        },
        "valid boolean true": {
            giveType:   "bool",
            giveValue:  "true",
            wantResult: true,
            wantError:  false,
        },
        "valid boolean false": {
            giveType:   "bool",
            giveValue:  "false",
            wantResult: false,
            wantError:  false,
        },
        "invalid boolean": {
            giveType:     "bool",
            giveValue:    "maybe",
            wantResult:   false,
            wantError:    true,
            wantErrorMsg: "invalid syntax",
        },
    }

    for name, tc := range tests {
        t.Run(name, func(t *testing.T) {
            t.Setenv("TEST_VAR", tc.giveValue)

            env := genv.New()
            var result interface{}
            var err error

            // Type-specific parsing
            switch tc.giveType {
            case "int":
                intResult := env.Var("TEST_VAR").NewInt()
                err = env.Parse()
                result = intResult
            case "bool":
                boolResult := env.Var("TEST_VAR").NewBool()
                err = env.Parse()
                result = boolResult
            }

            // Assertions
            if tc.wantError {
                assert.Error(t, err)
                if tc.wantErrorMsg != "" {
                    assert.Contains(t, err.Error(), tc.wantErrorMsg)
                }
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tc.wantResult, result)
            }
        })
    }
}
```

### Array Parsing Testing Pattern

```go
func TestArrayParsing(t *testing.T) {
    tests := map[string]struct {
        giveValue     string
        giveDelimiter string
        wantResult    []string
        wantError     bool
    }{
        "comma separated": {
            giveValue:     "a,b,c",
            giveDelimiter: ",",
            wantResult:    []string{"a", "b", "c"},
            wantError:     false,
        },
        "semicolon separated": {
            giveValue:     "a;b;c",
            giveDelimiter: ";",
            wantResult:    []string{"a", "b", "c"},
            wantError:     false,
        },
        "empty elements filtered": {
            giveValue:     "a,,b, ,c",
            giveDelimiter: ",",
            wantResult:    []string{"a", "b", "c"},
            wantError:     false,
        },
        "whitespace trimmed": {
            giveValue:     " a , b , c ",
            giveDelimiter: ",",
            wantResult:    []string{"a", "b", "c"},
            wantError:     false,
        },
        "single element": {
            giveValue:     "single",
            giveDelimiter: ",",
            wantResult:    []string{"single"},
            wantError:     false,
        },
        "empty string": {
            giveValue:     "",
            giveDelimiter: ",",
            wantResult:    []string{},
            wantError:     false,
        },
    }

    for name, tc := range tests {
        t.Run(name, func(t *testing.T) {
            t.Setenv("TEST_ARRAY", tc.giveValue)

            env := genv.New()
            result := env.Var("TEST_ARRAY").WithSplitKey(tc.giveDelimiter).NewStringSlice()

            err := env.Parse()

            if tc.wantError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tc.wantResult, result)
            }
        })
    }
}
```

## Architectural Benefits

### 1. Comprehensive Coverage
- **Multiple Scenarios**: Single test function covers many cases
- **Edge Cases**: Systematic testing of boundary conditions
- **Error Conditions**: Explicit testing of failure modes
- **Type Safety**: All scenarios tested with consistent types

### 2. Maintainable Test Code
- **Descriptive Names**: Test scenario names explain the test purpose
- **Structured Data**: Consistent field naming and organization
- **Reusable Pattern**: Same structure applicable across different test types
- **Easy Addition**: New scenarios added by extending test data

### 3. Clear Test Intent
- **Give/Want Naming**: Clear distinction between inputs and expected outputs
- **Scenario Descriptions**: Test names document expected behavior
- **Isolated Cases**: Each scenario tests specific behavior independently
- **Readable Assertions**: Straightforward verification logic

### 4. Debugging Support
- **Scenario Isolation**: Failures isolated to specific named scenarios
- **Clear Context**: Test names provide debugging context
- **Structured Output**: Test results clearly show which scenarios failed
- **Reproducible**: Each scenario can be run independently

## Usage Patterns

### Complex Configuration Testing

```go
func TestComplexConfiguration(t *testing.T) {
    tests := map[string]struct {
        giveEnvVars       map[string]string
        giveGlobalDefault string
        wantConfig        Config
        wantError         bool
        wantErrorContains string
    }{
        "valid complete configuration": {
            giveEnvVars: map[string]string{
                "DATABASE_URL": "postgres://localhost/test",
                "PORT":         "8080",
                "DEBUG":        "true",
                "FEATURES":     "feature1,feature2",
            },
            wantConfig: Config{
                DatabaseURL: "postgres://localhost/test",
                Port:        8080,
                Debug:       true,
                Features:    []string{"feature1", "feature2"},
            },
            wantError: false,
        },
        "defaults applied when allowed": {
            giveEnvVars: map[string]string{
                "DATABASE_URL": "postgres://localhost/test",
            },
            giveGlobalDefault: "true",
            wantConfig: Config{
                DatabaseURL: "postgres://localhost/test",
                Port:        8080,  // default
                Debug:       false, // default
                Features:    []string{"default1", "default2"}, // default
            },
            wantError: false,
        },
        "missing required variable": {
            giveEnvVars: map[string]string{
                "PORT":  "8080",
                "DEBUG": "true",
            },
            wantError:         true,
            wantErrorContains: "DATABASE_URL",
        },
    }

    for name, tc := range tests {
        t.Run(name, func(t *testing.T) {
            // Set up environment
            for key, value := range tc.giveEnvVars {
                t.Setenv(key, value)
            }
            if tc.giveGlobalDefault != "" {
                t.Setenv("GENV_ALLOW_DEFAULT", tc.giveGlobalDefault)
            }

            // Create configuration
            config, err := LoadConfig()

            // Verify results
            if tc.wantError {
                assert.Error(t, err)
                if tc.wantErrorContains != "" {
                    assert.Contains(t, err.Error(), tc.wantErrorContains)
                }
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tc.wantConfig, *config)
            }
        })
    }
}
```

### Custom Type Testing

```go
func TestCustomTypeParsing(t *testing.T) {
    type UserID string

    parseUserID := func(s string) (UserID, error) {
        if s == "" {
            return "", errors.New("UserID cannot be empty")
        }
        if !strings.HasPrefix(s, "user_") {
            return "", errors.New("UserID must start with 'user_'")
        }
        return UserID(s), nil
    }

    tests := map[string]struct {
        giveValue         string
        wantResult        UserID
        wantError         bool
        wantErrorContains string
    }{
        "valid user ID": {
            giveValue:  "user_123",
            wantResult: UserID("user_123"),
            wantError:  false,
        },
        "empty user ID": {
            giveValue:         "",
            wantResult:        UserID(""),
            wantError:         true,
            wantErrorContains: "cannot be empty",
        },
        "invalid prefix": {
            giveValue:         "admin_123",
            wantResult:        UserID(""),
            wantError:         true,
            wantErrorContains: "must start with 'user_'",
        },
    }

    for name, tc := range tests {
        t.Run(name, func(t *testing.T) {
            registry := genv.NewRegistry()
            genv.RegisterTypedParserOn(registry, parseUserID)

            env := genv.New(genv.WithRegistry(registry))

            t.Setenv("USER_ID", tc.giveValue)

            var result UserID
            env.Var("USER_ID").parseOne[UserID]() // Would need actual implementation

            err := env.Parse()

            if tc.wantError {
                assert.Error(t, err)
                if tc.wantErrorContains != "" {
                    assert.Contains(t, err.Error(), tc.wantErrorContains)
                }
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tc.wantResult, result)
            }
        })
    }
}
```

## Design Decisions

### Why Map-Based Test Structure?

1. **Named Scenarios**: Map keys provide meaningful test scenario names
2. **Readable Output**: Test failures show scenario names for easy debugging
3. **Flexible Organization**: Can group related scenarios together
4. **Easy Extension**: Adding new scenarios doesn't require changing test logic

### Give/Want Naming Convention

**Benefits of Consistent Naming**:
- **Clear Intent**: `give*` fields clearly indicate test inputs
- **Expected Results**: `want*` fields clearly indicate expected outputs
- **Searchable**: Easy to find input vs output fields in test data
- **Conventional**: Follows Go testing community standards

### Three-Phase Test Structure

```go
for name, tc := range tests {
    t.Run(name, func(t *testing.T) {
        // Arrange - Set up test conditions
        setupEnvironment(tc.giveInputs)

        // Act - Execute the code under test
        result, err := executeFunction(tc.giveParameters)

        // Assert - Verify expected outcomes
        verifyResults(tc.wantOutputs, result, err)
    })
}
```

## Testing Strategies

### Error Message Testing

```go
func TestErrorMessages(t *testing.T) {
    tests := map[string]struct {
        giveSetup         func(*testing.T)
        giveAction        func() error
        wantErrorContains []string
    }{
        "missing required variable": {
            giveSetup: func(t *testing.T) {
                // Don't set required environment variable
            },
            giveAction: func() error {
                env := genv.New()
                env.Var("REQUIRED_VAR").NewString()
                return env.Parse()
            },
            wantErrorContains: []string{"REQUIRED_VAR", "not set"},
        },
        "invalid integer format": {
            giveSetup: func(t *testing.T) {
                t.Setenv("INT_VAR", "not-a-number")
            },
            giveAction: func() error {
                env := genv.New()
                env.Var("INT_VAR").NewInt()
                return env.Parse()
            },
            wantErrorContains: []string{"INT_VAR", "invalid syntax"},
        },
    }

    for name, tc := range tests {
        t.Run(name, func(t *testing.T) {
            tc.giveSetup(t)

            err := tc.giveAction()

            assert.Error(t, err)
            for _, contains := range tc.wantErrorContains {
                assert.Contains(t, err.Error(), contains)
            }
        })
    }
}
```

### Performance Testing

```go
func TestParsingPerformance(t *testing.T) {
    tests := map[string]struct {
        giveVariableCount int
        giveSetup         func(*testing.T, int)
        wantMaxDuration   time.Duration
    }{
        "small configuration": {
            giveVariableCount: 10,
            giveSetup:         setupSmallConfig,
            wantMaxDuration:   1 * time.Millisecond,
        },
        "large configuration": {
            giveVariableCount: 1000,
            giveSetup:         setupLargeConfig,
            wantMaxDuration:   10 * time.Millisecond,
        },
    }

    for name, tc := range tests {
        t.Run(name, func(t *testing.T) {
            tc.giveSetup(t, tc.giveVariableCount)

            start := time.Now()
            err := parseConfiguration()
            duration := time.Since(start)

            assert.NoError(t, err)
            assert.Less(t, duration, tc.wantMaxDuration)
        })
    }
}
```

## Anti-Patterns to Avoid

1. **Over-Complex Test Data**: Keep test scenarios focused and simple
2. **Shared Mutable State**: Each test scenario should be independent
3. **Magic Values**: Use descriptive field names rather than indices or magic numbers
4. **Inconsistent Naming**: Stick to give/want naming convention throughout

## Performance Considerations

### Test Execution Speed
- Use `t.Parallel()` for independent scenarios
- Minimize environment setup overhead
- Cache expensive test data when appropriate

### Memory Usage
- Large test data structures can impact memory
- Consider table-driven subtests for very large test suites
- Clean up environment variables between tests

## Related Patterns

- **Behavioral Contracts**: Table-driven tests validate behavioral contracts
- **Parametrized Testing**: Multiple inputs with consistent validation logic
- **Test Data Builder**: Structured creation of test data
- **Scenario Testing**: Each table entry represents a test scenario