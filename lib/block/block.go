package block

import (
	"github.com/thrisp/marid/lib/env"
	"github.com/thrisp/marid/lib/template"
)

type Block interface {
	env.Command
	template.Templater
	Configuration
}

type block struct {
	env.Command
	template.Templater
	Configuration
}

func New(cnf ...Config) Block {
	b := &block{}
	b.Configuration = newConfiguration(b, cnf...)
	return b
}

var e *env.Env

func init() {
	e = env.Current
}
