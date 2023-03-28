package main

import (
	"fmt"
	"strings"
)

type Type interface {
	Generator
}

var types = map[string]Type{
	"struct": Structure{},
	"string": Alias{cname: "char*"},
	"int":    Alias{cname: "int"},
}

type Alias struct {
	alias string
	cname string
}

func (a Alias) Definition(_ string) string {
	def := &strings.Builder{}
	def.WriteString("typedef")
	def.WriteString(" ")
	def.WriteString(a.cname)
	def.WriteString(" ")
	def.WriteString(a.alias)
	def.WriteString(";")
	return def.String()
}

func (a Alias) Declaration(_ string) string {
	decl := &strings.Builder{}
	decl.WriteString(a.alias)
	return decl.String()
}

type Arg struct {
	label string
	typ   Type
}

func (a Arg) Format(pkg, sep string) string {
	decl := a.typ.Declaration(pkg)
	if al, ok := a.typ.(Alias); ok {
		pkg = ""
		if len(decl) == 0 {
			decl = al.cname
		}
	}
	return fmt.Sprintf("%s %s%s", decl, a.label, sep)
}

type Structure struct {
	label  string
	fields []Arg
}

func (s Structure) Definition(pkg string) string {
	def := &strings.Builder{}
	def.WriteString("typedef")
	def.WriteString(" ")
	def.WriteString("struct")
	def.WriteString(" ")
	label := s.label
	if len(pkg) > 0 {
		label = pkg + "_" + label
	}
	def.WriteString(label)
	def.WriteString("{")
	for _, f := range s.fields {
		def.WriteString(f.Format(pkg, ";"))
	}
	def.WriteString("}")
	def.WriteString(label)
	def.WriteString(";")
	return def.String()
}

func (s Structure) Declaration(pkg string) string {
	decl := &strings.Builder{}
	if len(pkg) > 0 {
		decl.WriteString(pkg + "_")
	}
	decl.WriteString(s.label)
	return decl.String()
}
