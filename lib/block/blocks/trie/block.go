package trie

import (
	"context"

	"github.com/thrisp/marid/lib/block"
	"github.com/thrisp/marid/lib/env"
	"github.com/thrisp/marid/lib/template"
)

type options struct {
	Directory, Package, ItemName string
}

var defaultOptions = options{".", "main", "Item"}

func mkFlagSet(o *options) *env.FlagSet {
	ret := env.NewFlagSet("xrror", env.PanicOnError)
	ret.StringVar(&o.Directory, "Directory", o.Directory, "")
	ret.StringVar(&o.Package, "Package", o.Package, "")
	ret.StringVar(&o.ItemName, "ItemName", o.ItemName, "The name of the item the trie will hold.")
	return ret
}

var tl template.Loader = template.MapLoader(tm)

var tm map[string]string = map[string]string{
	"trie": tt,
}

var e *env.Env

func trieBlock() block.Block {
	o := &defaultOptions
	fs := mkFlagSet(o)
	b := block.New(
		block.Templater(template.New(10, nil, tl)),
	)
	b.Add(block.Command(
		"",
		"trie",
		"generates a trie structure holding a designated item",
		50,
		func(c context.Context, v []string) env.ExitStatus {
			tb := "Block:Trie"
			td := template.Data(b, tb)
			err := b.Render(o.Directory, "trie", "trie", td.Data)
			if err != nil {
				e.Printf("block render error: %s", err)
				return env.ExitFailure
			}
			e.Printf("rendered %s", tb)
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
	e.Register(trieBlock())
}
