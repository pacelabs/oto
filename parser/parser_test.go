package parser

import (
	"bytes"
	"go/doc"
	"html/template"
	"strings"
	"testing"

	"github.com/matryer/is"
	"github.com/stretchr/testify/assert"
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
	is.Equal(def.MethodHasPagination(def.Services[0].Methods[0]), true)
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
	is.Equal(len(greetOutputObject.Fields), 3)
	is.Equal(greetOutputObject.Fields[0].Name, "Greetings")
	is.Equal(greetOutputObject.Fields[0].NameLowerCamel, "greetings")
	is.Equal(greetOutputObject.Fields[0].NameLowerSnake, "greetings")
	is.Equal(greetOutputObject.Fields[0].Type.TypeID, "github.com/meitner-se/oto/parser/testdata/services/pleasantries.Greeting")
	is.Equal(greetOutputObject.Fields[0].OmitEmpty, false)
	is.Equal(greetOutputObject.Fields[0].Type.TypeName, "Greeting")
	is.Equal(greetOutputObject.Fields[0].Type.Multiple, true)
	is.Equal(greetOutputObject.Fields[0].Type.Package, "")
	is.Equal(greetOutputObject.Fields[1].Name, "TotalCount")
	is.Equal(greetOutputObject.Fields[1].NameLowerCamel, "totalCount")
	is.Equal(greetOutputObject.Fields[1].NameLowerSnake, "total_count")
	is.Equal(greetOutputObject.Fields[1].OmitEmpty, false)
	is.Equal(greetOutputObject.Fields[1].Type.TypeName, "int64")
	is.Equal(greetOutputObject.Fields[1].Type.Multiple, false)
	is.Equal(greetOutputObject.Fields[1].Type.Package, "")
	is.Equal(greetOutputObject.Fields[2].Name, "Error")
	is.Equal(greetOutputObject.Fields[2].NameLowerCamel, "error")
	is.Equal(greetOutputObject.Fields[2].NameLowerSnake, "error")
	is.Equal(greetOutputObject.Fields[2].OmitEmpty, true)
	is.Equal(greetOutputObject.Fields[2].Type.TypeName, "string")
	is.Equal(greetOutputObject.Fields[2].Type.Multiple, false)
	is.Equal(greetOutputObject.Fields[2].Type.Package, "")

	example, err := def.ExampleJSON(*greetOutputObject)
	is.NoErr(err)
	is.Equal(string(example), `{"error":"something went wrong","greetings":[{"text":"Hello there"}],"total_count":334}`)

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
	is.Equal(welcomeInputObject.Fields[3].Example, struct{}{})
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

func Test_writeZodFieldModifiers(t *testing.T) {
	tt := []struct {
		name  string
		field Field
		want  string
	}{
		{
			name: "Optional field",
			field: Field{
				Type: FieldType{
					Multiple:      false,
					MultipleTimes: nil,
				},
				Metadata: map[string]interface{}{
					"optional": true,
				},
			},
			want: ".optional()",
		},
		{
			name: "Array field",
			field: Field{
				Type: FieldType{
					Multiple:      true,
					MultipleTimes: []struct{}{{}},
				},
				Metadata: map[string]interface{}{
					"optional": false,
				},
			},
			want: ".array()",
		},
		{
			name: "Nested array field",
			field: Field{
				Type: FieldType{
					Multiple:      true,
					MultipleTimes: []struct{}{{}, {}},
				},
				Metadata: map[string]interface{}{
					"optional": false,
				},
			},
			want: ".array().array()",
		},
		{
			name: "Nullable field",
			field: Field{
				Type: FieldType{
					Multiple:      false,
					MultipleTimes: nil,
				},
				Metadata: map[string]interface{}{
					"nullable": true,
				},
			},
			want: ".nullable()",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)

			builder := strings.Builder{}

			writeZodFieldModifiers(tc.field, &builder)

			is.Equal(builder.String(), tc.want)
		})
	}
}

func Test_writeNewLines(t *testing.T) {
	tt := []struct {
		name  string
		count int
		want  string
	}{
		{
			name:  "One line",
			count: 1,
			want:  "\n",
		},
		{
			name:  "Two lines",
			count: 2,
			want:  "\n\n",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)

			builder := strings.Builder{}

			writeNewLines(tc.count, &builder)

			is.Equal(builder.String(), tc.want)
		})
	}
}

