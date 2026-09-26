// Package edmx checks the provider's OData wire structs against a real
// OData V2 $metadata document. The document is tenant data and is not part of
// the repository: tests that use it skip themselves when it is not present.
package edmx

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

// EnvMetadataFile overrides where the $metadata document is read from.
const EnvMetadataFile = "SAP_INTEGRATION_SUITE_METADATA_FILE"

const defaultRelativePath = ".specs/cloudintegration-metadata.xml"

// Property is a structural or navigation property of an entity type.
type Property struct {
	Name       string
	Type       string
	MaxLength  string
	Navigation bool
}

// EntityType is one entity type with its key and properties.
type EntityType struct {
	Name       string
	BaseType   string
	Key        []string
	Properties map[string]Property
}

// FunctionImport is one OData V2 function import (service operation).
type FunctionImport struct {
	Name       string
	HTTPMethod string
	Parameters map[string]string
}

// Model is the parsed subset of a $metadata document the contract tests need.
type Model struct {
	Path            string
	EntityTypes     map[string]*EntityType
	EntitySets      map[string]string
	FunctionImports map[string]*FunctionImport
}

type xmlEdmx struct {
	Schemas []xmlSchema `xml:"DataServices>Schema"`
}

type xmlSchema struct {
	EntityTypes []struct {
		Name     string `xml:"Name,attr"`
		BaseType string `xml:"BaseType,attr"`
		Key      []struct {
			Name string `xml:"Name,attr"`
		} `xml:"Key>PropertyRef"`
		Properties []struct {
			Name      string `xml:"Name,attr"`
			Type      string `xml:"Type,attr"`
			MaxLength string `xml:"MaxLength,attr"`
		} `xml:"Property"`
		Navigation []struct {
			Name string `xml:"Name,attr"`
		} `xml:"NavigationProperty"`
	} `xml:"EntityType"`
	Containers []struct {
		EntitySets []struct {
			Name       string `xml:"Name,attr"`
			EntityType string `xml:"EntityType,attr"`
		} `xml:"EntitySet"`
		FunctionImports []struct {
			Name       string `xml:"Name,attr"`
			HTTPMethod string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices/metadata HttpMethod,attr"`
			Parameters []struct {
				Name string `xml:"Name,attr"`
				Type string `xml:"Type,attr"`
			} `xml:"Parameter"`
		} `xml:"FunctionImport"`
	} `xml:"EntityContainer"`
}

