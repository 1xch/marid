package env

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"time"
)

type Commander struct {
	name        string
	groups      commandGroups
	output      io.Writer
	instruction func()
	q           []*pop
	Writer
}

type commandGroup struct {
	name     string
	priority int
	sortBy   string
	commands []Command
}

func NewCommandGroup(name string, priority int, cs ...Command) *commandGroup {
	return &commandGroup{name, priority, "", cs}
}

func RegisterGroup(name string, priority int) {
	Base.RegisterGroup(name, priority)
}

func (c *Commander) RegisterGroup(name string, priority int) {
	c.groups = append(c.groups, NewCommandGroup(name, priority))
}

func getGroup(name string) *commandGroup {
	return Base.getGroup(name)
}

func (c *Commander) getGroup(name string) *commandGroup {
	for _, g := range c.groups {
		if name == g.name {
			return g
		}
	}
	return nil
}

func (g *commandGroup) SortBy(s string) {
	g.sortBy = s
	sort.Sort(g)
}

func (g commandGroup) Len() int {
	return len(g.commands)
}

func (g commandGroup) Less(i, j int) bool {
	switch g.sortBy {
	case "alpha":
		return g.commands[i].Tag() < g.commands[j].Tag()
	default:
		return g.commands[i].Priority() > g.commands[j].Priority()
	}
	return false
}

func (g commandGroup) Swap(i, j int) {
	g.commands[i], g.commands[j] = g.commands[j], g.commands[i]
}

func groupUsage(w io.Writer, g *commandGroup) {
	g.SortBy("alpha")
	for _, cmd := range g.commands {
		cmd.Use()
	}
}

type commandGroups []*commandGroup

func (p commandGroups) Len() int {
	return len(p)
}

func (p commandGroups) Less(i, j int) bool {
	return p[i].priority < p[j].priority
}

func (p commandGroups) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func NewCommander(name string) *Commander {
	c := &Commander{
		name:   name,
		groups: make(commandGroups, 0),
		output: os.Stderr,
	}
	c.RegisterGroup("", 0)
	return c
}

func (c *Commander) Instruction() {
	if c.instruction == nil {
		defaultCommanderInstruction(c)
	} else {
		c.instruction()
	}
}

const instructionString = `%s [OPTIONS...] {COMMAND} ...

`

func defaultCommanderInstruction(c *Commander) {
	fmt.Fprintf(c.output, instructionString, c.name)

	sort.Sort(c.groups)
	for _, cg := range c.groups {
		groupUsage(c.output, cg)
	}
}

func (c *Commander) Out() io.Writer {
	return c.output
}

func (c *Commander) SetOut(w io.Writer) {
	c.output = w
}

func isCommand(c *Commander, s string) (Command, bool) {
	for _, group := range c.groups {
		for _, cmd := range group.commands {
			if s == cmd.Tag() {
				return cmd, true
			}
		}
	}
	return nil, false
}

func (c *Commander) Get(k string) Command {
	for _, group := range c.groups {
		for _, cmd := range group.commands {
			if k == cmd.Tag() {
				return cmd
			}
		}
	}
	return nil
}

func Register(cmd Command) {
	Base.Register(cmd)
}

func (c *Commander) Register(cmds ...Command) {
	for _, cmd := range cmds {
		g := c.getGroup(cmd.Group())
		g.commands = append(g.commands, cmd)
	}
}

type pop struct {
	start, stop int
	c           Command
	v           []string
}

type pops []*pop

func (p pops) Len() int           { return len(p) }
func (p pops) Less(i, j int) bool { return p[i].c.Priority() < p[j].c.Priority() }
func (p pops) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func queue(c *Commander, arguments []string) {
	var ps pops

	for i, v := range arguments {
		if cmd, exists := isCommand(c, v); exists {
			a := &pop{i, 0, cmd, nil}
			ps = append(ps, a)
		}
	}

	li := len(ps) - 1
	la := len(arguments)
	for i, v := range ps {
		if i+1 <= li {
			nx := ps[i+1]
			v.stop = nx.start
		} else {
			v.stop = la
		}
	}

	for _, p := range ps {
		p.v = arguments[p.start:p.stop]
		c.q = append(c.q, p)
	}

	sort.Sort(ps)
}

