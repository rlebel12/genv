package goenvvars

type environment uint8

const (
	Dev environment = iota
	Prod
	Test
)

func CurrentEnv() environment {
	return currentEnv
}

func IsDev() bool {
	return CurrentEnv() == Dev
}

func IsProd() bool {
	return CurrentEnv() == Prod
}

func IsTest() bool {
	return CurrentEnv() == Test
}

var environments = map[string]environment{
	"DEVELOPMENT": Dev,
	"DEV":         Dev,
	"PRODUCTION":  Prod,
	"PROD":        Prod,
	"TEST":        Test,
}

var currentEnv = Dev

func init() {
	updateCurrentEnv()
}

func updateCurrentEnv() {
	envStr, err := New("ENV").
		Fallback("DEVELOPMENT", OverrideAllow(func() bool { return true })).
		String()
	if err != nil {
		// This should never happen
		panic(err)
	}

	env, ok := environments[envStr]
	if !ok {
		panic("Invalid environment: " + envStr)
	}
	currentEnv = env
}
