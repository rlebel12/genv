package goenvvars

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewGenv(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		genv, err := NewGenv()
		assert.NoError(t, err)
		assert.NotNil(t, genv)
		assert.True(t, genv.defaultAllowFallback())
	})

	t.Run("InvalidEnvironment", func(t *testing.T) {
		t.Setenv("ENV", "INVALID")
		_, err := NewGenv()
		assert.Error(t, err)
	})

}

func TestConstructor(t *testing.T) {
	for name, test := range map[string]struct {
		fn func(ev *Genv, key string, opts ...envVarOpt) *envVar
	}{
		"New": {(*Genv).New},
		"Env": {(*Genv).Env},
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
				"WithOptions": {"val", []envVarOpt{func(e *envVar) { e.value = "opts" }}, "opts", true},
			} {
				t.Run(name, func(t *testing.T) {
					const key = "TEST_VAR"
					if test.value != "" {
						t.Setenv(key, test.value)
					}
					genv, _ := NewGenv()
					actual := fn(genv, key, test.opts...)
					expected := &envVar{
						key:   key,
						value: test.expectedValue,
						found: test.expectedFound,
					}
					// We cannot test function equality
					expected.allowFallback, actual.allowFallback = nil, nil
					assert.Equal(t, *expected, *actual)
				})
			}
		})
	}
}

func TestEnvironmentKey(t *testing.T) {
	genv, _ := NewGenv(EnvironmentKey("CUSTOM_ENV"))
	assert.Equal(t, "CUSTOM_ENV", genv.environmentKey)
}

func TestIsDev(t *testing.T) {
	for name, test := range map[string]struct {
		env      environment
		expected bool
	}{
		"Dev":  {Dev, true},
		"Prod": {Prod, false},
		"Test": {Test, false},
	} {
		t.Run(name, func(t *testing.T) {
			genv, err := NewGenv()
			genv.environment = test.env
			assert.NoError(t, err)
			assert.Equal(t, test.expected, genv.IsDev())
		})
	}
}

func TestIsProd(t *testing.T) {
	for name, test := range map[string]struct {
		env      environment
		expected bool
	}{
		"Dev":  {Dev, false},
		"Prod": {Prod, true},
		"Test": {Test, false},
	} {
		t.Run(name, func(t *testing.T) {
			genv, err := NewGenv()
			genv.environment = test.env
			assert.NoError(t, err)
			assert.Equal(t, test.expected, genv.IsProd())
		})
	}
}

func TestIsTest(t *testing.T) {
	for name, test := range map[string]struct {
		env      environment
		expected bool
	}{
		"Dev":  {Dev, false},
		"Prod": {Prod, false},
		"Test": {Test, true},
	} {
		t.Run(name, func(t *testing.T) {
			genv, err := NewGenv()
			genv.environment = test.env
			assert.NoError(t, err)
			assert.Equal(t, test.expected, genv.IsTest())
		})
	}
}

func TestValidate(t *testing.T) {
	t.Run("Required", func(t *testing.T) {
		t.Run("Present", func(t *testing.T) {
			t.Setenv("TEST_VAR", "val")
			genv, _ := NewGenv()
			ev := genv.New("TEST_VAR")
			assert.Nil(t, ev.validate())
		})

		t.Run("Absent", func(t *testing.T) {
			t.Setenv("TEST_VAR", "")
			genv, _ := NewGenv()
			ev := genv.New("TEST_VAR")
			assert.Error(t, ev.validate())
		})
	})

	t.Run("Optional", func(t *testing.T) {
		t.Run("Present", func(t *testing.T) {
			t.Setenv("TEST_VAR", "val")
			genv, _ := NewGenv()
			ev := genv.New("TEST_VAR").Optional()
			assert.Nil(t, ev.validate())
		})

		t.Run("Absent", func(t *testing.T) {
			t.Setenv("TEST_VAR", "")
			genv, _ := NewGenv()
			ev := genv.New("TEST_VAR").Optional()
			assert.Nil(t, ev.validate())
		})
	})
}

func TestOptional(t *testing.T) {
	t.Run("Required", func(t *testing.T) {
		genv, _ := NewGenv()
		ev := genv.New("TEST_VAR")
		assert.Equal(t, false, ev.optional)
	})

	t.Run("Optional", func(t *testing.T) {
		genv, _ := NewGenv()
		ev := genv.New("TEST_VAR").Optional()
		assert.Equal(t, true, ev.optional)
	})
}

type MockFallbackOpt struct {
	mock.Mock
}

func (m *MockFallbackOpt) optFunc() {
	_ = m.Called()
}

