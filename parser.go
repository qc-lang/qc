package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Pkg struct {
	name    string
	builder *Builder
}

type Include struct {
	alias    string
	filename string
}

type Builder struct {
	types     map[string]Type
	includes  []Include
	functions []*Func
}

type Checker struct {
	*Pkg
	packages  []*Pkg
	tokenizer Tokenizer
	prevToken Token
	currToken Token

	filename string
}

func (c *Checker) Fatalf(pos Pos, format string, args ...any) {
	fmt.Printf("%s(%d:%d)", c.filename, pos.line, pos.column)
	fmt.Printf(format, args...)
	fmt.Println()
	os.Exit(1)
}

func (c *Checker) Next() (res Token) {
	token, err := c.tokenizer.Token()
	if err != nil && err != io.EOF {
		c.Fatalf(c.tokenizer.Pos, " found invalid token: %v", err)
	}
	c.prevToken, c.currToken = c.currToken, token
	return c.prevToken
}

func (c *Checker) Expect(kind TokenKind) Token {
	token := c.Next()
	if token.kind != kind {
		c.Fatalf(token.Pos, " expected token %v, got %v", kind, token.kind)
	}
	return token
}

func (c *Checker) Allow(kind TokenKind) bool {
	if c.currToken.kind == kind {
		c.Next()
		return true
	}
	return false
}

func (c *Checker) Current() TokenKind {
	if c.currToken.kind == Comment {
		c.Next()
		return c.Current()
	}
	return c.currToken.kind
}

func (c *Checker) ForeignDecl() {
	c.Expect(Foreign)
	c.Expect(Import)
	tok := c.Next()
	alias := tok.text
	if tok.kind == Identifier {
		tok = c.Expect(String)
		alias = ""
	}
	c.builder.includes = append(c.builder.includes, Include{
		alias:    alias,
		filename: tok.text,
	})
}

func (c *Checker) ImportDecl() {
	c.Expect(Import)
	tok := c.Expect(String)
	d, err := os.ReadDir(fmt.Sprintf("./lib/%s", tok.text))
	if os.IsNotExist(err) {
		c.Fatalf(tok.Pos, " unknown package %s", tok.text)
	}
	pkg := &Pkg{
		name: tok.text,
		builder: &Builder{
			types: map[string]Type{},
		},
	}
	for _, f := range d {
		if f.IsDir() {
			continue
		}
		filename := fmt.Sprintf("./lib/%s/%s", tok.text, f.Name())
		dat, err := os.ReadFile(filename)
		if err != nil {
			panic(err)
		}

		ch := &Checker{
			tokenizer: NewTokenizer(string(dat)),
			filename:  filename,
			Pkg:       pkg,
		}
		ch.Next()
		ch.Expect(Package)
		ch.name = ch.Expect(Identifier).text
		ch.Next()

	decls:
		for {
			switch c2 := ch.Current(); c2 {
			case Foreign:
				ch.ForeignDecl()
			case Fn:
				ch.FunctionDecl()
			case TypeDef:
				ch.TypeDefDecl()
			case Import:
				ch.ImportDecl()
			default:
				break decls
			}
		}
		ch.Allow(Semicolon)
		ch.Expect(EOF)
	}
	c.packages = append(c.packages, pkg)
}

func (c *Checker) FunctionCall(pkg *Pkg) string {
	var args []string
	var filename string

	if pkg == nil {
		filename = c.Next().text
		c.Expect(LeftParen)
		for tok := c.Next(); tok.kind != RightParen; tok = c.Next() {
			txt := tok.text
			if tok.kind == String {
				txt = fmt.Sprintf(`"%s"`, tok.text)
			}
			args = append(args, txt)
		}
		c.Next()
		return fmt.Sprintf("%s(%s)", filename, strings.Join(args, ","))
	}

	n := c.Next().text
	for _, f := range pkg.builder.functions {
		if n == f.label {
			filename = f.label
			c.Expect(LeftParen)
			for range f.args {
				tk := c.Next()
				txt := tk.text
				if tk.kind == String {
					txt = fmt.Sprintf(`"%s"`, tk.text)
				}
				args = append(args, txt)
			}
			c.Expect(RightParen)
		}
	}
	return fmt.Sprintf("%s_%s(%s)", pkg.name, filename, strings.Join(args, ","))
}

