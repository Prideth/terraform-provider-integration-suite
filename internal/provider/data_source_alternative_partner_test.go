package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestAlternativePartnerDataSource_SchemaRequiredComputed(t *testing.T) {
	d := NewAlternativePartnerDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}

	cases := []struct {
		name               string
		required, computed bool
	}{
		{"id", false, true},
		{"agency", true, false},
		{"scheme", true, false},
		{"external_id", true, false},
		{"partner_id", false, true},
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
}
