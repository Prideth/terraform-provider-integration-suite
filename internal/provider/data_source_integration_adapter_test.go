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
		{"type", false, true},
		{"application", false, true},
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

	if _, hasPackageID := resp.Schema.Attributes["package_id"]; hasPackageID {
		t.Error("package_id must not exist on the data source: GET-by-id returning it reliably was not confirmed")
	}
}
