# Behavioral Contract Testing Pattern

## Overview

The Behavioral Contract Testing Pattern validates that different implementations or configurations maintain consistent behavior across varying contexts. In genv, this pattern ensures registry isolation works correctly and that different parser configurations behave predictably while maintaining the same interface contracts.

## Pattern Structure

### Core Components

1. **Contract Definition**: Explicit behavioral expectations for interfaces
2. **Multiple Implementations**: Different configurations tested against same contract
3. **Isolation Verification**: Independent behavior validation across instances
4. **Behavioral Consistency**: Same inputs produce expected outputs regardless of implementation

### Key Files and Locations

- `genv_test.go:481-532` - Registry isolation contract testing
- `genv_test.go:951-1048` - Mixed parser configuration contracts
- `genv_test.go:1050-1151` - Complex validation contract testing

## Implementation Details

### Registry Isolation Contract

```go
func TestRegistryIsolationContract(t *testing.T) {
    // Contract: Different registries should have independent behavior
    // without affecting each other's parsing capabilities

    type CustomType1 string
    type CustomType2 string

    parser1 := func(s string) (CustomType1, error) {
        return CustomType1("type1_" + s), nil
    }

    parser2 := func(s string) (CustomType2, error) {
        return CustomType2("type2_" + s), nil
    }

    // Create isolated registries
    registry1 := genv.NewRegistry()
    registry2 := genv.NewRegistry()

    // Register different parsers in each registry
    genv.RegisterTypedParserOn(registry1, parser1)
    genv.RegisterTypedParserOn(registry2, parser2)

    // Create independent Genv instances
    env1 := genv.New(genv.WithRegistry(registry1))
    env2 := genv.New(genv.WithRegistry(registry2))

    t.Setenv("TEST_VAR", "test_value")

    // Contract validation: Each environment should only know about its own types
    var result1 CustomType1
    var result2 CustomType2

    env1.Var("TEST_VAR").parseOne[CustomType1](&result1) // Should work
    env2.Var("TEST_VAR").parseOne[CustomType2](&result2) // Should work

    // Parse both environments
    err1 := env1.Parse()
    err2 := env2.Parse()

    // Contract assertions
    assert.NoError(t, err1, "registry1 should parse CustomType1")
    assert.NoError(t, err2, "registry2 should parse CustomType2")
    assert.Equal(t, CustomType1("type1_test_value"), result1)
    assert.Equal(t, CustomType2("type2_test_value"), result2)

    // Cross-contamination test: env1 should not know about CustomType2
    env1_crossTest := genv.New(genv.WithRegistry(registry1))
    var crossResult CustomType2
    env1_crossTest.Var("TEST_VAR").parseOne[CustomType2](&crossResult)

    err := env1_crossTest.Parse()
    assert.Error(t, err, "registry1 should not parse CustomType2")
    assert.Contains(t, err.Error(), "no parser registered")
}
```

### API Consistency Contract

```go
func TestAPIConsistencyContract(t *testing.T) {
    // Contract: Pointer API and Value API should produce identical results
    // for the same environment configuration

    testCases := []struct {
        envVar   string
        envValue string
        typeName string
    }{
        {"STRING_VAR", "test_string", "string"},
        {"INT_VAR", "42", "int"},
        {"BOOL_VAR", "true", "bool"},
        {"FLOAT_VAR", "3.14", "float64"},
    }

    for _, tc := range testCases {
        t.Run(tc.typeName+"_consistency", func(t *testing.T) {
            t.Setenv(tc.envVar, tc.envValue)

            // Test Pointer API
            env1 := genv.New()
            var ptrResult interface{}

            switch tc.typeName {
            case "string":
                var s string
                env1.Var(tc.envVar).String(&s)
                ptrResult = s
            case "int":
                var i int
                env1.Var(tc.envVar).Int(&i)
                ptrResult = i
            case "bool":
                var b bool
                env1.Var(tc.envVar).Bool(&b)
                ptrResult = b
            case "float64":
                var f float64
                env1.Var(tc.envVar).Float64(&f)
                ptrResult = f
            }

            err1 := env1.Parse()

            // Test Value API
            env2 := genv.New()
            var valueResult interface{}

            switch tc.typeName {
            case "string":
                valueResult = env2.Var(tc.envVar).NewString()
            case "int":
                valueResult = env2.Var(tc.envVar).NewInt()
            case "bool":
                valueResult = env2.Var(tc.envVar).NewBool()
            case "float64":
                valueResult = env2.Var(tc.envVar).NewFloat64()
            }

            err2 := env2.Parse()

            // Contract assertions: Both APIs should behave identically
            assert.Equal(t, err1, err2, "Error behavior should be consistent")
            if err1 == nil && err2 == nil {
                assert.Equal(t, ptrResult, valueResult, "Results should be identical")
            }
        })
    }
}
```

