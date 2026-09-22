package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

func scriptCollectionSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := NewScriptCollectionResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

func TestScriptCollectionResource_SchemaRequiredOptionalComputed(t *testing.T) {
	s := scriptCollectionSchema(t).Schema

	cases := []struct {
		name               string
		required, optional bool
		computed           bool
	}{
		{"id", false, false, true},
		{"package_id", true, false, false},
		{"script_collection_id", true, false, false},
		{"name", true, false, false},
		{"content", false, true, true},
		{"content_hash", false, true, true},
		{"version", false, false, true},
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
		if attr.IsOptional() != c.optional {
			t.Errorf("%s.Optional = %v, want %v", c.name, attr.IsOptional(), c.optional)
		}
		if attr.IsComputed() != c.computed {
			t.Errorf("%s.Computed = %v, want %v", c.name, attr.IsComputed(), c.computed)
		}
	}
}

func TestScriptCollectionResource_ImportState(t *testing.T) {
	r := NewScriptCollectionResource().(resource.ResourceWithImportState)

	resp := &resource.ImportStateResponse{State: newTestState(t, scriptCollectionSchema(t).Schema)}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "UTILITIES/shared-scripts"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ImportState() produced diagnostics: %v", resp.Diagnostics)
	}

	var packageID, scriptCollectionID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRoot("package_id"), &packageID)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRoot("script_collection_id"), &scriptCollectionID)...)
	if packageID.ValueString() != "UTILITIES" {
		t.Errorf("package_id = %q, want %q", packageID.ValueString(), "UTILITIES")
	}
	if scriptCollectionID.ValueString() != "shared-scripts" {
		t.Errorf("script_collection_id = %q, want %q", scriptCollectionID.ValueString(), "shared-scripts")
	}
}

func TestScriptCollectionToModel_PreservesPreviousContentFields(t *testing.T) {
	previous := scriptCollectionModel{
		Content:     types.StringValue("${path.module}/script-collections/shared-scripts.zip"),
		ContentHash: types.StringValue("deadbeef"),
	}
	sc := &cloudintegration.ScriptCollection{
		ID:      "shared-scripts",
		Name:    "Shared Scripts",
		Version: "1.0.1",
	}

	got := scriptCollectionToModel("UTILITIES", sc, previous)

	if got.ID.ValueString() != "UTILITIES/shared-scripts" {
		t.Errorf("ID = %q, want %q", got.ID.ValueString(), "UTILITIES/shared-scripts")
	}
	if got.Content != previous.Content {
		t.Errorf("Content = %v, want it preserved from the prior state (%v)", got.Content, previous.Content)
	}
	if got.ContentHash != previous.ContentHash {
		t.Errorf("ContentHash = %v, want it preserved from the prior state (%v)", got.ContentHash, previous.ContentHash)
	}
	if got.Version.ValueString() != "1.0.1" {
		t.Errorf("Version = %q, want %q", got.Version.ValueString(), "1.0.1")
	}
}
