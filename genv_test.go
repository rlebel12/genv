package genv

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewGenv(t *testing.T) {
	genv := New()
	assert.NotNil(t, genv)
	gotAllow, gotErr := genv.allowDefault(genv)
	assert.NoError(t, gotErr)
	assert.False(t, gotAllow)
	assert.Equal(t, ",", genv.splitKey)
}

func TestWithDefaultSplitKey(t *testing.T) {
	genv := New(WithSplitKey(";"))
	assert.Equal(t, ";", genv.splitKey)
}

func TestWithAllowDefault(t *testing.T) {
	genv := New(WithAllowDefault(func(*Genv) (bool, error) { return true, nil }))
	gotAllow, gotErr := genv.allowDefault(genv)
	assert.NoError(t, gotErr)
	assert.True(t, gotAllow)
}

func TestNew(t *testing.T) {
	for name, test := range map[string]struct {
		value         string
		opts          []Opt[Var]
		expectedValue string
		expectedFound bool
	}{
		"Defined":     {"val", nil, "val", true},
		"Undefined":   {"", nil, "", false},
		"WithOptions": {"val", []Opt[Var]{func(e *Var) { e.value = "opts" }}, "opts", true},
	} {
		t.Run(name, func(t *testing.T) {
			const key = "TEST_VAR"
			if test.value != "" {
				t.Setenv(key, test.value)
			}
			genv := New()
			actual := genv.Var(key, test.opts...)
			expected := &Var{
				key:      key,
				value:    test.expectedValue,
				found:    test.expectedFound,
				splitKey: ",",
				genv:     genv,
			}
				expected.allowDefault, actual.allowDefault = nil, nil
			assert.Equal(t, *expected, *actual)
		})
	}
}

func TestOptional(t *testing.T) {
	t.Run("Required", func(t *testing.T) {
		genv := New()
		ev := genv.Var("TEST_VAR")
		assert.Equal(t, false, ev.optional)
	})

	t.Run("Optional", func(t *testing.T) {
		genv := New()
		ev := genv.Var("TEST_VAR").Optional()
		assert.Equal(t, true, ev.optional)
	})
}

func TestWithSplitKey(t *testing.T) {
	genv := New(WithSplitKey(","))
	actual := genv.Var("TEST_VAR").
		Default("123;456", genv.WithAllowDefaultAlways()).
		NewInts(genv.WithSplitKey(";"))
	err := genv.Parse()
	assert.NoError(t, err)
	assert.Equal(t, []int{123, 456}, *actual)
}

type MockDefaultOpt struct {
	mock.Mock
}

func (m *MockDefaultOpt) optFunc() {
	_ = m.Called()
}

func TestDefault(t *testing.T) {
	allow := func(*Genv) (bool, error) { return true, nil }
	disallow := func(*Genv) (bool, error) { return false, nil }
	for name, test := range map[string]struct {
		found         bool
		opts          []allowFunc
		expectedValue string
		wantErr       error
	}{
		"Found":              {true, nil, "val", nil},
		"FoundAllowed":       {true, []allowFunc{allow}, "val", nil},
		"FoundDisallowed":    {true, []allowFunc{disallow}, "val", nil},
		"NotFound":           {false, nil, "default", nil},
		"NotFoundAllowed":    {false, []allowFunc{allow}, "default", nil},
		"NotFoundDisallowed": {false, []allowFunc{disallow}, "", ErrRequiredEnvironmentVariable},
	} {
		t.Run(name, func(t *testing.T) {
			genv := New(WithAllowDefault(func(*Genv) (bool, error) { return true, nil }))

			if test.found {
				t.Setenv("TEST_VAR", "val")
			}

			customOpt := new(MockDefaultOpt)
			customOpt.On("optFunc")
			opts := []Opt[fallback]{func(fb *fallback) { customOpt.optFunc() }}
			fallbackOpts := make([]Opt[fallback], len(test.opts))
			for i, opt := range test.opts {
				fallbackOpts[i] = genv.WithAllowDefault(opt)
			}
			opts = append(opts, fallbackOpts...)

			actual := genv.Var("TEST_VAR").Default("default", opts...).NewString()
			err := genv.Parse()
			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expectedValue, *actual)
			customOpt.AssertExpectations(t)
		})
	}
}

func TestString(t *testing.T) {
	for name, test := range map[string]struct {
		giveValue string
		wantValue string
		wantErr   bool
	}{
		"Valid":   {"val", "val", false},
		"Invalid": {"", "", true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("TEST_VAR", test.giveValue)
			env := New()
			got := env.Var("TEST_VAR").NewString()
			gotErr := env.Parse()
			assert.Equal(t, test.wantValue, *got)
			assert.Equal(t, test.wantErr, gotErr != nil)
		})
	}
}

