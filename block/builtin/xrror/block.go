package xrror

import (
	f "flag"
	"strings"

	"github.com/thrisp/marid/block"
	"github.com/thrisp/marid/flag"
	"github.com/thrisp/marid/loader"
)

var Block block.Block = block.BasicBlock(
	"xrror",
	fs,
	lr,
	[]string{"xrror"},
)

var fs flag.Flagset = flag.NewFlagset("xrror", mkFlagSet())

var (
	ErrorName         string
	Letter            string
	ErrorFunctionName string
)

func mkFlagSet() *f.FlagSet {
	ret := f.NewFlagSet("xrror", f.PanicOnError)
	ret.StringVar(&ErrorName, "ErrorName", "xrror", "")
	ret.StringVar(&Letter, "Letter", strings.ToLower(string(ErrorName[0])), "")
	ret.StringVar(&ErrorFunctionName, "ErrorFunctionName", "Xrror", "")
	return ret
}

var lr loader.Loader = loader.MapLoader(ml)

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
