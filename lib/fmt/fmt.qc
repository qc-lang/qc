package fmt

foreign import "stdio.h"

fn println(s string) {
    C.puts(s)
}

fn printf(s string) {
    C.printf(s)
}