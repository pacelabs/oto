package parser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Example generates an object that is a realistic example
// of this object.
// Examples are read from the docs.
// This is experimental.
func (d *Definition) Example(o Object) (map[string]interface{}, error) {
	obj := make(map[string]interface{})
	for _, field := range o.Fields {
		if field.Type.IsObject {
			subobj, err := d.Object(strings.TrimPrefix(field.Type.ObjectName, "*"))
			if err != nil {
				return nil, fmt.Errorf("Object(%q): %w", field.Type.ObjectName, err)
			}
			example, err := d.Example(*subobj)
			if err != nil {
				return nil, err
			}
			obj[field.NameLowerSnake] = example
			if field.Type.Multiple {
				// turn it into an array
				obj[field.NameLowerSnake] = []interface{}{obj[field.NameLowerSnake]}
			}
			continue
		}
		obj[field.NameLowerSnake] = field.Example
		if field.Type.Multiple {
			// turn it into an array
			obj[field.NameLowerSnake] = []interface{}{obj[field.NameLowerSnake], obj[field.NameLowerSnake], obj[field.NameLowerSnake]}
		}
	}
	return obj, nil
}

func (d *Definition) ExampleJSON(o Object) ([]byte, error) {
	data, err := d.Example(o)
	if err != nil {
		return nil, err
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return dataJSON, nil
}
