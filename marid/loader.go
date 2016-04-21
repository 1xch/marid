package marid

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type LoaderSet []Loader

func NewLoaderSet() LoaderSet {
	return make([]Loader, 0)
}

type Loader interface {
	Load(string) (string, error)
	ListTemplates() []string
}

type BaseLoader struct {
	Errors         []error
	FileExtensions []string
}

func (b *BaseLoader) Load(name string) (string, error) {
	return "", NoLoadMethod
}

func (b *BaseLoader) ListTemplates() []string {
	return []string{"not implemented"}
}

func (b *BaseLoader) ValidExtension(ext string) bool {
	for _, extension := range b.FileExtensions {
		if extension == ext {
			return true
		}
	}
	return false
}

type dirLoader struct {
	BaseLoader
	Paths []string
}

func DirLoader(paths ...string) Loader {
	d := &dirLoader{}
	d.FileExtensions = append(d.FileExtensions, ".m")
	for _, p := range paths {
		p, err := filepath.Abs(filepath.Clean(p))
		if err != nil {
			d.Errors = append(d.Errors, PathError(p))
		}
		d.Paths = append(d.Paths, p)
	}
	return d
}

func (l *dirLoader) Load(name string) (string, error) {
	for _, p := range l.Paths {
		f := filepath.Join(p, name)
		if l.ValidExtension(filepath.Ext(f)) {
			if _, err := os.Stat(f); err == nil {
				file, err := os.Open(f)
				r, err := ioutil.ReadAll(file)
				return string(r), err
			}
		}
	}
	return "", NoTemplateError(name)
}

func (l *dirLoader) ListTemplates() []string {
	var listing []string
	for _, p := range l.Paths {
		filepath.Walk(p, func(path string, _ os.FileInfo, _ error) (err error) {
			tem := filepath.Base(path)
			if l.ValidExtension(filepath.Ext(tem)) {
				listing = append(listing, tem)
			}
			return err
		})
	}
	return listing
}

type mapLoader struct {
	BaseLoader
	TemplateMap map[string]string
}

func MapLoader(tm ...map[string]string) Loader {
	m := &mapLoader{TemplateMap: make(map[string]string)}
	for _, t := range tm {
		for k, v := range t {
			m.TemplateMap[k] = v
		}
	}
	return m
}

func (l *mapLoader) Load(name string) (string, error) {
	if r, ok := l.TemplateMap[name]; ok {
		return string(r), nil
	}
	return "", NoTemplateError(name)
}

func (l *mapLoader) ListTemplates() []string {
	var listing []string
	for k, _ := range l.TemplateMap {
		listing = append(listing, k)
	}
	return listing
}
