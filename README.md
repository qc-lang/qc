# Qc

C transpiler prototype written in go!
don't expect it to be working well, it's very much incomplete and is missing a lot of stuff, but feel free to contribute.

# Discord
https://discord.gg/7SjjYNA2Xb

# Note
the tokenizer and parser were mostly inspired by [Blaise](https://github.com/gingerBill/blaise) by the creator of the [odin language](https://odin-lang.org)

# Basic Example
```go
package main

// The import keyword is used to import libraries
import "fmt"

// The def keyword is used to define a new types
type Foo string
type Bar int

// structs...
type FooBar struct {
 // name types
    Foo string
    Bar int
}

type BarFoo struct {
    Foo FooBar
}

fn main() {
  fmt.println("Hello, World!")
}
```
