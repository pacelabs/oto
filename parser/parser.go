package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/doc"
	"go/token"
	"go/types"
	"html/template"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/fatih/structtag"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

// ErrNotFound is returned when an Object is not found.
var ErrNotFound = errors.New("not found")

// Definition describes an Oto definition.
type Definition struct {
	// PackageName is the name of the package.
	PackageName string `json:"packageName"`
	// Services are the services described in this definition.
	Services []Service `json:"services"`
	// Objects are the structures that are used throughout this definition.
	Objects []Object `json:"objects"`
	// Imports is a map of Go imports that should be imported into
	// Go code.
	Imports map[string]string `json:"imports"`
}

// Object looks up an object by name. Returns ErrNotFound error
// if it cannot find it.
func (d *Definition) Object(name string) (*Object, error) {
	for i := range d.Objects {
		obj := &d.Objects[i]
		if obj.Name == name {
			return obj, nil
		}
	}
	return nil, ErrNotFound
}

// ObjectIsInput gets whether this object is a method
// input (request) type or not.\
// Returns true if any method.InputObject.ObjectName matches
// name.
func (d *Definition) ObjectIsInput(name string) bool {
	for _, service := range d.Services {
		for _, method := range service.Methods {
			if method.InputObject.ObjectName == name {
				return true
			}
		}
	}
	return false
}

// MethodHasPagination checks if the object given by name, has pagination.
// The object has pagination if it is an output object and has a field named TotalCount of the type int64
// and the input object has query.
func (d *Definition) MethodHasPagination(method Method) bool {
	outObj, err := d.Object(method.OutputObject.TypeName)
	if err != nil {
		panic(err)
	}

	inObj, err := d.Object(method.InputObject.TypeName)
	if err != nil {
		panic(err)
	}

	// Should be an output object and input object
	if !d.ObjectIsOutput(outObj.Name) || !d.ObjectIsInput(inObj.Name) {
		return false
	}

	outputHasTotalCount := false

	for _, field := range outObj.Fields {
		// Should have a field named TotalCount of the type int64.
		if field.Name == "TotalCount" && field.Type.CleanObjectName == "int64" {
			outputHasTotalCount = true
			break
		}
	}

	inputHasQuery := false
	for _, field := range inObj.Fields {
		// Should have a field named Query
		if field.Name == "Query" {
			inputHasQuery = true
			break
		}
	}

	return outputHasTotalCount && inputHasQuery
}

// ObjectIsOutput gets whether this object is a method
// output (response) type or not.
// Returns true if any method.OutputObject.ObjectName matches
// name.
func (d *Definition) ObjectIsOutput(name string) bool {
	for _, service := range d.Services {
		for _, method := range service.Methods {
			if method.OutputObject.ObjectName == name {
				return true
			}
		}
	}
	return false
}

func (d *Definition) ZodEndpointSchema() template.HTML {
	// Store the objects that has been generated
	generated := make(map[string]struct{})

	builder := &strings.Builder{}
	builder.WriteString("import { z } from \"zod\";")
	writeNewLines(1, builder)
	builder.WriteString("import ZodTypes from \"./zod_types.gen\";")
	writeNewLines(2, builder)

	for _, object := range d.Objects {
		d.writeZodEndpointSchemaObject(object.Name, builder, generated)
	}

	return template.HTML(builder.String())
}

func getTypeNameForZod(fieldType string) string {
	if !strings.HasPrefix(fieldType, "types.") {
		panic("invalid field type: " + fieldType)
	}

	return "ZodTypes." + strings.TrimPrefix(fieldType, "types.")
}

func removePackagePrefix(variable string) string {
	if strings.Contains(variable, ".") {
		variable = strings.TrimPrefix(filepath.Ext(variable), ".")
	}

	return variable
}

func getRecursiveFields(objectFields []Field, objectName string) []Field {
	recursiveFields := make([]Field, 0)
	for _, field := range objectFields {
		if field.Type.IsObject && removePackagePrefix(field.Type.CleanObjectName) == objectName {
			recursiveFields = append(recursiveFields, field)
		}
	}

	return recursiveFields
}

func getExtendedFields(objectFields []Field) []string {
	extendedFields := make([]string, 0)
	for _, field := range objectFields {
		if _, ok := field.Metadata["extend"]; ok {
			extendedFields = append(extendedFields, camelizeDown(removePackagePrefix(field.Type.CleanObjectName))+"Schema")
		}
	}

	return extendedFields
}

