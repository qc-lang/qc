package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type _Package struct {
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
	functions []*_Function
}

type Checker struct {
	*_Package
	packages  []*_Package
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
	pkg := &_Package{
		name: tok.text,
		builder: &Builder{
			types: map[string]Type{},
		},
	}
	for _, f := range d {
		if f.IsDir() {
			continue
		}
		fname := fmt.Sprintf("./lib/%s/%s", tok.text, f.Name())
		dat, err := os.ReadFile(fname)
		if err != nil {
			panic(err)
		}

		ch := &Checker{
			tokenizer: NewTokenizer(string(dat)),
			filename:  fname,
			_Package:  pkg,
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

func (c *Checker) FunctionCall(pkg *_Package) string {
	var args []string
	var fname string

	if pkg == nil {
		fname = c.Next().text
		c.Expect(LeftParen)
		for tok := c.Next(); tok.kind != RightParen; tok = c.Next() {
			txt := tok.text
			if tok.kind == String {
				txt = fmt.Sprintf(`"%s"`, tok.text)
			}
			args = append(args, txt)
		}
		c.Next()
		return fmt.Sprintf("%s(%s)", fname, strings.Join(args, ","))
	}

	n := c.Next().text
	for _, f := range pkg.builder.functions {
		if n == f.name {
			fname = f.name
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
	return fmt.Sprintf("%s_%s(%s)", pkg.name, fname, strings.Join(args, ","))
}

func (c *Checker) FunctionDecl() {
	c.Expect(Fn)
	name := c.Expect(Identifier).text
	c.Expect(LeftParen)
	var args []_Argument
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
	c.builder.functions = append(c.builder.functions, &_Function{
		name:    name,
		args:    args,
		body:    body,
		returns: returns,
	})
	c.Next()
}

func (c *Checker) TypeDefDecl() {
	c.Expect(TypeDef)
	name := c.Expect(Identifier).text
	typ := c.Type(c.Next())
	if t, ok := typ.(SimpleType); ok {
		t.name = name
		typ = t
		c.Next()
	} else {
		switch typ.(type) {
		case _Struct:
			c.Expect(LeftBrace)
			var fields []_Argument
			for c.Current() != RightBrace {
				fields = append(fields, c.Argument())
				c.Allow(Semicolon)
			}
			c.Expect(RightBrace)
			typ = _Struct{
				name:   name,
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

func (c *Checker) Argument() _Argument {
	name := c.Expect(Identifier)
	kind := c.Expect(Identifier)
	typ := c.Type(kind)
	arg := _Argument{
		name: name.text,
		typ:  typ,
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
		_Package:  &_Package{},
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
	pkgs := append([]*_Package{c._Package}, c.packages...)
	var lines []string
	for _, pkg := range pkgs {
		b := pkg.builder
		for _, include := range b.includes {
			lines = append(lines, fmt.Sprintf("#include <%s>", include.filename))
		}
		for _, typ := range b.types {
			lines = append(lines, typ.Def(pkg.name))
		}
		for _, f := range b.functions {
			lines = append(lines, f.Def(pkg.name))
		}
	}
	return strings.Join(lines, "\n")
}
