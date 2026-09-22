package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestScriptCollectionDeploymentResource_SchemaRequiredOptionalComputed(t *testing.T) {
	r := NewScriptCollectionDeploymentResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}

	cases := []struct {
		name               string
		required, computed bool
	}{
		{"id", false, true},
		{"package_id", true, false},
		{"script_collection_id", true, false},
		{"script_collection_version", true, false},
		{"status", false, true},
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

	if _, ok := resp.Schema.Blocks["timeouts"]; !ok {
		t.Error("expected a \"timeouts\" block")
	}
}

func TestScriptCollectionDeploymentResource_ImportState(t *testing.T) {
	r := NewScriptCollectionDeploymentResource().(resource.ResourceWithImportState)

	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)

	resp := &resource.ImportStateResponse{State: newTestState(t, schemaResp.Schema)}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "UTILITIES/shared-scripts"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ImportState() produced diagnostics: %v", resp.Diagnostics)
	}
}
