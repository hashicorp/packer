[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/txgruppi/parseargs-go)
[![Codeship](https://img.shields.io/codeship/173b62f0-bcc9-0133-0239-6e8926ac3d5c/master.svg?style=flat-square)](https://codeship.com/projects/136367)
[![Codecov](https://img.shields.io/codecov/c/github/txgruppi/parseargs-go/master.svg?style=flat-square)](https://codecov.io/github/txgruppi/parseargs-go)
[![Go Report Card](https://img.shields.io/badge/go_report-A+-brightgreen.svg?style=flat-square)](https://goreportcard.com/report/github.com/txgruppi/parseargs-go)

# `parseargs-go`

This is a port of the [parserargs.js](https://github.com/txgruppi/parseargs.js) project to [Go](https://golang.org).

What about parsing arguments allowing quotes in them? But beware that this library will not parse flags (-- and -), flags will be returned as simple strings.

## Installation

`go get -u github.com/txgruppi/parseargs-go`

## Example

```go
package main

import (
  "fmt"
  "log"

  "github.com/txgruppi/parseargs-go"
)

func main() {
  setInRedis := `set name "Put your name here"`
  parsed, err := parseargs.Parse(setInRedis)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("%#v\n", parsed) // []string{"set", "name", "Put your name here"}
}
```

## Tests

```
go get -u -t github.com/txgruppi/parseargs-go
cd $GOPATH/src/github.com/txgruppi/parseargs-go
go test ./...
```

## License

MIT