func TestManyString(t *testing.T) {
	for name, test := range map[string]struct {
		value    string
		splitKey string
		expected []string
		wantErr  bool
	}{
		"Valid":   {"foo,bar,baz", ",", []string{"foo", "bar", "baz"}, false},
		"Invalid": {"", ",", nil, true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("TEST_VAR", test.value)
			env := New()
			got := env.Var("TEST_VAR").NewStrings()
			gotErr := env.Parse()
			assert.Equal(t, test.expected, *got)
			assert.Equal(t, test.wantErr, gotErr != nil)
		})
	}
}

func TestBool(t *testing.T) {
	for name, test := range map[string]struct {
		value    string
		expected bool
		wantErr  bool
	}{
		"ValidTrue":  {"true", true, false},
		"ValidFalse": {"false", false, false},
		"Invalid":    {"", false, true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("TEST_VAR", test.value)
			env := New()
			got := env.Var("TEST_VAR").NewBool()
			gotErr := env.Parse()
			assert.Equal(t, test.expected, *got)
			assert.Equal(t, test.wantErr, gotErr != nil)
		})
	}
}

func TestManyBool(t *testing.T) {
	for name, test := range map[string]struct {
		value    string
		splitKey string
		expected []bool
		wantErr  bool
	}{
		"Valid":   {"true,false", ",", []bool{true, false}, false},
		"Invalid": {"", ",", nil, true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("TEST_VAR", test.value)
			env := New()
			got := env.Var("TEST_VAR").NewBools()
			gotErr := env.Parse()
			assert.Equal(t, test.expected, *got)
			assert.Equal(t, test.wantErr, gotErr != nil)
		})
	}
}

func TestInt(t *testing.T) {
	for name, test := range map[string]struct {
		value    string
		expected int
		wantErr  bool
	}{
		"Valid":   {"123", 123, false},
		"Invalid": {"", 0, true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("TEST_VAR", test.value)
			env := New()
			got := env.Var("TEST_VAR").NewInt()
			gotErr := env.Parse()
			assert.Equal(t, test.expected, *got)
			assert.Equal(t, test.wantErr, gotErr != nil)
		})
	}
}

func TestManyInt(t *testing.T) {
	for name, test := range map[string]struct {
		value    string
		splitKey string
		expected []int
		wantErr  bool
	}{
		"Valid":   {"123,456", ",", []int{123, 456}, false},
		"Invalid": {"", ",", nil, true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("TEST_VAR", test.value)
			env := New()
			got := env.Var("TEST_VAR").NewInts()
			gotErr := env.Parse()
			assert.Equal(t, test.expected, *got)
			assert.Equal(t, test.wantErr, gotErr != nil)
		})
	}
}

func TestFloat64(t *testing.T) {
	for name, test := range map[string]struct {
		value    string
		expected float64
		wantErr  bool
	}{
		"Valid":   {"123.456", 123.456, false},
		"Invalid": {"", 0.0, true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("TEST_VAR", test.value)
			env := New()
			got := env.Var("TEST_VAR").NewFloat64()
			gotErr := env.Parse()
			assert.Equal(t, test.expected, *got)
			assert.Equal(t, test.wantErr, gotErr != nil)
		})
	}
}

func TestManyFloat64(t *testing.T) {
	for name, test := range map[string]struct {
		value    string
		splitKey string
		expected []float64
		wantErr  bool
	}{
		"Valid":   {"123.456,456.789", ",", []float64{123.456, 456.789}, false},
		"Invalid": {"", ",", nil, true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("TEST_VAR", test.value)
			env := New()
			got := env.Var("TEST_VAR").NewFloat64s()
			gotErr := env.Parse()
			assert.Equal(t, test.expected, *got)
			assert.Equal(t, test.wantErr, gotErr != nil)
		})
	}
}

func TestURL(t *testing.T) {
	for name, test := range map[string]struct {
		value    string
		expected string
		wantErr  bool
	}{
		"Valid":   {"https://example.com:8080", "https://example.com:8080", false},
		"Invalid": {"http://invalid url", "", true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("TEST_VAR", test.value)
			env := New()
			got := env.Var("TEST_VAR").NewURL()
			gotErr := env.Parse()
			assert.Equal(t, test.expected, got.String())
			assert.Equal(t, test.wantErr, gotErr != nil)
		})
	}
}