func execute(c *Commander, ctx context.Context, cmd Command, arguments []string) ExitStatus {
	err := cmd.Parse(arguments)
	if err != nil {
		return ExitUsageError
	}
	return cmd.Execute(ctx, arguments)
}

func (c *Commander) Execute(ctx context.Context, arguments []string) ExitStatus {
	if len(arguments) < 1 {
		c.Instruction()
		return ExitUsageError
	}

	queue(c, arguments)
	for _, p := range c.q {
		cmd := p.c
		args := p.v[1:]
		exit := execute(c, ctx, cmd, args)
		switch exit {
		case ExitSuccess:
			os.Exit(0)
			return exit
		case ExitFailure:
			os.Exit(-1)
			return exit
		case ExitUsageError:
			c.Instruction()
			return exit
		default:
			continue
		}
	}

	c.Instruction()
	return ExitUsageError
}

var Base *Commander

type Command interface {
	Group() string
	Tag() string
	Priority() int
	Use()
	Execute(context.Context, []string) ExitStatus
	Flagger
}

type command struct {
	group, tag, use string
	priority        int
	efn             ExecutionFunc
	*FlagSet
}

func NewCommand(group, tag, use string, priority int, efn ExecutionFunc, fs *FlagSet) Command {
	return &command{group, tag, use, priority, efn, fs}
}

func (c *command) Group() string {
	return c.group
}

func (c *command) Tag() string {
	return c.tag
}

func (c *command) Priority() int {
	return c.priority
}

func (c *command) useHead() string {
	return fmt.Sprintf("%s [<flags>]:\n", c.tag)
}

func (c *command) Use() {
	w := c.Out()
	fmt.Fprint(w, "-----\n")
	fmt.Fprintf(w, "%s %s\n", c.useHead(), c.use)
	fmt.Fprint(w, "\n")
	c.PrintDefaults()
	fmt.Fprint(w, "\n")
}

type ExecutionFunc func(context.Context, []string) ExitStatus

func (c *command) Execute(ctx context.Context, v []string) ExitStatus {
	if c.efn != nil {
		return c.efn(ctx, v)
	}
	return ExitFailure
}

type ExitStatus int

const (
	ExitSuccess ExitStatus = iota
	ExitFailure
	ExitUsageError
	ExitNo
)

type Flag struct {
	Name     string // name as it appears on command line
	Usage    string // help message
	Value    Value  // value as set
	DefValue string // default value (as text); for usage message
}

type Value interface {
	String() string
	Set(string) error
}

type Getter interface {
	Value
	Get() interface{}
}

type boolValue bool

func newBoolValue(val bool, p *bool) *boolValue {
	*p = val
	return (*boolValue)(p)
}

func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	*b = boolValue(v)
	return err
}

func (b *boolValue) Get() interface{} { return bool(*b) }

func (b *boolValue) String() string { return fmt.Sprintf("%v", *b) }

func (b *boolValue) IsBoolFlag() bool { return true }

type boolFlag interface {
	Value
	IsBoolFlag() bool
}

type intValue int

func newIntValue(val int, p *int) *intValue {
	*p = val
	return (*intValue)(p)
}

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = intValue(v)
	return err
}

func (i *intValue) Get() interface{} { return int(*i) }

func (i *intValue) String() string { return fmt.Sprintf("%v", *i) }

type int64Value int64

func newInt64Value(val int64, p *int64) *int64Value {
	*p = val
	return (*int64Value)(p)
}

func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = int64Value(v)
	return err
}

func (i *int64Value) Get() interface{} { return int64(*i) }

func (i *int64Value) String() string { return fmt.Sprintf("%v", *i) }

type uintValue uint

func newUintValue(val uint, p *uint) *uintValue {
	*p = val
	return (*uintValue)(p)
}

func (i *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	*i = uintValue(v)
	return err
}

func (i *uintValue) Get() interface{} { return uint(*i) }

func (i *uintValue) String() string { return fmt.Sprintf("%v", *i) }

type uint64Value uint64

func newUint64Value(val uint64, p *uint64) *uint64Value {
	*p = val
	return (*uint64Value)(p)
}

