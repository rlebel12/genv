# goevars
[![Test](https://github.com/rlebel12/goevars/actions/workflows/test.yml/badge.svg)](https://github.com/rlebel12/goevars/actions/workflows/test.yml)

A small package to help work with environment variables in Go.

## Installation
```console
go get github.com/rlebel12/goevars
```

## Required & Optional Variables

Environment variables can be required or optional. The goal of `Required` is to make it easy to work with environment variables in development while also ensuring that necessary configurations are handled in production. If an environment variable is not necessary, or otherwise should not cause a panic in production, `Optional` can be used as an alternative. 

Variables are set like so:
```go
var StringVar = goevars.Required("STRING", "default").String()
var BoolVar = goevars.Required("BOOL", "false").Bool()
var FloatVar = goevars.Required("FLOAT", "0.0").Float()
var OptionalIntVar = goevars.Optional("OPTIONAL_INT", "0").Int()
```

If the variable is optional, the fallback value will always be used. If the variable is required, the fallback value will only be used in non-production environments. Otherwise, the function will panic.

If the value from an environment variable cannot be parsed into the specified type, the function will panic.

## Variable Presence

It is also possible to check whether an environment variable has been set to a non-empty value:
```go
var PresenceVar = goevars.Presence("PRESENCE")
```