package xrror

import (
	"flag"

	. "github.com/thrisp/marid/b"
	. "github.com/thrisp/marid/f"
	. "github.com/thrisp/marid/l"
)

var Xrror Block = BasicBlock(
	"xrror",
	fs,
	lr,
	[]string{"xrror"},
)

var fs Flags = NewFlag("xrror", mkFlagSet())

var (
	ErrorName         string
	Letter            string
	ErrorFunctionName string
)

func mkFlagSet() *flag.FlagSet {
	ret := flag.NewFlagSet("xrror", flag.PanicOnError)
	ret.StringVar(&ErrorName, "ErrorName", "xrror", "")
	ret.StringVar(&Letter, "Letter", string(ErrorName[0]), "")
	ret.StringVar(&ErrorFunctionName, "ErrorFunctionName", "Xrror", "")
	return ret
}

var lr Loader = MapLoader(ml)

var ml map[string]string = map[string]string{
	"xrror": tmpl,
}

var tmpl string = `package {{.PackageName}}

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
`