func TestManyURL(t *testing.T) {
	for name, test := range map[string]struct {
		value    string
		splitKey string
		expected []string
		wantErr  bool
	}{
		"Valid":   {"https://example.com:8080,https://example.com:8081", ",", []string{"https://example.com:8080", "https://example.com:8081"}, false},
		"Invalid": {"http://invalid url", ",", nil, true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("TEST_VAR", test.value)
			env := New()
			got := env.Var("TEST_VAR").NewURLs()
			gotErr := env.Parse()
			for i, want := range test.expected {
				assert.Equal(t, want, (*got)[i].String())
			}
			assert.Equal(t, test.wantErr, gotErr != nil)
		})
	}
}

func TestUUID(t *testing.T) {
	giveIDRaw := "123e4567-e89b-12d3-a456-426614174000"
	for name, test := range map[string]struct {
		value    string
		expected uuid.UUID
		wantErr  bool
	}{
		"Valid":   {giveIDRaw, uuid.MustParse(giveIDRaw), false},
		"Invalid": {"invalid uuid", uuid.Nil, true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("TEST_VAR", test.value)
			env := New()
			got := env.Var("TEST_VAR").NewUUID()
			gotErr := env.Parse()
			assert.Equal(t, test.wantErr, gotErr != nil)
			assert.Equal(t, test.expected, *got)
		})
	}
}

func TestManyUUID(t *testing.T) {
	giveIDRaw := "123e4567-e89b-12d3-a456-426614174000"
	for name, test := range map[string]struct {
		value    string
		splitKey string
		expected []uuid.UUID
		wantErr  bool
	}{
		"Valid":   {giveIDRaw + "," + giveIDRaw, ",", []uuid.UUID{uuid.MustParse(giveIDRaw), uuid.MustParse(giveIDRaw)}, false},
		"Invalid": {"invalid uuid", ",", nil, true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("TEST_VAR", test.value)
			env := New()
			got := env.Var("TEST_VAR").NewUUIDs()
			gotErr := env.Parse()
			for i, want := range test.expected {
				assert.Equal(t, want, (*got)[i])
			}
			assert.Equal(t, test.wantErr, gotErr != nil)
		})
	}
}

func TestOptionalEmpty(t *testing.T) {
	env := New()
	got := env.Var("TEST_VAR").Optional().NewString()
	assert.NoError(t, env.Parse())
	assert.Equal(t, "", *got)
}

func TestParseManyNoSplitKey(t *testing.T) {
	env := New()
	got := env.Var("TEST_VAR").NewInts(env.WithSplitKey(""))
	assert.Error(t, env.Parse())
	assert.Nil(t, *got)
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	assert.NotNil(t, registry)

	parser, exists := registry.get(reflect.TypeOf(""))
	assert.False(t, exists)
	assert.Equal(t, Parser{}, parser)
}

func TestNewDefaultRegistry(t *testing.T) {
	registry := NewDefaultRegistry()
	assert.NotNil(t, registry)

	testCases := []any{
		"",          // string
		false,       // bool
		0,           // int
		0.0,         // float64
		url.URL{},   // url.URL
		uuid.UUID{}, // uuid.UUID
	}

	for _, testCase := range testCases {
		targetType := reflect.TypeOf(testCase)
		parser, exists := registry.get(targetType)
		assert.True(t, exists, "parser should exist for type %s", targetType)
		assert.NotNil(t, parser.ParseFn)
		assert.Equal(t, targetType, parser.TargetType())
	}
}

func TestParserRegistryRegisterTypedParser(t *testing.T) {
	type customType string

	registry := NewRegistry(
		WithParser(func(s string) (customType, error) {
			return customType("custom:" + s), nil
		}),
	)

	var zero customType
	targetType := reflect.TypeOf(zero)
	parser, exists := registry.get(targetType)
	assert.True(t, exists)

	result, err := parser.Parse("test")
	assert.NoError(t, err)
	assert.Equal(t, customType("custom:test"), result)
}

func TestGenvWithCustomRegistry(t *testing.T) {
	type customType string

	registry := NewRegistry(
		WithParser(func(s string) (customType, error) {
			return customType("parsed:" + s), nil
		}),
	)

	t.Setenv("CUSTOM_VAR", "value")
	env := New(WithRegistry(registry))

	_ = env.Var("CUSTOM_VAR").NewString()
	err := env.Parse()
	assert.Error(t, err) // No string parser in empty registry
}

func TestGenvWithDefaultRegistry(t *testing.T) {
	t.Setenv("TEST_VAR", "hello")
	env := New()
	got := env.Var("TEST_VAR").NewString()
	err := env.Parse()
	assert.NoError(t, err)
	assert.Equal(t, "hello", *got)
}

