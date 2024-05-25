package genv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewGenv(t *testing.T) {
	genv := New()
	assert.NotNil(t, genv)
	assert.False(t, genv.allowDefault(genv))
	assert.Equal(t, ",", genv.splitKey)
}

func TestWithDefaultSplitKey(t *testing.T) {
	genv := New(WithSplitKey(";"))
	assert.Equal(t, ";", genv.splitKey)
}

func TestWithAllowDefault(t *testing.T) {
	genv := New(WithAllowDefault(func(*Genv) bool { return true }))
	assert.True(t, genv.allowDefault(genv))
}

func TestConstructor(t *testing.T) {
	for name, test := range map[string]struct {
		fn func(ev *Genv, key string, opts ...envVarOpt) *Var
	}{
		"New": {(*Genv).Var},
	} {
		t.Run(name, func(t *testing.T) {
			fn := test.fn
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
					actual := fn(genv, key, test.opts...)
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
		})
	}
}

func TestValidate(t *testing.T) {
	t.Run("Required", func(t *testing.T) {
		t.Run("Present", func(t *testing.T) {
			t.Setenv("TEST_VAR", "val")
			genv := New()
			ev := genv.Var("TEST_VAR")
			assert.Nil(t, ev.validate())
		})

		t.Run("Absent", func(t *testing.T) {
			t.Setenv("TEST_VAR", "")
			genv := New()
			ev := genv.Var("TEST_VAR")
			assert.Error(t, ev.validate())
		})
	})

	t.Run("Optional", func(t *testing.T) {
		t.Run("Present", func(t *testing.T) {
			t.Setenv("TEST_VAR", "val")
			genv := New()
			ev := genv.Var("TEST_VAR").Optional()
			assert.Nil(t, ev.validate())
		})

		t.Run("Absent", func(t *testing.T) {
			t.Setenv("TEST_VAR", "")
			genv := New()
			ev := genv.Var("TEST_VAR").Optional()
			assert.Nil(t, ev.validate())
		})
	})
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
	assert.Equal(t, []int{123, 456}, actual)
}

type MockDefaultOpt struct {
	mock.Mock
}

func (m *MockDefaultOpt) optFunc() {
	_ = m.Called()
}

func TestDefault(t *testing.T) {
	allow := func(*Genv) bool { return true }
	disallow := func(*Genv) bool { return false }
	for name, test := range map[string]struct {
		found         bool
		opts          []func(*Genv) bool
		expectedValue string
	}{
		"Found":              {true, nil, "val"},
		"FoundAllowed":       {true, []func(*Genv) bool{allow}, "val"},
		"FoundDisallowed":    {true, []func(*Genv) bool{disallow}, "val"},
		"NotFound":           {false, nil, "default"},
		"NotFoundAllowed":    {false, []func(*Genv) bool{allow}, "default"},
		"NotFoundDisallowed": {false, []func(*Genv) bool{disallow}, ""},
	} {
		t.Run(name, func(t *testing.T) {
			genv := New(WithAllowDefault(func(*Genv) bool { return true }))

			if test.found {
				t.Setenv("TEST_VAR", "val")
			}

			customOpt := new(MockDefaultOpt)
			customOpt.On("optFunc")
			opts := []defaultOpt{func(fb *fallback) { customOpt.optFunc() }}
			fallbackOpts := make([]defaultOpt, len(test.opts))
			for i, opt := range test.opts {
				fallbackOpts[i] = genv.WithAllowDefault(opt)
			}
			opts = append(opts, fallbackOpts...)
			actual := (*Var).Default(
				genv.Var("TEST_VAR"),
				"default",
				opts...,
			)
			assert.Equal(t, test.expectedValue, actual.value)
			customOpt.AssertExpectations(t)
		})
	}
}

