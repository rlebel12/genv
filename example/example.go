package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/rlebel12/genv"
)

func main() {
	// Example 1: Basic usage (backward compatible)
	slog.Info("=== Basic Example ===")
	settings, err := NewSettings()
	if err != nil {
		slog.Error("new settings", "error", err.Error())
		os.Exit(1)
	}
	slog.Info("Basic Example Results",
		"StringVar", settings.StringVar,
		"IntVar", settings.IntVar,
		"BoolVar", settings.BoolVar,
		"AlwaysDefaultStringVar", settings.AlwaysDefaultStringVar,
		"OptionalFloatVar", settings.OptionalFloatVar,
		"AdvancedURLVar", settings.AdvancedURLVar,
		"ManyIntVar", settings.ManyIntVar,
	)

	// Example 2: Custom registry with custom types
	slog.Info("=== Custom Registry Example ===")
	customSettings, err := NewCustomRegistrySettings()
	if err != nil {
		slog.Error("custom registry settings", "error", err.Error())
		os.Exit(1)
	}
	slog.Info("Custom Registry Results",
		"UserID", customSettings.UserID,
		"Department", customSettings.Department,
		"ValidatedEmail", customSettings.ValidatedEmail,
	)

	// Example 3: Registry isolation
	slog.Info("=== Registry Isolation Example ===")
	DemonstrateRegistryIsolation()

	// Example 4: Advanced custom type features
	slog.Info("=== Advanced Custom Type Features ===")
	advancedSettings, err := NewAdvancedCustomTypeSettings()
	if err != nil {
		slog.Error("advanced custom type settings", "error", err.Error())
		os.Exit(1)
	}
	slog.Info("Advanced Features Results",
		"Priorities", fmt.Sprintf("%v", advancedSettings.Priorities),
		"OptionalServiceName", advancedSettings.OptionalServiceName,
		"LogLevel", advancedSettings.LogLevel,
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

	var s Settings

	env.Var("STRING_VAR").String(&s.StringVar)
	env.Var("INT_VAR").Int(&s.IntVar)
	env.Var("BOOL_VAR").Bool(&s.BoolVar)
	env.Var("ALWAYS_DEFAULT_STRING_VAR").
		String(&s.AlwaysDefaultStringVar).
		Default("default value", env.WithAllowDefaultAlways())
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

	if err := env.Parse(); err != nil {
		return Settings{}, fmt.Errorf("parse env: %w", err)
	}
	return s, nil
}

// Custom types for demonstration
type UserID string
type Department string
type ValidatedEmail string
type Priority int
type ServiceName string
type ExampleLogLevel string

// Custom settings using custom types
type CustomSettings struct {
	UserID         UserID
	Department     Department
	ValidatedEmail ValidatedEmail
}

// Advanced settings demonstrating slices, optional, and default behavior
type AdvancedCustomSettings struct {
	Priorities          []Priority
	OptionalServiceName ServiceName
	LogLevel            ExampleLogLevel
}

// NewCustomRegistrySettings demonstrates custom type parsing with registries
func NewCustomRegistrySettings() (CustomSettings, error) {
	// Create a custom registry with built-in parsers
	registry := genv.NewDefaultRegistry()

	// Register custom UserID parser
	genv.RegisterTypedParserOn(registry, func(s string) (UserID, error) {
		if s == "" {
			return "", errors.New("UserID cannot be empty")
		}
		if !strings.HasPrefix(s, "user_") {
			return UserID("user_" + s), nil
		}
		return UserID(s), nil
	})

	// Register custom Department parser with validation
	genv.RegisterTypedParserOn(registry, func(s string) (Department, error) {
		validDepts := map[string]bool{
			"engineering": true,
			"marketing":   true,
			"sales":       true,
			"hr":          true,
		}

		dept := strings.ToLower(strings.TrimSpace(s))
		if !validDepts[dept] {
			return "", fmt.Errorf("invalid department: %s (valid: engineering, marketing, sales, hr)", s)
		}
		return Department(dept), nil
	})

	// Register validated email parser
	genv.RegisterTypedParserOn(registry, func(s string) (ValidatedEmail, error) {
		if s == "" {
			return "", errors.New("email cannot be empty")
		}
		if !strings.Contains(s, "@") || !strings.Contains(s, ".") {
			return "", fmt.Errorf("invalid email format: %s", s)
		}
		return ValidatedEmail(strings.ToLower(s)), nil
	})

	// Create Genv with custom registry
	env := genv.New(
		genv.WithRegistry(registry),
		genv.WithAllowDefault(func(*genv.Genv) (bool, error) { return true, nil }),
	)

	var s CustomSettings

	genv.Type(env.Var("CUSTOM_USER_ID").Default("demo123"), &s.UserID)
	genv.Type(env.Var("CUSTOM_DEPARTMENT").Default("engineering"), &s.Department)
	genv.Type(env.Var("CUSTOM_EMAIL").Default("demo@example.com"), &s.ValidatedEmail)

	if err := env.Parse(); err != nil {
		return CustomSettings{}, fmt.Errorf("parse custom env: %w", err)
	}

	return s, nil
}

// DemonstrateRegistryIsolation shows how different registries can have different behavior
func DemonstrateRegistryIsolation() {
	// Production registry with strict validation for a custom type
	prodRegistry := genv.NewRegistry() // Start with empty registry
	genv.RegisterTypedParserOn(prodRegistry, func(s string) (string, error) {
		if len(s) < 3 {
			return "", errors.New("production strings must be at least 3 characters")
		}
		return s, nil
	})
	genv.RegisterTypedParserOn(prodRegistry, func(s string) (int, error) {
		// Production int parsing with range validation
		val, err := strconv.Atoi(s)
		if err != nil {
			return 0, err
		}
		if val < 0 || val > 1000 {
			return 0, errors.New("production integers must be between 0-1000")
		}
		return val, nil
	})

	// Development registry with lenient validation
	devRegistry := genv.NewRegistry() // Start with empty registry
	genv.RegisterTypedParserOn(devRegistry, func(s string) (string, error) {
		if s == "" {
			return "dev-default", nil // Provide default in dev
		}
		return s, nil
	})
	genv.RegisterTypedParserOn(devRegistry, func(s string) (int, error) {
		// Development int parsing that's more forgiving
		if s == "" {
			return 42, nil // Default value in dev
		}
		return strconv.Atoi(s)
	})

	_ = genv.New(genv.WithRegistry(prodRegistry))
	_ = genv.New(genv.WithRegistry(devRegistry))

	slog.Info("Registry isolation demonstrated",
		"prodRegistry", "strict validation (strings ≥3 chars, ints 0-1000)",
		"devRegistry", "lenient validation (provides dev defaults)",
		"isolation", "each Genv instance uses completely different parsers")
}

// NewAdvancedCustomTypeSettings demonstrates advanced custom type features
func NewAdvancedCustomTypeSettings() (AdvancedCustomSettings, error) {
	// Create registry with advanced custom types
	registry := genv.NewDefaultRegistry()

	// Register Priority parser (enum-like behavior)
	genv.RegisterTypedParserOn(registry, func(s string) (Priority, error) {
		switch strings.ToLower(strings.TrimSpace(s)) {
		case "low":
			return Priority(1), nil
		case "medium":
			return Priority(2), nil
		case "high":
			return Priority(3), nil
		case "critical":
			return Priority(4), nil
		default:
			return Priority(0), fmt.Errorf("invalid priority: %s (valid: low, medium, high, critical)", s)
		}
	})

	// Register ServiceName parser with prefix handling
	genv.RegisterTypedParserOn(registry, func(s string) (ServiceName, error) {
		if s == "" {
			return ServiceName(""), nil // Allow empty for optional
		}
		if len(s) < 3 {
			return "", errors.New("service name must be at least 3 characters")
		}
		if !strings.HasPrefix(s, "svc-") {
			return ServiceName("svc-" + s), nil
		}
		return ServiceName(s), nil
	})

	// Register LogLevel parser with strict validation
	genv.RegisterTypedParserOn(registry, func(s string) (ExampleLogLevel, error) {
		level := strings.ToUpper(strings.TrimSpace(s))
		switch level {
		case "DEBUG", "INFO", "WARN", "ERROR":
			return ExampleLogLevel(level), nil
		default:
			return "", fmt.Errorf("invalid log level: %s (valid: DEBUG, INFO, WARN, ERROR)", s)
		}
	})

	// Create Genv with custom registry and allow defaults
	env := genv.New(
		genv.WithRegistry(registry),
		genv.WithAllowDefault(func(*genv.Genv) (bool, error) { return true, nil }),
		genv.WithSplitKey("|"), // Use different delimiter for this example
	)

	var s AdvancedCustomSettings

	// Demonstrate slice parsing with custom types
	genv.Types(env.Var("TASK_PRIORITIES").Default("medium|high|low"), &s.Priorities)

	// Demonstrate optional custom type (won't error if missing/empty)
	genv.Type(env.Var("SERVICE_NAME").Optional(), &s.OptionalServiceName)

	// Demonstrate custom type with default value
	genv.Type(env.Var("LOG_LEVEL").Default("INFO"), &s.LogLevel)

	if err := env.Parse(); err != nil {
		return AdvancedCustomSettings{}, fmt.Errorf("parse advanced env: %w", err)
	}

	return s, nil
}