func TestRegistryIsolation(t *testing.T) {
	registry1 := NewRegistry()
	registry2 := NewRegistry()

	type customType1 string
	type customType2 string

	WithParser(func(s string) (customType1, error) {
		return customType1("registry1:" + s), nil
	})

	WithParser(func(s string) (customType2, error) {
		return customType2("registry2:" + s), nil
	})

	var zero1 customType1
	var zero2 customType2

	_, exists1in1 := registry1.get(reflect.TypeOf(zero1))
	_, exists2in1 := registry1.get(reflect.TypeOf(zero2))
	_, exists1in2 := registry2.get(reflect.TypeOf(zero1))
	_, exists2in2 := registry2.get(reflect.TypeOf(zero2))

	assert.True(t, exists1in1)  // registry1 has customType1
	assert.False(t, exists2in1) // registry1 doesn't have customType2
	assert.False(t, exists1in2) // registry2 doesn't have customType1
	assert.True(t, exists2in2)  // registry2 has customType2
}

// Extended Registry API Tests

func TestParserRegistryDuplicateRegistration(t *testing.T) {
	registry := NewRegistry()
	type testType string

	WithParser(func(s string) (testType, error) {
		return testType("first:" + s), nil
	})

	assert.Panics(t, func() {
		WithParser(func(s string) (testType, error) {
			return testType("second:" + s), nil
		})
	})
}


func TestRegistryCloneBehavior(t *testing.T) {
	registry := NewRegistry()
	type cloneTestType string
	WithParser(func(s string) (cloneTestType, error) {
		return cloneTestType("original:" + s), nil
	})

	env := New(WithRegistry(registry))

	clonedEnv := env.Clone()
	assert.Equal(t, env.registry, clonedEnv.registry)

	var zero cloneTestType
	parser, exists := clonedEnv.registry.get(reflect.TypeOf(zero))
	assert.True(t, exists)

	result, err := parser.Parse("test")
	assert.NoError(t, err)
	assert.Equal(t, cloneTestType("original:test"), result)
}

func TestRegistryEmptyRegistryWithBuiltinTypes(t *testing.T) {
	emptyRegistry := NewRegistry()
	t.Setenv("TEST_VAR", "hello")
	env := New(WithRegistry(emptyRegistry))

	_ = env.Var("TEST_VAR").NewString()
	err := env.Parse()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no parser registered for type string")
}

func TestRegistryWithTime(t *testing.T) {
	registry := NewDefaultRegistry()
	var zero time.Time
	parser, exists := registry.get(reflect.TypeOf(zero))
	assert.True(t, exists)

	timeStr := "2023-01-01T12:00:00Z"
	result, err := parser.Parse(timeStr)
	assert.NoError(t, err)
	expected, _ := time.Parse(time.RFC3339, timeStr)
	assert.Equal(t, expected, result)
}

// Custom Type Integration Tests

func TestCustomTypeEndToEnd(t *testing.T) {
	type UserID string

	registry := NewDefaultRegistry()
	WithParser(func(s string) (UserID, error) {
		if s == "" {
			return "", errors.New("UserID cannot be empty")
		}
		if !strings.HasPrefix(s, "user_") {
			return UserID("user_" + s), nil
		}
		return UserID(s), nil
	})

	t.Setenv("USER_ID", "12345")
	env := New(WithRegistry(registry))
	var userID UserID
	Type(env.Var("USER_ID"), &userID)
	err := env.Parse()
	assert.NoError(t, err)
	assert.Equal(t, UserID("user_12345"), userID)

	t.Setenv("USER_ID_2", "user_98765")
	var userID2 UserID
	Type(env.Var("USER_ID_2"), &userID2)
	err = env.Parse()
	assert.NoError(t, err)
	assert.Equal(t, UserID("user_98765"), userID2)

	t.Setenv("USER_ID_EMPTY", "")
	var userID3 UserID
	Type(env.Var("USER_ID_EMPTY"), &userID3)
	err = env.Parse()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "environment variable is empty")
}

func TestCustomTypeWithNewTypePattern(t *testing.T) {
	type Department string

	registry := NewDefaultRegistry()
	WithParser(func(s string) (Department, error) {
		validDepts := []string{"engineering", "marketing", "sales", "hr"}
		dept := strings.ToLower(strings.TrimSpace(s))
		for _, valid := range validDepts {
			if dept == valid {
				return Department(dept), nil
			}
		}
		return "", fmt.Errorf("invalid department: %s", s)
	})

	t.Setenv("DEPARTMENT", "Engineering")
	env := New(WithRegistry(registry))
	dept := NewType[Department](env.Var("DEPARTMENT"))
	err := env.Parse()
	assert.NoError(t, err)
	assert.Equal(t, Department("engineering"), *dept)

	t.Setenv("INVALID_DEPT", "finance")
	invalidDept := NewType[Department](env.Var("INVALID_DEPT"))
	err = env.Parse()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid department: finance")
	_ = invalidDept // Use variable to avoid unused warning
}