func TestEVarString(t *testing.T) {
	for _, test := range []struct {
		name     string
		value    string
		expected string
		panics   bool
	}{
		{"Valid", "val", "val", false},
		{"Invalid", "", "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ev := Var{key: "TEST_VAR", value: test.value}
			if test.panics {
				assert.Panics(t, func() { _ = ev.String() })
			} else {
				assert.Equal(t, test.expected, ev.String())
			}
		})
	}
}

func TestEVarTryString(t *testing.T) {
	for _, test := range []struct {
		name     string
		value    string
		expected string
		err      bool
	}{
		{"Valid", "val", "val", false},
		{"Invalid", "", "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ev := Var{key: "TEST_VAR", value: test.value}
			actual, err := ev.TryString()
			if test.err {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expected, actual)
			}
		})
	}
}

func TestManyEvarString(t *testing.T) {
	t.Run(("Valid"), func(t *testing.T) {
		ev := &Var{key: "TEST_VAR", value: "val1,val2", splitKey: ","}
		assert.Equal(t, []string{"val1", "val2"}, ev.ManyString())
	})

	t.Run(("Invalid"), func(t *testing.T) {
		ev := &Var{key: "TEST_VAR", value: ""}
		assert.Panics(t, func() { ev.ManyString() })
	})

	t.Run(("Empty"), func(t *testing.T) {
		ev := &Var{key: "TEST_VAR", value: "", optional: true}
		assert.Empty(t, ev.ManyString())
	})
}

func TestTryManyEvarString(t *testing.T) {
	for _, test := range []struct {
		name     string
		value    string
		optional bool
		expected []string
		err      bool
	}{
		{"Valid", "val1,val2", false, []string{"val1", "val2"}, false},
		{"Empty", "", false, []string{}, true},
		{"Optional", "", true, []string{}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			ev := &Var{key: "TEST_VAR", value: test.value, splitKey: ","}
			if test.optional {
				ev = ev.Optional()
			}
			actual, err := ev.TryManyString()
			if test.err {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expected, actual)
			}
		})
	}
}

func TestEVarBool(t *testing.T) {
	t.Run(("Valid"), func(t *testing.T) {
		ev := Var{key: "TEST_VAR", value: "true"}
		assert.True(t, ev.Bool())
		ev.value = "false"
		assert.False(t, ev.Bool())
	})

	t.Run(("Invalid"), func(t *testing.T) {
		ev := Var{key: "TEST_VAR", value: "invalid"}
		assert.Panics(t, func() { ev.Bool() })
	})
}

func TestEVarTryBool(t *testing.T) {
	for _, test := range []struct {
		name     string
		value    string
		expected bool
		err      bool
	}{
		{"ValidTrue", "true", true, false},
		{"ValidFalse", "false", false, false},
		{"Missing", "", false, true},
		{"Invalid", "invalid", false, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ev := Var{key: "TEST_VAR", value: test.value}
			actual, err := ev.TryBool()
			if test.err {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expected, actual)
			}
		})
	}
}

func TestManyEvarBool(t *testing.T) {
	t.Run(("Valid"), func(t *testing.T) {
		ev := &Var{key: "TEST_VAR", value: "true,false", splitKey: ","}
		assert.Equal(t, []bool{true, false}, ev.ManyBool())
	})

	t.Run(("Invalid"), func(t *testing.T) {
		ev := &Var{key: "TEST_VAR", value: "invalid"}
		assert.Panics(t, func() { ev.ManyBool() })
	})

	t.Run(("Empty"), func(t *testing.T) {
		ev := &Var{key: "TEST_VAR", value: "", optional: true}
		assert.Empty(t, ev.ManyBool())
	})
}

func TestTryManyEvarBool(t *testing.T) {
	for _, test := range []struct {
		name     string
		value    string
		optional bool
		expected []bool
		err      bool
	}{
		{"Valid", "true,false", false, []bool{true, false}, false},
		{"Empty", "", false, []bool{}, true},
		{"Optional", "", true, []bool{}, false},
		{"Invalid", "invalid", false, []bool{}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ev := &Var{key: "TEST_VAR", value: test.value, splitKey: ","}
			if test.optional {
				ev = ev.Optional()
			}
			actual, err := ev.TryManyBool()
			if test.err {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expected, actual)
			}
		})
	}
}

