package main

import (
	"io"
	"os"
	"os/exec"
)

func main() {
	args := os.Args[1:]

	if len(args) < 2 {
		panic("not enough arguments")
	}

	var stderr io.Writer
	if len(args) >= 3 {
		switch args[2] {
		case "-d":
			stderr = os.Stderr
		}
	}

	switch arg := args[0]; arg {
	case "build":
		Parse(args[1])
		gcc("_out.c", stderr)
		cmd("rm -f _out.c")
	case "run":
		Parse(args[1])
		gcc("_out.c", stderr)
		cmd("./a.out")
		cmd("rm -f _out.c")
		cmd("rm -f ./a.out")
	case "translate":
		Parse(args[1])
	}
}

func gcc(file string, stderr io.Writer) {
	cmd := exec.Command("gcc", file)
	cmd.Stdout = os.Stdout
	cmd.Stderr = stderr
	cmd.Run()
}

func cmd(args string) {
	cmd := exec.Command("/bin/sh", []string{"-c", args}...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
