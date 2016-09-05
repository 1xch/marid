package flip

import (
	"context"

	"github.com/thrisp/marid/lib/block"
	"github.com/thrisp/marid/lib/env"
	"github.com/thrisp/marid/lib/template"
)

type options struct {
	Directory, Package, BackQuote string
}

var defaultOptions = options{".", "main", "`"}

func mkFlagSet(o *options) *env.FlagSet {
	ret := env.NewFlagSet("xrror", env.PanicOnError)
	ret.StringVar(&o.Directory, "Directory", o.Directory, "")
	ret.StringVar(&o.Package, "Package", o.Package, "")
	return ret
}

var fl template.Loader = template.MapLoader(fm)

var fm map[string]string = map[string]string{
	"flip": ft,
}

var e *env.Env

func flipBlock() block.Block {
	o := &defaultOptions
	fs := mkFlagSet(o)
	b := block.New(
		block.Templater(template.New(10, nil, fl)),
	)
	b.Add(block.Command(
		"",
		"flip",
		"generates an embeddable flag line parser for multiple commands",
		50,
		func(c context.Context, v []string) env.ExitStatus {
			tb := "Block:Flip"
			td := template.Data(b, tb)
			err := b.Render(o.Directory, "flip", "flip", td.Data)
			if err != nil {
				e.Printf("block render error: %s", err)
				return env.ExitFailure
			}
			e.Printf("rendered  %s", tb)
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
	e.Register(flipBlock())
}
