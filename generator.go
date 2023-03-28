package main

import (
	"fmt"
	"strings"
)

type Func struct {
	label string
	ret   Type
	args  []Arg
	body  []string
}

func (f Func) Gen(pkg string) string {
	if f.ret == nil {
		f.ret = Alias{alias: "void"}
	}
	var args []string
	for _, arg := range f.args {
		args = append(args, arg.Format(pkg, ","))
	}
	if len(args) >= 1 {
		args[len(args)-1] = strings.TrimSuffix(args[len(args)-1], ",")
	}

	mPkg := pkg + "_"
	if f.label == "main" {
		mPkg = ""
	}

	rPkg := pkg + "_"
	if _, ok := f.ret.(Alias); ok {
		rPkg = ""
	}

	var body []string
	for _, line := range f.body {
		body = append(body, line+";")
	}

	return fmt.Sprintf("%s %s%s(%s) {%s}", f.ret.Declaration(rPkg), mPkg, f.label, strings.Join(args, ""), strings.Join(body, ""))
}

type Generator interface {
	Declaration(pkg string) string
	Definition(pkg string) string
}