func getMergeString(extendedFields []string) string {
	mergeString := ""
	for _, field := range extendedFields {
		mergeString += ".merge(" + field + ")"
	}

	return mergeString
}

func (d *Definition) writeZodEndpointSchemaObject(objectName string, builder *strings.Builder, generated map[string]struct{}) {
	objectName = removePackagePrefix(objectName)

	// Skip if it has already been generated
	if _, ok := generated[objectName]; ok {
		return
	}

	generated[objectName] = struct{}{}

	object, err := d.Object(objectName)
	if err != nil {
		panic("cannot get object to generate zod endpoint schema for object " + objectName + " " + err.Error())
	}

	for _, field := range object.Fields {
		if _, ok := field.Metadata["exclude"]; ok {
			continue
		}

		if field.Type.IsObject {
			d.writeZodEndpointSchemaObject(field.Type.CleanObjectName, builder, generated)
		}

		if field.Type.IsMap {
			if _, err := d.Object(field.Type.Map.ElementType); err == nil {
				d.writeZodEndpointSchemaObject(field.Type.Map.ElementType, builder, generated)
			}
		}
	}

	recursiveFields := getRecursiveFields(object.Fields, objectName)

	if len(recursiveFields) > 0 {
		fmt.Fprintf(builder, "const %sBaseSchema = ", object.NameLowerCamel)
		d.writeZodBaseObject(object.Fields, objectName, builder)
		builder.WriteString(";")
		writeNewLines(2, builder)
	}

	extendedFields := getExtendedFields(object.Fields)

	if len(recursiveFields) > 0 {
		writeRecursiveType(recursiveFields, object, builder)
		writeNewLines(2, builder)
		writeExtendedRecursiveZodObject(recursiveFields, object.Name, builder)
	} else {
		fmt.Fprintf(builder, "export const %sSchema = ", camelizeDown(object.Name))
		d.writeZodBaseObject(object.Fields, objectName, builder)
	}

	if len(extendedFields) > 0 {
		mergeString := getMergeString(extendedFields)

		fmt.Fprintf(builder, "%s", mergeString)
	}

	builder.WriteString(";")
	writeNewLines(2, builder)
}

func writeRecursiveType(recursiveFields []Field, object *Object, builder *strings.Builder) {
	fmt.Fprintf(builder, "type %sRecursive = z.infer<typeof %sBaseSchema> & {", object.Name, object.NameLowerCamel)
	writeNewLines(1, builder)

	for _, field := range recursiveFields {
		builder.WriteString("\t")

		builder.WriteString(field.NameLowerSnake)

		if optional, ok := field.Metadata["optional"]; ok {
			if optional.(bool) {
				builder.WriteString("?")
			}
		}

		builder.WriteString(": ")
		fmt.Fprintf(builder, "%sRecursive", object.Name)

		if field.Type.Multiple {
			for i := 0; i < len(field.Type.MultipleTimes); i++ {
				builder.WriteString("[]")
			}
		}

		if nullable, ok := field.Metadata["nullable"]; ok {
			if nullable.(bool) {
				builder.WriteString(" | null")
			}
		}

		builder.WriteString(";")
		writeNewLines(1, builder)
	}
	builder.WriteString("};")
}

func writeExtendedRecursiveZodObject(fields []Field, objectName string, builder *strings.Builder) {
	fmt.Fprintf(builder, "export const %sSchema: z.ZodType<%sRecursive> = %sBaseSchema.extend({", camelizeDown(objectName), objectName, camelizeDown(objectName))
	for _, field := range fields {
		writeNewLines(1, builder)
		builder.WriteString("\t")
		builder.WriteString(field.NameLowerSnake + ": ")
		builder.WriteString("z.lazy(() => ")
		builder.WriteString(camelizeDown(objectName) + "Schema")
		builder.WriteString(")")
		writeZodFieldModifiers(field, builder)
		builder.WriteString(",")
		writeNewLines(1, builder)
	}

	builder.WriteString("})")
}