// Parse reads a $metadata document.
func Parse(data []byte) (*Model, error) {
	var doc xmlEdmx
	if err := xml.Unmarshal(data, &doc); err != nil { //nolint:gosec // G709: test-only helper decoding a developer-supplied $metadata file into a fixed struct
		return nil, fmt.Errorf("edmx: %w", err)
	}

	m := &Model{
		EntityTypes:     map[string]*EntityType{},
		EntitySets:      map[string]string{},
		FunctionImports: map[string]*FunctionImport{},
	}
	for _, s := range doc.Schemas {
		for _, et := range s.EntityTypes {
			t := &EntityType{Name: et.Name, BaseType: unqualified(et.BaseType), Properties: map[string]Property{}}
			for _, k := range et.Key {
				t.Key = append(t.Key, k.Name)
			}
			for _, p := range et.Properties {
				t.Properties[p.Name] = Property{Name: p.Name, Type: p.Type, MaxLength: p.MaxLength}
			}
			for _, n := range et.Navigation {
				t.Properties[n.Name] = Property{Name: n.Name, Navigation: true}
			}
			m.EntityTypes[et.Name] = t
		}
		for _, c := range s.Containers {
			for _, es := range c.EntitySets {
				m.EntitySets[es.Name] = unqualified(es.EntityType)
			}
			for _, fi := range c.FunctionImports {
				f := &FunctionImport{Name: fi.Name, HTTPMethod: fi.HTTPMethod, Parameters: map[string]string{}}
				for _, p := range fi.Parameters {
					f.Parameters[p.Name] = p.Type
				}
				m.FunctionImports[fi.Name] = f
			}
		}
	}
	for _, et := range m.EntityTypes {
		if err := m.resolveBase(et, map[string]bool{}); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func unqualified(name string) string {
	if i := strings.LastIndex(name, "."); i >= 0 {
		return name[i+1:]
	}
	return name
}

// Load returns the $metadata model, or skips the test when no document is
// available. It looks at $SAP_INTEGRATION_SUITE_METADATA_FILE first, then for
// .specs/cloudintegration-metadata.xml in the working directory and its parents.
func Load(t *testing.T) *Model {
	t.Helper()
	return LoadFrom(t, EnvMetadataFile, defaultRelativePath)
}

// LoadFrom is Load for another service's $metadata document: it reads the
// file named by the environment variable env, or rel (a path relative to
// the working directory or one of its parents), and skips the test when
// neither exists.
func LoadFrom(t *testing.T, env, rel string) *Model {
	t.Helper()

	path := os.Getenv(env)
	if path == "" {
		path = findUpwards(rel)
	}
	if path == "" {
		t.Skipf("no $metadata document found; set %s or place it at %s", env, rel)
	}

	data, err := os.ReadFile(path) //nolint:gosec // G304: path comes from the developer's own environment
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	m, err := Parse(data)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	m.Path = path
	return m
}

func findUpwards(rel string) string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		candidate := filepath.Join(dir, rel)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// EntityTypeOf returns the entity type behind an entity set, failing the test
// if the set does not exist.
func (m *Model) EntityTypeOf(t *testing.T, entitySet string) *EntityType {
	t.Helper()
	name, ok := m.EntitySets[entitySet]
	if !ok {
		t.Errorf("entity set %q does not exist in %s", entitySet, m.Path)
		return nil
	}
	et, ok := m.EntityTypes[name]
	if !ok {
		t.Errorf("entity type %q of entity set %q is not defined", name, entitySet)
		return nil
	}
	return et
}

// AssertStruct checks a struct the provider decodes from SAP's responses:
// every JSON field of v, including fields of embedded structs, must be a
// property of the entity set's type. A field may map a navigation property
// only when its Go type can decode what SAP returns for one: a
// {"__deferred": ...} link without $expand, {"results": [...]} with it.
// That means a struct with a "results" field (v2.ExpandedCollection) or
// json.RawMessage; a plain slice or string fails on every real response.
func (m *Model) AssertStruct(t *testing.T, entitySet string, v any) {
	t.Helper()
	m.assertFields(t, entitySet, v, true)
}

// AssertWriteStruct checks a request body that is only ever encoded, never
// decoded, for example a create with a deep insert. Navigation properties
// may take any shape there.
func (m *Model) AssertWriteStruct(t *testing.T, entitySet string, v any) {
	t.Helper()
	m.assertFields(t, entitySet, v, false)
}

func (m *Model) assertFields(t *testing.T, entitySet string, v any, decoded bool) {
	t.Helper()
	et := m.EntityTypeOf(t, entitySet)
	if et == nil {
		return
	}
	for _, field := range jsonFields(reflect.TypeOf(v)) {
		p, ok := et.Properties[field.name]
		if !ok {
			t.Errorf("%T: JSON field %q is not a property of %s (entity set %s)", v, field.name, et.Name, entitySet)
			continue
		}
		if !decoded {
			continue
		}
		if p.Navigation {
			if !decodesNavigation(field.typ) {
				t.Errorf("%T: JSON field %q maps the navigation property %s.%s as %s, which cannot decode SAP's "+
					"{\"__deferred\": ...} or {\"results\": [...]} object; drop it from the read struct or use "+
					"v2.ExpandedCollection with $expand", v, field.name, et.Name, field.name, field.typ)
			}
			continue
		}
		if ok, want := decodesEdm(p.Type, field); !ok {
			t.Errorf("%T: JSON field %q is %s, but %s.%s is %s, which OData V2 JSON sends as %s",
				v, field.name, field.typ, et.Name, field.name, p.Type, want)
		}
	}
}

var (
	unmarshalerType = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
	timeType        = reflect.TypeOf(time.Time{})
	numberType      = reflect.TypeOf(json.Number(""))
)

// decodesEdm reports whether a Go field can decode a property's value as
// OData V2 JSON represents it: Edm.Int64 and Edm.Decimal as strings ("51"),
// Edm.DateTime as "/Date(ms)/", smaller integers, doubles and booleans as
// JSON numbers and booleans. The second result describes the JSON shape for
// the error message.
func decodesEdm(edmType string, f jsonField) (bool, string) {
	typ := f.typ
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	kind := typ.Kind()
	isInt := kind >= reflect.Int && kind <= reflect.Uint64
	isFloat := kind == reflect.Float32 || kind == reflect.Float64

	switch edmType {
	case "Edm.DateTime", "Edm.DateTimeOffset", "Edm.Time":
		// time.Time implements json.Unmarshaler but only reads RFC 3339.
		if typ == timeType {
			return false, `a string such as "/Date(1790424405369)/"`
		}
	}
	if typ == reflect.TypeOf(json.RawMessage(nil)) || kind == reflect.Interface ||
		reflect.PointerTo(typ).Implements(unmarshalerType) {
		return true, ""
	}

	switch edmType {
	case "Edm.String", "Edm.Guid", "Edm.DateTime", "Edm.DateTimeOffset", "Edm.Time":
		return kind == reflect.String, "a JSON string"
	case "Edm.Int64", "Edm.Decimal":
		return kind == reflect.String || typ == numberType || ((isInt || isFloat) && f.quoted),
			`a JSON string such as "51" (use a string, json.Number or the ",string" tag option)`
	case "Edm.Int32", "Edm.Int16", "Edm.Byte", "Edm.SByte", "Edm.Double", "Edm.Single":
		return isInt || isFloat || typ == numberType, "a JSON number"
	case "Edm.Boolean":
		return kind == reflect.Bool, "a JSON boolean"
	case "Edm.Binary":
		return kind == reflect.String || (kind == reflect.Slice && typ.Elem().Kind() == reflect.Uint8), "a base64 string"
	default:
		// A complex type arrives as a JSON object.
		return kind == reflect.Struct || kind == reflect.Map, "a JSON object"
	}
}

// decodesNavigation reports whether a Go type can hold a navigation
// property's JSON object.
func decodesNavigation(typ reflect.Type) bool {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ == reflect.TypeOf(json.RawMessage(nil)) {
		return true
	}
	if typ.Kind() != reflect.Struct {
		return false
	}
	for _, f := range jsonFields(typ) {
		if f.name == "results" || f.name == "__deferred" {
			return true
		}
	}
	return false
}

// AssertKey checks the entity set's key properties and their EDM types, given
// as consecutive name/type pairs, for example "Id", "Edm.Int64".
func (m *Model) AssertKey(t *testing.T, entitySet string, nameTypePairs ...string) {
	t.Helper()
	et := m.EntityTypeOf(t, entitySet)
	if et == nil {
		return
	}
	var want []string
	for i := 0; i+1 < len(nameTypePairs); i += 2 {
		want = append(want, nameTypePairs[i])
		if got := et.Properties[nameTypePairs[i]].Type; got != nameTypePairs[i+1] {
			t.Errorf("%s key %s has type %q, want %q", et.Name, nameTypePairs[i], got, nameTypePairs[i+1])
		}
	}
	if strings.Join(et.Key, ",") != strings.Join(want, ",") {
		t.Errorf("%s key = %v, want %v", et.Name, et.Key, want)
	}
}

// AssertFunctionImport checks that a function import exists with exactly the
// given parameter names.
func (m *Model) AssertFunctionImport(t *testing.T, name, httpMethod string, params ...string) {
	t.Helper()
	fi, ok := m.FunctionImports[name]
	if !ok {
		t.Errorf("function import %q does not exist in %s", name, m.Path)
		return
	}
	if httpMethod != "" && !strings.EqualFold(fi.HTTPMethod, httpMethod) {
		t.Errorf("function import %s uses %s, want %s", name, fi.HTTPMethod, httpMethod)
	}
	if len(fi.Parameters) != len(params) {
		t.Errorf("function import %s has parameters %v, want %v", name, fi.Parameters, params)
	}
	for _, p := range params {
		if _, ok := fi.Parameters[p]; !ok {
			t.Errorf("function import %s has no parameter %q (has %v)", name, p, fi.Parameters)
		}
	}
}

// JSONFields lists the JSON names of a struct type's exported fields, flattening
// embedded structs the way encoding/json does and skipping "-" tags.
func JSONFields(typ reflect.Type) []string {
	fields := jsonFields(typ)
	names := make([]string, 0, len(fields))
	for _, f := range fields {
		names = append(names, f.name)
	}
	return names
}

type jsonField struct {
	name   string
	typ    reflect.Type
	quoted bool // the ",string" tag option
}

func jsonFields(typ reflect.Type) []jsonField {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	var fields []jsonField
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		tag := f.Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		if name == "-" {
			continue
		}
		if f.Anonymous && name == "" {
			fields = append(fields, jsonFields(f.Type)...)
			continue
		}
		if !f.IsExported() {
			continue
		}
		if name == "" {
			name = f.Name
		}
		fields = append(fields, jsonField{name: name, typ: f.Type, quoted: strings.Contains(tag, ",string")})
	}
	return fields
}

// resolveBase copies inherited key and properties from the BaseType chain
// into et, so lookups on a derived type see everything the service accepts.
func (m *Model) resolveBase(et *EntityType, seen map[string]bool) error {
	if et.BaseType == "" {
		return nil
	}
	if seen[et.Name] {
		return fmt.Errorf("edmx: BaseType cycle at %s", et.Name)
	}
	seen[et.Name] = true
	base, ok := m.EntityTypes[et.BaseType]
	if !ok {
		return fmt.Errorf("edmx: %s has unknown BaseType %s", et.Name, et.BaseType)
	}
	if err := m.resolveBase(base, seen); err != nil {
		return err
	}
	if len(et.Key) == 0 {
		et.Key = append([]string(nil), base.Key...)
	}
	for name, p := range base.Properties {
		if _, own := et.Properties[name]; !own {
			et.Properties[name] = p
		}
	}
	et.BaseType = ""
	return nil
}

// AssertMaxLength checks the MaxLength facet of a property of the entity
// set's type, including inherited properties.
func (m *Model) AssertMaxLength(t *testing.T, entitySet, property string, want int) {
	t.Helper()
	et := m.EntityTypeOf(t, entitySet)
	if et == nil {
		return
	}
	p, ok := et.Properties[property]
	if !ok {
		t.Errorf("%s has no property %q", et.Name, property)
		return
	}
	if got := strconv.Itoa(want); p.MaxLength != got {
		t.Errorf("%s.%s MaxLength = %q, want %q", et.Name, property, p.MaxLength, got)
	}
}
