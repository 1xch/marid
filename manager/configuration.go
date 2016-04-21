package manager

import (
	"sort"

	"github.com/thrisp/marid/block"
	"github.com/thrisp/marid/loader"
)

type ConfigFn func(*manager) error

type Config interface {
	Order() int
	Configure(*manager) error
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

func (c config) Configure(m *manager) error {
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
	m          *manager
	configured bool
	list       configList
}

func newConfiguration(m *manager, conf ...Config) *configuration {
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

func configure(m *manager, conf ...Config) error {
	for _, c := range conf {
		err := c.Configure(m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *configuration) Configure() error {
	DefaultLogr.PrintIf("configuring...")
	sort.Sort(c.list)

	err := configure(c.m, c.list...)
	if err == nil {
		c.configured = true
		DefaultLogr.PrintIf("configured")
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

func setBufferPool(m *manager) error {
	if m.bufferPool == nil {
		m.bufferPool = newBufferPool(m.bufferPoolSize)
	}
	return nil
}

func setLogger(m *manager) error {
	if m.Logr == nil {
		m.Logr = newLogr(m.verbose)
	}
	return nil
}

func Verbose(is bool) Config {
	return DefaultConfig(func(m *manager) error {
		m.verbose = is
		return nil
	})
}

func Loaders(l ...loader.Loader) Config {
	return DefaultConfig(func(m *manager) error {
		m.loaders = append(m.loaders, l...)
		return nil
	})
}

func Blocks(b ...block.Block) Config {
	return DefaultConfig(func(m *manager) error {
		for _, bk := range b {
			m.blocks[bk.Tag()] = bk
			m.loaders = append(m.loaders, bk.Loaders()...)
			m.addFuncs(bk.Funcs())
		}
		return nil
	})
}
