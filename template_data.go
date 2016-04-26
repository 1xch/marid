package marid

import (
	"flag"
)

type TemplateData struct {
	Data map[string]interface{}
}

func NewTemplateData(b Block, fs *flag.FlagSet) *TemplateData {
	ret := &TemplateData{
		Data: make(map[string]interface{}),
	}

	fn := func(fl *flag.Flag) {
		ret.Data[fl.Name] = fl.Value
	}
	fs.VisitAll(fn)

	ret.Data["PackageName"] = b.Package()
	ret.Data["Block"] = b.Tag()

	return ret
}
