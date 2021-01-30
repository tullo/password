## Golang Password Generator

[![GoDoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/tullo/password/password)
[![GitHub Actions](https://img.shields.io/github/workflow/status/tullo/password/Test?style=flat-square)](https://github.com/tullo/password/actions?query=workflow%3ATest)

This library implements generation of random passwords with provided
requirements as described by  [AgileBits
1Password](https://discussions.agilebits.com/discussion/23842/how-random-are-the-generated-passwords)
in pure Golang. The algorithm is commonly used when generating website
passwords.

The library uses crypto/rand for added randomness.

Sample example passwords this library may generate:

```text
0N[k9PhDqmmfaO`p_XHjVv`HTq|zsH4XiH8umjg9JAGJ#\Qm6lZ,28XF4{X?3sHj
7@90|0H7!4p\,c<!32:)0.9N
UlYuRtgqyWEivlXnLeBpZvIQ
Q795Im1VR5h363s48oZGaLDa
wpvbxlsc
```

## Installation

```sh
$ go get -u github.com/tullo/password/password
```

## Usage

```golang
package main

import (
  "log"

  "github.com/tullo/password/password"
)

func main() {
  // Generate a password that is 64 characters long with 10 digits, 10 symbols,
  // allowing upper and lower case letters, disallowing repeat characters.
  res, err := password.Generate(64, 10, 10, false, false)
  if err != nil {
    log.Fatal(err)
  }
  log.Printf(res)
}
```

See the [GoDoc](https://pkg.go.dev/github.com/tullo/password) for more
information.

## Testing

For testing purposes, instead of accepting a `*password.StatefulGenerator`
struct, accept a `password.Generator` interface:

```go
// func MyFunc(p *password.Generator)
func MyFunc(p password.Generator) {
  // ...
}
```

Then, in tests, use a mocked password generator with stubbed data:

```go
func TestMyFunc(t *testing.T) {
  gen := password.NewMockPasswordGenerator("canned-response", false)
  MyFunc(gen)
}
```

In this example, the mock generator will always return the value
"canned-response", regardless of the provided parameters.

## License

This code is licensed under the MIT license.