func (d *Definition) writeZodBaseObject(fields []Field, objectName string, builder *strings.Builder) {
	builder.WriteString("z.object({")
	writeNewLines(1, builder)

	for _, field := range fields {
		// Field is excluded
		if _, ok := field.Metadata["exclude"]; ok {
			continue
		}

		// Field is an extended field, we handle this separately
		if _, ok := field.Metadata["extend"]; ok {
			continue
		}

		// Field is a recursive field, we handle this separately
		if removePackagePrefix(field.Type.CleanObjectName) == objectName {
			continue
		}

		builder.WriteString("\t")
		builder.WriteString(field.NameLowerSnake + ": ")

		switch {
		case field.Type.IsObject:
			writeZodObject(field, builder)
		case field.Type.IsMap:
			d.writeZodRecord(field, builder)
		case field.Metadata["options"] != nil:
			writeZodEnum(field, builder)
		default:
			if customTypeName, ok := field.Metadata["type"].(string); ok {
				builder.WriteString(getTypeNameForZod(customTypeName))
			} else {
				builder.WriteString("z." + field.Type.JSType + "()")
			}
		}

		writeZodFieldModifiers(field, builder)

		if removePackagePrefix(field.Type.CleanObjectName) == objectName {
			builder.WriteString(")")
		}

		builder.WriteString(",")
		writeNewLines(1, builder)
	}

	builder.WriteString("})")
}

func writeZodObject(field Field, builder *strings.Builder) {
	builder.WriteString(camelizeDown(removePackagePrefix(field.Type.CleanObjectName)) + "Schema")
}

func (d *Definition) writeZodRecord(field Field, builder *strings.Builder) {
	builder.WriteString("z.record(")
	builder.WriteString("z." + field.Type.Map.KeyTypeTS + "(), ")

	_, err := d.Object(field.Type.Map.ElementType)
	if err == nil {
		builder.WriteString(camelizeDown(field.Type.Map.ElementType) + "Schema")
	} else {
		builder.WriteString("z." + field.Type.Map.ElementTypeTS + "()")
	}

	if field.Type.Map.ElementIsMultiple {
		builder.WriteString(".array()")
	}

	builder.WriteString(")")
}

func writeZodEnum(field Field, builder *strings.Builder) {
	options := make([]string, 0, len(field.Metadata["options"].([]interface{})))

	for _, option := range field.Metadata["options"].([]interface{}) {
		if s, ok := option.(string); ok {
			options = append(options, "\""+s+"\"")
		}
	}

	builder.WriteString("z.enum([" + strings.Join(options, ", ") + "])")
}

func writeNewLines(count int, builder *strings.Builder) {
	for i := 0; i < count; i++ {
		builder.WriteString("\n")
	}
}

func writeZodFieldModifiers(field Field, builder *strings.Builder) {
	if field.Type.Multiple {
		for i := 0; i < len(field.Type.MultipleTimes); i++ {
			builder.WriteString(".array()")
		}
	}

	if nullable, ok := field.Metadata["nullable"]; ok {
		if nullable.(bool) {
			builder.WriteString(".nullable()")
		}
	}

	if optional, ok := field.Metadata["optional"]; ok {
		if optional.(bool) {
			builder.WriteString(".optional()")
		}
	}
}

// Service describes a service, akin to an interface in Go.
type Service struct {
	Name    string   `json:"name"`
	Methods []Method `json:"methods"`
	Comment string   `json:"comment"`
	// Metadata are typed key/value pairs extracted from the
	// comments.
	Metadata map[string]interface{} `json:"metadata"`
}

// Method describes a method that a Service can perform.
type Method struct {
	Name           string    `json:"name"`
	NameLowerCamel string    `json:"nameLowerCamel"`
	NameLowerSnake string    `json:"nameLowerSnake"`
	InputObject    FieldType `json:"inputObject"`
	OutputObject   FieldType `json:"outputObject"`
	Comment        string    `json:"comment"`
	// Metadata are typed key/value pairs extracted from the
	// comments.
	Metadata map[string]interface{} `json:"metadata"`
}

// Object describes a data structure that is part of this definition.
type Object struct {
	TypeID         string  `json:"typeID"`
	Name           string  `json:"name"`
	NameLowerCamel string  `json:"nameLowerCamel"`
	NameLowerSnake string  `json:"nameLowerSnake"`
	Imported       bool    `json:"imported"`
	Fields         []Field `json:"fields"`
	Comment        string  `json:"comment"`
	// Metadata are typed key/value pairs extracted from the
	// comments.
	Metadata map[string]interface{} `json:"metadata"`
}

