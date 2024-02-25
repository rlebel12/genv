package goenvvars

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper function to test individual fields, since we cannot compare
// functions for equality.
func expectEnvVarEqual(t *testing.T, expected, actual *envVar) {
	assert := assert.New(t)
	assert.Equal(expected.key, actual.key)
	assert.Equal(expected.value, actual.value)
	assert.Equal(expected.found, actual.found)
	assert.Equal(expected.optional, actual.optional)
}

func TestNew(t *testing.T) {
	t.Run("Defined", func(t *testing.T) {
		t.Setenv("TEST_VAR", "val")
		actual := New("TEST_VAR")
		expected := &envVar{
			key:      "TEST_VAR",
			value:    "val",
			found:    true,
			optional: false,
		}
		expectEnvVarEqual(t, expected, actual)
	})

	t.Run("Undefined", func(t *testing.T) {
		actual := New("TEST_VAR")
		expected := &envVar{
			key:      "TEST_VAR",
			value:    "",
			found:    false,
			optional: false,
		}
		expectEnvVarEqual(t, expected, actual)
	})

	t.Run("Options", func(t *testing.T) {
		value_opt := func(e *envVar) {
			e.value = "opts"
		}
		found_opt := func(e *envVar) {
			e.found = true
		}
		actual := New("TEST_VAR", value_opt, found_opt)
		expected := &envVar{
			key:      "TEST_VAR",
			value:    "opts",
			found:    true,
			optional: false,
		}
		expectEnvVarEqual(t, expected, actual)
	})
}

func TestValidate(t *testing.T) {
	t.Run("Required", func(t *testing.T) {
		t.Run("Present", func(t *testing.T) {
			t.Setenv("TEST_VAR", "val")
			ev := New("TEST_VAR")
			assert.Nil(t, ev.validate())
		})

		t.Run("Absent", func(t *testing.T) {
			t.Setenv("TEST_VAR", "")
			ev := New("TEST_VAR")
			assert.Error(t, ev.validate())
		})
	})

	t.Run("Optional", func(t *testing.T) {
		t.Run("Present", func(t *testing.T) {
			t.Setenv("TEST_VAR", "val")
			ev := New("TEST_VAR").Optional()
			assert.Nil(t, ev.validate())
		})

		t.Run("Absent", func(t *testing.T) {
			t.Setenv("TEST_VAR", "")
			ev := New("TEST_VAR").Optional()
			assert.Nil(t, ev.validate())
		})
	})
}

func TestOptional(t *testing.T) {
	t.Run("Required", func(t *testing.T) {
		ev := New("TEST_VAR")
		assert.Equal(t, false, ev.optional)
	})

	t.Run("Optional", func(t *testing.T) {
		ev := New("TEST_VAR").Optional()
		assert.Equal(t, true, ev.optional)
	})
}

type MockFallbackOpt struct {
	mock.Mock
}

func (m *MockFallbackOpt) optFunc() {
	_ = m.Called()
	return
}

func fallbackOptForTest(m *MockFallbackOpt) fallbackOpt {
	return func(fb *fallback) {
		m.optFunc()
	}
}

func TestFallback(t *testing.T) {
	t.Run("Options", func(t *testing.T) {
		opt := new(MockFallbackOpt)
		opt.On("optFunc")
		New("TEST_VAR").Fallback("fallback", fallbackOptForTest(opt))
		opt.AssertExpectations(t)
	})

	t.Run("Found", func(t *testing.T) {
		t.Setenv("TEST_VAR", "val")
		ev := New("TEST_VAR").Fallback("fallback")
		assert.Equal(t, "val", ev.value)
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Run("AllowFallback", func(t *testing.T) {
			ev := New("TEST_VAR").
				Fallback("fallback", OverrideAllow(func() bool { return true }))
			assert.Equal(t, "fallback", ev.value)
		})

		t.Run("DisallowFallback", func(t *testing.T) {
			ev := New("TEST_VAR").
				Fallback("fallback", OverrideAllow(func() bool { return false }))
			assert.Equal(t, "", ev.value)
		})
	})
}

func TestAllowAlways(t *testing.T) {
	DefaultAllowFallback = func() bool { return false }
	actual := New("TEST_VAR").Fallback("fallback", AllowAlways()).String()
	assert.Equal(t, "fallback", actual)
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

func TestDefaultAllowFallback(t *testing.T) {
	t.Run("Dev", func(t *testing.T) {
		t.Setenv("ENV", "DEVELOPMENT")
		updateCurrentEnv()
		assert.True(t, defaultAllowFallback())
	})

	t.Run("Test", func(t *testing.T) {
		t.Setenv("ENV", "TEST")
		updateCurrentEnv()
		assert.True(t, defaultAllowFallback())
	})

	t.Run("Prod", func(t *testing.T) {
		t.Setenv("ENV", "PRODUCTION")
		updateCurrentEnv()
		assert.False(t, defaultAllowFallback())
	})
}
