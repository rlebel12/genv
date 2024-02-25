package goenvvars

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
