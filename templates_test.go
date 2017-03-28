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
