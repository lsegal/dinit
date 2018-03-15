# DInit

[![GoDoc](https://godoc.org/github.com/lsegal/dinit?status.svg)](https://godoc.org/github.com/lsegal/dinit)
[![Build Status](https://travis-ci.org/lsegal/dinit.svg?branch=master)](https://travis-ci.org/lsegal/dinit)

DInit is a library to initialize structs using Dependenci Injection (DI).

## Installation

```sh
go get -u github.com/lsegal/dinit
```

## Usage

Use `dinit.Init(val1, val2, ...)` where arguments are either struct values,
references, bare functions, or initializer functions that return struct values.

DInit will then call your functions in the appropriate order to ensure that
all functions have their arguments filled by the values provided in the
`Init()` call.

For example, in the code below we have a `log.Logger` object, a `client` and
`service`. The `service` depends on `log.Logger,` and `client` depends on
both `service` and `log.Logger`. When we call their initializers, only 1
instance of client and server are created:

```go
func newClient(l *log.Logger, svc lister) *client { /* ... */ }
func newService(l *log.Logger) *service { /* ... */ }

func main() {
  l := log.New(os.Stdout, "", log.Lshortfile)
  useClient := func(c *client) { c.PrintPeople() }
  dinit.Init(newClient, useClient, newService, l)

  // Output:
  // main.go:40: Initializing service
  // main.go:35: Initializing client
  // main.go:30: Client asked for a list of people
  // main.go:18: People: Sarah, Bob, André
}
```

You can see the full (runnable) example in `examples/client_server/main.go`.

## Copyright & License

Copyright (c) 2018 Loren Segal. All rights reserved.

This project is licensed under the MIT license, see `LICENSE` for details.