func (c *Checker) FunctionDecl() {
	c.Expect(Fn)
	name := c.Expect(Identifier).text
	c.Expect(LeftParen)
	var args []Arg
	for c.Current() != RightParen {
		args = append(args, c.Argument())
		c.Allow(Comma)
	}
	c.Expect(RightParen)
	var returns Type
	if c.Current() != LeftBrace {
		returns = c.Type(c.Expect(Identifier))
	}
	c.Expect(LeftBrace)

	var body []string

	for tok := c.Next(); tok.kind != RightBrace; tok = c.Next() {
		tok := tok
		if tok.kind == Identifier && c.Next().kind == Dot {
			if tok.text == "C" {
				body = append(body, c.FunctionCall(nil))
				break
			}
			for _, pkg := range c.packages {
				if pkg.name == tok.text {
					body = append(body, c.FunctionCall(pkg))
					break
				}
			}
		}
	}
	c.builder.functions = append(c.builder.functions, &Func{
		label: name,
		args:  args,
		body:  body,
		ret:   returns,
	})
	for _, f := range c.builder.functions {
		fmt.Println(f.label)
	}
	c.Next()
}

func (c *Checker) TypeDefDecl() {
	c.Expect(TypeDef)
	name := c.Expect(Identifier).text
	typ := c.Type(c.Next())
	if t, ok := typ.(Alias); ok {
		t.alias = name
		typ = t
		c.Next()
	} else {
		switch typ.(type) {
		case Structure:
			c.Expect(LeftBrace)
			var fields []Arg
			for c.Current() != RightBrace {
				fields = append(fields, c.Argument())
				c.Allow(Semicolon)
			}
			c.Expect(RightBrace)
			typ = Structure{
				label:  name,
				fields: fields,
			}
		}
	}
	c.builder.types[name] = typ
}

func (c *Checker) Type(kind Token) Type {
	var typ Type
	for n, t := range types {
		if n == kind.text {
			typ = t
			break
		}
	}
	for n, t := range c.builder.types {
		if n == kind.text {
			typ = t
			break
		}
	}
	if typ == nil {
		c.Fatalf(kind.Pos, " unknown type %s", kind.text)
	}
	return typ
}

func (c *Checker) Argument() Arg {
	name := c.Expect(Identifier)
	kind := c.Expect(Identifier)
	typ := c.Type(kind)
	arg := Arg{
		label: name.text,
		typ:   typ,
	}
	return arg
}

func Parse(filename string) {
	dat, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	c := &Checker{
		tokenizer: NewTokenizer(string(dat)),
		filename:  filename,
		Pkg:       &Pkg{},
	}
	c.builder = &Builder{
		types: map[string]Type{},
	}
	c.Next()
	c.Expect(Package)
	c.name = c.Expect(Identifier).text
	c.Next()

decls:
	for {
		switch c2 := c.Current(); c2 {
		case Foreign:
			c.ForeignDecl()
		case Fn:
			c.FunctionDecl()
		case TypeDef:
			c.TypeDefDecl()
		case Import:
			c.ImportDecl()
		default:
			break decls
		}
	}
	c.Allow(Semicolon)
	c.Expect(EOF)

	data := []byte(c.writeC())
	err = os.WriteFile("debug/_out.c", data, 0644)
}

func (c *Checker) writeC() string {
	pkgs := append([]*Pkg{c.Pkg}, c.packages...)
	var lines []string
	for _, pkg := range pkgs {
		b := pkg.builder
		for _, include := range b.includes {
			lines = append(lines, fmt.Sprintf("#include <%s>", include.filename))
		}
	}
	for _, pkg := range pkgs {
		b := pkg.builder
		for _, typ := range b.types {
			lines = append(lines, typ.Definition(pkg.name))
		}
	}
	for _, pkg := range pkgs {
		b := pkg.builder
		for _, f := range b.functions {
			lines = append(lines, f.Gen(pkg.name))
		}
	}
	return strings.Join(lines, "\n")
}
