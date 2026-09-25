package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestIntegrationAdapterDataSource_SchemaRequiredComputed(t *testing.T) {
	d := NewIntegrationAdapterDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}

	cases := []struct {
		name               string
		required, computed bool
	}{
		{"id", true, false},
		{"name", false, true},
		{"package_id", false, true},
		{"description", false, true},
		{"version", false, true},
	}

	for _, c := range cases {
		attr, ok := resp.Schema.Attributes[c.name]
		if !ok {
			t.Errorf("missing attribute %q", c.name)
			continue
		}
		if attr.IsRequired() != c.required {
			t.Errorf("%s.Required = %v, want %v", c.name, attr.IsRequired(), c.required)
		}
		if attr.IsComputed() != c.computed {
			t.Errorf("%s.Computed = %v, want %v", c.name, attr.IsComputed(), c.computed)
		}
	}

	for _, removed := range []string{"type", "application"} {
		if _, ok := resp.Schema.Attributes[removed]; ok {
			t.Errorf("%s must not exist: it is not a property of IntegrationAdapterDesigntimeArtifact", removed)
		}
	}
}
