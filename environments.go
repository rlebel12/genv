package goevars

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
	var ok bool
	envStr := Optional("ENV", "DEVELOPMENT").String()
	currentEnv, ok = environments[envStr]
	if !ok {
		panic("Invalid environment: " + envStr)
	}
}
