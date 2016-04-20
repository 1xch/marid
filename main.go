package main

import (
	"fmt"
	"os"

	"github.com/thrisp/marid/b"
	"github.com/thrisp/marid/b/configuration"
	"github.com/thrisp/marid/b/xrror"
	"github.com/thrisp/marid/m"
)

var (
	block         string
	blockArgs     []string
	version       bool
	verbose       bool
	defaultBlocks []b.Block = []b.Block{
		xrror.Xrror,
		configuration.Configuration,
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
			block = id[i+1]
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
	marid := m.New(m.Verbose(verbose), m.Blocks(defaultBlocks...))
	if err := marid.Configure(); err != nil {
		marid.Fatalf("Marid configuration error: %s", err)
	}
	marid.PrintIf("configured -- %t", marid.Configured())
	if block != "" {
		var flags []string
		flags = append(flags, blockArgs...)
		if err := marid.Do(block, flags); err != nil {
			marid.Fatalf("Marid do error: %s", err)
		}
		marid.PrintIf("exiting...")
		os.Exit(0)
	}
	marid.Fatalf("no block specified! exiting.")
}
