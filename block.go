package marid

import (
	"flag"
)

type BlockSet struct {
	b map[string]Block
}

func NewBlockSet() *BlockSet {
	return &BlockSet{make(map[string]Block)}
}

func (b *BlockSet) AddBlocks(bs ...Block) {
	for _, nb := range bs {
		b.b[nb.Tag()] = nb
	}
}

func (b *BlockSet) GetBlock(tag string) (Block, error) {
	if bl, ok := b.b[tag]; ok {
		return bl, nil
	}
	return nil, NoBlockError(tag)
}

func (b *BlockSet) GetBlocks() map[string]Block {
	return b.b
}

type Block interface {
	Tag() string
	Flags() *flag.FlagSet
	Loaders() []Loader
	Funcs() map[string]interface{}
	Templates() []string
	Directory() string
	Package() string
}

type block struct {
	tag       string
	flags     *flag.FlagSet
	loaders   []Loader
	funcs     map[string]interface{}
	templates []string
	directory string
	pckge     string
}

func NewBlock(t string,
	f *flag.FlagSet,
	lr Loader,
	fn map[string]interface{},
	tm []string,
	d string,
	p string) Block {
	return &block{t, f, []Loader{lr}, fn, tm, d, p}
}

func BasicBlock(t string, f *flag.FlagSet, l Loader, tm []string) Block {
	return NewBlock(t, f, l, nil, tm, ".", "main")
}

func (b *block) Tag() string {
	return b.tag
}

func (b *block) Flags() *flag.FlagSet {
	return b.flags
}

func (b *block) Loaders() []Loader {
	return b.loaders
}

func (b *block) Funcs() map[string]interface{} {
	return b.funcs
}

func (b *block) Templates() []string {
	return b.templates
}

func (b *block) Directory() string {
	return b.directory
}

func (b *block) Package() string {
	return b.pckge
}
