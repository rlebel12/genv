# goenvvars
[![Test](https://github.com/rlebel12/goenvvars/actions/workflows/test.yml/badge.svg)](https://github.com/rlebel12/goenvvars/actions/workflows/test.yml)

A small package to help work with environment variables in Go.

## Installation
```console
go get github.com/rlebel12/goenvvars
```

## Usage

First, ensure that the package is imported:
```go
import "github.com/rlebel12/goenvvars"
```

### Basic
In its most basic form, the package can be used to retrieve environment variables and then parse them into specified types:

```go
var StringVar = goenvvars.NewEnvVar("STRING_VAR").String()
var BoolVar = goenvvars.NewEnvVar("BOOL_VAR").Bool()
var IntVar = goenvvars.NewEnvVar("INT_VAR").Int()
var FloatVar = goenvvars.NewEnvVar("FLOAT_VAR").Float()
```

If the value from an environment variable cannot be parsed into the specified type, the function will panic.

### Optional Variables
By default, the package will panic if an environment variable is absent (either because the environment variable is not defined, or because it was set to an empty string). However, it is possible to specify that a variable is optional to prevent the panic behavior:

```go
var OptionalVar = goenvvars.NewEnvVar("OPTIONAL_VAR", Optional())
```

### Fallbacks
You can specify a fallback value to use if the environment variable is absent:

```go
var FallbackVar = goenvvars.NewEnvVar("FALLBACK_VAR", Fallback("fallback value"))
```

This is intended to be used to allow speed and ease of development while ensuring that all environment variables are defined before deploying to production. Thus, the default behavior is to forbid fallbacks when the `ENV` environment variable is set to `PRODUCTION` or `PROD`. This behavior can be overridden in two ways.

#### Global Override

Override the behavior for all subsequent invocations of `goenvvars.NewEnvVar` by calling changing the value of `goenvvars.DefaultAllowFallback`:

```go
goenvvars.DefaultAllowFallback = func() bool { return true }
```

#### Individual Override

Override the behavior for individual environment variables by passing an `OverrideAllowFallback` option to `Fallback`:
    
```go
var FallbackVar = goenvvars.NewEnvVar(
    "FALLBACK_VAR",
    Fallback(
        "fallback value",
        OverrideAllowFallback(func() bool { return true }),
    ),
)
```

This approach takes priority over the global override.

### Variable Presence
It is also possible to simply check whether an environment variable has been set to a non-empty value:

```go
var PresenceVar = goenvvars.Presence("PRESENCE")
```