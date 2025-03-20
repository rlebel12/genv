# genv

[![Go Reference](https://pkg.go.dev/badge/github.com/rlebel12/genv.svg)](https://pkg.go.dev/github.com/rlebel12/genv)
[![Test](https://github.com/rlebel12/genv/actions/workflows/test.yml/badge.svg)](https://github.com/rlebel12/genv/actions/workflows/test.yml)

A small package to help work with environment variables in Go.

## Installation

```console
go get github.com/rlebel12/genv
```

## Usage

First, ensure that the package is imported:

```go
import "github.com/rlebel12/genv"
```

### Basic

In its most basic form, the package can be used to retrieve environment variables and then parse them into specified types:

```go
env := genv.New()

StringVar := env.Var("STRING_VAR").NewString()
BoolVar := env.Var("BOOL_VAR").NewBool()
IntVar := env.Var("INT_VAR").NewInt()
FloatVar := env.Var("FLOAT_VAR").NewFloat64()
URLVar := env.Var("URL_VAR").NewURL()

if err := env.Parse(); err != nil {
    slog.Error("env parse", "error", err.Error())
}

// Or instead, using pointers:

type Example struct {
    StringVar string
    IntVar    int
}

var example Example
env.Var("STRING_VAR").String(&example.StringVar)
env.Var("INT_VAR").Int(&example.IntVar)

if err := env.Parse(); err != nil {
    slog.Error("env parse", "error", err.Error())
}
```

### Optional Variables

By default, parsing will fail if an environment variable is absent (either because the environment variable is not defined, or because it was set to an empty string). However, it is possible to specify that a variable is optional to prevent the failure and default to the zero value:

```go
var OptionalVar = genv.Var("OPTIONAL_VAR").Optional()
```

### Defaults

You can specify a default value to use if the environment variable is absent:

```go
var DefaultVar = genv.Var("DEFAULT_VAR").Default("default value")
```

This is intended to be used to allow speed and ease of development while ensuring that all environment variables are defined before deploying to production. Thus, the default behavior is to forbid defaults unless the `GENV_ALLOW_DEFAULT` environment variable evaluates to `true`. This behavior can be overridden in two ways.

#### Allow Defaults: Global Override

Override the behavior for all subsequent invocations of `genv.Var`:

```go
var genv := genv.New(
    genv.WithAllowDefault(func() bool { return true }),
)
```

#### Allow Defaults: Individual Override

Override the behavior for individual environment variables by passing an `WithAllowDefault` option to `Default`:

```go
var DefaultVar = genv.Var("DEFAULT_VAR").Default(
    "default value",
    genv.WithAllowDefault(func() bool { return true }),
)
```

This approach takes priority over the global override.

### Combining Options

Options can be chained together. For example, it is possible to declare that an environment variable is both
optional and has a default value. This means that the default value will be used if allowed and necessary, and the program
will not panic if the final value is still absent.

```go
var OptionalDefaultVar = genv.Var("OPTIONAL_DEFAULT_VAR").
    Default("default value").
    Optional()
```

### Example

See the `example` package for a more complete demonstration of how this package can be used.