func (i *uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	*i = uint64Value(v)
	return err
}

func (i *uint64Value) Get() interface{} { return uint64(*i) }

func (i *uint64Value) String() string { return fmt.Sprintf("%v", *i) }

type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Get() interface{} { return string(*s) }

func (s *stringValue) String() string { return fmt.Sprintf("%s", *s) }

type float64Value float64

func newFloat64Value(val float64, p *float64) *float64Value {
	*p = val
	return (*float64Value)(p)
}

func (f *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	*f = float64Value(v)
	return err
}

func (f *float64Value) Get() interface{} { return float64(*f) }

func (f *float64Value) String() string { return fmt.Sprintf("%v", *f) }

type durationValue time.Duration

func newDurationValue(val time.Duration, p *time.Duration) *durationValue {
	*p = val
	return (*durationValue)(p)
}

func (d *durationValue) Set(s string) error {
	v, err := time.ParseDuration(s)
	*d = durationValue(v)
	return err
}

func (d *durationValue) Get() interface{} { return time.Duration(*d) }

func (d *durationValue) String() string { return (*time.Duration)(d).String() }

type ErrorHandling int

const (
	ContinueOnError ErrorHandling = iota
	ExitOnError
	PanicOnError
)

type FlagSet struct {
	usage         func()
	name          string
	parsed        bool
	actual        map[string]*Flag
	formal        map[string]*Flag
	args          []string
	errorHandling ErrorHandling
	output        io.Writer
}

func NewFlagSet(name string, errorHandling ErrorHandling) *FlagSet {
	return &FlagSet{
		name:          name,
		errorHandling: errorHandling,
	}
}

type Flagger interface {
	PrintDefaults()
	Usage()
	Parse([]string) error
	Parsed() bool
	FlagSetter
}

type FlagSetter interface {
	Lookup(string) *Flag
	Set(string, string) error
	Visit(func(*Flag))
	VisitAll(func(*Flag))
	FlagVars
}

type Writer interface {
	Out() io.Writer
	SetOut(io.Writer)
}

type FlagVars interface {
	Var(Value, string, string)
}

func sortFlags(flags map[string]*Flag) []*Flag {
	list := make(sort.StringSlice, len(flags))
	i := 0
	for _, f := range flags {
		list[i] = f.Name
		i++
	}
	list.Sort()
	result := make([]*Flag, len(list))
	for i, name := range list {
		result[i] = flags[name]
	}
	return result
}

func (f *FlagSet) Out() io.Writer {
	if f.output == nil {
		return os.Stderr
	}
	return f.output
}

func (f *FlagSet) SetOut(output io.Writer) {
	f.output = output
}

func (f *FlagSet) VisitAll(fn func(*Flag)) {
	for _, flag := range sortFlags(f.formal) {
		fn(flag)
	}
}

func (f *FlagSet) Visit(fn func(*Flag)) {
	for _, flag := range sortFlags(f.actual) {
		fn(flag)
	}
}

func (f *FlagSet) Lookup(name string) *Flag {
	if fl, ok := f.formal[name]; ok {
		return fl
	}
	return nil
}

func (f *FlagSet) Set(name, value string) error {
	flag, ok := f.formal[name]
	if !ok {
		return fmt.Errorf("no such flag -%v", name)
	}
	err := flag.Value.Set(value)
	if err != nil {
		return err
	}
	if f.actual == nil {
		f.actual = make(map[string]*Flag)
	}
	f.actual[name] = flag
	return nil
}

func isZeroValue(value string) bool {
	switch value {
	case "false":
		return true
	case "":
		return true
	case "0":
		return true
	}
	return false
}

func UnquoteUsage(flag *Flag) (name string, usage string) {
	// Look for a back-quoted name, but avoid the strings package.
	usage = flag.Usage
	for i := 0; i < len(usage); i++ {
		if usage[i] == '`' {
			for j := i + 1; j < len(usage); j++ {
				if usage[j] == '`' {
					name = usage[i+1 : j]
					usage = usage[:i] + name + usage[j+1:]
					return name, usage
				}
			}
			break // Only one back quote; use type name.
		}
	}
	// No explicit name, so use type if we can find one.
	name = "value"
	switch flag.Value.(type) {
	case boolFlag:
		name = ""
	case *durationValue:
		name = "duration"
	case *float64Value:
		name = "float"
	case *intValue, *int64Value:
		name = "int"
	case *stringValue:
		name = "string"
	case *uintValue, *uint64Value:
		name = "uint"
	}
	return
}

