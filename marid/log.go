package marid

import (
	"log"
	"os"
)

type Logr interface {
	Fatalf(string, ...interface{})
	Panicf(string, ...interface{})
	Printf(string, ...interface{})
	PrintIf(string, ...interface{})
}

type logr struct {
	verbose bool
	*log.Logger
}

func newLogr(verbose bool) Logr {
	return &logr{
		verbose: verbose,
		Logger:  log.New(os.Stdout, "marid: ", log.Ldate|log.Lmicroseconds|log.LUTC),
	}
}

func (l *logr) PrintIf(format string, v ...interface{}) {
	if l.verbose {
		l.Printf(format, v...)
	}
}
