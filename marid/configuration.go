package marid

import (
	"sort"
)

type ConfigFn func(*marid) error

type Config interface {
	Order() int
	Configure(*marid) error
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

func (c config) Configure(m *marid) error {
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
	m          *marid
	configured bool
	list       configList
}

func newConfiguration(m *marid, conf ...Config) *configuration {
	c := &configuration{
		m:    m,
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

func configure(m *marid, conf ...Config) error {
	for _, c := range conf {
		err := c.Configure(m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *configuration) Configure() error {
	sort.Sort(c.list)

	err := configure(c.m, c.list...)
	if err == nil {
		c.configured = true
	}

	return err
}

func (c *configuration) Configured() bool {
	return c.configured
}

var builtIns = []Config{
	config{1000, setBufferPool},
	config{1001, setLogger},
}

func setBufferPool(m *marid) error {
	if m.bufferPool == nil {
		m.bufferPool = newBufferPool(m.bufferPoolSize)
	}
	return nil
}

func setLogger(m *marid) error {
	if m.Logr == nil {
		m.Logr = newLogr(m.verbose)
	}
	return nil
}

func Verbose(is bool) Config {
	return DefaultConfig(func(m *marid) error {
		m.verbose = is
		return nil
	})
}

func Loaders(l ...Loader) Config {
	return DefaultConfig(func(m *marid) error {
		m.loaders = append(m.loaders, l...)
		return nil
	})
}

func Blocks(b ...Block) Config {
	return DefaultConfig(func(m *marid) error {
		for _, bk := range b {
			m.blocks[bk.Tag()] = bk
			m.loaders = append(m.loaders, bk.Loader())
			m.addFuncs(bk.Funcs())
		}
		return nil
	})
}