func Test_writeZodEnum(t *testing.T) {
	tt := []struct {
		name  string
		field Field
		want  string
	}{
		{
			name: "Enum field",
			field: Field{
				Metadata: map[string]interface{}{
					"options": []interface{}{"one", "two", "three"},
				},
			},
			want: "z.enum([\"one\", \"two\", \"three\"])",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)
			builder := strings.Builder{}

			writeZodEnum(tc.field, &builder)

			is.Equal(builder.String(), tc.want)
		})
	}
}

func Test_writeZodRecord(t *testing.T) {
	tt := []struct {
		name       string
		definition Definition
		field      Field
		want       string
	}{
		{
			name:       "String to string record",
			definition: Definition{},
			field: Field{
				Type: FieldType{
					Map: FieldTypeMap{
						KeyTypeTS:     "string",
						ElementType:   "string",
						ElementTypeTS: "string",
					},
				},
			},
			want: "z.record(z.string(), z.string())",
		},
		{
			name:       "String to string array",
			definition: Definition{},
			field: Field{
				Type: FieldType{
					Map: FieldTypeMap{
						KeyTypeTS:         "string",
						ElementType:       "string",
						ElementTypeTS:     "string",
						ElementIsMultiple: true,
					},
				},
			},
			want: "z.record(z.string(), z.string().array())",
		},
		{
			name: "String to object record",
			definition: Definition{
				Objects: []Object{
					{
						Name: "greeting",
					},
				},
			},
			field: Field{
				Type: FieldType{
					Map: FieldTypeMap{
						KeyTypeTS:   "string",
						ElementType: "greeting",
					},
				},
			},
			want: "z.record(z.string(), greetingSchema)",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)
			builder := strings.Builder{}

			tc.definition.writeZodRecord(tc.field, &builder)

			is.Equal(builder.String(), tc.want)
		})
	}
}

func Test_writeZodObject(t *testing.T) {
	tt := []struct {
		name       string
		definition Definition
		field      Field
		want       string
	}{
		{
			name: "Object",
			field: Field{
				Type: FieldType{
					CleanObjectName: "Greeting",
				},
			},
			want: "greetingSchema",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)
			builder := strings.Builder{}

			writeZodObject(tc.field, &builder)

			is.Equal(builder.String(), tc.want)
		})
	}
}

