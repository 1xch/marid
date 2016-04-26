package main

import (
	"fmt"
	"os"

	"github.com/thrisp/marid"
	"github.com/thrisp/marid/blocks/configuration"
	"github.com/thrisp/marid/blocks/xrror"
)

var (
	blockArg      string
	blockArgs     []string
	version       bool
	verbose       bool
	defaultBlocks []marid.Block = []marid.Block{
		xrror.Block,
		configuration.Block,
	}
)

func index(s []string) map[int]string {
	indexd := make(map[int]string)
	for i, in := range s {
		indexd[i] = in
	}
	return indexd
}

func cull(s []string, id map[int]string) map[int]string {
	var toDelete []int
	add := func(d int) { toDelete = append(toDelete, d) }
	for i, label := range s {
		switch label {
		case "-block", "-b":
			add(i)
			blockArg = id[i+1]
			add(i + 1)
		case "-version", "-v":
			add(i)
			fmt.Println(fmtVersion())
			os.Exit(0)
		case "-verbose", "-vv":
			add(i)
			verbose = true
		}
	}
	for _, d := range toDelete {
		delete(id, d)
	}
	return id
}

func list(si map[int]string) []string {
	var ret []string
	for _, v := range si {
		ret = append(ret, v)
	}
	return ret
}

func parse(s []string) {
	blockArgs = list(cull(s, index(s)))
}

func init() {
	parse(os.Args[1:])
}

func main() {
	if verbose {
		marid.DefaultLogr.PrintIf("starting...")
	}
	m := marid.New(marid.Verbose(verbose), marid.Blocks(defaultBlocks...))
	if err := m.Configure(); err != nil {
		m.Fatalf("configuration error: %s", err)
	}

	if blockArg != "" {
		if err := m.Do(blockArg, blockArgs); err != nil {
			m.Fatalf("do error: %s", err)
		}
		m.PrintIf("done.")
		os.Exit(0)
	}
	m.Fatalf("no block specified! exiting.")
}
