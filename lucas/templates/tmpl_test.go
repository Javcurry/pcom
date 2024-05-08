package templates_test

import (
	"fmt"
	"html/template"
	"os"
	"testing"
)

func TestTmplWith(t *testing.T) {

	type TestS struct {
		S string
	}
	tmpl := `{{with .S}}test{{end}}`
	tm, err := template.New("test").Parse(tmpl)
	fmt.Println(err)
	s := TestS{}
	err = tm.Execute(os.Stdout, s)
	fmt.Println(err)
	s.S = "test"
	err = tm.Execute(os.Stdout, s)
	fmt.Println(err)
}
