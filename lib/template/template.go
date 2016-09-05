package template

import (
	"fmt"
	"go/format"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/thrisp/marid/lib/env"
)

type Templater interface {
	Fetch(string) (*template.Template, error)
	Render(string, string, string, interface{}) error //dir, file, template, data
	FuncSet
	Loaders
}

type templater struct {
	*bufferPool
	FuncSet
	Loaders
}

func New(b int, f FuncSet, l ...Loader) Templater {
	bp := newBufferPool(b)
	fs := newFuncSet()
	if f != nil {
		fs.SetFuncs(f)
	}
	ls := NewLoaders()
	l = append(l, baseLoader)
	ls.SetLoaders(l...)
	return &templater{
		bufferPool: bp,
		FuncSet:    fs,
		Loaders:    ls,
	}
}

func (t *templater) Fetch(tpl string) (*template.Template, error) {
	return t.assemble(tpl)
}

func (t *templater) Render(dir, file, tpl string, data interface{}) error {
	var err error
	var tmpl *template.Template

	tmpl, err = t.assemble(tpl)
	if err != nil {
		return err
	}

	b := t.get()

	err = tmpl.Execute(b, data)
	if err != nil {
		return RenderError(err)
	}

	var src []byte
	src, err = format.Source(b.Bytes())
	if err != nil {
		return InvalidGoCodeError(err)
	}

	t.put(b)

	output := strings.ToLower(fmt.Sprintf("%s.go", file))
	outputPath := filepath.Join(dir, output)
	if wErr := ioutil.WriteFile(outputPath, src, 0644); wErr != nil {
		return RenderError(wErr)
	}
	return err
}

type Node struct {
	Name string
	Src  string
}

var (
	reExtendsTag  *regexp.Regexp = regexp.MustCompile("{{ extends [\"']?([^'\"}']*)[\"']? }}")
	reIncludeTag  *regexp.Regexp = regexp.MustCompile(`{{ include ["']?([^"]*)["']? }}`)
	reDefineTag   *regexp.Regexp = regexp.MustCompile("{{ ?define \"([^\"]*)\" ?\"?([a-zA-Z0-9]*)?\"? ?}}")
	reTemplateTag *regexp.Regexp = regexp.MustCompile("{{ ?template \"([^\"]*)\" ?([^ ]*)? ?}}")
)

func (t *templater) assemble(tpl string) (*template.Template, error) {
	stack := []*Node{}

	err := t.add(&stack, tpl)

	if err != nil {
		return nil, err
	}

	blocks := map[string]string{}
	blockId := 0

	for _, node := range stack {
		var errInReplace error = nil
		node.Src = reIncludeTag.ReplaceAllStringFunc(node.Src, func(raw string) string {
			parsed := reIncludeTag.FindStringSubmatch(raw)
			templatePath := parsed[1]
			subTpl, err := getTemplate(t, templatePath)
			if err != nil {
				errInReplace = err
				return "[error]"
			}
			return subTpl
		})
		if errInReplace != nil {
			return nil, errInReplace
		}
	}

	for _, node := range stack {
		node.Src = reDefineTag.ReplaceAllStringFunc(node.Src, func(raw string) string {
			parsed := reDefineTag.FindStringSubmatch(raw)
			blockName := fmt.Sprintf("BLOCK_%d", blockId)
			blocks[parsed[1]] = blockName

			blockId += 1
			return "{{ define \"" + blockName + "\" }}"
		})
	}

	var rootTemplate *template.Template

	for i, node := range stack {
		node.Src = reTemplateTag.ReplaceAllStringFunc(node.Src, func(raw string) string {
			parsed := reTemplateTag.FindStringSubmatch(raw)
			origName := parsed[1]
			replacedName, ok := blocks[origName]

			dot := "."
			if len(parsed) == 3 && len(parsed[2]) > 0 {
				dot = parsed[2]
			}
			if ok {
				return fmt.Sprintf(`{{ template "%s" %s }}`, replacedName, dot)
			} else {
				return ""
			}
		})

		var thisTemplate *template.Template

		if i == 0 {
			thisTemplate = template.New(node.Name)
			rootTemplate = thisTemplate
		} else {
			thisTemplate = rootTemplate.New(node.Name)
		}

		thisTemplate.Funcs(t.GetFuncs())

		_, err := thisTemplate.Parse(node.Src)
		if err != nil {
			return nil, err
		}
	}

	return rootTemplate, nil
}

func getTemplate(t *templater, tpl string) (string, error) {
	for _, l := range t.GetLoaders() {
		tmpl, err := l.Load(tpl)
		if err == nil {
			return tmpl, nil
		}
	}
	return "", NoTemplateError(tpl)
}

func (t *templater) add(stack *[]*Node, tpl string) error {
	tplSrc, err := getTemplate(t, tpl)

	if err != nil {
		return err
	}

	if len(tplSrc) < 1 {
		return EmptyTemplateError(tpl)
	}

	extendsMatches := reExtendsTag.FindStringSubmatch(tplSrc)
	if len(extendsMatches) == 2 {
		err := t.add(stack, extendsMatches[1])
		if err != nil {
			return err
		}
		tplSrc = reExtendsTag.ReplaceAllString(tplSrc, "")
	}

	node := &Node{
		Name: tpl,
		Src:  tplSrc,
	}

	*stack = append((*stack), node)

	return nil
}

type funcSet struct {
	f map[string]interface{}
}

func newFuncSet() *funcSet {
	return &funcSet{make(map[string]interface{})}
}

type FuncSet interface {
	SetFuncs(...FuncSet)
	GetFuncs() map[string]interface{}
}

func (f *funcSet) SetFuncs(fs ...FuncSet) {
	for _, s := range fs {
		for k, fn := range s.GetFuncs() {
			f.f[k] = fn
		}
	}
}

func (f *funcSet) GetFuncs() map[string]interface{} {
	return f.f
}

type TemplateData struct {
	Data map[string]interface{}
}

func Data(fs env.Flagger, additional ...string) *TemplateData {
	ret := &TemplateData{
		Data: make(map[string]interface{}),
	}

	fn := func(fl *env.Flag) {
		ret.Data[fl.Name] = fl.Value
	}
	fs.VisitAll(fn)

	for _, v := range additional {
		spl := strings.Split(v, ":")
		if len(spl) == 2 {
			ret.Data[spl[0]] = spl[1]
		}
	}

	return ret
}
