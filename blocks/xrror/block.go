package xrror

import (
	"flag"
	"strings"

	"github.com/thrisp/marid"
)

var Block marid.Block = marid.BasicBlock(
	"xrror",
	mkFlagSet(),
	xl,
	[]string{"xrror"},
)

var (
	ErrorName         string
	Letter            string
	ErrorFunctionName string
)

func mkFlagSet() *flag.FlagSet {
	ret := flag.NewFlagSet("xrror", flag.PanicOnError)
	ret.StringVar(&ErrorName, "ErrorName", "xrror", "")
	ret.StringVar(&Letter, "Letter", strings.ToLower(string(ErrorName[0])), "")
	ret.StringVar(&ErrorFunctionName, "ErrorFunctionName", "Xrror", "")
	return ret
}

var xl marid.Loader = marid.MapLoader(em)

var em map[string]string = map[string]string{
	"xrror": et,
}

var et string = `{{ extends "block_base" }}
{{ define "block_root" }}package {{.PackageName}}

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