### Default Value Contract

```go
func TestDefaultValueContract(t *testing.T) {
    // Contract: Default value behavior should be consistent across
    // different global settings and variable configurations

    testScenarios := []struct {
        name                string
        globalDefaultSetting string
        variableConfig      func(*genv.Var) *genv.Var
        environmentValue    string
        expectedResult      string
        expectDefault       bool
    }{
        {
            name:                "global defaults enabled, use default",
            globalDefaultSetting: "true",
            variableConfig:      func(v *genv.Var) *genv.Var { return v.Default("default") },
            environmentValue:    "",
            expectedResult:      "default",
            expectDefault:       true,
        },
        {
            name:                "global defaults disabled, no default",
            globalDefaultSetting: "false",
            variableConfig:      func(v *genv.Var) *genv.Var { return v.Default("default") },
            environmentValue:    "",
            expectedResult:      "",
            expectDefault:       false,
        },
        {
            name:                "force default always overrides global",
            globalDefaultSetting: "false",
            variableConfig:      func(v *genv.Var) *genv.Var {
                return v.Default("default").WithAllowDefaultAlways()
            },
            environmentValue:    "",
            expectedResult:      "default",
            expectDefault:       true,
        },
    }

    for _, scenario := range testScenarios {
        t.Run(scenario.name, func(t *testing.T) {
            // Set up global default setting
            if scenario.globalDefaultSetting != "" {
                t.Setenv("GENV_ALLOW_DEFAULT", scenario.globalDefaultSetting)
            }

            // Set up environment value
            if scenario.environmentValue != "" {
                t.Setenv("TEST_VAR", scenario.environmentValue)
            }

            env := genv.New()
            result := scenario.variableConfig(env.Var("TEST_VAR")).NewString()

            err := env.Parse()

            // Contract validation
            if scenario.expectDefault || scenario.environmentValue != "" {
                assert.NoError(t, err, "Should parse successfully when default expected or env value provided")
                assert.Equal(t, scenario.expectedResult, result, "Result should match expected value")
            } else {
                // Should error when no default and no environment value
                assert.Error(t, err, "Should error when no default allowed and no environment value")
            }
        })
    }
}
```

### Parser Extension Contract

```go
func TestParserExtensionContract(t *testing.T) {
    // Contract: Custom parsers should integrate seamlessly with built-in parsers
    // and maintain the same error handling and validation behavior

    type EmailAddress string

    parseEmail := func(s string) (EmailAddress, error) {
        if s == "" {
            return "", errors.New("email cannot be empty")
        }
        if !strings.Contains(s, "@") {
            return "", errors.New("email must contain @")
        }
        return EmailAddress(s), nil
    }

    registry := genv.NewDefaultRegistry() // Start with built-in parsers
    genv.RegisterTypedParserOn(registry, parseEmail)

    env := genv.New(genv.WithRegistry(registry))

    tests := []struct {
        name           string
        setup          func(*testing.T)
        expectError    bool
        validateResult func(*testing.T, EmailAddress, string, bool, error)
    }{
        {
            name: "valid email with built-in string",
            setup: func(t *testing.T) {
                t.Setenv("EMAIL", "test@example.com")
                t.Setenv("NAME", "John Doe")
            },
            expectError: false,
            validateResult: func(t *testing.T, email EmailAddress, name string, debug bool, err error) {
                assert.NoError(t, err)
                assert.Equal(t, EmailAddress("test@example.com"), email)
                assert.Equal(t, "John Doe", name)
            },
        },
        {
            name: "invalid email format",
            setup: func(t *testing.T) {
                t.Setenv("EMAIL", "invalid-email")
                t.Setenv("NAME", "John Doe")
            },
            expectError: true,
            validateResult: func(t *testing.T, email EmailAddress, name string, debug bool, err error) {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), "must contain @")
            },
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            tc.setup(t)

            var email EmailAddress
            var name string
            var debug bool

            // Mix custom and built-in types
            env.Var("EMAIL").parseOne[EmailAddress](&email)
            env.Var("NAME").String(&name)
            env.Var("DEBUG").Default("false").Bool(&debug)

            err := env.Parse()

            tc.validateResult(t, email, name, debug, err)
        })
    }
}
```

## Architectural Benefits

### 1. Interface Compliance Verification
- **Consistent Behavior**: All implementations follow same behavioral contracts
- **Regression Detection**: Changes that break contracts are caught immediately
- **API Stability**: Interface consistency maintained across different configurations

### 2. Cross-Configuration Validation
- **Implementation Independence**: Different configurations tested with same expectations
- **Isolation Verification**: Components don't interfere with each other
- **Behavioral Predictability**: Same inputs produce consistent outputs

