package goevars

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequiredProvided(t *testing.T) {
	t.Setenv("TEST_VAR", "val")
	assert.Equal(t, "val", string(Required("TEST_VAR", "")))
}

func TestRequiredAllowFallback(t *testing.T) {
	assert.Equal(t, "fallback", string(Required("TEST_VAR", "fallback")))
}

func TestRequiredDisallowFallback(t *testing.T) {
	t.Setenv("ENV", "PRODUCTION")
	updateCurrentEnv()
	assert.Panics(t, func() { Required("TEST_VAR", "") })
}

func TestOptionalProvided(t *testing.T) {
	t.Setenv("TEST_VAR", "val")
	assert.Equal(t, "val", string(Optional("TEST_VAR", "fallback")))
}

func TestOptionalAllowFallback(t *testing.T) {
	assert.Equal(t, "fallback", string(Optional("TEST_VAR", "fallback")))
}

func TestEVarString(t *testing.T) {
	assert.Equal(t, "val", eVar("val").String())
}

func TestEvarBoolValid(t *testing.T) {
	assert.True(t, eVar("true").Bool())
	assert.False(t, eVar("false").Bool())
}

func TestEvarBoolInvalid(t *testing.T) {
	assert.Panics(t, func() { eVar("invalid").Bool() })
}

func TestEvarIntValid(t *testing.T) {
	assert.Equal(t, 123, eVar("123").Int())
}

func TestEvarIntInvalid(t *testing.T) {
	assert.Panics(t, func() { eVar("invalid").Int() })
}

func TestEvarFloat64Valid(t *testing.T) {
	assert.Equal(t, 123.456, eVar("123.456").Float64())
}

func TestEvarFloat64Invalid(t *testing.T) {
	assert.Panics(t, func() { eVar("invalid").Float64() })
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
