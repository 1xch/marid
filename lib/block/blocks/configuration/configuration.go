package configuration

var ct string = `{{ extends "block_base" }}
{{ define "block_root"}}package {{.PackageName}}

import (
	"sort"
)

type ConfigFn func(*{{.Configurable}}) error

type Config interface {
	Order() int
	Configure(*{{.Configurable}}) error
}

type config struct {
	order int
	fn    ConfigFn
}

func DefaultConfig(fn ConfigFn) Config {
	return config{50, fn}
}

func NewConfig(order int, fn ConfigFn) Config {
	return config{order, fn}
}

func (c config) Order() int {
	return c.order
}

func (c config) Configure({{.Letter}} *{{.Configurable}}) error {
	return c.fn({{.Letter}})
}

type configList []Config

func (c configList) Len() int {
	return len(c)
}

func (c configList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c configList) Less(i, j int) bool {
	return c[i].Order() < c[j].Order()
}

type Configuration interface {
	Add(...Config)
	AddFn(...ConfigFn)
	Configure() error
	Configured() bool
}

type configuration struct {
	{{.Letter}}          *{{.Configurable}}
	configured bool
	list       configList
}

func newConfiguration({{.Letter}} *{{.Configurable}}, conf ...Config) *configuration {
	c := &configuration{
		{{.Letter}}:    {{.Letter}},
		list: builtIns,
	}
	c.Add(conf...)
	return c
}

func (c *configuration) Add(conf ...Config) {
	c.list = append(c.list, conf...)
}

func (c *configuration) AddFn(fns ...ConfigFn) {
	for _, fn := range fns {
		c.list = append(c.list, DefaultConfig(fn))
	}
}

func configure({{.Letter}} *{{.Configurable}}, conf ...Config) error {
	for _, c := range conf {
		err := c.Configure({{.Letter}})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *configuration) Configure() error {
	sort.Sort(c.list)

	err := configure(c.{{.Letter}}, c.list...)
	if err == nil {
		c.configured = true
	}

	return err
}

func (c *configuration) Configured() bool {
	return c.configured
}

var builtIns = []Config{
	//config{int, function},
}
{{ end }}
`
