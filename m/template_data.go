package m

import (
	"flag"

	. "github.com/thrisp/marid/b"
	. "github.com/thrisp/marid/f"
)

type TemplateData struct {
	Data map[string]interface{}
}

func NewTemplateData(b Block, f Flags) *TemplateData {
	ret := &TemplateData{
		Data: make(map[string]interface{}),
	}

	fn := func(fl *flag.Flag) {
		ret.Data[fl.Name] = fl.Value
	}
	f.VisitAll(fn)

	ret.Data["PackageName"] = b.Package()

	return ret
}
