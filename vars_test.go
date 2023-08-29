package goenvvars

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper function to test individual fields, since we cannot compare
// functions for equality.
func expectEnvVarEqual(t *testing.T, expected, actual *EnvVar) {
	assert := assert.New(t)
	assert.Equal(expected.Key, actual.Key)
	assert.Equal(expected.Value, actual.Value)
	assert.Equal(expected.Found, actual.Found)
	assert.Equal(expected.Optional, actual.Optional)
}

func TestNewEnvVar(t *testing.T) {
	t.Run("Defined", func(t *testing.T) {
		t.Run("Present", func(t *testing.T) {
			t.Run("Required", func(t *testing.T) {
				t.Setenv("TEST_VAR", "val")
				actual := New("TEST_VAR")
				expected := &EnvVar{
					Key:      "TEST_VAR",
					Value:    "val",
					Found:    true,
					Optional: false,
				}
				expectEnvVarEqual(t, expected, actual)
			})

			t.Run("Optional", func(t *testing.T) {
				t.Setenv("TEST_VAR", "val")
				actual := New("TEST_VAR", Optional())
				expected := &EnvVar{
					Key:      "TEST_VAR",
					Value:    "val",
					Found:    true,
					Optional: true,
				}
				expectEnvVarEqual(t, expected, actual)
			})
		})

		t.Run("Empty", func(t *testing.T) {
			t.Run("Required", func(t *testing.T) {
				t.Setenv("TEST_VAR", "")
				assert.Panics(t, func() { New("TEST_VAR") })
			})

			t.Run("Optional", func(t *testing.T) {
				t.Setenv("TEST_VAR", "")
				actual := New("TEST_VAR", Optional())
				expected := &EnvVar{
					Key:      "TEST_VAR",
					Value:    "",
					Found:    true,
					Optional: true,
				}
				expectEnvVarEqual(t, expected, actual)
			})
		})
	})

	t.Run("Undefined", func(t *testing.T) {
		t.Run("Required", func(t *testing.T) {
			assert.Panics(t, func() { New("TEST_VAR") })
		})

		t.Run("Optional", func(t *testing.T) {
			actual := New("TEST_VAR", Optional())
			expected := &EnvVar{
				Key:      "TEST_VAR",
				Value:    "",
				Found:    false,
				Optional: true,
			}
			expectEnvVarEqual(t, expected, actual)
		})
	})
}

func TestOptional(t *testing.T) {
	const key = "TEST_VAR"
	fb := Fallback("fallback")
	DefaultAllowFallback = func() bool { return true }
	assert.Equal(t, false, New(key, fb).Optional)
	assert.Equal(t, true, New(key, fb, Optional()).Optional)
}

type MockFallbackOpt struct {
	mock.Mock
}

func (m *MockFallbackOpt) optFunc() {
	_ = m.Called()
	return
}

func fallbackOptForTest(m *MockFallbackOpt) fallbackOpt {
	return func(e *EnvVar) {
		m.optFunc()
	}
}

func TestFallback(t *testing.T) {
	t.Run("Options", func(t *testing.T) {
		opt := new(MockFallbackOpt)
		opt.On("optFunc")
		New("TEST_VAR", Fallback("fallback", fallbackOptForTest(opt)))
		opt.AssertExpectations(t)
	})

	t.Run("Found", func(t *testing.T) {
		t.Setenv("TEST_VAR", "val")
		ev := New("TEST_VAR", Fallback("fallback"))
		assert.Equal(t, "val", ev.Value)
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Run("AllowFallback", func(t *testing.T) {
			ev := New("TEST_VAR", Fallback(
				"fallback",
				OverrideAllowFallback(func() bool { return true }),
			))
			assert.Equal(t, "fallback", ev.Value)
		})

		t.Run("DisallowFallback", func(t *testing.T) {
			assert.Panics(t, func() {
				New("TEST_VAR", Fallback(
					"fallback",
					OverrideAllowFallback(func() bool { return false }),
				))
			})
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
	ev := EnvVar{Key: "TEST_VAR", Value: "val"}
	assert.Equal(t, "val", ev.String())
}

func TestEVarBool(t *testing.T) {
	t.Run(("Valid"), func(t *testing.T) {
		ev := EnvVar{Key: "TEST_VAR", Value: "true"}
		assert.True(t, ev.Bool())
		ev.Value = "false"
		assert.False(t, ev.Bool())
	})

	t.Run(("Invalid"), func(t *testing.T) {
		ev := EnvVar{Key: "TEST_VAR", Value: "invalid"}
		assert.Panics(t, func() { ev.Bool() })
	})
}

func TestEvarInt(t *testing.T) {
	t.Run(("Valid"), func(t *testing.T) {
		ev := EnvVar{Key: "TEST_VAR", Value: "123"}
		assert.Equal(t, 123, ev.Int())
	})

	t.Run(("Invalid"), func(t *testing.T) {
		ev := EnvVar{Key: "TEST_VAR", Value: "invalid"}
		assert.Panics(t, func() { ev.Int() })
	})
}

func TestEvarFloat64(t *testing.T) {
	t.Run(("Valid"), func(t *testing.T) {
		ev := EnvVar{Key: "TEST_VAR", Value: "123.456"}
		assert.Equal(t, 123.456, ev.Float64())
	})

	t.Run(("Invalid"), func(t *testing.T) {
		ev := EnvVar{Key: "TEST_VAR", Value: "invalid"}
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
