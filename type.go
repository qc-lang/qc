package main

import (
	"fmt"
	"strings"
)

var types = map[string]Type{
	"struct": _Struct{},
	"string": SimpleType{c: "char*"},
	"int":    SimpleType{c: "int"},
}

type SimpleType struct {
	name string
	c    string
}

func (s SimpleType) C(pkg string) string {
	return s.c
}

func (s SimpleType) Def(pkg string) string {
	return fmt.Sprintf("typedef %s %s_%s;", s.c, pkg, s.name)
}

type Type interface {
	C(pkg string) string
	Def(pkg string) string
}

// Argument ...
type _Argument struct {
	name string
	typ  Type
}

func (f _Argument) C(pkg string) string {
	return fmt.Sprintf("%s %s", f.typ.C(pkg), f.name)
}

func (f _Argument) Def(pkg string) string {
	return fmt.Sprintf("%s %s_%s", f.typ.C(pkg), pkg, f.name)
}

// Struct ...
type _Struct struct {
	name   string
	fields []_Argument
}

func (s _Struct) C(pkg string) string {
	return fmt.Sprintf("struct %s_%s", pkg, s.name)
}

func (s _Struct) Def(pkg string) string {
	var fields []string
	for _, field := range s.fields {
		fields = append(fields, field.C(pkg)+";")
	}
	return fmt.Sprintf("typedef struct %s_%s {%s};", pkg, s.name, strings.Join(fields, ""))
}

// Function ...
type _Function struct {
	name    string
	returns Type
	args    []_Argument
	body    []string
}

func (f _Function) C(pkg string) string {
	return fmt.Sprintf("%s %s_%s();", f.returns.C(pkg), pkg, f.name)
}

func (f _Function) Def(pkg string) string {
	if f.returns == nil {
		f.returns = SimpleType{"void", "void"}
	}
	var args []string
	for _, arg := range f.args {
		args = append(args, arg.C(pkg))
	}
	mPkg := pkg + "_"
	if f.name == "main" {
		mPkg = ""
	}
	var body []string
	for _, line := range f.body {
		body = append(body, line+";")
	}

	return fmt.Sprintf("%s %s%s(%s) {%s}", f.returns.C(pkg), mPkg, f.name, strings.Join(args, ", "), strings.Join(body, ""))
}
