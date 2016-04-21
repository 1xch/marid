package flag

import (
	f "flag"
)

type Flagset interface {
	Tag() string
	Parse([]string) error
	Parsed() bool
	VisitAll(func(*f.Flag))
}

type flagSet struct {
	tag string
	*f.FlagSet
}

func NewFlagset(tag string, set *f.FlagSet) Flagset {
	return &flagSet{tag, set}
}

func (f *flagSet) Tag() string {
	return f.tag
}
