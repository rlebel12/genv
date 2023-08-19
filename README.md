# goevars

A small package to help work with environment variables in Go. Variables are to be initialized like so:
    
```go
var StringVar = Required("STRING", "default").String()
var BoolVar = Required("BOOL", "false").Bool()
var OptionalIntVar = Optional("OPTIONAL_INT", "0").Int()
```

If the value from an environment variable cannot be parsed into the specified type, the function will panic. If the environment variable is not set, the default value will be used.

If the variable is optional, the fallback value will always be used. If the variable is required, the fallback value will only be used in non-production environments. Otherwise, the function will panic.