package main

import (
	"log/slog"
	"math/rand"
	"net/url"

	"github.com/rlebel12/goenvvars/v3"
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
	genv := goenvvars.New(
		goenvvars.WithAllowDefault(func(*goenvvars.Genv) bool {
			return false
		}),
		goenvvars.WithSplitKey(";"),
	)

	return &Example{
		StringVar: genv.Var("STRING_VAR").String(), // Required
		IntVar:    genv.Var("INT_VAR").Int(),       // Required
		BoolVar:   genv.Var("BOOL_VAR").Bool(),     // Required
		AlwaysDefaultStringVar: genv.Var("ALWAYS_DEFAULT_STRING_VAR").
			Default("default value", genv.WithAllowDefaultAlways()).
			String(),
		OptionalFloatVar: genv.Var("OPTIONAL_FLOAT_VAR").Optional().Float64(),
		AdvancedURLVar: genv.Var("ADVANCED_URL_VAR").
			Default(
				"https://example.com",
				genv.WithAllowDefault(func(*goenvvars.Genv) bool {
					return rand.Float32() < 0.5 // 50% chance to use the default
				}),
			).
			Optional().
			URL(),
		ManyIntVar: genv.
			Var("MANY_INT_VAR").
			Optional().
			Default("123;456;", genv.WithAllowDefaultAlways()).
			ManyInt(),
	}
}
