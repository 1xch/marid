package b

import (
	. "github.com/thrisp/marid/f"
	. "github.com/thrisp/marid/l"
)

type BlockSet map[string]Block

func NewBlockSet() BlockSet {
	return make(BlockSet)
}

type Block interface {
	Tag() string
	Flags() Flags
	Loader() Loader
	Funcs() map[string]interface{}
	Templates() []string
	Directory() string
	Package() string
}

type block struct {
	tag       string
	flags     Flags
	loader    Loader
	funcs     map[string]interface{}
	templates []string
	directory string
	pckge     string
}

func NewBlock(t string,
	f Flags,
	lr Loader,
	fn map[string]interface{},
	tm []string,
	d string,
	p string) Block {
	return &block{t, f, lr, fn, tm, d, p}
}

func BasicBlock(t string, f Flags, l Loader, tm []string) Block {
	return NewBlock(t, f, l, nil, tm, ".", "main")
}

func (b *block) Tag() string {
	return b.tag
}

func (b *block) Flags() Flags {
	return b.flags
}

func (b *block) Loader() Loader {
	return b.loader
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
