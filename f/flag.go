package f

import "flag"

type FlagSet map[string]Flags

func NewFlagSet() FlagSet {
	return make(FlagSet)
}

type Flags interface {
	Tag() string
	Parse([]string) error
	Parsed() bool
	VisitAll(func(*flag.Flag))
}

type flags struct {
	tag string
	*flag.FlagSet
}

func NewFlag(tag string, set *flag.FlagSet) Flags {
	return &flags{tag, set}
}

func (f *flags) Tag() string {
	return f.tag
}
