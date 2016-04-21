package manager

import (
	f "flag"

	"github.com/thrisp/marid/block"
	"github.com/thrisp/marid/flag"
)

type TemplateData struct {
	Data map[string]interface{}
}

func NewTemplateData(b block.Block, fs flag.Flags) *TemplateData {
	ret := &TemplateData{
		Data: make(map[string]interface{}),
	}

	fn := func(fl *f.Flag) {
		ret.Data[fl.Name] = fl.Value
	}
	fs.VisitAll(fn)

	ret.Data["PackageName"] = b.Package()

	return ret
}
