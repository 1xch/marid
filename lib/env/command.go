package env

import (
	"context"

	"github.com/Laughs-In-Flowers/ifriit/lib/flip"
)

type Options struct {
	LogFormatter string
}

var DefaultOptions = Options{"null"}

func MkFlagSet(fs *FlagSet, o *Options) {
	fs.StringVar(&o.LogFormatter, "logFormatter", o.LogFormatter, "Sets the environment logger formatter.")
}

type execution func(*Options)

func logLoading(o *Options) {
	if o.LogFormatter != "null" {
		switch o.LogFormatter {
		case "text", "stdout":
			Current.SwapFormatter(GetFormatter("text"))
		}
		Current.Printf("switching to log formatter: %s", o.LogFormatter)
	}
}

var executing = []execution{
	logLoading,
}

func Execute(o *Options, c context.Context, a []string) flip.ExitStatus {
	for _, fn := range executing {
		fn(o)
	}
	return flip.ExitNo
}
