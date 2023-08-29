package goenvvars

import (
	"errors"
	"strconv"
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
			assert.Nil(t, ev.validate())
		})

		t.Run("Absent", func(t *testing.T) {
			t.Setenv("TEST_VAR", "")
			ev := New("TEST_VAR")
			err := ev.validate()
			if assert.Error(t, err) {
				assert.Equal(t, errors.New("Missing required environment variable: "+ev.key), err)
			}
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
	actual, _ := ev.String()
	assert.Equal(t, "val", actual)
}

func TestEVarBool(t *testing.T) {
	t.Run(("Valid"), func(t *testing.T) {
		ev := envVar{key: "TEST_VAR", value: "true"}
		actual, _ := ev.Bool()
		assert.True(t, actual)

		ev.value = "false"
		actual, _ = ev.Bool()
		assert.False(t, actual)
	})

	t.Run(("Invalid"), func(t *testing.T) {
		ev := envVar{key: "TEST_VAR", value: "invalid"}
		actual, err := ev.Bool()
		if assert.Error(t, err) {
			_, ok := err.(*strconv.NumError)
			assert.True(t, ok)
		}
		assert.Equal(t, false, actual)
	})
}

func TestEvarInt(t *testing.T) {
	t.Run(("Valid"), func(t *testing.T) {
		ev := envVar{key: "TEST_VAR", value: "123"}
		actual, _ := ev.Int()
		assert.Equal(t, 123, actual)
	})

	t.Run(("Invalid"), func(t *testing.T) {
		ev := envVar{key: "TEST_VAR", value: "invalid"}
		actual, err := ev.Int()
		if assert.Error(t, err) {
			_, ok := err.(*strconv.NumError)
			assert.True(t, ok)
		}
		assert.Equal(t, 0, actual)
	})
}

func TestEvarFloat64(t *testing.T) {
	t.Run(("Valid"), func(t *testing.T) {
		ev := envVar{key: "TEST_VAR", value: "123.456"}
		actual, _ := ev.Float64()
		assert.Equal(t, 123.456, actual)
	})

	t.Run(("Invalid"), func(t *testing.T) {
		ev := envVar{key: "TEST_VAR", value: "invalid"}
		actual, err := ev.Float64()
		if assert.Error(t, err) {
			_, ok := err.(*strconv.NumError)
			assert.True(t, ok)
		}
		assert.Equal(t, 0.0, actual)
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
