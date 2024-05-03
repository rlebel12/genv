package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net/url"
	"os"

	"github.com/rlebel12/goenvvars/v3"
)

func main() {
	_, err := NewExample()
	if err != nil {
		slog.Error("failed to create example", "error", err)
		os.Exit(1)
	}
}

type Example struct {
	StringVar               string
	IntVar                  int
	BoolVar                 bool
	AlwaysFallbackStringVar string
	OptionalFloatVar        float64
	AdvancedURLVar          *url.URL
	ManyIntVar              []int
}

func NewExample() (example *Example, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to create genv: %w", r.(error))
		}
	}()

	genv, err := goenvvars.New(goenvvars.DefaultAllowFallback(func() bool {
		return false
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to create genv: %w", err)
	}

	example = &Example{
		StringVar: genv.New("STRING_VAR").String(), // Required
		IntVar:    genv.New("INT_VAR").Int(),       // Required
		BoolVar:   genv.New("BOOL_VAR").Bool(),     // Required
		AlwaysFallbackStringVar: genv.New("ALWAYS_FALLBACK_STRING_VAR").
			Fallback("fallback value", goenvvars.AllowAlways()).
			String(),
		OptionalFloatVar: genv.New("OPTIONAL_FLOAT_VAR").Optional().Float64(),
		AdvancedURLVar: genv.New("ADVANCED_URL_VAR").
			Fallback(
				"https://example.com",
				goenvvars.OverrideAllow(func() bool {
					return rand.Float32() < 0.5 // 50% chance to use the fallback
				}),
			).
			Optional().
			URL(),
		ManyIntVar: genv.
			New("MANY_INT_VAR").
			Optional().
			Fallback("123,456,", goenvvars.AllowAlways()).
			ManyInt(),
	}
	return
}