func TestEvarInt(t *testing.T) {
	t.Run(("Valid"), func(t *testing.T) {
		ev := Var{key: "TEST_VAR", value: "123"}
		assert.Equal(t, 123, ev.Int())
	})

	t.Run(("Invalid"), func(t *testing.T) {
		ev := Var{key: "TEST_VAR", value: "invalid"}
		assert.Panics(t, func() { ev.Int() })
	})
}

func TestEvarTryInt(t *testing.T) {
	for _, test := range []struct {
		name     string
		value    string
		optional bool
		expected int
		err      bool
	}{
		{"Valid", "123", false, 123, false},
		{"Empty", "", false, 0, true},
		{"Optional", "", true, 0, false},
		{"Invalid", "invalid", false, 0, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ev := Var{key: "TEST_VAR", value: test.value}
			if test.optional {
				ev = *ev.Optional()
			}
			actual, err := ev.TryInt()
			if test.err {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expected, actual)
			}
		})
	}
}

func TestManyEvarInt(t *testing.T) {
	t.Run(("Valid"), func(t *testing.T) {
		ev := &Var{key: "TEST_VAR", value: "123,456", splitKey: ","}
		assert.Equal(t, []int{123, 456}, ev.ManyInt())
	})

	t.Run(("Invalid"), func(t *testing.T) {
		ev := &Var{key: "TEST_VAR", value: "invalid", splitKey: ","}
		assert.Panics(t, func() { ev.ManyInt() })
	})

	t.Run(("Empty"), func(t *testing.T) {
		ev := &Var{key: "TEST_VAR", value: "", optional: true, splitKey: ","}
		assert.Empty(t, ev.ManyInt())
	})
}

func TestTryManyEvarInt(t *testing.T) {
	for _, test := range []struct {
		name     string
		value    string
		optional bool
		expected []int
		err      bool
	}{
		{"Valid", "123,456", false, []int{123, 456}, false},
		{"Empty", "", false, []int{}, true},
		{"Optional", "", true, []int{}, false},
		{"Invalid", "invalid", false, []int{}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ev := &Var{key: "TEST_VAR", value: test.value, splitKey: ","}
			if test.optional {
				ev = ev.Optional()
			}
			actual, err := ev.TryManyInt()
			if test.err {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expected, actual)
			}
		})
	}
}

func TestEvarFloat64(t *testing.T) {
	t.Run(("Valid"), func(t *testing.T) {
		ev := Var{key: "TEST_VAR", value: "123.456"}
		assert.Equal(t, 123.456, ev.Float64())
	})

	t.Run(("Invalid"), func(t *testing.T) {
		ev := Var{key: "TEST_VAR", value: "invalid"}
		assert.Panics(t, func() { ev.Float64() })
	})
}

func TestEvarTryFloat64(t *testing.T) {
	for _, test := range []struct {
		name     string
		value    string
		optional bool
		expected float64
		err      bool
	}{
		{"Valid", "123.456", false, 123.456, false},
		{"Empty", "", false, 0.0, true},
		{"Optional", "", true, 0.0, false},
		{"Invalid", "invalid", false, 0.0, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ev := Var{key: "TEST_VAR", value: test.value}
			if test.optional {
				ev = *ev.Optional()
			}
			actual, err := ev.TryFloat64()
			if test.err {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expected, actual)
			}
		})
	}
}

