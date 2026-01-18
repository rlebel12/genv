package genv

import (
	"net/url"
	"reflect"
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
	type customType1 string
	type customType2 string

	registry1 := NewRegistry(
		WithParser(func(s string) (customType1, error) {
			return customType1("registry1:" + s), nil
		}),
	)

	registry2 := NewRegistry(
		WithParser(func(s string) (customType2, error) {
			return customType2("registry2:" + s), nil
		}),
	)

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
	type testType string

	// Should panic when registering same type twice
	assert.Panics(t, func() {
		NewRegistry(
			WithParser(func(s string) (testType, error) {
				return testType("first:" + s), nil
			}),
			WithParser(func(s string) (testType, error) {
				return testType("second:" + s), nil
			}),
		)
	})
}

func TestRegistryCloneBehavior(t *testing.T) {
	type cloneTestType string

	registry := NewRegistry(
		WithParser(func(s string) (cloneTestType, error) {
			return cloneTestType("original:" + s), nil
		}),
	)

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

	registry := NewDefaultRegistry(
		WithParser(func(s string) (UserID, error) {
			return UserID("user_" + s), nil
		}),
	)

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