### 3. Integration Testing
- **End-to-End Validation**: Complete workflows tested for behavioral consistency
- **Component Interaction**: Verify components work together as expected
- **Real-World Scenarios**: Test actual usage patterns and combinations

### 4. Quality Assurance
- **Contract Documentation**: Tests serve as behavioral specification
- **Breaking Change Detection**: Contract violations identified automatically
- **Confidence Building**: Comprehensive behavioral validation increases confidence

## Design Decisions

### Why Contract-Based Testing?

1. **Behavioral Focus**: Tests focus on behavior rather than implementation details
2. **Multiple Implementations**: Same contracts apply to different configurations
3. **Integration Validation**: Ensures components work together correctly
4. **Regression Prevention**: Behavioral contracts prevent breaking changes

### Contract Definition Strategy

**Explicit Contracts**: Define expected behavior clearly in test names and assertions
- Contract violations should cause test failures
- Contracts should be technology-agnostic when possible
- Contract tests should be independent of implementation details

### Testing Multiple Implementations

```go
// Pattern for testing multiple implementations against same contract
func TestImplementationContract(t *testing.T) {
    implementations := []struct {
        name   string
        setup  func() *genv.Genv
    }{
        {
            name: "default registry",
            setup: func() *genv.Genv {
                return genv.New()
            },
        },
        {
            name: "custom registry",
            setup: func() *genv.Genv {
                registry := genv.NewRegistry()
                // Add custom parsers
                return genv.New(genv.WithRegistry(registry))
            },
        },
    }

    for _, impl := range implementations {
        t.Run(impl.name, func(t *testing.T) {
            env := impl.setup()

            // Apply same contract tests to all implementations
            validateParsingContract(t, env)
            validateErrorHandlingContract(t, env)
            validateDefaultValueContract(t, env)
        })
    }
}
```

## Testing Strategies

### Contract Violation Detection

```go
func TestContractViolationDetection(t *testing.T) {
    // This test should fail if contracts are violated

    env1 := genv.New()
    env2 := genv.New()

    // Same configuration should produce same results
    result1 := configureEnvironment(env1, standardConfig)
    result2 := configureEnvironment(env2, standardConfig)

    // Contract: Identical configurations should produce identical results
    assert.Equal(t, result1, result2, "Contract violation: identical configs produced different results")
}
```

### Cross-Implementation Consistency

```go
func TestCrossImplementationConsistency(t *testing.T) {
    testData := []struct {
        config   Config
        expected Result
    }{
        // Test data representing behavioral contracts
    }

    implementations := getAllImplementations()

    for _, impl := range implementations {
        for _, test := range testData {
            t.Run(fmt.Sprintf("%s_%s", impl.name, test.name), func(t *testing.T) {
                result := impl.execute(test.config)

                // Contract: All implementations should produce same result for same config
                assert.Equal(t, test.expected, result,
                    "Implementation %s violated behavioral contract", impl.name)
            })
        }
    }
}
```

### Behavioral Regression Testing

```go
func TestBehavioralRegression(t *testing.T) {
    // Known good behaviors that must be maintained
    knownBehaviors := []struct {
        description string
        test        func(*testing.T)
    }{
        {
            description: "empty env var with default uses default",
            test: func(t *testing.T) {
                env := genv.New()
                result := env.Var("EMPTY_VAR").Default("default").NewString()

                err := env.Parse()
                assert.NoError(t, err)
                assert.Equal(t, "default", result)
            },
        },
        {
            description: "invalid int format produces descriptive error",
            test: func(t *testing.T) {
                t.Setenv("INT_VAR", "not-an-int")

                env := genv.New()
                env.Var("INT_VAR").NewInt()

                err := env.Parse()
                assert.Error(t, err)
                assert.Contains(t, err.Error(), "INT_VAR")
                assert.Contains(t, err.Error(), "invalid syntax")
            },
        },
    }

    for _, behavior := range knownBehaviors {
        t.Run(behavior.description, behavior.test)
    }
}
```

## Performance Considerations

### Contract Test Efficiency
- Run contract tests in parallel when possible
- Cache expensive setup operations
- Use table-driven tests for multiple contract scenarios

### Memory Usage
- Clean up test environments between contract validations
- Avoid memory leaks in long-running contract test suites
- Consider resource cleanup in contract test teardown

## Anti-Patterns to Avoid

1. **Implementation-Specific Contracts**: Contracts should be implementation-agnostic
2. **Over-Specific Contracts**: Focus on behavior, not implementation details
3. **Weak Assertions**: Contract violations should cause clear test failures
4. **Contract Overlap**: Avoid redundant contract testing across different test suites

## Related Patterns

- **Table-Driven Tests**: Contract scenarios often use table-driven structure
- **Interface Segregation**: Contracts validate specific interface behaviors
- **Integration Testing**: Contract tests often serve as integration tests
- **Behavioral Testing**: Focus on behavior rather than state verification