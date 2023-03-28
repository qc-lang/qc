package main

// The import keyword is used to import libraries
import "fmt"

// The def keyword is used to define a new type
type Foo string
type Bar int

// structs...
type FooBar struct {
 // name type
    Foo string
    Bar int
}

type BarFoo struct {
    Foo FooBar
}

fn main() {
  fmt.println("Hello, World!")
}