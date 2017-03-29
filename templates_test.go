package cobalt_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ardanlabs/cobalt"
)

func Test_TemplateHelloWorld(t *testing.T) {
	tmp := cobalt.DefaultTemplates()
	tmp.Directory = "_testdata/templates"

	var buf bytes.Buffer

	if err := tmp.Execute(&buf, "hello", "world"); err != nil {
		t.Fatalf("Error should be nil, was %v", err)
	}

	want := "Body: Hello, world!"
	if got := strings.TrimSpace(buf.String()); got != want {
		t.Errorf("Got:  %s", got)
		t.Errorf("Want: %s", want)
	}
}

func Test_TemplateFuncs(t *testing.T) {
	tmp := cobalt.DefaultTemplates()
	tmp.Directory = "_testdata/templates"
	tmp.Funcs = map[string]interface{}{
		"upper": strings.ToUpper,
		"split": func(s string) []string {
			return strings.Split(s, "")
		},
		"join": func(s []string) string {
			return strings.Join(s, "-")
		},
	}

	var buf bytes.Buffer

	if err := tmp.Execute(&buf, "funcs", "world"); err != nil {
		t.Fatalf("Error should be nil, was %v", err)
	}

	want := "Body: W-O-R-L-D"
	if got := strings.TrimSpace(buf.String()); got != want {
		t.Errorf("Got:  %s", got)
		t.Errorf("Want: %s", want)
	}
}
