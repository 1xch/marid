package marid

import (
	"fmt"
	"go/format"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type Marid interface {
	Configuration
	Logr
	Doer
	Templater
}

type Doer interface {
	Do(string, []string) error
}

type Templater interface {
	Render(string, string, interface{}) error
	Fetch(string) (*template.Template, error)
}

type marid struct {
	Configuration
	Logr
	*settings
	*bufferPool
	loaders LoaderSet
	blocks  BlockSet
	funcs   map[string]interface{}
}

func New(cnf ...Config) Marid {
	m := &marid{
		settings: defaultSettings(),
		loaders:  NewLoaderSet(),
		blocks:   NewBlockSet(),
	}
	m.Configuration = newConfiguration(m, cnf...)
	return m
}

func (m *marid) addFuncs(fns map[string]interface{}) {
	for k, fn := range fns {
		m.funcs[k] = fn
	}
}

func (m *marid) render(t *template.Template, d interface{}, dir, file string) error {
	m.PrintIf("rendering...")
	b := m.get()
	var src []byte
	var err error

	if err = t.Execute(b, d); err != nil {
		return RenderError(err)
	}

	src = b.Bytes()
	src, err = format.Source(src)
	if err != nil {
		return InvalidGoCodeError(err)
	}

	output := strings.ToLower(fmt.Sprintf("%s.go", file))
	outputPath := filepath.Join(dir, output)
	if err := ioutil.WriteFile(outputPath, src, 0644); err != nil {
		return RenderError(err)
	}

	m.put(b)
	m.PrintIf("rendered.")
	return nil
}

func (m *marid) Do(bl string, fl []string) error {
	m.PrintIf("Doing block %s with args %s", bl, fl)
	if blk, ok := m.blocks[bl]; ok {
		fls := blk.Flags()
		fpErr := fls.Parse(fl)
		if fpErr != nil {
			return fpErr
		}
		td := NewTemplateData(blk, fls)
		var rErr error
		for _, t := range blk.Templates() {
			if tmpl, err := m.Fetch(t); err == nil {
				rErr = m.render(tmpl, td.Data, blk.Directory(), t)
				if rErr != nil {
					return rErr
				}
			}
		}
		if fpErr == nil && rErr == nil {
			m.PrintIf("done doing block")
			return nil
		}
	}
	return NoBlockError(bl)
}

func (m *marid) Render(t, dir string, data interface{}) error {
	if tmpl, err := m.Fetch(t); err == nil {
		return m.render(tmpl, data, t, dir)
	}
	return NoTemplateError(t)
}

func (m *marid) Fetch(t string) (*template.Template, error) {
	return m.assemble(t)
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

func (m *marid) assemble(t string) (*template.Template, error) {
	m.PrintIf("assembling...%s", t)
	stack := []*Node{}

	err := m.add(&stack, t)

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
			subTpl, err := getTemplate(m, templatePath)
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

		thisTemplate.Funcs(m.funcs)

		_, err := thisTemplate.Parse(node.Src)
		if err != nil {
			return nil, err
		}
	}

	m.PrintIf("assembled %s...", t)
	return rootTemplate, nil
}

func getTemplate(m *marid, t string) (string, error) {
	for _, l := range m.loaders {
		tmpl, err := l.Load(t)
		if err == nil {
			return tmpl, nil
		}
	}
	return "", NoTemplateError(t)
}

func (m *marid) add(stack *[]*Node, t string) error {
	m.PrintIf("adding %s...", t)
	tplSrc, err := getTemplate(m, t)

	if err != nil {
		return err
	}

	if len(tplSrc) < 1 {
		return EmptyTemplateError(t)
	}

	extendsMatches := reExtendsTag.FindStringSubmatch(tplSrc)
	if len(extendsMatches) == 2 {
		err := m.add(stack, extendsMatches[1])
		if err != nil {
			return err
		}
		tplSrc = reExtendsTag.ReplaceAllString(tplSrc, "")
	}

	node := &Node{
		Name: t,
		Src:  tplSrc,
	}

	*stack = append((*stack), node)

	m.PrintIf("added...%s", t)
	return nil
}