func Test_writeZodBaseObject(t *testing.T) {
	tt := []struct {
		name       string
		definition Definition
		fields     []Field
		objectName string
		want       string
	}{
		{
			name:       "Excluded field",
			definition: Definition{},
			fields: []Field{
				{
					Metadata: map[string]interface{}{
						"exclude": true,
					},
				},
			},
			objectName: "GreetRequest",
			want: `z.object({
})`,
		},
		{
			name:       "Extended field",
			definition: Definition{},
			fields: []Field{
				{
					Metadata: map[string]interface{}{
						"extend": true,
					},
				},
			},
			objectName: "GreetRequest",
			want: `z.object({
})`,
		},
		{
			name:       "Basic types",
			definition: Definition{},
			fields: []Field{
				{
					NameLowerSnake: "string",
					Metadata: map[string]interface{}{
						"type": "types.String",
					},
				},
				{
					NameLowerSnake: "int",
					Metadata: map[string]interface{}{
						"type": "types.Int",
					},
				},
				{
					NameLowerSnake: "int_16",
					Metadata: map[string]interface{}{
						"type": "types.Int16",
					},
				},
				{
					NameLowerSnake: "int_32",
					Metadata: map[string]interface{}{
						"type": "types.Int32",
					},
				},
				{
					NameLowerSnake: "int_64",
					Metadata: map[string]interface{}{
						"type": "types.Int64",
					},
				},
				{
					NameLowerSnake: "float_64",
					Metadata: map[string]interface{}{
						"type": "types.Float64",
					},
				},
				{
					NameLowerSnake: "bool",
					Metadata: map[string]interface{}{
						"type": "types.Bool",
					},
				},
				{
					NameLowerSnake: "time",
					Metadata: map[string]interface{}{
						"type": "types.Time",
					},
				},
				{
					NameLowerSnake: "date",
					Metadata: map[string]interface{}{
						"type": "types.Date",
					},
				},
				{
					NameLowerSnake: "timestamp",
					Metadata: map[string]interface{}{
						"type": "types.Timestamp",
					},
				},
				{
					NameLowerSnake: "uuid",
					Metadata: map[string]interface{}{
						"type": "types.UUID",
					},
				},
				{
					NameLowerSnake: "rich_text",
					Metadata: map[string]interface{}{
						"type": "types.RichText",
					},
				},
			},
			objectName: "GreetRequest",
			want: `z.object({
	string: ZodTypes.String,
	int: ZodTypes.Int,
	int_16: ZodTypes.Int16,
	int_32: ZodTypes.Int32,
	int_64: ZodTypes.Int64,
	float_64: ZodTypes.Float64,
	bool: ZodTypes.Bool,
	time: ZodTypes.Time,
	date: ZodTypes.Date,
	timestamp: ZodTypes.Timestamp,
	uuid: ZodTypes.UUID,
	rich_text: ZodTypes.RichText,
})`,
		},
		{
			name:       "Object",
			definition: Definition{},
			fields: []Field{
				{
					Name:           "Greeting",
					NameLowerSnake: "greeting",
					Type: FieldType{
						IsObject:        true,
						CleanObjectName: "Greeting",
					},
				},
			},
			objectName: "GreetRequest",
			want: `z.object({
	greeting: greetingSchema,
})`,
		},
		{
			name: "Object",
			definition: Definition{
				Objects: []Object{
					{
						Name: "Greeting",
					},
				},
			},
			fields: []Field{
				{
					Name:           "Greeting",
					NameLowerSnake: "greeting",
					Type: FieldType{
						IsMap: true,
						Map: FieldTypeMap{
							KeyTypeTS:   "string",
							ElementType: "Greeting",
						},
					},
				},
			},
			objectName: "GreetRequest",
			want: `z.object({
	greeting: z.record(z.string(), greetingSchema),
})`,
		},
		{
			name:       "Enum",
			definition: Definition{},
			fields: []Field{
				{
					NameLowerSnake: "greeting_options",
					Metadata: map[string]interface{}{
						"options": []interface{}{"one", "two", "three"},
					},
				},
			},
			objectName: "GreetRequest",
			want: `z.object({
	greeting_options: z.enum(["one", "two", "three"]),
})`,
		},
		{
			name:       "Modifiers",
			definition: Definition{},
			fields: []Field{
				{
					NameLowerSnake: "string",
					Metadata: map[string]interface{}{
						"type":     "types.String",
						"nullable": true,
					},
				},
				{
					NameLowerSnake: "string",
					Metadata: map[string]interface{}{
						"type":     "types.String",
						"optional": true,
					},
				},
				{
					NameLowerSnake: "string",
					Metadata: map[string]interface{}{
						"type": "types.String",
					},
					Type: FieldType{
						Multiple:      true,
						MultipleTimes: []struct{}{{}},
					},
				},
				{
					NameLowerSnake: "string",
					Metadata: map[string]interface{}{
						"type":     "types.String",
						"nullable": true,
						"optional": true,
					},
					Type: FieldType{
						Multiple:      true,
						MultipleTimes: []struct{}{{}},
					},
				},
			},
			objectName: "GreetRequest",
			want: `z.object({
	string: ZodTypes.String.nullable(),
	string: ZodTypes.String.optional(),
	string: ZodTypes.String.array(),
	string: ZodTypes.String.array().nullable().optional(),
})`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			builder := strings.Builder{}

			tc.definition.writeZodBaseObject(tc.fields, tc.objectName, &builder)

			assert.Equal(tc.want, builder.String())
		})
	}
}