// Custom Type Slice Tests

func TestCustomTypeSlices(t *testing.T) {
	type Color string

	registry := NewDefaultRegistry()
	WithParser(func(s string) (Color, error) {
		validColors := []string{"red", "green", "blue", "yellow"}
		color := strings.ToLower(strings.TrimSpace(s))
		for _, valid := range validColors {
			if color == valid {
				return Color(color), nil
			}
		}
		return "", fmt.Errorf("invalid color: %s", s)
	})

	t.Setenv("COLORS", "red,GREEN, Blue ")
	env := New(WithRegistry(registry))
	colors := NewTypes[Color](env.Var("COLORS"))
	err := env.Parse()
	assert.NoError(t, err)
	assert.Equal(t, []Color{"red", "green", "blue"}, *colors)

	t.Setenv("INVALID_COLORS", "red,purple,blue")
	invalidColors := NewTypes[Color](env.Var("INVALID_COLORS"))
	err = env.Parse()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid color: purple")
	_ = invalidColors

	t.Setenv("EMPTY_COLORS", "")
	emptyColors := NewTypes[Color](env.Var("EMPTY_COLORS").Optional())
	err = env.Parse()
	assert.NoError(t, err)
	assert.NotNil(t, emptyColors)
	assert.Len(t, *emptyColors, 0)
}

func TestCustomTypeSlicesWithCustomDelimiter(t *testing.T) {
	type Priority int

	// Create registry with Priority parser
	registry := NewDefaultRegistry()
	WithParser(func(s string) (Priority, error) {
		switch strings.ToLower(s) {
		case "low":
			return Priority(1), nil
		case "medium":
			return Priority(2), nil
		case "high":
			return Priority(3), nil
		case "critical":
			return Priority(4), nil
		default:
			return Priority(0), fmt.Errorf("invalid priority: %s", s)
		}
	})

	t.Setenv("PRIORITIES", "low|medium|high")
	env := New(WithRegistry(registry), WithSplitKey("|"))
	priorities := NewTypes[Priority](env.Var("PRIORITIES"))
	err := env.Parse()
	assert.NoError(t, err)
	assert.Equal(t, []Priority{1, 2, 3}, *priorities)
}

// Custom Type Optional and Default Tests

func TestCustomTypeOptional(t *testing.T) {
	type Status string

	registry := NewDefaultRegistry()
	WithParser(func(s string) (Status, error) {
		if s == "" {
			return "", errors.New("status cannot be empty")
		}
		return Status(strings.ToLower(s)), nil
	})

	env := New(WithRegistry(registry))
	status := NewType[Status](env.Var("MISSING_STATUS").Optional())
	err := env.Parse()
	assert.NoError(t, err)
	assert.Equal(t, Status(""), *status) // Zero value

	requiredStatus := NewType[Status](env.Var("MISSING_REQUIRED_STATUS"))
	err = env.Parse()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "environment variable is empty or unset")
	_ = requiredStatus
}

func TestCustomTypeWithDefault(t *testing.T) {
	type Environment string

	registry := NewDefaultRegistry()
	WithParser(func(s string) (Environment, error) {
		validEnvs := []string{"development", "staging", "production"}
		env := strings.ToLower(strings.TrimSpace(s))
		for _, valid := range validEnvs {
			if env == valid {
				return Environment(env), nil
			}
		}
		return "", fmt.Errorf("invalid environment: %s", s)
	})

	env := New(
		WithRegistry(registry),
		WithAllowDefault(func(*Genv) (bool, error) { return true, nil }),
	)
	environment := NewType[Environment](
		env.Var("APP_ENV").Default("development"),
	)
	err := env.Parse()
	assert.NoError(t, err)
	assert.Equal(t, Environment("development"), *environment)

	t.Setenv("EXPLICIT_ENV", "production")
	explicitEnv := NewType[Environment](
		env.Var("EXPLICIT_ENV").Default("development"),
	)
	err = env.Parse()
	assert.NoError(t, err)
	assert.Equal(t, Environment("production"), *explicitEnv)
}

