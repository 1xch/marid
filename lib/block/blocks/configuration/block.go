package configuration

import (
	"context"
	"strings"

	"github.com/thrisp/marid/lib/block"
	"github.com/thrisp/marid/lib/env"
	"github.com/thrisp/marid/lib/template"
)

type options struct {
	Directory, Package, Configurable, Letter string
}

var defaultOptions = options{".", "main", "Configurable", "c"}

func mkFlagSet(o *options) *env.FlagSet {
	ret := env.NewFlagSet("xrror", env.PanicOnError)
	ret.StringVar(&o.Directory, "Directory", o.Directory, "")
	ret.StringVar(&o.Package, "Package", o.Package, "")
	ret.StringVar(&o.Configurable, "Configurable", o.Configurable, "")
	ret.StringVar(&o.Letter, "Letter", strings.ToLower(string(o.Configurable[0])), "")
	return ret
}

var cl template.Loader = template.MapLoader(cm)

var cm map[string]string = map[string]string{
	"configuration": ct,
}

var e *env.Env

func configurationBlock() block.Block {
	o := &defaultOptions
	fs := mkFlagSet(o)
	b := block.New(
		block.Templater(template.New(10, nil, cl)),
	)
	b.Add(block.Command(
		"",
		"configuration",
		"generates a Configuration interface",
		50,
		func(c context.Context, v []string) env.ExitStatus {
			var bl string = "Block:Configuration"
			td := template.Data(b, bl)
			err := b.Render(o.Directory, "configuration", "configuration", td.Data)
			if err != nil {
				e.Printf("block render error: %s", err)
				return env.ExitFailure
			}
			e.Printf("rendered %s", bl)
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
	e.Register(configurationBlock())
}
