package x

import "fmt"

type maridError struct {
	err  string
	vals []interface{}
}

func (m *maridError) Error() string {
	return fmt.Sprintf(m.err, m.vals...)
}

func (m *maridError) Out(vals ...interface{}) *maridError {
	m.vals = vals
	return m
}

func Mrror(err string) *maridError {
	return &maridError{err: err}
}

var (
	EmptyTemplateError = Mrror("empty template named %s").Out
	NoTemplateError    = Mrror("no template named %s").Out
	PathError          = Mrror("path: %s returned error").Out
	NoLoadMethod       = Mrror("load method not implemented")
	RenderError        = Mrror("render error: %s").Out
	InvalidGoCodeError = Mrror("error formatting go code: invalid Go generated: %s\ncompile the package to analyze the error").Out
	NoBlockError       = Mrror("no block named %s available").Out
)
