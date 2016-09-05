package env

import (
	"os"
	"path"
)

type Env struct {
	*Commander
	Logger
}

func New() *Env {
	return &Env{
		Commander: Base,
		Logger:    NewLogger(os.Stdout, LInfo, DefaultNullFormatter()),
	}
}

var Current *Env

func init() {
	name := path.Base(os.Args[0])
	Base = NewCommander(name)
	localFormatters = make(map[string]Formatter)
	SetFormatter(
		"text",
		DefaultTextFormatter(name),
	)
	Current = New()
}
