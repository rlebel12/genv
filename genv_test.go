package genv

import (
	"testing"

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
		opts          []envVarOpt
		expectedValue string
		expectedFound bool
	}{
		"Defined":     {"val", nil, "val", true},
		"Undefined":   {"", nil, "", false},
		"WithOptions": {"val", []envVarOpt{func(e *Var) { e.value = "opts" }}, "opts", true},
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
		ManyInt(genv.WithSplitKey(";"))
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
			opts := []fallbackOpt{func(fb *fallback) { customOpt.optFunc() }}
			fallbackOpts := make([]fallbackOpt, len(test.opts))
			for i, opt := range test.opts {
				fallbackOpts[i] = genv.WithAllowDefault(opt)
			}
			opts = append(opts, fallbackOpts...)

			actual := genv.Var("TEST_VAR").Default("default", opts...).String()
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
			got := env.Var("TEST_VAR").String()
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
			got := env.Var("TEST_VAR").ManyString()
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
			got := env.Var("TEST_VAR").Bool()
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
			got := env.Var("TEST_VAR").ManyBool()
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
			got := env.Var("TEST_VAR").Int()
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
			got := env.Var("TEST_VAR").ManyInt()
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
			got := env.Var("TEST_VAR").Float64()
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
			got := env.Var("TEST_VAR").ManyFloat64()
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
		"Valid":   {"http://example.com:8080", "http://example.com:8080", false},
		"Invalid": {"http://invalid url", "", true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("TEST_VAR", test.value)
			env := New()
			got := env.Var("TEST_VAR").URL()
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
		"Valid":   {"http://example.com:8080,http://example.com:8081", ",", []string{"http://example.com:8080", "http://example.com:8081"}, false},
		"Invalid": {"http://invalid url", ",", nil, true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("TEST_VAR", test.value)
			env := New()
			got := env.Var("TEST_VAR").ManyURL()
			gotErr := env.Parse()
			for i, want := range test.expected {
				assert.Equal(t, want, (*got)[i].String())
			}
			assert.Equal(t, test.wantErr, gotErr != nil)
		})
	}
}

func TestOptionalEmpty(t *testing.T) {
	env := New()
	got := env.Var("TEST_VAR").Optional().String()
	assert.NoError(t, env.Parse())
	assert.Equal(t, "", *got)
}

func TestParseManyNoSplitKey(t *testing.T) {
	env := New()
	got := env.Var("TEST_VAR").ManyInt(env.WithSplitKey(""))
	assert.Error(t, env.Parse())
	assert.Nil(t, *got)
}
