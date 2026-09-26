package edmx

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"
)

const sample = `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://schemas.microsoft.com/ado/2007/06/edmx" Version="1.0">
  <edmx:DataServices xmlns:m="http://schemas.microsoft.com/ado/2007/08/dataservices/metadata" m:DataServiceVersion="2.0">
    <Schema xmlns="http://schemas.microsoft.com/ado/2008/09/edm" Namespace="com.example">
      <EntityType Name="Parameter">
        <Key><PropertyRef Name="Pid"/><PropertyRef Name="Id"/></Key>
        <Property Name="Pid" Type="Edm.String" Nullable="false"/>
        <Property Name="Id" Type="Edm.String" Nullable="false"/>
      </EntityType>
      <EntityType Name="StringParameter" BaseType="com.example.Parameter">
        <Property Name="Value" Type="Edm.String"/>
      </EntityType>
      <EntityType Name="Policy">
        <Key><PropertyRef Name="Id"/></Key>
        <Property Name="Id" Type="Edm.Int64" Nullable="false"/>
        <NavigationProperty Name="Refs" Relationship="com.example.r" FromRole="a" ToRole="b"/>
      </EntityType>
      <EntityContainer Name="C" m:IsDefaultEntityContainer="true">
        <EntitySet Name="StringParameters" EntityType="com.example.StringParameter"/>
        <EntitySet Name="Policies" EntityType="com.example.Policy"/>
        <FunctionImport Name="Deploy" ReturnType="Edm.String" m:HttpMethod="POST">
          <Parameter Name="Id" Type="Edm.String"/>
          <Parameter Name="Version" Type="Edm.String"/>
        </FunctionImport>
      </EntityContainer>
    </Schema>
  </edmx:DataServices>
</edmx:Edmx>`

func TestParse_ResolvesBaseTypesAndContainers(t *testing.T) {
	m, err := Parse([]byte(sample))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	sp := m.EntityTypes["StringParameter"]
	if sp == nil {
		t.Fatal("StringParameter missing")
	}
	if !reflect.DeepEqual(sp.Key, []string{"Pid", "Id"}) {
		t.Errorf("inherited key = %v, want [Pid Id]", sp.Key)
	}
	var names []string
	for n := range sp.Properties {
		names = append(names, n)
	}
	sort.Strings(names)
	if !reflect.DeepEqual(names, []string{"Id", "Pid", "Value"}) {
		t.Errorf("properties = %v, want inherited Id, Pid plus own Value", names)
	}

	if !m.EntityTypes["Policy"].Properties["Refs"].Navigation {
		t.Error("navigation property not recorded")
	}
	if m.EntitySets["Policies"] != "Policy" {
		t.Errorf("entity set type = %q, want unqualified Policy", m.EntitySets["Policies"])
	}
	fi := m.FunctionImports["Deploy"]
	if fi == nil || fi.HTTPMethod != "POST" || len(fi.Parameters) != 2 {
		t.Errorf("function import = %+v", fi)
	}
}

func TestJSONFields_FlattensEmbeddedAndSkipsDash(t *testing.T) {
	type inner struct {
		A string `json:"A"`
	}
	type outer struct {
		inner
		B      string `json:"B,omitempty"`
		Hidden string `json:"-"`
		C      string
		d      string
	}
	_ = outer{}.d
	got := JSONFields(reflect.TypeOf(outer{}))
	if !reflect.DeepEqual(got, []string{"A", "B", "C"}) {
		t.Errorf("JSONFields = %v, want [A B C]", got)
	}
}

// A navigation property arrives as {"__deferred": ...} or, with $expand,
// {"results": [...]}. Plain strings and slices, as used by the API product,
// key value map and runtime artifact reads before September 2026, cannot
// decode either and must be rejected in read structs.
func TestDecodesNavigation(t *testing.T) {
	type expanded struct {
		Results []string `json:"results"`
	}
	cases := []struct {
		name string
		v    any
		want bool
	}{
		{"string", "", false},
		{"slice", []struct{ Name string }{}, false},
		{"plain struct", struct{ Name string }{}, false},
		{"expanded collection", expanded{}, true},
		{"pointer to expanded collection", &expanded{}, true},
		{"raw JSON", json.RawMessage(nil), true},
	}
	for _, c := range cases {
		if got := decodesNavigation(reflect.TypeOf(c.v)); got != c.want {
			t.Errorf("%s: decodesNavigation = %v, want %v", c.name, got, c.want)
		}
	}
}
