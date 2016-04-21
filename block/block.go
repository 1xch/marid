package block

import (
	"github.com/thrisp/marid/flag"
	"github.com/thrisp/marid/loader"
)

type BlockSet map[string]Block

func NewBlockSet() BlockSet {
	return make(BlockSet)
}

type Block interface {
	Tag() string
	Flags() flag.Flagset
	Loaders() []loader.Loader
	Funcs() map[string]interface{}
	Templates() []string
	Directory() string
	Package() string
}

type block struct {
	tag       string
	flags     flag.Flagset
	loaders   []loader.Loader
	funcs     map[string]interface{}
	templates []string
	directory string
	pckge     string
}

func NewBlock(t string,
	f flag.Flagset,
	lr loader.Loader,
	fn map[string]interface{},
	tm []string,
	d string,
	p string) Block {
	return &block{t, f, []loader.Loader{lr}, fn, tm, d, p}
}

func BasicBlock(t string, f flag.Flagset, l loader.Loader, tm []string) Block {
	return NewBlock(t, f, l, nil, tm, ".", "main")
}

func (b *block) Tag() string {
	return b.tag
}

func (b *block) Flags() flag.Flagset {
	return b.flags
}

func (b *block) Loaders() []loader.Loader {
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