func (f *FlagSet) PrintDefaults() {
	f.VisitAll(func(flag *Flag) {
		s := fmt.Sprintf("  -%s", flag.Name) // Two spaces before -; see next two comments.
		name, usage := UnquoteUsage(flag)
		if len(name) > 0 {
			s += " " + name
		}
		// Boolean flags of one ASCII letter are so common we
		// treat them specially, putting their usage on the same line.
		if len(s) <= 4 { // space, space, '-', 'x'.
			s += "\t"
		} else {
			// Four spaces before the tab triggers good alignment
			// for both 4- and 8-space tab stops.
			s += "\n    \t"
		}
		s += usage
		if !isZeroValue(flag.DefValue) {
			if _, ok := flag.Value.(*stringValue); ok {
				// put quotes on the value
				s += fmt.Sprintf(" (default %q)", flag.DefValue)
			} else {
				s += fmt.Sprintf(" (default %v)", flag.DefValue)
			}
		}
		fmt.Fprint(f.Out(), s, "\n")
	})
}

func defaultUsage(f *FlagSet) {
	if f.name == "" {
		fmt.Fprintf(f.Out(), "Usage:\n")
	} else {
		fmt.Fprintf(f.Out(), "Usage of %s:\n", f.name)
	}
	f.PrintDefaults()
}

func (f *FlagSet) NFlag() int { return len(f.actual) }

func (f *FlagSet) Arg(i int) string {
	if i < 0 || i >= len(f.args) {
		return ""
	}
	return f.args[i]
}

func (f *FlagSet) NArg() int { return len(f.args) }

func (f *FlagSet) Args() []string { return f.args }

func (f *FlagSet) BoolVar(p *bool, name string, value bool, usage string) {
	f.Var(newBoolValue(value, p), name, usage)
}

func (f *FlagSet) Bool(name string, value bool, usage string) *bool {
	p := new(bool)
	f.BoolVar(p, name, value, usage)
	return p
}

func (f *FlagSet) IntVar(p *int, name string, value int, usage string) {
	f.Var(newIntValue(value, p), name, usage)
}

func (f *FlagSet) Int(name string, value int, usage string) *int {
	p := new(int)
	f.IntVar(p, name, value, usage)
	return p
}

func (f *FlagSet) Int64Var(p *int64, name string, value int64, usage string) {
	f.Var(newInt64Value(value, p), name, usage)
}

func (f *FlagSet) Int64(name string, value int64, usage string) *int64 {
	p := new(int64)
	f.Int64Var(p, name, value, usage)
	return p
}

func (f *FlagSet) UintVar(p *uint, name string, value uint, usage string) {
	f.Var(newUintValue(value, p), name, usage)
}

func (f *FlagSet) Uint(name string, value uint, usage string) *uint {
	p := new(uint)
	f.UintVar(p, name, value, usage)
	return p
}

func (f *FlagSet) Uint64Var(p *uint64, name string, value uint64, usage string) {
	f.Var(newUint64Value(value, p), name, usage)
}

func (f *FlagSet) Uint64(name string, value uint64, usage string) *uint64 {
	p := new(uint64)
	f.Uint64Var(p, name, value, usage)
	return p
}

func (f *FlagSet) StringVar(p *string, name string, value string, usage string) {
	f.Var(newStringValue(value, p), name, usage)
}

func (f *FlagSet) String(name string, value string, usage string) *string {
	p := new(string)
	f.StringVar(p, name, value, usage)
	return p
}

func (f *FlagSet) Float64Var(p *float64, name string, value float64, usage string) {
	f.Var(newFloat64Value(value, p), name, usage)
}

func (f *FlagSet) Float64(name string, value float64, usage string) *float64 {
	p := new(float64)
	f.Float64Var(p, name, value, usage)
	return p
}

