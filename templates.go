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

func (t Templates) lookup(layout, name string) (*template.Template, error) {
	if !t.Development {
		if tmp, ok := t.cache[layout+name]; ok {
			return tmp, nil
		}
	}

	var tmp *template.Template

	if layout != "" {
		// Compile layout to get a base template
		l, err := template.New(t.Layout + t.Extension).Funcs(t.Funcs).ParseFiles(filepath.Join(t.Directory, t.Layout+t.Extension))
		if err != nil {
			return nil, err
		}

		// Add specified file to the base template
		n := filepath.Join(t.Directory, name+t.Extension)
		tmp, err = l.ParseFiles(n)
		if err != nil {
			return nil, err
		}
	} else {
		var err error

		tmp, err = template.New(filepath.Base(name) + t.Extension).Funcs(t.Funcs).ParseFiles(filepath.Join(t.Directory, name+t.Extension))
		if err != nil {
			return nil, err
		}
	}

	t.cache[layout+name] = tmp

	return tmp, nil
}

// Execute will load the layout and the named template then execute them
// against the provided writer.
func (t Templates) Execute(w io.Writer, name string, data interface{}) error {

	tmp, err := t.lookup(t.Layout, name)
	if err != nil {
		return err
	}

	return tmp.Execute(w, data)
}

// ExecuteOnly will load the named template ignoring the layout file. It is
// then executed against the provided writer.
func (t Templates) ExecuteOnly(w io.Writer, name string, data interface{}) error {

	tmp, err := t.lookup("", name)
	if err != nil {
		return err
	}

	return tmp.Execute(w, data)
}
