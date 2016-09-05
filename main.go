package main

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/thrisp/marid/lib/env"

	_ "github.com/thrisp/marid/lib/block/blocks/configuration"
	_ "github.com/thrisp/marid/lib/block/blocks/flip"
	_ "github.com/thrisp/marid/lib/block/blocks/log"
	_ "github.com/thrisp/marid/lib/block/blocks/trie"
	_ "github.com/thrisp/marid/lib/block/blocks/xrror"
)

const topUse = `Top level flag usage.`

func TopCommand() env.Command {
	fs := env.NewFlagSet("", env.ContinueOnError)

	eo := &env.DefaultOptions
	env.MkFlagSet(fs, eo)

	return env.NewCommand(
		"top",
		"marid",
		topUse,
		0,
		func(c context.Context, a []string) env.ExitStatus {
			env.Execute(eo, c, a)
			return env.ExitNo
		},
		fs,
	)
}

var (
	pkgVersion     *version
	versionPackage string = path.Base(os.Args[0])
	versionTag     string = "No Tag"
	versionHash    string = "No Hash"
	versionDate    string = "No Date"
	fullVersion    bool
)

type version struct {
	pkge, tag, hash, date string
}

func Version(p, t, h, d string) *version {
	return &version{p, t, h, d}
}

func (v *version) Default() string {
	return fmt.Sprintf("%s %s", v.pkge, v.tag)
}

func (v *version) Full() string {
	return fmt.Sprintf("%s %s(%s %s)", v.pkge, v.tag, v.hash, v.date)
}

func printVersion(c context.Context, a []string) env.ExitStatus {
	var p string
	if fullVersion {
		p = pkgVersion.Full()
	} else {
		p = pkgVersion.Default()
	}
	fmt.Println(p)
	return env.ExitSuccess
}

const versionUse = `Prints the package version and exits.`

func VersionCommand() env.Command {
	fs := env.NewFlagSet("version", env.PanicOnError)
	fs.BoolVar(&fullVersion, "full", false, "print full version information containing package name, tag, hash and date")
	return env.NewCommand("top", "version", versionUse, 1, printVersion, fs)
}

var (
	e *env.Env
)

func init() {
	pkgVersion = Version(versionPackage, versionTag, versionHash, versionDate)
	e = env.Current
	e.RegisterGroup("top", -1)
	e.Register(TopCommand(), VersionCommand())
}

func main() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "e", e)
	e.Execute(ctx, os.Args)
	os.Exit(0)
}
