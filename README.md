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

Optionally, use an alias for brevity (used in the examples below):
```go
import ev "github.com/rlebel12/goenvvars"
```


### Basic
In its most basic form, the package can be used to retrieve environment variables and then parse them into specified types:

```go
var StringVar = ev.New("STRING_VAR").String()
var BoolVar = ev.New("BOOL_VAR").Bool()
var IntVar = ev.New("INT_VAR").Int()
var FloatVar = ev.New("FLOAT_VAR").Float()
```

If the value from an environment variable cannot be parsed into the specified type, the function will panic.

### Optional Variables
By default, the package will panic if an environment variable is absent (either because the environment variable is not defined, or because it was set to an empty string). However, it is possible to specify that a variable is optional to prevent the panic behavior:

```go
var OptionalVar = ev.New("OPTIONAL_VAR", ev.Optional())
```

### Fallbacks
You can specify a fallback value to use if the environment variable is absent:

```go
var FallbackVar = ev.New("FALLBACK_VAR", ev.Fallback("fallback value"))
```

This is intended to be used to allow speed and ease of development while ensuring that all environment variables are defined before deploying to production. Thus, the default behavior is to forbid fallbacks when the `ENV` environment variable is set to `PRODUCTION` or `PROD`. This behavior can be overridden in two ways.

#### Allow Fallbacks: Global Override

Override the behavior for all subsequent invocations of `goenvvars.New` by calling changing the value of `goenvvars.DefaultAllowFallback`:

```go
ev.DefaultAllowFallback = func() bool { return true }
```

#### Allow Fallbacks: Individual Override

Override the behavior for individual environment variables by passing an `OverrideAllowFallback` option to `Fallback`:
    
```go
var FallbackVar = ev.New(
    "FALLBACK_VAR",
    ev.Fallback(
        "fallback value",
        ev.OverrideAllowFallback(func() bool { return true }),
    ),
)
```

This approach takes priority over the global override.

### Variable Presence
It is also possible to simply check whether an environment variable has been set to a non-empty value:

```go
var PresenceVar = ev.Presence("PRESENCE")
```