func (f *FlagSet) DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	f.Var(newDurationValue(value, p), name, usage)
}

func (f *FlagSet) Duration(name string, value time.Duration, usage string) *time.Duration {
	p := new(time.Duration)
	f.DurationVar(p, name, value, usage)
	return p
}

func (f *FlagSet) Var(value Value, name string, usage string) {
	// Remember the default value as a string; it won't change.
	flag := &Flag{name, usage, value, value.String()}
	_, alreadythere := f.formal[name]
	if alreadythere {
		var msg string
		if f.name == "" {
			msg = fmt.Sprintf("flag redefined: %s", name)
		} else {
			msg = fmt.Sprintf("%s flag redefined: %s", f.name, name)
		}
		fmt.Fprintln(f.Out(), msg)
		panic(msg) // Happens only if flags are declared with identical names
	}
	if f.formal == nil {
		f.formal = make(map[string]*Flag)
	}
	f.formal[name] = flag
}

func (f *FlagSet) failOnly(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	fmt.Fprintln(f.Out(), err)
	return err
}

func (f *FlagSet) failf(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	fmt.Fprintln(f.Out(), err)
	f.Usage()
	return err
}

func (f *FlagSet) Usage() {
	if f.usage == nil {
		defaultUsage(f)
	} else {
		f.usage()
	}
}

func (f *FlagSet) parseOne() (bool, error) {
	if len(f.args) == 0 {
		return false, nil
	}
	s := f.args[0]
	if len(s) == 0 || s[0] != '-' || len(s) == 1 {
		return false, nil
	}
	numMinuses := 1
	if s[1] == '-' {
		numMinuses++
		if len(s) == 2 { // "--" terminates the flags
			f.args = f.args[1:]
			return false, nil
		}
	}
	name := s[numMinuses:]
	if len(name) == 0 || name[0] == '-' || name[0] == '=' {
		return false, f.failf("bad flag syntax: %s", s)
	}

	// it's a flag. does it have an argument?
	f.args = f.args[1:]
	hasValue := false
	value := ""
	for i := 1; i < len(name); i++ { // equals cannot be first
		if name[i] == '=' {
			value = name[i+1:]
			hasValue = true
			name = name[0:i]
			break
		}
	}

	m := f.formal
	var flag *Flag
	var exists bool
	flag, exists = m[name]
	if !exists {
		if name == "help" || name == "h" { // special case for nice help message.
			f.usage()
			return false, ErrHelp
		}
		return false, f.failOnly("flag provided but not defined: -%s\n", name)
	}

	if fv, ok := flag.Value.(boolFlag); ok && fv.IsBoolFlag() { // special case: doesn't need an arg
		if hasValue {
			if err := fv.Set(value); err != nil {
				return false, f.failf("invalid boolean value %q for -%s: %v", value, name, err)
			}
		} else {
			if err := fv.Set("true"); err != nil {
				return false, f.failf("invalid boolean flag %s: %v", name, err)
			}
		}
	} else {
		// It must have a value, which might be the next argument.
		if !hasValue && len(f.args) > 0 {
			// value is the next arg
			hasValue = true
			value, f.args = f.args[0], f.args[1:]
		}
		if !hasValue {
			return false, f.failf("flag needs an argument: -%s", name)
		}
		if err := flag.Value.Set(value); err != nil {
			return false, f.failf("invalid value %q for flag -%s: %v", value, name, err)
		}
	}
	if f.actual == nil {
		f.actual = make(map[string]*Flag)
	}
	f.actual[name] = flag
	return true, nil
}

func (f *FlagSet) Parse(arguments []string) error {
	f.parsed = true
	f.args = arguments
	for {
		seen, err := f.parseOne()
		if seen {
			continue
		}
		if err == nil {
			break
		}
		switch f.errorHandling {
		case ContinueOnError:
			return err
		case ExitOnError:
			os.Exit(2)
		case PanicOnError:
			panic(err)
		}
	}
	return nil
}

func (f *FlagSet) Parsed() bool {
	return f.parsed
}

func (f *FlagSet) Init(name string, errorHandling ErrorHandling) {
	f.name = name
	f.errorHandling = errorHandling
}

var ErrHelp = errors.New("flag: help requested")
