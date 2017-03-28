package cobalt

import (
	"html/template"
	"io"
	"path/filepath"
)

// Templates handles the compiling and execution of templates.
type Templates struct {
	Directory   string // The directory holding template files.
	Extension   string // The file extension on templates.
	Layout      string // The name of the base template that holds
	Development bool   // Set to true to enable recompilation on each request
	Funcs       template.FuncMap

	cache map[string]*template.Template
}

// DefaultTemplates creates a Templates set with default values.
func DefaultTemplates() Templates {
	return Templates{
		Directory:   "templates",
		Extension:   ".tmpl",
		Layout:      "_layout",
		Development: false,
		Funcs:       make(template.FuncMap),
		cache:       make(map[string]*template.Template),
	}
}

func (t Templates) lookup(name string) (*template.Template, error) {
	if !t.Development {
		if tmp, ok := t.cache[name]; ok {
			return tmp, nil
		}
	}

	l, err := template.New(t.Layout + t.Extension).Funcs(t.Funcs).ParseFiles(filepath.Join(t.Directory, t.Layout+t.Extension))
	if err != nil {
		return nil, err
	}

	n := filepath.Join(t.Directory, name+t.Extension)
	tmp, err := l.ParseFiles(n)
	if err != nil {
		return nil, err
	}

	t.cache[name] = tmp

	return tmp, nil
}

// Execute will load the named template and execute it against the provided
// writer.
func (t Templates) Execute(w io.Writer, name string, data interface{}) error {

	tmp, err := t.lookup(name)
	if err != nil {
		return err
	}

	return tmp.Execute(w, data)
}
