package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestAccessPolicyDataSource_SchemaOptionalComputed(t *testing.T) {
	d := NewAccessPolicyDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}

	cases := []struct {
		name               string
		optional, computed bool
	}{
		{"id", true, true},
		{"role_name", true, true},
		{"description", false, true},
	}

	if len(resp.Schema.Attributes) != len(cases) {
		t.Errorf("schema has %d attributes, want %d", len(resp.Schema.Attributes), len(cases))
	}
	for _, c := range cases {
		attr, ok := resp.Schema.Attributes[c.name]
		if !ok {
			t.Errorf("missing attribute %q", c.name)
			continue
		}
		if attr.IsOptional() != c.optional {
			t.Errorf("%s.Optional = %v, want %v", c.name, attr.IsOptional(), c.optional)
		}
		if attr.IsComputed() != c.computed {
			t.Errorf("%s.Computed = %v, want %v", c.name, attr.IsComputed(), c.computed)
		}
	}
}

func TestAccessPolicyDataSource_RequiresExactlyOneLookupKey(t *testing.T) {
	d, ok := NewAccessPolicyDataSource().(datasource.DataSourceWithConfigValidators)
	if !ok {
		t.Fatal("data source must declare config validators")
	}
	if n := len(d.ConfigValidators(context.Background())); n != 1 {
		t.Errorf("ConfigValidators() returned %d validators, want 1 (ExactlyOneOf id, role_name)", n)
	}
}
