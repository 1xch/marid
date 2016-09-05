package log

var lt string = `{{ extends "block_base" }}
{{ define "block_root" }}package {{.NamePkg}} 

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

type StdLogger interface {
	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Fatalln(...interface{})
	Panic(...interface{})
	Panicf(string, ...interface{})
	Panicln(...interface{})
	Print(...interface{})
	Printf(string, ...interface{})
	Println(...interface{})
}

type Mutex interface {
	Lock()
	Unlock()
}

type Logger interface {
	io.Writer
	StdLogger
	Log(Level, Entry)
	Mutex
	Level() Level
	Formatter
	FormatterSwapper
	Hooks
}

type logger struct {
	io.Writer
	level Level
	Formatter
	Hooks
	sync.Mutex
}

func NewLogger(w io.Writer, l Level, f Formatter) Logger {
	return &logger{
		Writer:    w,
		level:     l,
		Formatter: f,
		Hooks:     &hooks{},
	}
}

func (l *logger) Level() Level {
	return l.level
}

func (l *logger) Log(lv Level, e Entry) {
	log(lv, e)
}

func log(lv Level, e Entry) {
	if err := e.Fire(lv, e); err != nil {
		e.Lock()
		fmt.Fprintf(os.Stderr, "log: Failed to fire hook -- %v\n", err)
		e.Unlock()
	}

	reader, err := e.Read()
	if err != nil {
		e.Lock()
		fmt.Fprintf(os.Stderr, "log: Failed to obtain reader -- %v\n", err)
		e.Unlock()
	}

	e.Lock()
	defer e.Unlock()

	_, err = io.Copy(e, reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "log: Failed to write -- %v\n", err)
	}

	if lv == LFatal {
		os.Exit(1)
	}

	if lv <= LPanic {
		panic(&e)
	}
}

func (l *logger) Fatal(v ...interface{}) {
	if l.level >= LFatal {
		log(LFatal, newEntry(l, mkFields(0, v...)...))
		os.Exit(1)
	}
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	if l.level >= LFatal {
		log(LFatal, newEntry(l, mkFormatFields(format, v...)...))
		os.Exit(1)
	}
}

func (l *logger) Fatalln(v ...interface{}) {
	if l.level >= LFatal {
		log(LFatal, newEntry(l, mkFields(0, v...)...))
		os.Exit(1)
	}
}

func (l *logger) Panic(v ...interface{}) {
	if l.level >= LPanic {
		log(LPanic, newEntry(l, mkFields(0, v...)...))
	}
	panic(fmt.Sprint(v...))
}

func (l *logger) Panicf(format string, v ...interface{}) {
	if l.level >= LPanic {
		log(LPanic, newEntry(l, mkFormatFields(format, v...)...))
	}
	panic(fmt.Sprintf(format, v...))
}

func (l *logger) Panicln(v ...interface{}) {
	l.Panic(v...)
}

func (l *logger) Print(v ...interface{}) {
	if l.level >= LError {
		log(LInfo, newEntry(l, mkFields(0, v...)...))
	}
}

func (l *logger) Printf(format string, v ...interface{}) {
	if l.level >= LError {
		log(LInfo, newEntry(l, mkFormatFields(format, v...)...))
	}
}

func (l *logger) Println(v ...interface{}) {
	if l.level >= LError {
		log(LInfo, newEntry(l, mkFields(0, v...)...))
	}
}

func (l *logger) SwapFormatter(f Formatter) {
	if f != nil {
		l.Lock()
		l.Formatter = f
		l.Unlock()
		return
	}
	l.Fatalf("Formatter must not be nil.")
}

type Entry interface {
	Logger
	Fielder
	Reader
	Created() time.Time
}

type Fielder interface {
	Fields() []Field
}

type Field struct {
	Order int
	Key   string
	Value interface{}
}

func mkFields(index int, v ...interface{}) []Field {
	var ret []Field
	for i, vv := range v {
		idx := i + index
		ret = append(ret, Field{idx, fmt.Sprintf("Field%d", idx), vv})
	}
	return ret
}

func mkFormatFields(format string, v ...interface{}) []Field {
	var ret []Field
	ret = append(ret, Field{1, "Format", format})
	ret = append(ret, mkFields(1, v...)...)
	return ret
}

type FieldsSort []Field

func (f FieldsSort) Len() int {
	return len(f)
}

func (f FieldsSort) Less(i, j int) bool {
	return f[i].Order < f[j].Order
}

func (f FieldsSort) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

type Reader interface {
	Read() (*bytes.Buffer, error)
}

type entry struct {
	created time.Time
	Reader
	Logger
	fields []Field
}

func newEntry(l Logger, f ...Field) Entry {
	return &entry{
		created: time.Now(),
		Logger:  l,
		fields:  f,
	}
}

func (e *entry) Read() (*bytes.Buffer, error) {
	s, err := e.Format(e)
	return bytes.NewBuffer(s), err
}

func (e *entry) Fields() []Field {
	return e.fields
}

func (e *entry) Created() time.Time {
	return e.created
}

type Level int

const (
	LUnrecognized Level = iota
	LPanic
	LFatal
	LError
	LWarn
	LInfo
	LDebug
)

var stringToLevel = map[string]Level{
	"panic": LPanic,
	"fatal": LFatal,
	"error": LError,
	"warn":  LWarn,
	"info":  LInfo,
	"debug": LDebug,
}

func StringToLevel(lv string) Level {
	if level, ok := stringToLevel[strings.ToLower(lv)]; ok {
		return level
	}
	return LUnrecognized
}

func (l Level) String() string {
	switch l {
	case LPanic:
		return "panic"
	case LFatal:
		return "fatal"
	case LError:
		return "error"
	case LWarn:
		return "warn"
	case LInfo:
		return "info"
	case LDebug:
		return "debug"
	}
	return "unrecognized"
}

func (lv Level) Color() func(io.Writer, ...interface{}) {
	switch lv {
	case LPanic:
		return red
	case LFatal:
		return magenta
	case LError:
		return cyan
	case LWarn:
		return yellow
	case LInfo:
		return green
	case LDebug:
		return blue
	}
	return white
}

type Formatter interface {
	Format(Entry) ([]byte, error)
}

type FormatterSwapper interface {
	SwapFormatter(Formatter)
}

type formatters map[string]Formatter

var hasFormatters formatters

func SetFormatter(k string, f Formatter) {
	hasFormatters[k] = f
}

func GetFormatter(k string) Formatter {
	if f, ok := hasFormatters[k]; ok {
		return f
	}
	return &NullFormatter{}
}

type NullFormatter struct{}

func DefaultNullFormatter() Formatter {
	return &NullFormatter{}
}

func (n *NullFormatter) Format(e Entry) ([]byte, error) {
	return nil, nil
}

type TextFormatter struct {
	Name            string
	TimestampFormat string
	Sort            bool
}

func DefaultTextFormatter(name string) Formatter {
	return &TextFormatter{
		name,
		time.StampNano,
		false,
	}
}

func (t *TextFormatter) Format(e Entry) ([]byte, error) {
	fs := e.Fields()
	var keys []string = make([]string, 0, len(fs))
	for _, k := range fs {
		keys = append(keys, k.Key)
	}

	if t.Sort {
		sort.Strings(keys)
	}

	timestampFormat := t.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.StampNano
	}

	b := &bytes.Buffer{}

	t.formatFields(b, e, keys, timestampFormat)

	b.WriteByte('\n')

	return b.Bytes(), nil
}

func (t *TextFormatter) formatFields(b *bytes.Buffer, e Entry, keys []string, timestampFormat string) {
	lvl := e.Level()
	lvlColor := lvl.Color()
	lvlText := strings.ToUpper(lvl.String())
	lvlColor(b, fmt.Sprintf("%s ", lvlText))

	black(b, fmt.Sprintf("[%s] ", t.Name))

	timestamp := time.Now().Format(timestampFormat)
	blue(b, fmt.Sprintf("%s ", timestamp))

	fds := FieldsSort(e.Fields())
	format(b, fds)
}

func formatTo(fds []Field) (bool, string, []interface{}) {
	var formattable bool
	var f string
	var ff []interface{}
	for _, fd := range fds {
		if fd.Key == "Format" {
			formattable = true
			f = fd.Value.(string)
		} else {
			ff = append(ff, fd.Value)
		}
	}
	return formattable, f, ff
}

func format(b *bytes.Buffer, fds FieldsSort) {
	sort.Sort(fds)
	formattable, f, ff := formatTo(fds)
	if formattable {
		fmt.Fprintf(b, f, ff...)
	} else {
		for _, v := range ff {
			fmt.Fprintf(b, "%s", v)
		}
	}
}

type color struct {
	params []Attribute
}

// Should work in most terminals.
// See github.com/mattn/go-colorable for tweaking tips by os.
func Color(value ...Attribute) func(io.Writer, ...interface{}) {
	c := &color{params: make([]Attribute, 0)}
	c.Add(value...)
	return c.Fprint
}

func (c *color) Add(value ...Attribute) *color {
	c.params = append(c.params, value...)
	return c
}

func (c *color) Fprint(w io.Writer, a ...interface{}) {
	c.wrap(w, a...)
}

func (c *color) Fprintf(w io.Writer, f string, a ...interface{}) {
	c.wrap(w, fmt.Sprintf(f, a...))
}

func (c *color) sequence() string {
	format := make([]string, len(c.params))
	for i, v := range c.params {
		format[i] = strconv.Itoa(int(v))
	}

	return strings.Join(format, ";")
}

func (c *color) wrap(w io.Writer, a ...interface{}) {
	if c.noColor() {
		fmt.Fprint(w, a...)
	}

	c.format(w)
	fmt.Fprint(w, a...)
	c.unformat(w)
}

func (c *color) format(w io.Writer) {
	fmt.Fprintf(w, "%s[%sm", escape, c.sequence())
}

func (c *color) unformat(w io.Writer) {
	fmt.Fprintf(w, "%s[%dm", escape, Reset)
}

var NoColor = !IsTerminal(os.Stdout.Fd())

const ioctlReadTermios = syscall.TCGETS

// IsTerminal return true if the file descriptor is terminal.
// see github.com/mattn/go-isatty
// You WILL want to change this if you are using an os other than a Linux variant.
func IsTerminal(fd uintptr) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, ioctlReadTermios, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}

func (c *color) noColor() bool {
	return NoColor
}

const escape = "\x1b"

type Attribute int

const (
	Reset Attribute = iota
	Bold
	Faint
	Italic
	Underline
	BlinkSlow
	BlinkRapid
	ReverseVideo
	Concealed
	CrossedOut
)

const (
	FgBlack Attribute = iota + 30
	FgRed
	FgGreen
	FgYellow
	FgBlue
	FgMagenta
	FgCyan
	FgWhite
)

const (
	FgHiBlack Attribute = iota + 90
	FgHiRed
	FgHiGreen
	FgHiYellow
	FgHiBlue
	FgHiMagenta
	FgHiCyan
	FgHiWhite
)

const (
	BgBlack Attribute = iota + 40
	BgRed
	BgGreen
	BgYellow
	BgBlue
	BgMagenta
	BgCyan
	BgWhite
)

const (
	BgHiBlack Attribute = iota + 100
	BgHiRed
	BgHiGreen
	BgHiYellow
	BgHiBlue
	BgHiMagenta
	BgHiCyan
	BgHiWhite
)

var (
	black   = Color(FgHiBlack)
	red     = Color(FgHiRed)
	green   = Color(FgHiGreen)
	yellow  = Color(FgHiYellow)
	blue    = Color(FgHiBlue)
	magenta = Color(FgHiMagenta)
	cyan    = Color(FgHiCyan)
	white   = Color(FgHiWhite)
)

type Hook interface {
	On() []Level
	Fire(Entry) error
}

type Hooks interface {
	AddHook(Hook)
	Fire(Level, Entry) error
}

type hooks struct {
	has map[Level][]Hook
}

func (h *hooks) AddHook(hk Hook) {
	for _, lv := range hk.On() {
		h.has[lv] = append(h.has[lv], hk)
	}
}

func (h *hooks) Fire(lv Level, e Entry) error {
	for _, hook := range h.has[lv] {
		if err := hook.Fire(e); err != nil {
			return err
		}
	}
	return nil
}
{{ end }}
`