// Field describes the field inside an Object.
type Field struct {
	Name           string              `json:"name"`
	NameLowerCamel string              `json:"nameLowerCamel"`
	NameLowerSnake string              `json:"nameLowerSnake"`
	Type           FieldType           `json:"type"`
	OmitEmpty      bool                `json:"omitEmpty"`
	Comment        string              `json:"comment"`
	Tag            string              `json:"tag"`
	ParsedTags     map[string]FieldTag `json:"parsedTags"`
	Example        interface{}         `json:"example"`
	// Metadata are typed key/value pairs extracted from the
	// comments.
	Metadata map[string]interface{} `json:"metadata"`
}

// FieldTag is a parsed tag.
// For more information, see Struct Tags in Go.
type FieldTag struct {
	// Value is the value of the tag.
	Value string `json:"value"`
	// Options are the options for the tag.
	Options []string `json:"options"`
}

// FieldType holds information about the type of data that this
// Field stores.
type FieldType struct {
	TypeID     string `json:"typeID"`
	TypeName   string `json:"typeName"`
	ObjectName string `json:"objectName"`
	// CleanObjectName is the ObjectName with * removed
	// for pointer types.
	CleanObjectName      string `json:"cleanObjectName"`
	ObjectNameLowerCamel string `json:"objectNameLowerCamel"`
	ObjectNameLowerSnake string `json:"objectNameLowerSnake"`
	Multiple             bool   `json:"multiple"`
	MultipleTimes        []struct{}
	Package              string       `json:"package"`
	IsObject             bool         `json:"isObject"`
	JSType               string       `json:"jsType"`
	TSType               string       `json:"tsType"`
	SwiftType            string       `json:"swiftType"`
	IsMap                bool         `json:"is_map"`
	Map                  FieldTypeMap `json:"map"`
}

type FieldTypeMap struct {
	KeyType           string
	KeyTypeJS         string `json:"keyTypeJS"`
	KeyTypeTS         string `json:"keyTypeTS"`
	KeyTypeSwift      string `json:"keyTypeSwift"`
	ElementType       string `json:"ElementType"`
	ElementTypeJS     string `json:"elementTypeJS"`
	ElementTypeTS     string `json:"elementTypeTS"`
	ElementTypeSwift  string `json:"elementTypeSwift"`
	ElementIsMultiple bool   `json:"elementIsMultiple"`
}

// IsOptional returns true for pointer types (optional).
func (f FieldType) IsOptional() bool {
	return strings.HasPrefix(f.ObjectName, "*")
}

// Parser parses Oto Go definition packages.
type Parser struct {
	Verbose bool

	ExcludeInterfaces []string

	patterns []string
	def      Definition

	// outputObjects marks output object names.
	outputObjects map[string]struct{}
	// objects marks object names.
	objects map[string]struct{}

	// docs are the docs for extracting comments.
	docs *doc.Package
}

// New makes a fresh parser using the specified patterns.
// The patterns should be the args passed into the tool (after any flags)
// and will be passed to the underlying build system.
func New(patterns ...string) *Parser {
	return &Parser{
		patterns: patterns,
	}
}

// Parse parses the files specified, returning the definition.
func (p *Parser) Parse() (Definition, error) {
	cfg := &packages.Config{
		Mode:  packages.NeedTypes | packages.NeedName | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedName | packages.NeedSyntax,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, p.patterns...)
	if err != nil {
		return p.def, err
	}
	p.outputObjects = make(map[string]struct{})
	p.objects = make(map[string]struct{})
	var excludedObjectsTypeIDs []string
	for _, pkg := range pkgs {
		p.docs, err = doc.NewFromFiles(pkg.Fset, pkg.Syntax, "")
		if err != nil {
			panic(err)
		}
		p.def.PackageName = pkg.Name
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			switch item := obj.Type().Underlying().(type) {
			case *types.Interface:
				s, err := p.parseService(pkg, obj, item)
				if err != nil {
					return p.def, err
				}
				if isInSlice(p.ExcludeInterfaces, name) {
					for _, method := range s.Methods {
						excludedObjectsTypeIDs = append(excludedObjectsTypeIDs, method.InputObject.TypeID)
						excludedObjectsTypeIDs = append(excludedObjectsTypeIDs, method.OutputObject.TypeID)
					}
					continue
				}
				p.def.Services = append(p.def.Services, s)
			case *types.Struct:
				p.parseObject(pkg, obj, item)
			}
		}
	}
	// remove any excluded objects
	nonExcludedObjects := make([]Object, 0, len(p.def.Objects))
	for _, object := range p.def.Objects {
		excluded := false
		for _, excludedTypeID := range excludedObjectsTypeIDs {
			if object.TypeID == excludedTypeID {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}
		nonExcludedObjects = append(nonExcludedObjects, object)
	}
	p.def.Objects = nonExcludedObjects
	// sort services
	sort.Slice(p.def.Services, func(i, j int) bool {
		return p.def.Services[i].Name < p.def.Services[j].Name
	})
	if err := p.addOutputFields(); err != nil {
		return p.def, err
	}
	return p.def, nil
}

