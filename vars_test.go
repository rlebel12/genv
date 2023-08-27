package goenvvars

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPresencePresent(t *testing.T) {
	t.Setenv("TEST_VAR", "val")
	assert.True(t, Presence("TEST_VAR"))
}

func TestPresenceAbsent(t *testing.T) {
	assert.False(t, Presence("TEST_VAR"))
}

func TestPresenceEmpty(t *testing.T) {
	t.Setenv("TEST_VAR", "")
	assert.False(t, Presence("TEST_VAR"))
}

func TestEVarString(t *testing.T) {
	ev := envVarData{key: "TEST_VAR", value: "val"}
	assert.Equal(t, "val", ev.String())
}

func TestEvarBoolValid(t *testing.T) {
	ev := envVarData{key: "TEST_VAR", value: "true"}
	assert.True(t, ev.Bool())
	ev.value = "false"
	assert.False(t, ev.Bool())
}

func TestEvarBoolInvalid(t *testing.T) {
	ev := envVarData{key: "TEST_VAR", value: "invalid"}
	assert.Panics(t, func() { ev.Bool() })
}

func TestEvarIntValid(t *testing.T) {
	ev := envVarData{key: "TEST_VAR", value: "123"}
	assert.Equal(t, 123, ev.Int())
}

func TestEvarIntInvalid(t *testing.T) {
	ev := envVarData{key: "TEST_VAR", value: "invalid"}
	assert.Panics(t, func() { ev.Int() })
}

func TestEvarFloat64Valid(t *testing.T) {
	ev := envVarData{key: "TEST_VAR", value: "123.456"}
	assert.Equal(t, 123.456, ev.Float64())
}

func TestEvarFloat64Invalid(t *testing.T) {
	ev := envVarData{key: "TEST_VAR", value: "invalid"}
	assert.Panics(t, func() { ev.Float64() })
}

func TestAllowFallbacksDev(t *testing.T) {
	t.Setenv("ENV", "DEVELOPMENT")
	updateCurrentEnv()
	assert.True(t, allowFallbacks())
}

func TestAllowFallbacksTest(t *testing.T) {
	t.Setenv("ENV", "TEST")
	updateCurrentEnv()
	assert.True(t, allowFallbacks())
}

func TestAllowFallbacksProd(t *testing.T) {
	t.Setenv("ENV", "PRODUCTION")
	updateCurrentEnv()
	assert.False(t, allowFallbacks())
}
