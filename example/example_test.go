package main

import (
	"net/url"
	"reflect"
	"testing"
)

func TestNewSettings(t *testing.T) {
	tests := map[string]struct {
		giveEnv map[string]string
		want    Settings
		wantErr bool
	}{
		"success": {
			giveEnv: map[string]string{
				"STRING_VAR":          "string",
				"INT_VAR":             "42",
				"BOOL_VAR":            "true",
				"DATABASE_NAME":       "db",
				"MANY_CUSTOM_STRINGS": "a;b;c",
				"PORT":                "2233",
				"IS_PRODUCTION":       "true",
				"TIMEOUT_MS":          "123.45",
			},
			want: Settings{
				StringVar:              "string",
				IntVar:                 42,
				BoolVar:                true,
				AlwaysDefaultStringVar: "default value",
				AdvancedURLVar:         url.URL{Scheme: "https", Host: "example.com"},
				ManyIntVar:             []int{123, 456},
				DatabaseName:           "db",
				ManyCustomStrings:      []CustomString{"a", "b", "c"},
				Port:                   2233,
				IsProduction:           true,
				TimeoutMs:              123.45,
			},
			wantErr: false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			for k, v := range test.giveEnv {
				t.Setenv(k, v)
			}

			got, err := NewSettings()
			if (err != nil) != test.wantErr {
				t.Errorf("NewSettings() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("NewSettings() = %v, want %v", got, test.want)
			}
		})
	}
}