func (p *Parser) parseService(pkg *packages.Package, obj types.Object, interfaceType *types.Interface) (Service, error) {
	var s Service
	s.Name = obj.Name()
	s.Comment = p.commentForType(s.Name)
	var err error
	s.Metadata, s.Comment, err = p.extractCommentMetadata(s.Comment)
	if err != nil {
		return s, p.wrapErr(errors.New("extract comment metadata"), pkg, obj.Pos())
	}
	if p.Verbose {
		fmt.Printf("%s ", s.Name)
	}
	l := interfaceType.NumMethods()
	for i := 0; i < l; i++ {
		m := interfaceType.Method(i)
		method, err := p.parseMethod(pkg, s.Name, m)
		if err != nil {
			return s, err
		}
		s.Methods = append(s.Methods, method)
	}
	return s, nil
}

func (p *Parser) parseMethod(pkg *packages.Package, serviceName string, methodType *types.Func) (Method, error) {
	var m Method
	m.Name = methodType.Name()
	m.NameLowerCamel = camelizeDown(m.Name)
	m.NameLowerSnake = snakeDown(m.Name)
	m.Comment = p.commentForMethod(serviceName, m.Name)
	var err error
	m.Metadata, m.Comment, err = p.extractCommentMetadata(m.Comment)
	if err != nil {
		return m, p.wrapErr(errors.New("extract comment metadata"), pkg, methodType.Pos())
	}
	sig := methodType.Type().(*types.Signature)
	inputParams := sig.Params()
	if inputParams.Len() != 1 {
		return m, p.wrapErr(errors.New("invalid method signature: expected Method(MethodRequest) MethodResponse"), pkg, methodType.Pos())
	}
	m.InputObject, err = p.parseFieldType(pkg, inputParams.At(0))
	if err != nil {
		return m, errors.Wrap(err, "parse input object type")
	}
	outputParams := sig.Results()
	if outputParams.Len() != 1 {
		return m, p.wrapErr(errors.New("invalid method signature: expected Method(MethodRequest) MethodResponse"), pkg, methodType.Pos())
	}
	m.OutputObject, err = p.parseFieldType(pkg, outputParams.At(0))
	if err != nil {
		return m, errors.Wrap(err, "parse output object type")
	}
	p.outputObjects[m.OutputObject.TypeName] = struct{}{}
	return m, nil
}

// parseObject parses a struct type and adds it to the Definition.
func (p *Parser) parseObject(pkg *packages.Package, o types.Object, v *types.Struct) error {
	var obj Object
	obj.Name = o.Name()
	obj.NameLowerCamel = camelizeDown(obj.Name)
	obj.NameLowerSnake = snakeDown(obj.Name)
	obj.Comment = p.commentForType(obj.Name)
	var err error
	obj.Metadata, obj.Comment, err = p.extractCommentMetadata(obj.Comment)
	if err != nil {
		return p.wrapErr(errors.New("extract comment metadata"), pkg, o.Pos())
	}
	if _, found := p.objects[obj.Name]; found {
		// if this has already been parsed, skip it
		return nil
	}
	if o.Pkg().Name() != pkg.Name {
		obj.Imported = true
	}
	typ := v.Underlying()
	st, ok := typ.(*types.Struct)
	if !ok {
		return p.wrapErr(errors.New(obj.Name+" must be a struct"), pkg, o.Pos())
	}
	obj.TypeID = o.Pkg().Path() + "." + obj.Name
	obj.Fields = []Field{}
	for i := 0; i < st.NumFields(); i++ {
		field, err := p.parseField(pkg, obj.Name, st.Field(i), st.Tag(i))
		if err != nil {
			return err
		}
		field.Tag = v.Tag(i)
		field.ParsedTags, err = p.parseTags(field.Tag)
		if err != nil {
			return errors.Wrap(err, "parse field tag")
		}
		obj.Fields = append(obj.Fields, field)
	}
	p.def.Objects = append(p.def.Objects, obj)
	p.objects[obj.Name] = struct{}{}
	return nil
}