func Test_writeExtendedRecursiveZodObject(t *testing.T) {
	tt := []struct {
		name       string
		fields     []Field
		objectName string
		want       string
	}{
		{
			name: "Recursive object",
			fields: []Field{
				{
					NameLowerSnake: "Greeting",
				},
			},
			objectName: "Greeting",
			want: `export const greetingSchema: z.ZodType<GreetingRecursive> = greetingBaseSchema.extend({
	Greeting: z.lazy(() => greetingSchema),
})`,
		},
		{
			name: "Recursive array object",
			fields: []Field{
				{
					NameLowerSnake: "Greeting",
					Type: FieldType{
						Multiple:      true,
						MultipleTimes: []struct{}{{}},
					},
				},
			},
			objectName: "Greeting",
			want: `export const greetingSchema: z.ZodType<GreetingRecursive> = greetingBaseSchema.extend({
	Greeting: z.lazy(() => greetingSchema).array(),
})`,
		},
		{
			name: "Recursive nullable object",
			fields: []Field{
				{
					NameLowerSnake: "Greeting",
					Metadata: map[string]interface{}{
						"nullable": true,
					},
				},
			},
			objectName: "Greeting",
			want: `export const greetingSchema: z.ZodType<GreetingRecursive> = greetingBaseSchema.extend({
	Greeting: z.lazy(() => greetingSchema).nullable(),
})`,
		},
		{
			name: "Recursive optional object",
			fields: []Field{
				{
					NameLowerSnake: "Greeting",
					Metadata: map[string]interface{}{
						"optional": true,
					},
				},
			},
			objectName: "Greeting",
			want: `export const greetingSchema: z.ZodType<GreetingRecursive> = greetingBaseSchema.extend({
	Greeting: z.lazy(() => greetingSchema).optional(),
})`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			builder := strings.Builder{}

			writeExtendedRecursiveZodObject(tc.fields, tc.objectName, &builder)

			assert.Equal(tc.want, builder.String())
		})
	}
}

func Test_writeRecursiveType(t *testing.T) {
	tt := []struct {
		name   string
		fields []Field
		object Object
		want   string
	}{
		{
			name: "Recursive object",
			fields: []Field{
				{
					NameLowerSnake: "greeting",
				},
			},
			object: Object{
				Name:           "Greeting",
				NameLowerCamel: "greeting",
			},
			want: `type GreetingRecursive = z.infer<typeof greetingBaseSchema> & {
	greeting: GreetingRecursive;
};`,
		},
		{
			name: "Recursive array object",
			fields: []Field{
				{
					NameLowerSnake: "greeting",
					Type: FieldType{
						Multiple:      true,
						MultipleTimes: []struct{}{{}},
					},
				},
			},
			object: Object{
				Name:           "Greeting",
				NameLowerCamel: "greeting",
			},
			want: `type GreetingRecursive = z.infer<typeof greetingBaseSchema> & {
	greeting: GreetingRecursive[];
};`,
		},
		{
			name: "Recursive nullable object",
			fields: []Field{
				{
					NameLowerSnake: "greeting",
					Metadata: map[string]interface{}{
						"nullable": true,
					},
				},
			},
			object: Object{
				Name:           "Greeting",
				NameLowerCamel: "greeting",
			},
			want: `type GreetingRecursive = z.infer<typeof greetingBaseSchema> & {
	greeting: GreetingRecursive | null;
};`,
		},
		{
			name: "Recursive optional object",
			fields: []Field{
				{
					NameLowerSnake: "greeting",
					Metadata: map[string]interface{}{
						"optional": true,
					},
				},
			},
			object: Object{
				Name:           "Greeting",
				NameLowerCamel: "greeting",
			},
			want: `type GreetingRecursive = z.infer<typeof greetingBaseSchema> & {
	greeting?: GreetingRecursive;
};`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			builder := strings.Builder{}

			writeRecursiveType(tc.fields, &tc.object, &builder)

			assert.Equal(tc.want, builder.String())
		})
	}
}

func Test_writeZodEnpointSchemaObject(t *testing.T) {}

func Test_getMergeString(t *testing.T) {
	tt := []struct {
		name   string
		fields []string
		want   string
	}{
		{
			name:   "One field",
			fields: []string{"field1"},
			want:   ".merge(field1)",
		},
		{
			name:   "Two fields",
			fields: []string{"field1", "field2"},
			want:   ".merge(field1).merge(field2)",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			mergeString := getMergeString(tc.fields)

			assert.Equal(tc.want, mergeString)
		})
	}
}