func TestCustomTypeOptionalWithDefault(t *testing.T) {
	type LogLevel string

	registry := NewDefaultRegistry()
	WithParser(func(s string) (LogLevel, error) {
		level := strings.ToUpper(strings.TrimSpace(s))
		switch level {
		case "DEBUG", "INFO", "WARN", "ERROR":
			return LogLevel(level), nil
		default:
			return "", fmt.Errorf("invalid log level: %s", s)
		}
	})

	env := New(
		WithRegistry(registry),
		WithAllowDefault(func(*Genv) (bool, error) { return true, nil }),
	)
	logLevel := NewType[LogLevel](
		env.Var("LOG_LEVEL").
			Default("INFO").
			Optional(),
	)
	err := env.Parse()
	assert.NoError(t, err)
	assert.Equal(t, LogLevel("INFO"), *logLevel)

	t.Setenv("INVALID_LOG_LEVEL", "INVALID")
	invalidLogLevel := NewType[LogLevel](
		env.Var("INVALID_LOG_LEVEL").
			Default("WARN").
			Optional(),
	)
	err = env.Parse()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log level: INVALID")
	_ = invalidLogLevel
}

// Registry Isolation and Performance Tests

func TestRegistryConcurrentAccess(t *testing.T) {
	registry1 := NewDefaultRegistry()
	registry2 := NewDefaultRegistry()

	type ConcurrentType1 string
	type ConcurrentType2 string

	WithParser(func(s string) (ConcurrentType1, error) {
		return ConcurrentType1("reg1:" + s), nil
	})

	WithParser(func(s string) (ConcurrentType2, error) {
		return ConcurrentType2("reg2:" + s), nil
	})

	const numGoroutines = 100
	results1 := make(chan string, numGoroutines)
	results2 := make(chan string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			env := New(WithRegistry(registry1))
			t.Setenv(fmt.Sprintf("CONCURRENT_VAR1_%d", id), fmt.Sprintf("value%d", id))
			result := NewType[ConcurrentType1](env.Var(fmt.Sprintf("CONCURRENT_VAR1_%d", id)))
			err := env.Parse()
			if err != nil {
				results1 <- "error"
			} else {
				results1 <- string(*result)
			}
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			env := New(WithRegistry(registry2))
			t.Setenv(fmt.Sprintf("CONCURRENT_VAR2_%d", id), fmt.Sprintf("value%d", id))
			result := NewType[ConcurrentType2](env.Var(fmt.Sprintf("CONCURRENT_VAR2_%d", id)))
			err := env.Parse()
			if err != nil {
				results2 <- "error"
			} else {
				results2 <- string(*result)
			}
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		result1 := <-results1
		result2 := <-results2
		assert.Contains(t, result1, "reg1:value")
		assert.Contains(t, result2, "reg2:value")
	}
}

func TestRegistryMemoryIsolation(t *testing.T) {
	type IsolationType string

	registry1 := NewRegistry()
	registry2 := NewRegistry()
	registry3 := NewRegistry()

	WithParser(func(s string) (IsolationType, error) {
		return IsolationType("registry1:" + s), nil
	})

	WithParser(func(s string) (IsolationType, error) {
		return IsolationType("registry2:" + s), nil
	})

	WithParser(func(s string) (IsolationType, error) {
		return IsolationType("registry3:" + s), nil
	})

	var zero IsolationType
	targetType := reflect.TypeOf(zero)

	parser1, exists1 := registry1.get(targetType)
	parser2, exists2 := registry2.get(targetType)
	parser3, exists3 := registry3.get(targetType)

	assert.True(t, exists1)
	assert.True(t, exists2)
	assert.True(t, exists3)

	result1, _ := parser1.Parse("test")
	result2, _ := parser2.Parse("test")
	result3, _ := parser3.Parse("test")

	assert.Equal(t, IsolationType("registry1:test"), result1)
	assert.Equal(t, IsolationType("registry2:test"), result2)
	assert.Equal(t, IsolationType("registry3:test"), result3)
}

// Advanced Scenario Tests

