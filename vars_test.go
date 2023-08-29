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
}

func TestValidate(t *testing.T) {
	t.Run("Required", func(t *testing.T) {
		t.Run("Present", func(t *testing.T) {
			t.Setenv("TEST_VAR", "val")
			ev := New("TEST_VAR")
			ev.validate()
		})

		t.Run("Absent", func(t *testing.T) {
			t.Setenv("TEST_VAR", "")
			ev := New("TEST_VAR")
			assert.Panics(t, func() { ev.validate() })
		})
	})

	t.Run("Optional", func(t *testing.T) {
		t.Run("Present", func(t *testing.T) {
			t.Setenv("TEST_VAR", "val")
			ev := New("TEST_VAR").Optional()
			ev.validate()
		})

		t.Run("Absent", func(t *testing.T) {
			t.Setenv("TEST_VAR", "")
			ev := New("TEST_VAR").Optional()
			ev.validate()
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

func fallbackOptForTest(m *MockFallbackOpt) envVarOpt {
	return func(e *envVar) {
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
	ev := envVar{key: "TEST_VAR", value: "val"}
	assert.Equal(t, "val", ev.String())
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
