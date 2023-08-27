package goenvvars

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCurrentEnvDefault(t *testing.T) {
	assert.Equal(t, Dev, CurrentEnv())
}

func TestCurrentEnvSpecified(t *testing.T) {
	for eVarName, eVarValue := range environments {
		t.Run(eVarName, func(t *testing.T) {
			t.Setenv("ENV", eVarName)
			updateCurrentEnv()
			assert.Equal(t, eVarValue, CurrentEnv())
		})
	}
}

func TestInvalidEnv(t *testing.T) {
	t.Setenv("ENV", "INVALID")
	assert.Panics(t, func() { updateCurrentEnv() })
}

func TestIsDevTrue(t *testing.T) {
	t.Setenv("ENV", "DEVELOPMENT")
	updateCurrentEnv()
	assert.True(t, IsDev())
}

func TestIsDevFalse(t *testing.T) {
	t.Setenv("ENV", "PRODUCTION")
	updateCurrentEnv()
	assert.False(t, IsDev())
}

func TestIsProdTrue(t *testing.T) {
	t.Setenv("ENV", "PRODUCTION")
	updateCurrentEnv()
	assert.True(t, IsProd())
}

func TestIsProdFalse(t *testing.T) {
	t.Setenv("ENV", "DEVELOPMENT")
	updateCurrentEnv()
	assert.False(t, IsProd())
}

func TestIsTestTrue(t *testing.T) {
	t.Setenv("ENV", "TEST")
	updateCurrentEnv()
	assert.True(t, IsTest())
}

func TestIsTestFalse(t *testing.T) {
	t.Setenv("ENV", "DEVELOPMENT")
	updateCurrentEnv()
	assert.False(t, IsTest())
}
