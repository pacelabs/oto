package parser

import (
	"encoding/json"
	"fmt"
)

// Example generates an object that is a realistic example
// of this object.
// Examples are read from the docs.
// This is experimental.
func (d *Definition) Example(o Object) (map[string]interface{}, error) {
	obj := make(map[string]interface{})
	for _, field := range o.Fields {
		if field.Type.IsObject {
			subobj, err := d.Object(field.Type.CleanObjectName)
			if err != nil {
				if err == ErrNotFound {
					continue
				}
				return nil, fmt.Errorf("Object(%q): %w", field.Type.CleanObjectName, err)
			}
			if subobj.Name == o.Name {
				obj[field.NameLowerSnake] = struct{}{}
				continue
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
			obj[field.NameLowerSnake] = []interface{}{field.Example, field.Example, field.Example}
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
