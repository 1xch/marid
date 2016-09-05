package xrror

import (
	"context"
	"strings"

	"github.com/thrisp/marid/lib/block"
	"github.com/thrisp/marid/lib/env"
	"github.com/thrisp/marid/lib/template"
)

type options struct {
	Directory, Package, ErrorName, Letter, ErrorFunctionName string
}

var defaultOptions = options{".", "main", "xrror", "x", "Xrror"}

func mkFlagSet(o *options) *env.FlagSet {
	ret := env.NewFlagSet("xrror", env.PanicOnError)
	ret.StringVar(&o.Directory, "Directory", o.Directory, "")
	ret.StringVar(&o.Package, "Package", o.Package, "")
	ret.StringVar(&o.ErrorName, "ErrorName", o.ErrorName, "")
	ret.StringVar(&o.Letter, "Letter", strings.ToLower(string(o.ErrorName[0])), "")
	ret.StringVar(&o.ErrorFunctionName, "ErrorFunctionName", o.ErrorFunctionName, "")
	return ret
}

var xl template.Loader = template.MapLoader(em)

var em map[string]string = map[string]string{
	"xrror": et,
}

var et string = `{{ extends "block_base" }}
{{ define "block_root" }}package {{.Package}}

import(
	"fmt"
)

type {{.ErrorName}} struct {
	base  string
	vals []interface{}
}

func ({{.Letter}} *{{.ErrorName}}) Error() string {
	return fmt.Sprintf("%s", fmt.Sprintf({{.Letter}}.base, {{.Letter}}.vals...))
}

func ({{.Letter}} *{{.ErrorName}}) Out(vals ...interface{}) *{{.ErrorName}} {
	{{.Letter}}.vals = vals
	return {{.Letter}}
}

func {{.ErrorFunctionName}}(base string) *{{.ErrorName}} {
	return &{{.ErrorName}}{base: base}
}
{{ end }}
`

var e *env.Env

func xrrorBlock() block.Block {
	o := &defaultOptions
	fs := mkFlagSet(o)
	b := block.New(
		block.Templater(template.New(10, nil, xl)),
	)
	b.Add(block.Command(
		"",
		"xrror",
		"generates a custom, multi use error struct",
		50,
		func(c context.Context, v []string) env.ExitStatus {
			td := template.Data(b, "Block:Xrror")
			err := b.Render(o.Directory, "xrror", "xrror", td.Data)
			if err != nil {
				e.Printf("block render error: %s", err)
				return env.ExitFailure
			}
			e.Printf("rendered block Xrror")
			return env.ExitSuccess
		},
		fs,
	))
	err := b.Configure()
	if err != nil {
		e.Fatal(err.Error())
	}
	return b
}

func init() {
	e = env.Current
	e.Register(xrrorBlock())
}