func TestMixedBuiltinAndCustomTypes(t *testing.T) {
	type ServiceConfig struct {
		Port        int
		Host        string
		ServiceName string
		Debug       bool
		Timeout     time.Duration
	}

	type Duration time.Duration
	type ServiceName string

	registry := NewDefaultRegistry()
	
	WithParser(func(s string) (Duration, error) {
		d, err := time.ParseDuration(s)
		if err != nil {
			return Duration(0), err
		}
		return Duration(d), nil
	})

	WithParser(func(s string) (ServiceName, error) {
		if len(s) < 3 {
			return "", errors.New("service name must be at least 3 characters")
		}
		if !strings.HasPrefix(s, "svc-") {
			return ServiceName("svc-" + s), nil
		}
		return ServiceName(s), nil
	})

	t.Setenv("SERVICE_PORT", "8080")
	t.Setenv("SERVICE_HOST", "localhost")
	t.Setenv("SERVICE_NAME", "api")
	t.Setenv("SERVICE_DEBUG", "true")
	t.Setenv("SERVICE_TIMEOUT", "30s")

	env := New(WithRegistry(registry))
	var config ServiceConfig

	env.Var("SERVICE_PORT").Int(&config.Port)
	env.Var("SERVICE_HOST").String(&config.Host)
	Type(env.Var("SERVICE_NAME"), (*ServiceName)(&config.ServiceName))
	env.Var("SERVICE_DEBUG").Bool(&config.Debug)
	Type(env.Var("SERVICE_TIMEOUT"), (*Duration)(&config.Timeout))

	err := env.Parse()
	assert.NoError(t, err)

	assert.Equal(t, 8080, config.Port)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, "svc-api", config.ServiceName)
	assert.Equal(t, true, config.Debug)
	assert.Equal(t, time.Duration(30*time.Second), time.Duration(config.Timeout))
}

func TestRegistryMigrationPattern(t *testing.T) {
	type UserID string
	type LegacyUserID string

	
	legacyRegistry := NewDefaultRegistry()
	WithParser(func(s string) (LegacyUserID, error) {
		return LegacyUserID("legacy:" + s), nil
	})

	newRegistry := NewDefaultRegistry()
	WithParser(func(s string) (UserID, error) {
		return UserID("new:" + s), nil
	})

	t.Setenv("LEGACY_USER", "user123")
	legacyEnv := New(WithRegistry(legacyRegistry))
	var legacyResult LegacyUserID
	Type(legacyEnv.Var("LEGACY_USER"), &legacyResult)
	err := legacyEnv.Parse()
	assert.NoError(t, err)
	assert.Equal(t, LegacyUserID("legacy:user123"), legacyResult)

	t.Setenv("NEW_USER", "user456")
	newEnv := New(WithRegistry(newRegistry))
	var newResult UserID
	Type(newEnv.Var("NEW_USER"), &newResult)
	err = newEnv.Parse()
	assert.NoError(t, err)
	assert.Equal(t, UserID("new:user456"), newResult)

	assert.NotEqual(t, string(legacyResult), string(newResult))
	
	t.Setenv("WRONG_TYPE", "test")
	wrongEnv := New(WithRegistry(newRegistry))
	var wrongResult LegacyUserID
	Type(wrongEnv.Var("WRONG_TYPE"), &wrongResult)
	err = wrongEnv.Parse()
	assert.Error(t, err) // Should fail because LegacyUserID not registered in newRegistry
	assert.Contains(t, err.Error(), "no parser registered for type")
}

func TestComplexValidationScenarios(t *testing.T) {
	type EmailAddress string
	type PhoneNumber string
	type PersonID string

	registry := NewDefaultRegistry()

	WithParser(func(s string) (EmailAddress, error) {
		if !strings.Contains(s, "@") {
			return "", errors.New("invalid email format")
		}
		if !strings.Contains(s, ".") {
			return "", errors.New("email missing domain")
		}
		return EmailAddress(strings.ToLower(s)), nil
	})

	WithParser(func(s string) (PhoneNumber, error) {
		cleaned := strings.ReplaceAll(s, "-", "")
		cleaned = strings.ReplaceAll(cleaned, " ", "")
		if len(cleaned) < 10 {
			return "", errors.New("phone number too short")
		}
		return PhoneNumber(cleaned), nil
	})

	WithParser(func(s string) (PersonID, error) {
		if len(s) < 5 {
			return "", errors.New("person ID too short")
		}
		if !strings.HasPrefix(s, "P") && !strings.HasPrefix(s, "E") {
			return "", errors.New("person ID must start with P or E")
		}
		return PersonID(strings.ToUpper(s)), nil
	})

	testCases := []struct {
		name        string
		email       string
		phone       string
		personID    string
		expectError bool
		errorText   string
	}{
		{
			name:        "All valid",
			email:       "test@example.com",
			phone:       "123-456-7890",
			personID:    "P12345",
			expectError: false,
		},
		{
			name:        "Invalid email",
			email:       "invalid-email",
			phone:       "123-456-7890", 
			personID:    "P12345",
			expectError: true,
			errorText:   "invalid email format",
		},
		{
			name:        "Invalid phone",
			email:       "test@example.com",
			phone:       "123",
			personID:    "P12345",
			expectError: true,
			errorText:   "phone number too short",
		},
		{
			name:        "Invalid person ID",
			email:       "test@example.com",
			phone:       "123-456-7890",
			personID:    "X12345",
			expectError: true,
			errorText:   "person ID must start with P or E",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("TEST_EMAIL", tc.email)
			t.Setenv("TEST_PHONE", tc.phone)
			t.Setenv("TEST_PERSON_ID", tc.personID)

			env := New(WithRegistry(registry))
			email := NewType[EmailAddress](env.Var("TEST_EMAIL"))
			phone := NewType[PhoneNumber](env.Var("TEST_PHONE"))
			personID := NewType[PersonID](env.Var("TEST_PERSON_ID"))

			err := env.Parse()

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorText)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, EmailAddress("test@example.com"), *email)
				assert.Equal(t, PhoneNumber("1234567890"), *phone)
				assert.Equal(t, PersonID("P12345"), *personID)
			}
		})
	}
}

