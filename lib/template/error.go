package template

import "fmt"

type templateError struct {
	err  string
	vals []interface{}
}

func (t *templateError) Error() string {
	return fmt.Sprintf(t.err, t.vals...)
}

func (t *templateError) Out(vals ...interface{}) *templateError {
	t.vals = vals
	return t
}

func Trror(err string) *templateError {
	return &templateError{err: err}
}

var (
	EmptyTemplateError = Trror("empty template named %s").Out
	NoTemplateError    = Trror("no template named %s").Out
	PathError          = Trror("path: %s returned error").Out
	NoLoadMethod       = Trror("load method not implemented")
	RenderError        = Trror("render error: %s").Out
	InvalidGoCodeError = Trror("error formatting go code: invalid Go generated: %s\ncompile the package to analyze the error").Out
	NoBlockError       = Trror("no block named %s available").Out
)
