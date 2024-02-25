package goenvvars

import "fmt"

type environment uint8

const (
	Dev environment = iota
	Prod
	Test
)

func newEnvironment(genv *Genv) (environment, error) {
	envStr, err := genv.New("ENV").
		Fallback("DEVELOPMENT", OverrideAllow(func() bool { return true })).
		TryString()
	if err != nil {
		return 0, fmt.Errorf("invalid environment - must be a string: %w", err)
	}
	env, ok := environments()[envStr]
	if !ok {
		return 0, fmt.Errorf("invalid environment value: %w", err)
	}
	return env, nil
}

func environments() map[string]environment {
	return map[string]environment{
		"DEVELOPMENT": Dev,
		"DEV":         Dev,
		"PRODUCTION":  Prod,
		"PROD":        Prod,
		"TEST":        Test,
	}
}
