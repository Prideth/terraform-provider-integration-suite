package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func integrationAdapterDeploymentSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := NewIntegrationAdapterDeploymentResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

func TestIntegrationAdapterDeploymentResource_SchemaRequiredComputed(t *testing.T) {
	s := integrationAdapterDeploymentSchema(t).Schema

	cases := []struct {
		name               string
		required, computed bool
	}{
		{"id", false, true},
		{"adapter_id", true, false},
		{"status", false, true},
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

// TestIntegrationAdapterDeploymentResource_NoVersionAttribute pins down a
// deliberate difference from every other *_deployment resource in this
// provider: there is no adapter_version attribute, since SAP's confirmed
// DeployIntegrationAdapterDesigntimeArtifact action takes only Id, with no
// Version query parameter.
func TestIntegrationAdapterDeploymentResource_NoVersionAttribute(t *testing.T) {
	s := integrationAdapterDeploymentSchema(t).Schema

	if _, ok := s.Attributes["adapter_version"]; ok {
		t.Error("adapter_version must not exist: the confirmed deploy action takes no Version parameter")
	}
}

func TestIntegrationAdapterDeploymentResource_AdapterIDForcesReplace(t *testing.T) {
	s := integrationAdapterDeploymentSchema(t).Schema

	attr, ok := s.Attributes["adapter_id"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("adapter_id is not a StringAttribute")
	}
	var mods []interface {
		Description(context.Context) string
	}
	for _, m := range attr.PlanModifiers {
		mods = append(mods, m)
	}
	if !hasRequiresReplace(mods) {
		t.Error("adapter_id has no RequiresReplace plan modifier")
	}
}

func TestIntegrationAdapterDeploymentResource_Update_AlwaysErrors(t *testing.T) {
	r := NewIntegrationAdapterDeploymentResource()

	var resp resource.UpdateResponse
	r.Update(context.Background(), resource.UpdateRequest{}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Error("Update() did not produce a diagnostic; adapter_id is RequiresReplace so Update should be unreachable")
	}
}

func TestIntegrationAdapterDeploymentResource_ImportState(t *testing.T) {
	r := NewIntegrationAdapterDeploymentResource().(resource.ResourceWithImportState)

	resp := &resource.ImportStateResponse{State: newTestState(t, integrationAdapterDeploymentSchema(t).Schema)}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "custom-sftp-extension"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ImportState() produced diagnostics: %v", resp.Diagnostics)
	}

	var adapterID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRoot("adapter_id"), &adapterID)...)
	if adapterID.ValueString() != "custom-sftp-extension" {
		t.Errorf("adapter_id = %q, want custom-sftp-extension", adapterID.ValueString())
	}
}
