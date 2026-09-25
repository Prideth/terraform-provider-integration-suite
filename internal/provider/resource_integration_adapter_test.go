package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func integrationAdapterSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := NewIntegrationAdapterResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

func TestIntegrationAdapterResource_SchemaRequiredComputed(t *testing.T) {
	s := integrationAdapterSchema(t).Schema

	cases := []struct {
		name               string
		required, computed bool
	}{
		{"id", true, false},
		{"package_id", true, false},
		{"name", true, false},
		{"description", false, true},
		{"content", false, true},
		{"content_hash", false, true},
		{"version", false, true},
	}

	for _, c := range cases {
		attr, ok := s.Attributes[c.name]
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

// TestIntegrationAdapterResource_EveryMutableFieldForcesReplace pins down
// the conservative "no confirmed in-place update" model: every attribute
// that can be set by a practitioner forces replacement, since SAP
// documents that importing a duplicate ID is rejected as an error and no
// public API for updating an existing adapter's metadata or content in
// place was confirmed.
func TestIntegrationAdapterResource_EveryMutableFieldForcesReplace(t *testing.T) {
	s := integrationAdapterSchema(t).Schema

	fields := []string{"id", "package_id", "name", "content", "content_hash"}
	for _, name := range fields {
		attr, ok := s.Attributes[name].(schema.StringAttribute)
		if !ok {
			t.Fatalf("attribute %q is not a StringAttribute", name)
		}
		var mods []interface {
			Description(context.Context) string
		}
		for _, m := range attr.PlanModifiers {
			mods = append(mods, m)
		}
		if !hasRequiresReplace(mods) {
			t.Errorf("%s has no RequiresReplace plan modifier", name)
		}
	}
}

func TestIntegrationAdapterResource_Update_AlwaysErrors(t *testing.T) {
	r := NewIntegrationAdapterResource()

	var resp resource.UpdateResponse
	r.Update(context.Background(), resource.UpdateRequest{}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Error("Update() did not produce a diagnostic; every field is RequiresReplace so Update should be unreachable")
	}
}

func TestIntegrationAdapterResource_ImportState(t *testing.T) {
	r := NewIntegrationAdapterResource().(resource.ResourceWithImportState)

	resp := &resource.ImportStateResponse{State: newTestState(t, integrationAdapterSchema(t).Schema)}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "ADAPTERS/custom-sftp-extension"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ImportState() produced diagnostics: %v", resp.Diagnostics)
	}

	var packageID, id types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRoot("package_id"), &packageID)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRootID(), &id)...)
	if packageID.ValueString() != "ADAPTERS" {
		t.Errorf("package_id = %q, want ADAPTERS", packageID.ValueString())
	}
	if id.ValueString() != "custom-sftp-extension" {
		t.Errorf("id = %q, want custom-sftp-extension", id.ValueString())
	}
}