func (p *Parser) parseTags(tag string) (map[string]FieldTag, error) {
	tags, err := structtag.Parse(tag)
	if err != nil {
		return nil, err
	}
	fieldTags := make(map[string]FieldTag)
	for _, tag := range tags.Tags() {
		fieldTags[tag.Key] = FieldTag{
			Value:   tag.Name,
			Options: tag.Options,
		}
	}
	return fieldTags, nil
}

func (p *Parser) parseField(pkg *packages.Package, objectName string, v *types.Var, tag string) (Field, error) {
	var f Field
	f.Name = v.Name()
	f.NameLowerCamel = camelizeDown(f.Name)
	f.NameLowerSnake = snakeDown(f.Name)
	// if it has a json tag, use that as the NameJSON.
	if tag != "" {
		fieldTag := reflect.StructTag(tag)
		jsonTag := fieldTag.Get("json")
		if jsonTag != "" {
			f.NameLowerCamel = strings.Split(jsonTag, ",")[0]
			f.NameLowerSnake = strings.Split(jsonTag, ",")[0]
		}
	}
	f.Comment = p.commentForField(objectName, f.Name)
	f.Metadata = map[string]interface{}{}
	if !v.Exported() {
		return f, p.wrapErr(errors.New(f.Name+" must be exported"), pkg, v.Pos())
	}
	var err error
	f.Metadata, f.Comment, err = p.extractCommentMetadata(f.Comment)
	if err != nil {
		return f, p.wrapErr(errors.New("extract comment metadata"), pkg, v.Pos())
	}
	f.Type, err = p.parseFieldType(pkg, v)
	if err != nil {
		return f, errors.Wrap(err, "parse type")
	}
	example, ok := f.Metadata["example"]
	if !ok {
		switch f.Type.TypeName {
		case "interface{}":
			example = struct{}{}
		case "map[string]interface{}":
			example = map[string]interface{}{
				"string": "value",
				"int":    88,
				"object": map[string]interface{}{},
			}
		case "string":
			example = "text"
		case "bool":
			example = true
		case "int", "int16", "int32", "int64",
			"uint", "uint16", "uint32", "uint64":
			example = 334
		case "float32", "float64":
			example = 1.235
		default:
			example = nil
		}
		if f.Type.Multiple {
			example = []interface{}{example}
		}
		if f.Type.IsObject && strings.HasPrefix(f.Type.TypeName, "*") {
			example = struct{}{}
		}
	}
	f.Example = example
	return f, nil
}