func Test_getExtendedFields(t *testing.T) {
	tt := []struct {
		name   string
		fields []Field
		want   []string
	}{
		{
			name: "One field extended",
			fields: []Field{
				{
					Type: FieldType{
						CleanObjectName: "Greeting",
					},
					Metadata: map[string]interface{}{
						"extend": true,
					},
				},
				{
					Type: FieldType{
						CleanObjectName: "SecondGreeting",
					},
				},
			},
			want: []string{"greetingSchema"},
		},
		{
			name: "Two fields extended",
			fields: []Field{
				{
					Type: FieldType{
						CleanObjectName: "Greeting",
					},
					Metadata: map[string]interface{}{
						"extend": true,
					},
				},
				{
					Type: FieldType{
						CleanObjectName: "SecondGreeting",
					},
					Metadata: map[string]interface{}{
						"extend": true,
					},
				},
			},
			want: []string{"greetingSchema", "secondGreetingSchema"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			extendedFields := getExtendedFields(tc.fields)

			assert.Equal(tc.want, extendedFields)
		})
	}
}

func Test_getRecursiveFields(t *testing.T) {
	tt := []struct {
		name       string
		objectName string
		fields     []Field
		want       []Field
	}{
		{
			name:       "One recursive field",
			objectName: "Greeting",
			fields: []Field{
				{
					Type: FieldType{
						IsObject:        true,
						CleanObjectName: "Greeting",
					},
				},
				{
					Type: FieldType{
						IsObject:        false,
						CleanObjectName: "SecondGreeting",
					},
				},
			},
			want: []Field{{
				Type: FieldType{
					IsObject:        true,
					CleanObjectName: "Greeting",
				},
			}},
		},
		{
			name:       "Two recursive fields",
			objectName: "Greeting",
			fields: []Field{
				{
					Type: FieldType{
						IsObject:        true,
						CleanObjectName: "Greeting",
					},
				},
				{
					Type: FieldType{
						IsObject:        true,
						CleanObjectName: "Greeting",
					},
				},
			},
			want: []Field{
				{
					Type: FieldType{
						IsObject:        true,
						CleanObjectName: "Greeting",
					},
				},
				{
					Type: FieldType{
						IsObject:        true,
						CleanObjectName: "Greeting",
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			extendedFields := getRecursiveFields(tc.fields, tc.objectName)

			assert.Equal(tc.want, extendedFields)
		})
	}
}

func Test_removePackagePrefix(t *testing.T) {
	tt := []struct {
		name       string
		objectName string
		want       string
	}{
		{
			name:       "With package prefix",
			objectName: "Service.Greeting",
			want:       "Greeting",
		},
		{
			name:       "Without package prefix",
			objectName: "Greeting",
			want:       "Greeting",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			objectName := removePackagePrefix(tc.objectName)

			assert.Equal(tc.want, objectName)
		})
	}
}

func Test_getTypeNameForZod(t *testing.T) {
	tt := []struct {
		name      string
		fieldType string
		want      string
	}{
		{
			name:      "String",
			fieldType: "types.String",
			want:      "ZodTypes.String",
		},
		{
			name:      "Int",
			fieldType: "types.Int",
			want:      "ZodTypes.Int",
		},
		{
			name:      "Int16",
			fieldType: "types.Int16",
			want:      "ZodTypes.Int16",
		},
		{
			name:      "Int64",
			fieldType: "types.Int64",
			want:      "ZodTypes.Int64",
		},
		{
			name:      "Date",
			fieldType: "types.Date",
			want:      "ZodTypes.Date",
		},
		{
			name:      "Timestamp",
			fieldType: "types.Timestamp",
			want:      "ZodTypes.Timestamp",
		},
		{
			name:      "Time",
			fieldType: "types.Time",
			want:      "ZodTypes.Time",
		},
		{
			name:      "UUID",
			fieldType: "types.UUID",
			want:      "ZodTypes.UUID",
		},
		{
			name:      "Bool",
			fieldType: "types.Bool",
			want:      "ZodTypes.Bool",
		},
		{
			name:      "JSON",
			fieldType: "types.JSON",
			want:      "ZodTypes.JSON",
		},
		{
			name:      "Float64",
			fieldType: "types.Float64",
			want:      "ZodTypes.Float64",
		},
		{
			name:      "RichText",
			fieldType: "types.RichText",
			want:      "ZodTypes.RichText",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			objectName := getTypeNameForZod(tc.fieldType)

			assert.Equal(tc.want, objectName)
		})
	}

	t.Run("Invalid type", func(t *testing.T) {
		assert := assert.New(t)

		assert.Panics(func() {
			getTypeNameForZod("InvalidType")
		})
	})
}

