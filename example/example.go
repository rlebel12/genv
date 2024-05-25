package main

import (
	"log/slog"
	"math/rand"
	"net/url"

	"github.com/rlebel12/genv"
)

func main() {
	example := NewExample()
	slog.Info("Example",
		"StringVar", example.StringVar,
		"IntVar", example.IntVar,
		"BoolVar", example.BoolVar,
		"AlwaysDefaultStringVar", example.AlwaysDefaultStringVar,
		"OptionalFloatVar", example.OptionalFloatVar,
		"AdvancedURLVar", example.AdvancedURLVar,
		"ManyIntVar", example.ManyIntVar,
	)
}

type Example struct {
	StringVar              string
	IntVar                 int
	BoolVar                bool
	AlwaysDefaultStringVar string
	OptionalFloatVar       float64
	AdvancedURLVar         *url.URL
	ManyIntVar             []int
}

func NewExample() *Example {
	env := genv.New(
		genv.WithAllowDefault(func(*genv.Genv) bool {
			return false
		}),
		genv.WithSplitKey(";"),
	)

	return &Example{
		StringVar: env.Var("STRING_VAR").String(), // Required
		IntVar:    env.Var("INT_VAR").Int(),       // Required
		BoolVar:   env.Var("BOOL_VAR").Bool(),     // Required
		AlwaysDefaultStringVar: env.Var("ALWAYS_DEFAULT_STRING_VAR").
			Default("default value", env.WithAllowDefaultAlways()).
			String(),
		OptionalFloatVar: env.Var("OPTIONAL_FLOAT_VAR").Optional().Float64(),
		AdvancedURLVar: env.Var("ADVANCED_URL_VAR").
			Default(
				"https://example.com",
				env.WithAllowDefault(func(*genv.Genv) bool {
					return rand.Float32() < 0.5 // 50% chance to use the default
				}),
			).
			Optional().
			URL(),
		ManyIntVar: env.
			Var("MANY_INT_VAR").
			Optional().
			Default("123;456;", env.WithAllowDefaultAlways()).
			ManyInt(),
	}
}
