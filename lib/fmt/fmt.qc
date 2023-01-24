package fmt

foreign import "stdio.h"

func printf(s: string) {
    C.printf(s)
}