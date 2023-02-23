package render

import (
	"os"
	"strings"
	"testing"

	"github.com/matryer/is"
	"github.com/pacedotdev/oto/parser"
)

func TestExmapleGolang(t *testing.T) {
	is := is.New(t)
	patterns := []string{"../parser/testdata/services/pleasantries"}
	parser := parser.New(patterns...)
	parser.Verbose = testing.Verbose()
	parser.ExcludeInterfaces = []string{"Ignorer"}
	def, err := parser.Parse()
	is.NoErr(err)
	inputObject, err := def.Object(def.Services[0].Methods[0].InputObject.ObjectName)
	is.NoErr(err) // get inputObject
	example := ObjectGolang(def, *inputObject, 0)
	err = os.WriteFile("./delete-me-example.go.notgo", []byte(example), 0666)
	is.NoErr(err) // write file

	for _, should := range []string{
		"// Package services contains services.",
		"package services",
	} {
		if !strings.Contains(example, should) {
			t.Errorf("missing: %s", should)
			is.Fail()
		}
	}
}
