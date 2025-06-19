package genv

import (
	"testing"

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
			// We cannot test function equality
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