func (p *Parser) parseFieldType(pkg *packages.Package, obj types.Object) (FieldType, error) {
	var ftype FieldType
	pkgPath := pkg.PkgPath
	resolver := func(other *types.Package) string {
		if other.Name() != pkg.Name {
			if p.def.Imports == nil {
				p.def.Imports = make(map[string]string)
			}
			p.def.Imports[other.Path()] = other.Name()
			ftype.Package = other.Path()
			pkgPath = other.Path()
			return other.Name()
		}
		return "" // no package prefix
	}

	typ := obj.Type()
	for {
		slice, ok := typ.(*types.Slice)
		if !ok {
			break
		}

		typ = slice.Elem()
		ftype.Multiple = true
		ftype.MultipleTimes = append(ftype.MultipleTimes, struct{}{})
	}

	originalTyp := typ
	pointerType, isPointer := typ.(*types.Pointer)
	if isPointer {
		typ = pointerType.Elem()
		if slice, ok := typ.(*types.Slice); ok {
			typ = slice.Elem()
			ftype.Multiple = true
		}
	}
	if named, ok := typ.(*types.Named); ok {
		if structure, ok := named.Underlying().(*types.Struct); ok {
			if !isPointer {
				if err := p.parseObject(pkg, named.Obj(), structure); err != nil {
					return ftype, err
				}
			}
			ftype.IsObject = true
		}
	}
	mapType, isMap := typ.(*types.Map)
	if isMap {
		keyType := mapType.Key()
		elementType := mapType.Elem()

		ftype.IsMap = true

		ftype.Map.KeyType = types.TypeString(keyType, resolver)
		ftype.Map.KeyTypeJS = ftype.Map.KeyType
		ftype.Map.KeyTypeSwift = ftype.Map.KeyType
		ftype.Map.KeyTypeTS = ftype.Map.KeyType

		switch ftype.Map.KeyType {
		case "interface{}":
			ftype.Map.KeyTypeJS = "any"
			ftype.Map.KeyTypeSwift = "Any"
			ftype.Map.KeyTypeTS = "any"
		case "map[string]interface{}":
			ftype.Map.KeyTypeJS = "object"
			ftype.Map.KeyTypeTS = "object"
			ftype.Map.KeyTypeSwift = "Any"
		case "string":
			ftype.Map.KeyTypeJS = "string"
			ftype.Map.KeyTypeSwift = "String"
			ftype.Map.KeyTypeTS = "string"
		case "bool":
			ftype.Map.KeyTypeJS = "boolean"
			ftype.Map.KeyTypeSwift = "Bool"
			ftype.Map.KeyTypeTS = "boolean"
		case "int", "int16", "int32", "int64",
			"uint", "uint16", "uint32", "uint64",
			"float32", "float64":
			ftype.Map.KeyTypeJS = "number"
			ftype.Map.KeyTypeSwift = "Double"
			ftype.Map.KeyTypeTS = "number"
		}

		ftype.Map.ElementType = types.TypeString(elementType, resolver)
		if slice, ok := elementType.(*types.Slice); ok {
			ftype.Map.ElementType = types.TypeString(slice.Elem(), resolver)
			ftype.Map.ElementIsMultiple = true
		}
		ftype.Map.ElementTypeJS = ftype.Map.ElementType
		ftype.Map.ElementTypeSwift = ftype.Map.ElementType
		ftype.Map.ElementTypeTS = ftype.Map.ElementType

		switch ftype.Map.ElementType {
		case "interface{}":
			ftype.Map.ElementTypeJS = "any"
			ftype.Map.ElementTypeSwift = "Any"
			ftype.Map.ElementTypeTS = "any"
		case "map[string]interface{}":
			ftype.Map.ElementTypeJS = "object"
			ftype.Map.ElementTypeTS = "object"
			ftype.Map.ElementTypeSwift = "Any"
		case "string":
			ftype.Map.ElementTypeJS = "string"
			ftype.Map.ElementTypeSwift = "String"
			ftype.Map.ElementTypeTS = "string"
		case "bool":
			ftype.Map.ElementTypeJS = "boolean"
			ftype.Map.ElementTypeSwift = "Bool"
			ftype.Map.ElementTypeTS = "boolean"
		case "int", "int16", "int32", "int64",
			"uint", "uint16", "uint32", "uint64",
			"float32", "float64":
			ftype.Map.ElementTypeJS = "number"
			ftype.Map.ElementTypeSwift = "Double"
			ftype.Map.ElementTypeTS = "number"
		}
	}
	// disallow nested structs
	switch typ.(type) {
	case *types.Struct:
		return ftype, p.wrapErr(errors.New("nested structs not supported (create another type instead)"), pkg, obj.Pos())
	}

	ftype.TypeName = types.TypeString(originalTyp, resolver)
	ftype.ObjectName = types.TypeString(originalTyp, func(other *types.Package) string { return "" })
	ftype.ObjectNameLowerCamel = camelizeDown(ftype.ObjectName)
	ftype.ObjectNameLowerSnake = snakeDown(ftype.ObjectName)
	ftype.TypeID = pkgPath + "." + ftype.ObjectName
	ftype.CleanObjectName = strings.TrimPrefix(types.TypeString(typ, resolver), "*")
	ftype.TSType = ftype.CleanObjectName
	ftype.JSType = ftype.CleanObjectName
	ftype.SwiftType = ftype.CleanObjectName
	if ftype.IsObject {
		ftype.JSType = "object"
		// ftype.SwiftType = "Any"
	} else {
		switch ftype.CleanObjectName {
		case "interface{}":
			ftype.JSType = "any"
			ftype.SwiftType = "Any"
			ftype.TSType = "any"
		case "map[string]interface{}":
			ftype.JSType = "object"
			ftype.TSType = "object"
			ftype.SwiftType = "Any"
		case "string":
			ftype.JSType = "string"
			ftype.SwiftType = "String"
			ftype.TSType = "string"
		case "bool":
			ftype.JSType = "boolean"
			ftype.SwiftType = "Bool"
			ftype.TSType = "boolean"
		case "int", "int16", "int32", "int64",
			"uint", "uint16", "uint32", "uint64",
			"float32", "float64":
			ftype.JSType = "number"
			ftype.SwiftType = "Double"
			ftype.TSType = "number"
		}
	}

	return ftype, nil
}

