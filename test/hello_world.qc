package main

// The import keyword is used to import libraries
import "fmt"

// The def keyword is used to define a new types
type Foo string
type Bar int

// structs...
type FooBar struct {
 // name types
    Foo Foo
    Bar int
}

type BarFoo struct {
    Foo FooBar
}

fn main() {
  fmt.println("Hello, World!")
}