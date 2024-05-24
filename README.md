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

### Defaults
You can specify a default value to use if the environment variable is absent:

```go
var DefaultVar = genv.New("DEFAULT_VAR").Default("default value")
```

This is intended to be used to allow speed and ease of development while ensuring that all environment variables are defined before deploying to production. Thus, the default behavior is to forbid defaults when the `ENV` environment variable is set to `PRODUCTION` or `PROD`. This behavior can be overridden in two ways.

#### Allow Defaults: Global Override

Override the behavior for all subsequent invocations of `goenvvars.New`:

```go
var genv, _ := goenvvars.New(
    goenvvars.DefaultAllowDefault(func() bool { return true }),
)
```

#### Allow Defaults: Individual Override

Override the behavior for individual environment variables by passing an `WithAllowDefault` option to `Default`:
    
```go
var DefaultVar = genv.New("DEFAULT_VAR").Default(
    "default value",
    goenvvars.WithAllowDefault(func() bool { return true }),
)
```

This approach takes priority over the global override.

### Combining Options
Options can be chained together. For example, it is possible to declare that an environment variable is both
optional and has a default value. This means that the default value will be used if allowed and necessary, and the program
will not panic if the final value is still absent.

```go
var OptionalDefaultVar = genv.New("OPTIONAL_DEFAULT_VAR").
    Default("default value").
    Optional()
```

### Example
See the `example` package for a more complete demonstration of how this package can be used.

### Variable Presence
It is also possible to simply check whether an environment variable has been set to a non-empty value:

```go
var PresenceVar = goenvvars.Presence("PRESENCE")
```