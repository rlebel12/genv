package main

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"github.com/rlebel12/genv"
)

func main() {
	settings, err := NewSettings()
	if err != nil {
		slog.Error("new settings", "error", err.Error())
		os.Exit(1)
	}
	slog.Info("Example",
		"StringVar", settings.StringVar,
		"IntVar", settings.IntVar,
		"BoolVar", settings.BoolVar,
		"AlwaysDefaultStringVar", settings.AlwaysDefaultStringVar,
		"OptionalFloatVar", settings.OptionalFloatVar,
		"AdvancedURLVar", settings.AdvancedURLVar,
		"ManyIntVar", settings.ManyIntVar,
		"DatabaseName", settings.DatabaseName,
		"Port", settings.Port,
		"IsProduction", settings.IsProduction,
		"TimeoutMs", settings.TimeoutMs,
	)
}

type (
	Settings struct {
		StringVar              string
		IntVar                 int
		BoolVar                bool
		AlwaysDefaultStringVar string
		OptionalFloatVar       float64
		AdvancedURLVar         url.URL
		ManyIntVar             []int
		DatabaseName           CustomString
		ManyCustomStrings      []CustomString
		Port                   Port
		IsProduction           IsProduction
		TimeoutMs              TimeoutMs
	}

	CustomString string
	Port         int
	IsProduction bool
	TimeoutMs    float64
)

func NewSettings() (Settings, error) {
	env := genv.New(
		genv.WithAllowDefault(func(*genv.Genv) (bool, error) {
			return false, nil
		}),
		genv.WithSplitKey(";"),
	)

	var s Settings

	genv.String(&s.StringVar, env.Var("STRING_VAR"))
	env.Var("INT_VAR").Int(&s.IntVar)
	env.Var("BOOL_VAR").Bool(&s.BoolVar)
	genv.String(&s.AlwaysDefaultStringVar, env.Var("ALWAYS_DEFAULT_STRING_VAR").
		Default("default value", env.WithAllowDefaultAlways()))
	env.Var("OPTIONAL_FLOAT_VAR").Float64(&s.OptionalFloatVar).Optional()
	env.Var("ADVANCED_URL_VAR").
		Default("https://example.com",
			env.WithAllowDefault(func(*genv.Genv) (bool, error) {
				clone := env.Clone()
				allow := clone.Var("ADVANCED_URL_VAR_ALLOW_DEFAULT").
					Default("true", clone.WithAllowDefaultAlways()).
					NewBool()
				if err := clone.Parse(); err != nil {
					return false, fmt.Errorf("parse ADVANCED_URL_VAR_ALLOW_DEFAULT: %w", err)
				}
				return *allow, nil
			}),
		).
		Optional().
		URL(&s.AdvancedURLVar)
	env.Var("MANY_INT_VAR").
		Optional().
		Default("123;456;", env.WithAllowDefaultAlways()).
		Ints(&s.ManyIntVar)

	genv.String(&s.DatabaseName, env.Var("DATABASE_NAME").Default("myapp_db", env.WithAllowDefaultAlways()))
	genv.Strings(&s.ManyCustomStrings, env.Var("MANY_CUSTOM_STRINGS").Default("a;b;c", env.WithAllowDefaultAlways()))
	genv.Int(env.Var("PORT").Default("8080", env.WithAllowDefaultAlways()), &s.Port)
	genv.Bool(env.Var("IS_PRODUCTION").Default("false", env.WithAllowDefaultAlways()), &s.IsProduction)
	genv.Float64(env.Var("TIMEOUT_MS").Default("5000.0", env.WithAllowDefaultAlways()), &s.TimeoutMs)

	if err := env.Parse(); err != nil {
		return Settings{}, fmt.Errorf("parse env: %w", err)
	}
	return s, nil
}
