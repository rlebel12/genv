package goenvvars

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCurrentEnv(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, Dev, CurrentEnv())
	})

	t.Run("Specified", func(t *testing.T) {
		for eVarName, eVarValue := range environments {
			t.Run(eVarName, func(t *testing.T) {
				t.Setenv("ENV", eVarName)
				updateCurrentEnv()
				assert.Equal(t, eVarValue, CurrentEnv())
			})
		}
		t.Run("Invalid", func(t *testing.T) {
			t.Setenv("ENV", "INVALID")
			assert.Panics(t, func() { updateCurrentEnv() })
		})
	})
}

func TestIsDev(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		t.Setenv("ENV", "DEVELOPMENT")
		updateCurrentEnv()
		assert.True(t, IsDev())
	})

	t.Run("False", func(t *testing.T) {
		t.Setenv("ENV", "PRODUCTION")
		updateCurrentEnv()
		assert.False(t, IsDev())
	})
}

func TestIsProd(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		t.Setenv("ENV", "PRODUCTION")
		updateCurrentEnv()
		assert.True(t, IsProd())
	})

	t.Run("False", func(t *testing.T) {
		t.Setenv("ENV", "DEVELOPMENT")
		updateCurrentEnv()
		assert.False(t, IsProd())
	})
}

func TestIsTest(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		t.Setenv("ENV", "TEST")
		updateCurrentEnv()
		assert.True(t, IsTest())
	})

	t.Run("False", func(t *testing.T) {
		t.Setenv("ENV", "DEVELOPMENT")
		updateCurrentEnv()
		assert.False(t, IsTest())
	})
}
