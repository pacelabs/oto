package parser

import (
	"bytes"
	"go/doc"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestParse(t *testing.T) {
	is := is.New(t)
	patterns := []string{"./testdata/services/pleasantries"}
	parser := New(patterns...)
	parser.Verbose = testing.Verbose()
	parser.ExcludeInterfaces = []string{"Ignorer"}
	def, err := parser.Parse()
	is.NoErr(err)

	is.Equal(def.PackageName, "pleasantries")
	is.Equal(len(def.Services), 3) // should be 3 services
	is.Equal(def.Services[0].Name, "GreeterService")
	is.Equal(def.Services[0].Metadata["strapline"], "A lovely greeter service") // custom metadata
	is.Equal(def.Services[0].Comment, `GreeterService is a polite API.
You will love it.`)
	is.Equal(len(def.Services[0].Methods), 2)
	is.Equal(def.Services[0].Methods[0].Name, "GetGreetings")
	is.Equal(def.Services[0].Methods[0].Metadata["featured"], false) // custom metadata
	is.Equal(def.Services[0].Methods[0].NameLowerCamel, "getGreetings")
	is.Equal(def.Services[0].Methods[0].NameLowerSnake, "get_greetings")
	is.Equal(def.Services[0].Methods[0].Comment, "GetGreetings gets a range of saved Greetings.")
	is.Equal(def.Services[0].Methods[0].InputObject.TypeName, "GetGreetingsRequest")
	is.Equal(def.Services[0].Methods[0].InputObject.Multiple, false)
	is.Equal(def.Services[0].Methods[0].InputObject.Package, "")
	is.Equal(def.Services[0].Methods[0].OutputObject.TypeName, "GetGreetingsResponse")
	is.Equal(def.Services[0].Methods[0].OutputObject.Multiple, false)
	is.Equal(def.Services[0].Methods[0].OutputObject.Package, "")

	is.Equal(def.Services[0].Methods[1].Name, "Greet")
	is.Equal(def.Services[0].Methods[1].Metadata["featured"], true) // custom metadata
	is.Equal(def.Services[0].Methods[1].NameLowerCamel, "greet")
	is.Equal(def.Services[0].Methods[1].NameLowerSnake, "greet")
	is.Equal(def.Services[0].Methods[1].Comment, "Greet creates a Greeting for one or more people.")
	is.Equal(def.Services[0].Methods[1].InputObject.TypeName, "GreetRequest")
	is.Equal(def.Services[0].Methods[1].InputObject.Multiple, false)
	is.Equal(def.Services[0].Methods[1].InputObject.Package, "")
	is.Equal(def.Services[0].Methods[1].OutputObject.TypeName, "GreetResponse")
	is.Equal(def.Services[0].Methods[1].OutputObject.Multiple, false)
	is.Equal(def.Services[0].Methods[1].OutputObject.IsObject, true)
	is.Equal(def.Services[0].Methods[1].OutputObject.Package, "")

	greetResponse, err := def.Object(def.Services[0].Methods[1].OutputObject.TypeName)
	is.NoErr(err)
	is.Equal(greetResponse.Fields[0].Name, "Greeting")
	is.Equal(greetResponse.Fields[0].Type.IsObject, true)

	formatCommentText := func(s string) string {
		var buf bytes.Buffer
		doc.ToText(&buf, s, "// ", "", 80)
		return buf.String()
	}
	greetResponseObject, err := def.Object(def.Services[0].Methods[1].OutputObject.TypeName)
	is.NoErr(err)
	actualComment := strings.TrimSpace(formatCommentText(greetResponseObject.Comment))
	is.Equal(actualComment, `// GreetResponse is the response object containing a person's greeting.`)

	greetInputObject, err := def.Object(def.Services[0].Methods[0].InputObject.TypeName)
	is.NoErr(err)
	is.Equal(greetInputObject.Name, "GetGreetingsRequest")
	is.Equal(greetInputObject.Comment, "GetGreetingsRequest is the request object for GreeterService.GetGreetings.")
	is.Equal(greetInputObject.Metadata["featured"], true) // custom metadata
	is.Equal(len(greetInputObject.Fields), 1)
	is.Equal(greetInputObject.Fields[0].Name, "Page")
	is.Equal(greetInputObject.Fields[0].NameLowerCamel, "page")
	is.Equal(greetInputObject.Fields[0].NameLowerSnake, "page")
	is.Equal(greetInputObject.Fields[0].Comment, "Page describes which page of data to get.")
	is.Equal(greetInputObject.Fields[0].OmitEmpty, false)
	is.Equal(greetInputObject.Fields[0].Type.TypeName, "services.Page")
	is.Equal(greetInputObject.Fields[0].Type.CleanObjectName, "services.Page")
	is.Equal(greetInputObject.Fields[0].Type.ObjectName, "Page")
	is.Equal(greetInputObject.Fields[0].Type.ObjectNameLowerCamel, "page")
	is.Equal(greetInputObject.Fields[0].Type.ObjectNameLowerSnake, "page")
	is.Equal(greetInputObject.Fields[0].Type.JSType, "object")
	is.Equal(greetInputObject.Fields[0].Type.TypeID, "github.com/meitner-se/oto/testdata/services.Page")
	is.Equal(greetInputObject.Fields[0].Type.TSType, "services.Page")
	is.Equal(greetInputObject.Fields[0].Type.IsObject, true)
	is.Equal(greetInputObject.Fields[0].Type.Multiple, false)
	is.Equal(greetInputObject.Fields[0].Type.Package, "github.com/meitner-se/oto/testdata/services")
	is.Equal(greetInputObject.Fields[0].Tag, `tagtest:"value,option1,option2"`)
	is.True(greetInputObject.Fields[0].ParsedTags != nil)
	is.Equal(greetInputObject.Fields[0].ParsedTags["tagtest"].Value, "value")
	is.Equal(len(greetInputObject.Fields[0].ParsedTags["tagtest"].Options), 2)
	is.Equal(greetInputObject.Fields[0].ParsedTags["tagtest"].Options[0], "option1")
	is.Equal(greetInputObject.Fields[0].ParsedTags["tagtest"].Options[1], "option2")

	greetOutputObject, err := def.Object(def.Services[0].Methods[0].OutputObject.TypeName)
	is.NoErr(err)
	is.Equal(greetOutputObject.Name, "GetGreetingsResponse")
	is.Equal(greetOutputObject.Comment, "GetGreetingsResponse is the respponse object for GreeterService.GetGreetings.")
	is.Equal(greetOutputObject.Metadata["featured"], false) // custom metadata
	is.Equal(len(greetOutputObject.Fields), 2)
	is.Equal(greetOutputObject.Fields[0].Name, "Greetings")
	is.Equal(greetOutputObject.Fields[0].NameLowerCamel, "greetings")
	is.Equal(greetOutputObject.Fields[0].NameLowerSnake, "greetings")
	is.Equal(greetOutputObject.Fields[0].Type.TypeID, "github.com/meitner-se/oto/parser/testdata/services/pleasantries.Greeting")
	is.Equal(greetOutputObject.Fields[0].OmitEmpty, false)
	is.Equal(greetOutputObject.Fields[0].Type.TypeName, "Greeting")
	is.Equal(greetOutputObject.Fields[0].Type.Multiple, true)
	is.Equal(greetOutputObject.Fields[0].Type.Package, "")
	is.Equal(greetOutputObject.Fields[1].Name, "Error")
	is.Equal(greetOutputObject.Fields[1].NameLowerCamel, "error")
	is.Equal(greetOutputObject.Fields[1].NameLowerSnake, "error")
	is.Equal(greetOutputObject.Fields[1].OmitEmpty, true)
	is.Equal(greetOutputObject.Fields[1].Type.TypeName, "string")
	is.Equal(greetOutputObject.Fields[1].Type.Multiple, false)
	is.Equal(greetOutputObject.Fields[1].Type.Package, "")

	example, err := def.ExampleJSON(*greetOutputObject)
	is.NoErr(err)
	is.Equal(string(example), `{"error":"something went wrong","greetings":[{"text":"Hello there"}]}`)

	is.Equal(def.Services[1].Name, "StrangeTypesService")
	strangeInputObj, err := def.Object(def.Services[1].Methods[0].InputObject.ObjectName)
	is.NoErr(err)
	is.Equal(strangeInputObj.Fields[0].Type.JSType, "any")
	is.Equal(strangeInputObj.Fields[0].Type.TSType, "any")

	is.Equal(def.Services[2].Name, "Welcomer")
	is.Equal(len(def.Services[2].Methods), 1)

	is.Equal(def.Services[2].Methods[0].InputObject.TypeName, "WelcomeRequest")
	is.Equal(def.Services[2].Methods[0].InputObject.Multiple, false)
	is.Equal(def.Services[2].Methods[0].InputObject.Package, "")
	is.Equal(def.Services[2].Methods[0].OutputObject.TypeName, "WelcomeResponse")
	is.Equal(def.Services[2].Methods[0].OutputObject.Multiple, false)
	is.Equal(def.Services[2].Methods[0].OutputObject.Package, "")

	welcomeInputObject, err := def.Object(def.Services[2].Methods[0].InputObject.TypeName)
	is.NoErr(err)
	is.Equal(welcomeInputObject.Name, "WelcomeRequest")
	is.Equal(len(welcomeInputObject.Fields), 4)

	example, err = def.ExampleJSON(*welcomeInputObject)
	is.NoErr(err)
	is.Equal(string(example), `{"customer_details":{"new_customer":true},"name":"John Smith","recipients":"your@email.com","times":3}`)

	is.Equal(welcomeInputObject.Fields[0].Name, "To")
	is.Equal(welcomeInputObject.Fields[0].Comment, "To is the address of the person to send the message to.")
	is.Equal(welcomeInputObject.Fields[0].Metadata["featured"], true)
	is.Equal(welcomeInputObject.Fields[0].NameLowerSnake, "recipients") // changed by json tag
	is.Equal(welcomeInputObject.Fields[0].NameLowerCamel, "recipients") // changed by json tag
	is.Equal(welcomeInputObject.Fields[0].OmitEmpty, false)
	is.Equal(welcomeInputObject.Fields[0].Type.TypeName, "string")
	is.Equal(welcomeInputObject.Fields[0].Type.Multiple, false)
	is.Equal(welcomeInputObject.Fields[0].Type.Package, "")
	is.Equal(welcomeInputObject.Fields[0].Example, "your@email.com")

	is.Equal(welcomeInputObject.Fields[1].Name, "Name")
	is.True(welcomeInputObject.Fields[0].Metadata != nil) // no metadata shouldn't be nil
	is.Equal(welcomeInputObject.Fields[1].NameLowerCamel, "name")
	is.Equal(welcomeInputObject.Fields[1].NameLowerSnake, "name")
	is.Equal(welcomeInputObject.Fields[1].OmitEmpty, false)
	is.Equal(welcomeInputObject.Fields[1].Type.TypeName, "*string")
	is.Equal(welcomeInputObject.Fields[1].Type.JSType, "string")
	is.Equal(welcomeInputObject.Fields[1].Type.TSType, "string")
	is.Equal(welcomeInputObject.Fields[1].Type.SwiftType, "String")
	is.Equal(welcomeInputObject.Fields[1].Type.Multiple, false)
	is.Equal(welcomeInputObject.Fields[1].Type.Package, "")
	is.Equal(welcomeInputObject.Fields[1].Example, "John Smith")

	is.Equal(welcomeInputObject.Fields[2].Example, float64(3))
	is.Equal(welcomeInputObject.Fields[2].Type.JSType, "number")
	is.Equal(welcomeInputObject.Fields[2].Type.TSType, "number")
	is.Equal(welcomeInputObject.Fields[2].Type.SwiftType, "Double")

	is.Equal(welcomeInputObject.Fields[3].Type.TypeName, "*CustomerDetails")
	is.Equal(welcomeInputObject.Fields[3].Type.JSType, "object")
	is.Equal(welcomeInputObject.Fields[3].Type.TSType, "CustomerDetails")
	is.Equal(welcomeInputObject.Fields[3].Example, nil)
	is.Equal(welcomeInputObject.Fields[3].Type.SwiftType, "CustomerDetails")

	welcomeOutputObject, err := def.Object(def.Services[2].Methods[0].OutputObject.TypeName)
	is.NoErr(err)
	is.Equal(welcomeOutputObject.Name, "WelcomeResponse")
	is.Equal(len(welcomeOutputObject.Fields), 2)

	example, err = def.ExampleJSON(*welcomeOutputObject)
	is.NoErr(err)
	is.Equal(string(example), `{"error":"something went wrong","message":"Welcome John Smith."}`)

	is.Equal(welcomeOutputObject.Fields[0].Name, "Message")
	is.Equal(welcomeOutputObject.Fields[0].NameLowerCamel, "message")
	is.Equal(welcomeOutputObject.Fields[0].NameLowerSnake, "message")
	is.Equal(welcomeOutputObject.Fields[0].Type.IsObject, false)
	is.Equal(welcomeOutputObject.Fields[0].OmitEmpty, false)
	is.Equal(welcomeOutputObject.Fields[0].Type.TypeName, "string")
	is.Equal(welcomeOutputObject.Fields[0].Type.Multiple, false)
	is.Equal(welcomeOutputObject.Fields[0].Type.Package, "")
	is.Equal(welcomeOutputObject.Fields[1].Name, "Error")
	is.Equal(welcomeOutputObject.Fields[1].NameLowerCamel, "error")
	is.Equal(welcomeOutputObject.Fields[1].NameLowerSnake, "error")
	is.Equal(welcomeOutputObject.Fields[1].OmitEmpty, true)
	is.Equal(welcomeOutputObject.Fields[1].Type.TypeName, "string")
	is.Equal(welcomeOutputObject.Fields[1].Type.Multiple, false)
	is.Equal(welcomeOutputObject.Fields[1].Type.Package, "")
	is.Equal(welcomeOutputObject.Fields[1].Type.JSType, "string")
	is.Equal(welcomeOutputObject.Fields[1].Type.TSType, "string")
	is.Equal(welcomeOutputObject.Fields[1].Type.SwiftType, "String")
	is.True(welcomeOutputObject.Metadata != nil)

	is.Equal(len(def.Objects), 11)
	for i := range def.Objects {
		switch def.Objects[i].Name {
		case "Greeting":
			is.Equal(len(def.Objects[i].Fields), 1)
			is.Equal(def.Objects[i].Imported, false)
		case "Page":
			is.Equal(def.Objects[i].TypeID, "github.com/meitner-se/oto/testdata/services.Page")
			is.Equal(len(def.Objects[i].Fields), 3)
			is.Equal(def.Objects[i].Imported, true)
		}
	}

	// b, err := json.MarshalIndent(def, "", "  ")
	// is.NoErr(err)
	// log.Println(string(b))
}

func TestFieldTypeIsOptional(t *testing.T) {
	is := is.New(t)

	f := FieldType{ObjectName: "*SomeType"}
	is.Equal(f.IsOptional(), true)
	f = FieldType{ObjectName: "SomeType"}
	is.Equal(f.IsOptional(), false)
}

func TestExtractCommentMetadata(t *testing.T) {
	is := is.New(t)

	p := &Parser{}
	p.Verbose = testing.Verbose()
	metadata, comment, err := p.extractCommentMetadata(`
		This is a comment
		example: "With an example"
		required: true
		monkey: 24
		Kind is one of: monthly, weekly, tags-monthly, tags-weekly
	`)
	is.NoErr(err)
	is.Equal(comment, "This is a comment")
	is.Equal(metadata["example"], "With an example")
	is.Equal(metadata["required"], true)
	is.Equal(metadata["monkey"], float64(24))
}

func TestObjectIsInputOutput(t *testing.T) {
	is := is.New(t)
	patterns := []string{"./testdata/services/pleasantries"}
	parser := New(patterns...)
	parser.Verbose = testing.Verbose()
	parser.ExcludeInterfaces = []string{"Ignorer"}
	def, err := parser.Parse()
	is.NoErr(err)

	is.Equal(def.ObjectIsInput("GreetRequest"), true)
	is.Equal(def.ObjectIsInput("GreetResponse"), false)
	is.Equal(def.ObjectIsOutput("GreetRequest"), false)
	is.Equal(def.ObjectIsOutput("GreetResponse"), true)
}

func TestParseNestedStructs(t *testing.T) {
	is := is.New(t)
	patterns := []string{"./testdata/nested-structs"}
	p := New(patterns...)
	p.Verbose = testing.Verbose()
	_, err := p.Parse()
	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "nested structs not supported"))
}

