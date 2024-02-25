# goenvvars
[![Go Reference](https://pkg.go.dev/badge/github.com/rlebel12/goenvvars/v3.svg)](https://pkg.go.dev/github.com/rlebel12/goenvvars/v3)
[![Test](https://github.com/rlebel12/goenvvars/actions/workflows/test.yml/badge.svg)](https://github.com/rlebel12/goenvvars/actions/workflows/test.yml)

A small package to help work with environment variables in Go.

## Installation
```console
go get github.com/rlebel12/goenvvars/v3
```

## Usage

First, ensure that the package is imported:
```go
import "github.com/rlebel12/goenvvars/v3"
```


### Basic
In its most basic form, the package can be used to retrieve environment variables and then parse them into specified types:

```go
var genv, _ := goenvvars.NewGenv()
var StringVar = genv.New("STRING_VAR").String()
var BoolVar = genv.New("BOOL_VAR").Bool()
var IntVar = genv.New("INT_VAR").Int()
var FloatVar = genv.New("FLOAT_VAR").Float()
var URLVar = genv.New("URL_VAR").URL()
```

If the value from an environment variable cannot be parsed into the specified type, the function will panic. Alternatively, the `Try*` functions can be used to return an error instead of panicking.
```go
myVar, err := genv.New("MY_VAR").TryString()
```

### Optional Variables
By default, the package will fail if an environment variable is absent (either because the environment variable is not defined, or because it was set to an empty string). However, it is possible to specify that a variable is optional to prevent the panic behavior:

```go
var OptionalVar = genv.New("OPTIONAL_VAR").Optional()
```

### Fallbacks
You can specify a fallback value to use if the environment variable is absent:

```go
var FallbackVar = genv.New("FALLBACK_VAR").Fallback("fallback value")
```

This is intended to be used to allow speed and ease of development while ensuring that all environment variables are defined before deploying to production. Thus, the default behavior is to forbid fallbacks when the `ENV` environment variable is set to `PRODUCTION` or `PROD`. This behavior can be overridden in two ways.

#### Allow Fallbacks: Global Override

Override the behavior for all subsequent invocations of `goenvvars.New`:

```go
var genv, _ := goenvvars.NewGenv(
    goenvvars.DefaultAllowFallback(func() bool { return true }),
)
```

#### Allow Fallbacks: Individual Override

Override the behavior for individual environment variables by passing an `OverrideAllow` option to `Fallback`:
    
```go
var FallbackVar = genv.New("FALLBACK_VAR").Fallback(
    "fallback value",
    goenvvars.OverrideAllow(func() bool { return true }),
)
```

This approach takes priority over the global override.

### Combining Options
Options can be chained together. For example, it is possible to declare that an environment variable is both
optional and has a fallback value. This means that the fallback value will be used if allowed and necessary, and the program
will not panic if the final value is still absent.

```go
var OptionalFallbackVar = genv.New("OPTIONAL_FALLBACK_VAR").
    Fallback("fallback value").
    Optional()
```

### Variable Presence
It is also possible to simply check whether an environment variable has been set to a non-empty value:

```go
var PresenceVar = goenvvars.Presence("PRESENCE")
```