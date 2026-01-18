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
	// Example 0: Simplified API with Bind/Parse (NEW!)
	slog.Info("=== Simplified Bind/Parse API Example ===")
	simplifiedSettings, err := NewSimplifiedSettings()
	if err != nil {
		slog.Error("simplified settings", "error", err.Error())
		os.Exit(1)
	}
	slog.Info("Simplified API Results",
		"AppName", simplifiedSettings.AppName,
		"Port", simplifiedSettings.Port,
		"Debug", simplifiedSettings.Debug,
		"Timeout", simplifiedSettings.Timeout,
		"DatabaseURL", simplifiedSettings.DatabaseURL,
		"Tags", simplifiedSettings.Tags,
	)

	// Example 0b: Using NewType[T]() for creating variables (NEW!)
	slog.Info("=== NewType[T]() Generic Function Example ===")
	DemonstrateNewMethod()

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

// SimplifiedSettings demonstrates the new simplified API
type SimplifiedSettings struct {
	AppName     string
	Port        int
	Debug       bool
	Timeout     float64
	DatabaseURL url.URL
	Tags        []string
}

// NewSimplifiedSettings shows the new simplified API using Bind() and Parse()
// where the type is automatically detected. No need for .String(), .Int(), etc.
func NewSimplifiedSettings() (SimplifiedSettings, error) {
	env := genv.New(
		genv.WithAllowDefault(func(*genv.Genv) (bool, error) { return true, nil }),
	)

	var s SimplifiedSettings

	// Simplified API: use Bind()/BindMany() with Parse() and type is inferred via generics!
	// All variables are registered and parsed in one call. Zero reflection overhead.
	err := genv.Parse(env,
		genv.Bind("APP_NAME", &s.AppName).Default("MyApp"),
		genv.Bind("PORT", &s.Port).Default("8080"),
		genv.Bind("DEBUG", &s.Debug).Default("false"),
		genv.Bind("TIMEOUT", &s.Timeout).Default("30.5"),
		genv.Bind("DATABASE_URL", &s.DatabaseURL).Default("https://db.example.com"),
		genv.BindMany("TAGS", &s.Tags).Default("api,web,production"),
	)
	if err != nil {
		return SimplifiedSettings{}, fmt.Errorf("parse env: %w", err)
	}
	return s, nil
}