func TestEvarManyFloat64(t *testing.T) {
	t.Run(("Valid"), func(t *testing.T) {
		ev := &Var{key: "TEST_VAR", value: "123.456,456.789", splitKey: ","}
		assert.Equal(t, []float64{123.456, 456.789}, ev.ManyFloat64())
	})

	t.Run(("Invalid"), func(t *testing.T) {
		ev := &Var{key: "TEST_VAR", value: "invalid", splitKey: ","}
		assert.Panics(t, func() { ev.ManyFloat64() })
	})

	t.Run(("Empty"), func(t *testing.T) {
		ev := &Var{key: "TEST_VAR", value: "", optional: true, splitKey: ","}
		assert.Empty(t, ev.ManyFloat64())
	})
}

func TestTryManyEvarFloat64(t *testing.T) {
	for _, test := range []struct {
		name     string
		value    string
		optional bool
		expected []float64
		err      bool
	}{
		{"Valid", "123.456,456.789", false, []float64{123.456, 456.789}, false},
		{"Empty", "", false, []float64{}, true},
		{"Optional", "", true, []float64{}, false},
		{"Invalid", "invalid", false, []float64{}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ev := &Var{key: "TEST_VAR", value: test.value, splitKey: ","}
			if test.optional {
				ev = ev.Optional()
			}
			actual, err := ev.TryManyFloat64()
			if test.err {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expected, actual)
			}
		})
	}
}

func TestEvarURL(t *testing.T) {
	t.Run(("Valid"), func(t *testing.T) {
		ev := Var{key: "TEST_VAR", value: "http://example.com:8080"}
		url := ev.URL()
		assert.Equal(t, "http", url.Scheme)
		assert.Equal(t, "example.com", url.Hostname())
		assert.Equal(t, "8080", url.Port())
		assert.Equal(t, "http://example.com:8080", ev.URL().String())
	})

	t.Run(("Invalid"), func(t *testing.T) {
		ev := Var{key: "TEST_VAR", value: "http://invalid url"}
		assert.Panics(t, func() { ev.URL() })
	})
}

func TestEvarTryURL(t *testing.T) {
	for _, test := range []struct {
		name     string
		value    string
		optional bool
		expected string
		err      bool
	}{
		{"Valid", "http://example.com:8080", false, "http://example.com:8080", false},
		{"Empty", "", false, "", true},
		{"Optional", "", true, "", false},
		{"Invalid", "http://invalid url", false, "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ev := Var{key: "TEST_VAR", value: test.value}
			if test.optional {
				ev = *ev.Optional()
			}
			actual, err := ev.TryURL()
			if test.err {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expected, actual.String())
			}
		})
	}
}

func TestManyEvarURL(t *testing.T) {
	t.Run(("Valid"), func(t *testing.T) {
		ev := &Var{key: "TEST_VAR", value: "http://example.com:8080,http://example.com:8081", splitKey: ","}
		urls := ev.ManyURL()
		assert.Equal(t, "http://example.com:8080", urls[0].String())
		assert.Equal(t, "http://example.com:8081", urls[1].String())
	})

	t.Run(("Invalid"), func(t *testing.T) {
		ev := &Var{key: "TEST_VAR", value: "http://invalid url", splitKey: ","}
		assert.Panics(t, func() { ev.ManyURL() })
	})

	t.Run(("Empty"), func(t *testing.T) {
		ev := &Var{key: "TEST_VAR", value: "", optional: true}
		assert.Empty(t, ev.ManyURL())
	})
}

func TestPresent(t *testing.T) {
	present := "present"
	empty := ""
	for name, test := range map[string]struct {
		val      *string
		expected bool
	}{
		"present": {&present, true},
		"empty":   {&empty, false},
		"absent":  {nil, false},
	} {
		t.Run(name, func(t *testing.T) {
			if test.val != nil {
				t.Setenv("TEST_VAR", *test.val)
			}
			actual := newGenv().Present("TEST_VAR")
			assert.Equal(t, test.expected, actual)
		})
	}
}

func newGenv() *Genv {
	return New(WithAllowDefault(func(*Genv) bool { return true }))
}
