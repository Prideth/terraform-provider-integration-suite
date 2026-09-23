package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestKeystoreEntryDataSource_SchemaRequiredComputed(t *testing.T) {
	d := NewKeystoreEntryDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}

	cases := []struct {
		name               string
		required, computed bool
	}{
		{"alias", true, false},
		{"hex_alias", false, true},
		{"key_type", false, true},
		{"key_size", false, true},
		{"valid_not_before", false, true},
		{"valid_not_after", false, true},
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

func TestKeystoreEntriesDataSource_Schema(t *testing.T) {
	d := NewKeystoreEntriesDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}

	entriesAttr, ok := resp.Schema.Attributes["entries"]
	if !ok {
		t.Fatal("missing attribute \"entries\"")
	}
	if !entriesAttr.IsComputed() {
		t.Error("entries must be Computed")
	}
}
