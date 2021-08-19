package parser

import (
	"testing"

	"github.com/matryer/is"
)

func TestObjectExample(t *testing.T) {
	is := is.New(t)

	obj1 := Object{
		Name: "obj1",
		Fields: []Field{
			{
				Name:           "Name",
				NameLowerSnake: "name",
				Example:        "Mat",
			},
			{
				Name:           "Project",
				NameLowerSnake: "project",
				Example:        "Respond",
			},
			{
				Name:           "SinceYear",
				NameLowerSnake: "sinceYear",
				Example:        2021,
			},
			{
				Name:           "Favourites",
				NameLowerSnake: "favourites",
				Type: FieldType{
					TypeName:        "obj2",
					ObjectName:      "obj2",
					CleanObjectName: "obj2",
					IsObject:        true,
				},
			},
		},
	}
	obj2 := Object{
		Name: "obj2",
		Fields: []Field{
			{
				Type:           FieldType{TypeName: "string", Multiple: true},
				NameLowerSnake: "languages",
				Example:        "Go",
			},
		},
	}
	def := &Definition{
		Objects: []Object{obj1, obj2},
	}
	example, err := def.Example(obj1)
	is.NoErr(err)
	is.True(example != nil)

	is.Equal(example["name"], "Mat")
	is.Equal(example["project"], "Respond")
	is.Equal(example["sinceYear"], 2021)
	is.True(example["favourites"] != nil)
	favourites, ok := example["favourites"].(map[string]interface{})
	is.True(ok) // Favourites map[string]interface{}
	languages, ok := favourites["languages"].([]interface{})
	is.True(ok) // Languages []interface{}
	is.Equal(len(languages), 3)

	exampleJSON, err := def.ExampleJSON(obj1)
	is.NoErr(err)
	is.Equal(string(exampleJSON), `{"favourites":{"languages":["Go","Go","Go"]},"name":"Mat","project":"Respond","sinceYear":2021}`)
}
