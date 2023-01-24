package main

// The import keyword is used to import libraries
import "fmt"

// The def keyword is used to define a new type
def Foo string
def Bar int

// structs...
def FooBar struct {
 // name type
    Foo: string
    Bar: int
}

def BarFoo struct {
    Foo: FooBar
}

func main() {
  fmt.printf("Hello, World!\n")
}