func TestParseMap(t *testing.T) {
	is := is.New(t)
	patterns := []string{"./testdata/maps"}
	p := New(patterns...)
	p.Verbose = testing.Verbose()
	def, err := p.Parse()
	is.NoErr(err)

	greetInputObject, err := def.Object("GreetRequest")
	is.NoErr(err)
	is.Equal(len(greetInputObject.Fields), 1)
	is.Equal(greetInputObject.Fields[0].Name, "GreetingMap")
	is.Equal(greetInputObject.Fields[0].Type.IsMap, true)
	is.Equal(greetInputObject.Fields[0].Type.Map.KeyType, "string")
	is.Equal(greetInputObject.Fields[0].Type.Map.KeyTypeJS, "string")
	is.Equal(greetInputObject.Fields[0].Type.Map.KeyTypeTS, "string")
	is.Equal(greetInputObject.Fields[0].Type.Map.KeyTypeSwift, "String")
	is.Equal(greetInputObject.Fields[0].Type.Map.ElementType, "int")
	is.Equal(greetInputObject.Fields[0].Type.Map.ElementTypeJS, "number")
	is.Equal(greetInputObject.Fields[0].Type.Map.ElementTypeTS, "number")
	is.Equal(greetInputObject.Fields[0].Type.Map.ElementTypeSwift, "Double")
	is.Equal(greetInputObject.Fields[0].Type.Map.ElementIsMultiple, false)

	greetOutputObject, err := def.Object("GreetResponse")
	is.NoErr(err)
	is.Equal(len(greetOutputObject.Fields), 2)
	is.Equal(greetOutputObject.Fields[0].Name, "Greeting")
	is.Equal(greetOutputObject.Fields[0].Type.IsMap, true)
	is.Equal(greetOutputObject.Fields[0].Type.Map.KeyType, "string")
	is.Equal(greetOutputObject.Fields[0].Type.Map.KeyTypeJS, "string")
	is.Equal(greetOutputObject.Fields[0].Type.Map.KeyTypeTS, "string")
	is.Equal(greetOutputObject.Fields[0].Type.Map.KeyTypeSwift, "String")
	is.Equal(greetOutputObject.Fields[0].Type.Map.ElementType, "string")
	is.Equal(greetOutputObject.Fields[0].Type.Map.ElementTypeJS, "string")
	is.Equal(greetOutputObject.Fields[0].Type.Map.ElementTypeTS, "string")
	is.Equal(greetOutputObject.Fields[0].Type.Map.ElementTypeSwift, "String")
	is.Equal(greetOutputObject.Fields[0].Type.Map.ElementIsMultiple, false)
	is.Equal(greetOutputObject.Fields[1].Name, "Error")
	is.Equal(greetOutputObject.Fields[1].Type.IsMap, false)

	greetMultipleInputObject, err := def.Object("GreetMultipleRequest")
	is.NoErr(err)
	is.Equal(len(greetMultipleInputObject.Fields), 1)
	is.Equal(greetMultipleInputObject.Fields[0].Name, "GreetingMap")
	is.Equal(greetMultipleInputObject.Fields[0].Type.IsMap, true)
	is.Equal(greetMultipleInputObject.Fields[0].Type.Map.KeyType, "string")
	is.Equal(greetMultipleInputObject.Fields[0].Type.Map.KeyTypeJS, "string")
	is.Equal(greetMultipleInputObject.Fields[0].Type.Map.KeyTypeTS, "string")
	is.Equal(greetMultipleInputObject.Fields[0].Type.Map.KeyTypeSwift, "String")
	is.Equal(greetMultipleInputObject.Fields[0].Type.Map.ElementType, "GreetRequest")
	is.Equal(greetMultipleInputObject.Fields[0].Type.Map.ElementTypeJS, "GreetRequest")
	is.Equal(greetMultipleInputObject.Fields[0].Type.Map.ElementTypeTS, "GreetRequest")
	is.Equal(greetMultipleInputObject.Fields[0].Type.Map.ElementTypeSwift, "GreetRequest")
	is.Equal(greetMultipleInputObject.Fields[0].Type.Map.ElementIsMultiple, true)

	greetMultipleOutputObject, err := def.Object("GreetMultipleResponse")
	is.NoErr(err)
	is.Equal(len(greetMultipleOutputObject.Fields), 2)
	is.Equal(greetMultipleOutputObject.Fields[0].Name, "Greeting")
	is.Equal(greetMultipleOutputObject.Fields[0].Type.IsMap, true)
	is.Equal(greetMultipleOutputObject.Fields[0].Type.Map.KeyType, "string")
	is.Equal(greetMultipleOutputObject.Fields[0].Type.Map.KeyTypeJS, "string")
	is.Equal(greetMultipleOutputObject.Fields[0].Type.Map.KeyTypeTS, "string")
	is.Equal(greetMultipleOutputObject.Fields[0].Type.Map.KeyTypeSwift, "String")
	is.Equal(greetMultipleOutputObject.Fields[0].Type.Map.ElementType, "GreetResponse")
	is.Equal(greetMultipleOutputObject.Fields[0].Type.Map.ElementTypeJS, "GreetResponse")
	is.Equal(greetMultipleOutputObject.Fields[0].Type.Map.ElementTypeTS, "GreetResponse")
	is.Equal(greetMultipleOutputObject.Fields[0].Type.Map.ElementTypeSwift, "GreetResponse")
	is.Equal(greetMultipleOutputObject.Fields[0].Type.Map.ElementIsMultiple, true)
	is.Equal(greetMultipleOutputObject.Fields[1].Name, "Error")
	is.Equal(greetMultipleOutputObject.Fields[1].Type.IsMap, false)
}