func Test_writeZodEndpointSchemaObject(t *testing.T) {
	tt := []struct {
		name       string
		definition Definition
		objectName string
		want       string
	}{
		{
			name: "Simple object",
			definition: Definition{
				Objects: []Object{
					{
						Name: "GreetRequest",
						Fields: []Field{
							{
								NameLowerSnake: "string",
								Metadata: map[string]interface{}{
									"type": "types.String",
								},
							},
						},
					},
				},
			},
			objectName: "GreetRequest",
			want: `export const greetRequestSchema = z.object({
	string: ZodTypes.String,
});

`,
		},
		{
			name: "With field object",
			definition: Definition{
				Objects: []Object{
					{
						Name: "GreetRequest",
						Fields: []Field{
							{
								NameLowerSnake: "query",
								Type: FieldType{
									IsObject:        true,
									CleanObjectName: "GreetRequestQuery",
								},
							},
						},
					},
					{
						Name: "GreetRequestQuery",
						Fields: []Field{
							{
								NameLowerSnake: "string",
								Metadata: map[string]interface{}{
									"type": "types.String",
								},
							},
						},
					},
				},
			},
			objectName: "GreetRequest",
			want: `export const greetRequestQuerySchema = z.object({
	string: ZodTypes.String,
});

export const greetRequestSchema = z.object({
	query: greetRequestQuerySchema,
});

`,
		},
		{
			name: "With extended object",
			definition: Definition{
				Objects: []Object{
					{
						Name: "GreetRequest",
						Fields: []Field{
							{
								NameLowerSnake: "string",
								Metadata: map[string]interface{}{
									"type": "types.String",
								},
							},
							{
								NameLowerSnake: "greeting",
								Type: FieldType{
									IsObject:        true,
									CleanObjectName: "Greeting",
								},
								Metadata: map[string]interface{}{
									"extend": true,
								},
							},
						},
					},
					{
						Name: "Greeting",
					},
				},
			},
			objectName: "GreetRequest",
			want: `export const greetingSchema = z.object({
});

export const greetRequestSchema = z.object({
	string: ZodTypes.String,
}).merge(greetingSchema);

`,
		},
		{
			name: "With recursive object",
			definition: Definition{
				Objects: []Object{
					{
						Name:           "Greeting",
						NameLowerCamel: "greeting",
						Fields: []Field{
							{
								NameLowerSnake: "greeting",
								Type: FieldType{
									IsObject:        true,
									CleanObjectName: "Greeting",
								},
							},
						},
					},
				},
			},
			objectName: "Greeting",
			want: `const greetingBaseSchema = z.object({
});

type GreetingRecursive = z.infer<typeof greetingBaseSchema> & {
	greeting: GreetingRecursive;
};

export const greetingSchema: z.ZodType<GreetingRecursive> = greetingBaseSchema.extend({
	greeting: z.lazy(() => greetingSchema),
});

`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			builder := strings.Builder{}

			tc.definition.writeZodEndpointSchemaObject(tc.objectName, &builder, make(map[string]struct{}))

			assert.Equal(tc.want, builder.String())
		})
	}
}

func Test_ZodEndpointSchema(t *testing.T) {
	tt := []struct {
		name       string
		definition Definition
		objectName string
		want       string
	}{
		{
			name: "Simple schema",
			definition: Definition{
				Objects: []Object{
					{
						Name: "GreetRequest",
						Fields: []Field{
							{
								NameLowerSnake: "string",
								Metadata: map[string]interface{}{
									"type": "types.String",
								},
							},
						},
					},
				},
			},
			objectName: "GreetRequest",
			want: `import { z } from "zod";
import ZodTypes from "./zod_types.gen";

export const greetRequestSchema = z.object({
	string: ZodTypes.String,
});

`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			html := tc.definition.ZodEndpointSchema()

			builder := strings.Builder{}

			builder.WriteString(tc.want)

			assert.Equal(template.HTML(builder.String()), html)
		})
	}
}