// Tests for the new VarFunc/Parse API
func TestVarFuncAPI_Basic(t *testing.T) {
	t.Setenv("APP_NAME", "TestApp")
	t.Setenv("PORT", "9000")
	t.Setenv("DEBUG", "true")

	env := New()
	var appName string
	var port int
	var debug bool

	err := Parse(env,
		Bind("APP_NAME", &appName),
		Bind("PORT", &port),
		Bind("DEBUG", &debug),
	)

	assert.NoError(t, err)
	assert.Equal(t, "TestApp", appName)
	assert.Equal(t, 9000, port)
	assert.True(t, debug)
}

func TestVarFuncAPI_WithOptions(t *testing.T) {
	t.Setenv("PORT", "8080")

	env := New(WithAllowDefault(func(*Genv) (bool, error) { return true, nil }))
	var port int
	var name string
	var timeout float64

	err := Parse(env,
		Bind("PORT", &port),
		Bind("NAME", &name).Default("DefaultName"),
		Bind("TIMEOUT", &timeout).Optional(),
	)

	assert.NoError(t, err)
	assert.Equal(t, 8080, port)
	assert.Equal(t, "DefaultName", name)
	assert.Equal(t, 0.0, timeout)
}

func TestVarFuncAPI_Slices(t *testing.T) {
	t.Setenv("TAGS", "a,b,c")
	t.Setenv("PORTS", "8080,8081,8082")

	env := New()
	var tags []string
	var ports []int

	err := Parse(env,
		BindMany("TAGS", &tags),
		BindMany("PORTS", &ports),
	)

	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, tags)
	assert.Equal(t, []int{8080, 8081, 8082}, ports)
}

func TestVarFuncAPI_CustomTypes(t *testing.T) {
	type UserID string

	registry := NewDefaultRegistry()
	WithParser(func(s string) (UserID, error) {
		return UserID("user_" + s), nil
	})

	t.Setenv("USER_ID", "12345")
	t.Setenv("USER_IDS", "1,2,3")

	env := New(WithRegistry(registry))
	var userID UserID
	var userIDs []UserID

	err := Parse(env,
		Bind("USER_ID", &userID),
		BindMany("USER_IDS", &userIDs),
	)

	assert.NoError(t, err)
	assert.Equal(t, UserID("user_12345"), userID)
	assert.Equal(t, []UserID{"user_1", "user_2", "user_3"}, userIDs)
}

func TestVarFuncAPI_ComplexScenario(t *testing.T) {
	type Config struct {
		AppName  string
		Port     int
		Debug    bool
		Timeout  float64
		Features []string
		APIKey   string
	}

	t.Setenv("APP_NAME", "MyApp")
	t.Setenv("PORT", "8080")
	t.Setenv("DEBUG", "true")
	t.Setenv("FEATURES", "auth,api,web")

	env := New(WithAllowDefault(func(*Genv) (bool, error) { return true, nil }))
	var cfg Config

	err := Parse(env,
		Bind("APP_NAME", &cfg.AppName),
		Bind("PORT", &cfg.Port),
		Bind("DEBUG", &cfg.Debug),
		Bind("TIMEOUT", &cfg.Timeout).Default("30.5"),
		BindMany("FEATURES", &cfg.Features),
		Bind("API_KEY", &cfg.APIKey).Default("default-key").Optional(),
	)

	assert.NoError(t, err)
	assert.Equal(t, "MyApp", cfg.AppName)
	assert.Equal(t, 8080, cfg.Port)
	assert.True(t, cfg.Debug)
	assert.Equal(t, 30.5, cfg.Timeout)
	assert.Equal(t, []string{"auth", "api", "web"}, cfg.Features)
	assert.Equal(t, "default-key", cfg.APIKey)
}