func TestFallingBack(t *testing.T) {
	for name, test := range map[string]struct {
		fn func(ev *envVar, value string, opts ...fallbackOpt) *envVar
	}{
		"Fallback": {(*envVar).Fallback},
		"Default":  {(*envVar).Default},
	} {
		fn := test.fn
		t.Run(name, func(t *testing.T) {
			allow := OverrideAllow(func() bool { return true })
			disallow := OverrideAllow(func() bool { return false })
			for name, test := range map[string]struct {
				found         bool
				opts          []fallbackOpt
				expectedValue string
			}{
				"Found":              {true, nil, "val"},
				"FoundAllowed":       {true, []fallbackOpt{allow}, "val"},
				"FoundDisallowed":    {true, []fallbackOpt{disallow}, "val"},
				"NotFound":           {false, nil, "fallback"},
				"NotFoundAllowed":    {false, []fallbackOpt{allow}, "fallback"},
				"NotFoundDisallowed": {false, []fallbackOpt{disallow}, ""},
			} {
				t.Run(name, func(t *testing.T) {
					genv, err := NewGenv(DefaultAllowFallback(func() bool { return true }))
					assert.NoError(t, err)

					if test.found {
						t.Setenv("TEST_VAR", "val")
					}

					customOpt := new(MockFallbackOpt)
					customOpt.On("optFunc")
					opts := []fallbackOpt{func(fb *fallback) { customOpt.optFunc() }}
					opts = append(opts, test.opts...)
					actual := fn(
						genv.New("TEST_VAR"),
						"fallback",
						opts...,
					)
					assert.Equal(t, test.expectedValue, actual.value)
					customOpt.AssertExpectations(t)
				})
			}
		})

		for name, test := range map[string]struct {
			found         bool
			expectedValue string
		}{
			"Found":    {true, "val"},
			"NotFound": {false, "fallback"},
		} {
			t.Run(name, func(t *testing.T) {
				if test.found {
					t.Setenv("TEST_VAR", "val")
				}
				genv, _ := NewGenv()
				actual := fn(genv.New("TEST_VAR"), "fallback", AllowAlways()).value
				assert.Equal(t, test.expectedValue, actual)
			})
		}
	}
}

func TestPresence(t *testing.T) {
	t.Run("Present", func(t *testing.T) {
		t.Setenv("TEST_VAR", "val")
		assert.True(t, Presence("TEST_VAR"))
	})

	t.Run("Absent", func(t *testing.T) {
		assert.False(t, Presence("TEST_VAR"))
	})

	t.Run("Empty", func(t *testing.T) {
		t.Setenv("TEST_VAR", "")
		assert.False(t, Presence("TEST_VAR"))
	})
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
			ev := envVar{key: "TEST_VAR", value: test.value}
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
			ev := envVar{key: "TEST_VAR", value: test.value}
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

func TestEVarBool(t *testing.T) {
	t.Run(("Valid"), func(t *testing.T) {
		ev := envVar{key: "TEST_VAR", value: "true"}
		assert.True(t, ev.Bool())
		ev.value = "false"
		assert.False(t, ev.Bool())
	})

	t.Run(("Invalid"), func(t *testing.T) {
		ev := envVar{key: "TEST_VAR", value: "invalid"}
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
			ev := envVar{key: "TEST_VAR", value: test.value}
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

func TestEvarInt(t *testing.T) {
	t.Run(("Valid"), func(t *testing.T) {
		ev := envVar{key: "TEST_VAR", value: "123"}
		assert.Equal(t, 123, ev.Int())
	})

	t.Run(("Invalid"), func(t *testing.T) {
		ev := envVar{key: "TEST_VAR", value: "invalid"}
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
			ev := envVar{key: "TEST_VAR", value: test.value}
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

func TestEvarFloat64(t *testing.T) {
	t.Run(("Valid"), func(t *testing.T) {
		ev := envVar{key: "TEST_VAR", value: "123.456"}
		assert.Equal(t, 123.456, ev.Float64())
	})

	t.Run(("Invalid"), func(t *testing.T) {
		ev := envVar{key: "TEST_VAR", value: "invalid"}
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
			ev := envVar{key: "TEST_VAR", value: test.value}
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

func TestEvarURL(t *testing.T) {
	t.Run(("Valid"), func(t *testing.T) {
		ev := envVar{key: "TEST_VAR", value: "http://example.com:8080"}
		url := ev.URL()
		assert.Equal(t, "http", url.Scheme)
		assert.Equal(t, "example.com", url.Hostname())
		assert.Equal(t, "8080", url.Port())
		assert.Equal(t, "http://example.com:8080", ev.URL().String())
	})

	t.Run(("Invalid"), func(t *testing.T) {
		ev := envVar{key: "TEST_VAR", value: "http://invalid url"}
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
			ev := envVar{key: "TEST_VAR", value: test.value}
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

func TestEnvironment(t *testing.T) {
	genv, err := NewGenv(DefaultAllowFallback(func() bool { return false }))
	assert.NoError(t, err)

	t.Run("Specified", func(t *testing.T) {
		envs := environments()
		assert.NotEmpty(t, envs)
		for value, expected := range envs {
			t.Run(value, func(t *testing.T) {
				t.Setenv("ENV", value)
				env, err := newEnvironment(genv)
				assert.NoError(t, err)
				assert.Equal(t, expected, env)
			})
		}

		for name, test := range map[string]struct {
			envValue string
		}{
			"NotString": {"5"},
			"Invalid":   {"INVALID"},
		} {
			t.Run(name, func(t *testing.T) {
				t.Setenv("ENV", test.envValue)
				_, err := newEnvironment(genv)
				assert.Error(t, err)
			})
		}
	})

	t.Run("Unspecified", func(t *testing.T) {
		env, err := newEnvironment(genv)
		assert.NoError(t, err)
		assert.Equal(t, Dev, env)
	})
}