// addOutputFields adds built-in fields to the response objects
// mentioned in p.outputObjects.
func (p *Parser) addOutputFields() error {
	errorField := Field{
		OmitEmpty:      true,
		Name:           "Error",
		NameLowerCamel: "error",
		NameLowerSnake: "error",
		Comment:        "Error is string explaining what went wrong. Empty if everything was fine.",
		Type: FieldType{
			TypeName:  "string",
			JSType:    "string",
			SwiftType: "String",
			TSType:    "string",
		},
		Metadata: map[string]interface{}{},
		Example:  "something went wrong",
	}
	for typeName := range p.outputObjects {
		obj, err := p.def.Object(typeName)
		if err != nil {
			// skip if we can't find it - it must be excluded
			continue
		}
		obj.Fields = append(obj.Fields, errorField)
	}
	return nil
}

func (p *Parser) wrapErr(err error, pkg *packages.Package, pos token.Pos) error {
	position := pkg.Fset.Position(pos)
	return errors.Wrap(err, position.String())
}

func isInSlice(slice []string, s string) bool {
	for i := range slice {
		if slice[i] == s {
			return true
		}
	}
	return false
}

func (p *Parser) lookupType(name string) *doc.Type {
	for i := range p.docs.Types {
		if p.docs.Types[i].Name == name {
			return p.docs.Types[i]
		}
	}
	return nil
}

func (p *Parser) commentForType(name string) string {
	typ := p.lookupType(name)
	if typ == nil {
		return ""
	}
	return cleanComment(typ.Doc)
}

func (p *Parser) commentForMethod(service, method string) string {
	typ := p.lookupType(service)
	if typ == nil {
		return ""
	}
	spec, ok := typ.Decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return ""
	}
	iface, ok := spec.Type.(*ast.InterfaceType)
	if !ok {
		return ""
	}
	var m *ast.Field
outer:
	for i := range iface.Methods.List {
		for _, name := range iface.Methods.List[i].Names {
			if name.Name == method {
				m = iface.Methods.List[i]
				break outer
			}
		}
	}
	if m == nil {
		return ""
	}
	return cleanComment(m.Doc.Text())
}

func (p *Parser) commentForField(typeName, field string) string {
	typ := p.lookupType(typeName)
	if typ == nil {
		return ""
	}
	spec, ok := typ.Decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return ""
	}
	obj, ok := spec.Type.(*ast.StructType)
	if !ok {
		return ""
	}
	var f *ast.Field
outer:
	for i := range obj.Fields.List {
		for _, name := range obj.Fields.List[i].Names {
			if name.Name == field {
				f = obj.Fields.List[i]
				break outer
			}
		}
	}
	if f == nil {
		return ""
	}
	return cleanComment(f.Doc.Text())
}

func cleanComment(s string) string {
	return strings.TrimSpace(s)
}

// metadataCommentRegex is the regex to pull key value metadata
// used since we can't simply trust lines that contain a colon
var metadataCommentRegex = regexp.MustCompile(`^.*: .*`)

// extractCommentMetadata extracts key value pairs from the comment.
// It returns a map of metadata, and the
// remaining comment string.
// Metadata fields should succeed the comment string.
func (p *Parser) extractCommentMetadata(comment string) (map[string]interface{}, string, error) {
	var lines []string
	metadata := make(map[string]interface{})
	s := bufio.NewScanner(strings.NewReader(comment))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if metadataCommentRegex.MatchString(line) {
			line = strings.TrimSpace(line)
			if line == "" {
				return metadata, strings.Join(lines, "\n"), nil
			}
			// SplitN is being used to ensure that colons can exist
			// in values by only splitting on the first colon in the line
			splitLine := strings.SplitN(line, ": ", 2)
			key := splitLine[0]
			value := strings.TrimSpace(splitLine[1])
			var val interface{}
			if err := json.Unmarshal([]byte(value), &val); err != nil {
				if p.Verbose {
					fmt.Printf("(skipping) failed to marshal JSON value (%s): %s\n", err, value)
				}
				continue
			}
			metadata[key] = val
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return metadata, strings.Join(lines, "\n"), nil
}
