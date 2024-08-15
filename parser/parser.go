package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/doc"
	"go/token"
	"go/types"
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

// Service describes a service, akin to an interface in Go.
type Service struct {
	Name    string   `json:"name"`
	Methods []Method `json:"methods"`
	Comment string   `json:"comment"`
	// Metadata are typed key/value pairs extracted from the
	// comments.
	Metadata map[string]interface{} `json:"metadata"`
}

func (s Service) MethodsByMetadata(field string) []MethodGroup {
	var groups []MethodGroup
	var currentVal interface{}
	for _, method := range s.Methods {
		val, ok := method.Metadata[field]
		if !ok {
			continue
		}
		if val != currentVal {
			groups = append(groups, MethodGroup{
				Metadata: method.Metadata,
			})
			currentVal = val
		}
		groups[len(groups)-1].Methods = append(groups[len(groups)-1].Methods, method)
	}
	// sort them by the field
	sort.Slice(groups, func(i, j int) bool {
		return fmt.Sprintf("%v", groups[i].Metadata[field]) < fmt.Sprintf("%v", groups[j].Metadata[field])
	})
	return groups
}

// Method describes a method that a Service can perform.
type Method struct {
	Name           string    `json:"name"`
	NameLowerCamel string    `json:"nameLowerCamel"`
	InputObject    FieldType `json:"inputObject"`
	OutputObject   FieldType `json:"outputObject"`
	Comment        string    `json:"comment"`
	// Metadata are typed key/value pairs extracted from the
	// comments.
	Metadata map[string]interface{} `json:"metadata"`
}

// Object describes a data structure that is part of this definition.
type Object struct {
	TypeID             string  `json:"typeID"`
	ObjectName         string  `json:"objectName"`
	ExternalObjectName string  `json:"externalObjectName"`
	Name               string  `json:"name"`
	Imported           bool    `json:"imported"`
	Fields             []Field `json:"fields"`
	Comment            string  `json:"comment"`
	// Metadata are typed key/value pairs extracted from the
	// comments.
	Metadata map[string]interface{} `json:"metadata"`
}

// Field describes the field inside an Object.
type Field struct {
	Name           string              `json:"name"`
	NameLowerCamel string              `json:"nameLowerCamel"`
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
	TypeID             string `json:"typeID"`
	TypeName           string `json:"typeName"`
	ObjectName         string `json:"objectName"`
	ExternalObjectName string `json:"externalObjectName"`
	// CleanObjectName is the ObjectName with * removed
	// for pointer types.
	CleanObjectName      string `json:"cleanObjectName"`
	ObjectNameLowerCamel string `json:"objectNameLowerCamel"`
	Multiple             bool   `json:"multiple"`
	Package              string `json:"package"`
	IsObject             bool   `json:"isObject"`
	JSType               string `json:"jsType"`
	TSType               string `json:"tsType"`
	SwiftType            string `json:"swiftType"`
	DartType             string `json:"dartType"`
}

// IsOptional returns true for pointer types (optional).
func (f FieldType) IsOptional() bool {
	return strings.HasPrefix(f.ObjectName, "*")
}

// Parser parses Oto Go definition packages.
type Parser struct {
	Verbose bool

	ExcludeInterfaces []string

	PackageName string

	patterns []string
	def      Definition

	// outputObjects marks output object names.
	outputObjects map[string]struct{}
	// objects marks object names.
	objects map[string]struct{}

	// SuppressErrorField suppresses the Error field in output objects.
	SuppressErrorField bool

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
	// sort objects
	sort.Slice(p.def.Objects, func(i, j int) bool {
		return p.def.Objects[i].Name < p.def.Objects[j].Name
	})
	if !p.SuppressErrorField {
		if err := p.addOutputFields(); err != nil {
			return p.def, err
		}
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

	obj.ObjectName = types.TypeString(o.Type(), func(other *types.Package) string { return "" })
	obj.ExternalObjectName = types.TypeString(o.Type(), func(other *types.Package) string { return p.PackageName })

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
	// if it has a json tag, use that as the NameJSON.
	if tag != "" {
		fieldTag := reflect.StructTag(tag)
		jsonTag := fieldTag.Get("json")
		if jsonTag != "" {
			f.NameLowerCamel = strings.Split(jsonTag, ",")[0]
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
	if example, ok := f.Metadata["example"]; ok {
		f.Example = example
	}
	f.Type, err = p.parseFieldType(pkg, v)
	if err != nil {
		return f, errors.Wrap(err, "parse type")
	}
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
	if slice, ok := obj.Type().(*types.Slice); ok {
		typ = slice.Elem()
		ftype.Multiple = true
	}
	isPointer := true
	originalTyp := typ
	pointerType, isPointer := typ.(*types.Pointer)
	if isPointer {
		typ = pointerType.Elem()
		isPointer = true
	}
	if named, ok := typ.(*types.Named); ok {
		if structure, ok := named.Underlying().(*types.Struct); ok {
			if err := p.parseObject(pkg, named.Obj(), structure); err != nil {
				return ftype, err
			}
			ftype.IsObject = true
		}
	}
	// disallow nested structs
	switch typ.(type) {
	case *types.Struct:
		return ftype, p.wrapErr(errors.New("nested structs not supported (create another type instead)"), pkg, obj.Pos())
	}
	ftype.TypeName = types.TypeString(originalTyp, resolver)
	ftype.ObjectName = types.TypeString(originalTyp, func(other *types.Package) string { return "" })
	ftype.ExternalObjectName = types.TypeString(originalTyp, func(other *types.Package) string { return p.PackageName })
	ftype.ObjectNameLowerCamel = camelizeDown(ftype.ObjectName)
	ftype.TypeID = pkgPath + "." + ftype.ObjectName
	ftype.CleanObjectName = strings.TrimPrefix(ftype.ObjectName, "*")
	ftype.TSType = ftype.CleanObjectName
	ftype.JSType = ftype.CleanObjectName
	ftype.SwiftType = ftype.CleanObjectName
	ftype.DartType = ftype.CleanObjectName
	if ftype.IsObject {
		ftype.JSType = "object"
		//ftype.SwiftType = "Any"
	} else {
		switch ftype.CleanObjectName {
		case "interface{}":
			ftype.JSType = "any"
			ftype.SwiftType = "Any"
			ftype.TSType = "object"
			ftype.DartType = "dynamic"
		case "map[string]interface{}":
			ftype.JSType = "object"
			ftype.TSType = "object"
			ftype.SwiftType = "Any"
			ftype.DartType = "Map<String, dynamic>"
		case "string":
			ftype.JSType = "string"
			ftype.SwiftType = "String"
			ftype.TSType = "string"
			ftype.DartType = "String"
		case "bool":
			ftype.JSType = "boolean"
			ftype.SwiftType = "Bool"
			ftype.TSType = "boolean"
			ftype.DartType = "bool"
		case "int", "int16", "int32", "int64",
			"uint", "uint16", "uint32", "uint64":
			ftype.JSType = "number"
			ftype.TSType = "number"
			ftype.SwiftType = "Int"
			ftype.DartType = "int"
		case "float32", "float64":
			ftype.JSType = "number"
			ftype.SwiftType = "Double"
			ftype.TSType = "number"
			ftype.DartType = "double"
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
		Comment:        "Error is string explaining what went wrong. Empty if everything was fine.",
		Type: FieldType{
			TypeName:  "string",
			JSType:    "string",
			SwiftType: "String",
			TSType:    "string",
			DartType:  "String",
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
	var metadata = make(map[string]interface{})
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
