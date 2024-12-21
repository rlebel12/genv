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
	)
}

type Settings struct {
	StringVar              string
	IntVar                 int
	BoolVar                bool
	AlwaysDefaultStringVar string
	OptionalFloatVar       float64
	AdvancedURLVar         url.URL
	ManyIntVar             []int
}

func NewSettings() (Settings, error) {
	env := genv.New(
		genv.WithAllowDefault(func(*genv.Genv) (bool, error) {
			return false, nil
		}),
		genv.WithSplitKey(";"),
	)

	var settings Settings

	env.Var("STRING_VAR").StringVar(&settings.StringVar)
	env.Var("INT_VAR").IntVar(&settings.IntVar)
	env.Var("BOOL_VAR").BoolVar(&settings.BoolVar)
	env.Var("ALWAYS_DEFAULT_STRING_VAR").
		Default("default value", env.WithAllowDefaultAlways()).
		StringVar(&settings.AlwaysDefaultStringVar)
	env.Var("OPTIONAL_FLOAT_VAR").Optional().Float64Var(&settings.OptionalFloatVar)
	env.Var("ADVANCED_URL_VAR").
		Default("https://example.com",
			env.WithAllowDefault(func(*genv.Genv) (bool, error) {
				clone := env.Clone()
				allow := clone.Var("ADVANCED_URL_VAR_ALLOW_DEFAULT").
					Default("true", clone.WithAllowDefaultAlways()).
					Bool()
				if err := clone.Parse(); err != nil {
					return false, fmt.Errorf("parse ADVANCED_URL_VAR_ALLOW_DEFAULT: %w", err)
				}
				return *allow, nil
			}),
		).
		Optional().
		URLVar(&settings.AdvancedURLVar)
	env.Var("MANY_INT_VAR").
		Optional().
		Default("123;456;", env.WithAllowDefaultAlways()).
		ManyIntVar(&settings.ManyIntVar)

	if err := env.Parse(); err != nil {
		return Settings{}, fmt.Errorf("parse env: %w", err)
	}
	return settings, nil
}
