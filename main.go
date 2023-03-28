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
		gcc("debug/_out.c", stderr)
		//cmd("del -f ./debug/_out.c")
	case "run":
		Parse(args[1])
		gcc("debug/_out.c", stderr)
		run("./debug/a.exe")
		//cmd("del -f debug/_out.c")
		//cmd("del -f debug/a.exe")
	case "translate":
		Parse(args[1])
	}
}

func gcc(file string, stderr io.Writer) {
	cmd := exec.Command("gcc", file, "-o", "debug/a.exe")
	cmd.Stdout = os.Stdout
	cmd.Stderr = stderr
	cmd.Run()
}

func cmd(args string) {
	cmd := exec.Command("cmd", []string{"/C", args}...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func run(file string) {
	cmd := exec.Command(file)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