// DemonstrateNewMethod shows using NewType[T]() for creating variables
func DemonstrateNewMethod() {
	env := genv.New(
		genv.WithAllowDefault(func(*genv.Genv) (bool, error) { return true, nil }),
	)

	// Use NewType[T]() for creating new variables - unified API for all types!
	// Works with built-in types AND custom types!
	apiKey := genv.NewType[string](env.Var("API_KEY").Default("demo-key-12345"))
	maxRetries := genv.NewType[int](env.Var("MAX_RETRIES").Default("3"))
	enableCache := genv.NewType[bool](env.Var("ENABLE_CACHE").Default("true"))
	apiTimeout := genv.NewType[float64](env.Var("API_TIMEOUT").Default("60.0"))

	if err := env.Parse(); err != nil {
		slog.Error("parse error", "error", err.Error())
		return
	}

	slog.Info("NewType[T]() Generic Function Results",
		"ApiKey", *apiKey,
		"MaxRetries", *maxRetries,
		"EnableCache", *enableCache,
		"ApiTimeout", *apiTimeout,
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
	// Create a custom registry with built-in parsers and custom parsers
	registry := genv.NewDefaultRegistry(
		// Register custom UserID parser
		genv.WithParser(func(s string) (UserID, error) {
			if s == "" {
				return "", errors.New("UserID cannot be empty")
			}
			if !strings.HasPrefix(s, "user_") {
				return UserID("user_" + s), nil
			}
			return UserID(s), nil
		}),
		// Register custom Department parser with validation
		genv.WithParser(func(s string) (Department, error) {
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
		}),
		// Register validated email parser
		genv.WithParser(func(s string) (ValidatedEmail, error) {
			if s == "" {
				return "", errors.New("email cannot be empty")
			}
			if !strings.Contains(s, "@") || !strings.Contains(s, ".") {
				return "", fmt.Errorf("invalid email format: %s", s)
			}
			return ValidatedEmail(strings.ToLower(s)), nil
		}),
	)

	// Create Genv with custom registry
	env := genv.New(
		genv.WithRegistry(registry),
		genv.WithAllowDefault(func(*genv.Genv) (bool, error) { return true, nil }),
	)

	var s CustomSettings

	// Simplified API works with custom types too!
	// Before: genv.Type(env.Var("CUSTOM_USER_ID").Default("demo123"), &s.UserID)
	// After:  genv.Bind("CUSTOM_USER_ID", &s.UserID).Default("demo123")
	err := genv.Parse(env,
		genv.Bind("CUSTOM_USER_ID", &s.UserID).Default("demo123"),
		genv.Bind("CUSTOM_DEPARTMENT", &s.Department).Default("engineering"),
		genv.Bind("CUSTOM_EMAIL", &s.ValidatedEmail).Default("demo@example.com"),
	)
	if err != nil {
		return CustomSettings{}, fmt.Errorf("parse custom env: %w", err)
	}

	return s, nil
}

// DemonstrateRegistryIsolation shows how different registries can have different behavior
func DemonstrateRegistryIsolation() {
	// Production registry with strict validation for a custom type
	prodRegistry := genv.NewRegistry(
		genv.WithParser(func(s string) (string, error) {
			if len(s) < 3 {
				return "", errors.New("production strings must be at least 3 characters")
			}
			return s, nil
		}),
		genv.WithParser(func(s string) (int, error) {
			// Production int parsing with range validation
			val, err := strconv.Atoi(s)
			if err != nil {
				return 0, err
			}
			if val < 0 || val > 1000 {
				return 0, errors.New("production integers must be between 0-1000")
			}
			return val, nil
		}),
	)

	// Development registry with lenient validation
	devRegistry := genv.NewRegistry(
		genv.WithParser(func(s string) (string, error) {
			if s == "" {
				return "dev-default", nil // Provide default in dev
			}
			return s, nil
		}),
		genv.WithParser(func(s string) (int, error) {
			// Development int parsing that's more forgiving
			if s == "" {
				return 42, nil // Default value in dev
			}
			return strconv.Atoi(s)
		}),
	)

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
	registry := genv.NewDefaultRegistry(
		// Register Priority parser (enum-like behavior)
		genv.WithParser(func(s string) (Priority, error) {
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
		}),
		// Register ServiceName parser with prefix handling
		genv.WithParser(func(s string) (ServiceName, error) {
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
		}),
		// Register LogLevel parser with strict validation
		genv.WithParser(func(s string) (ExampleLogLevel, error) {
			level := strings.ToUpper(strings.TrimSpace(s))
			switch level {
			case "DEBUG", "INFO", "WARN", "ERROR":
				return ExampleLogLevel(level), nil
			default:
				return "", fmt.Errorf("invalid log level: %s (valid: DEBUG, INFO, WARN, ERROR)", s)
			}
		}),
	)

	// Create Genv with custom registry and allow defaults
	env := genv.New(
		genv.WithRegistry(registry),
		genv.WithAllowDefault(func(*genv.Genv) (bool, error) { return true, nil }),
		genv.WithSplitKey("|"), // Use different delimiter for this example
	)

	var s AdvancedCustomSettings

	// Simplified API works with slices and custom types too!
	// Before: genv.Types(env.Var("TASK_PRIORITIES").Default("medium|high|low"), &s.Priorities)
	// After:  genv.BindMany("TASK_PRIORITIES", &s.Priorities).Default("medium|high|low")
	err := genv.Parse(env,
		genv.BindMany("TASK_PRIORITIES", &s.Priorities).Default("medium|high|low"),
		genv.Bind("SERVICE_NAME", &s.OptionalServiceName).Optional(),
		genv.Bind("LOG_LEVEL", &s.LogLevel).Default("INFO"),
	)
	if err != nil {
		return AdvancedCustomSettings{}, fmt.Errorf("parse advanced env: %w", err)
	}

	return s, nil
}
