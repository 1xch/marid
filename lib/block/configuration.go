package block

import (
	"sort"

	"github.com/thrisp/marid/lib/env"
	"github.com/thrisp/marid/lib/template"
)

type ConfigFn func(*block) error

type Config interface {
	Order() int
	Configure(*block) error
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

func (c config) Configure(m *block) error {
	return c.fn(m)
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
	b          *block
	configured bool
	list       configList
}

func newConfiguration(b *block, conf ...Config) *configuration {
	c := &configuration{
		b:    b,
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

func configure(b *block, conf ...Config) error {
	for _, c := range conf {
		err := c.Configure(b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *configuration) Configure() error {
	sort.Sort(c.list)

	err := configure(c.b, c.list...)
	if err == nil {
		c.configured = true
	}

	return err
}

func (c *configuration) Configured() bool {
	return c.configured
}

var builtIns = []Config{}

func Command(group, tag, use string, priority int, efn env.ExecutionFunc, fs *env.FlagSet) Config {
	return DefaultConfig(func(b *block) error {
		b.Command = env.NewCommand(group, tag, use, priority, efn, fs)
		return nil
	})
}

func Templater(t template.Templater) Config {
	return DefaultConfig(func(b *block) error {
		b.Templater = t
		return nil
	})
}
