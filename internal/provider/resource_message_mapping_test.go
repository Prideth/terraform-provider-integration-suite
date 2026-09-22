package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

// newTestState builds a null tfsdk.State matching s, the same starting
// point Terraform itself gives a resource's ImportState method before any
// SetAttribute call populates it.
func newTestState(t *testing.T, s schema.Schema) tfsdk.State {
	t.Helper()
	return tfsdk.State{
		Schema: s,
		Raw:    tftypes.NewValue(s.Type().TerraformType(context.Background()), nil),
	}
}

// hasRequiresReplace reports whether any of mods is a RequiresReplace-style
// plan modifier, detected by its documented description text rather than a
// type assertion, since stringplanmodifier.RequiresReplace() returns the
// unexported type behind RequiresReplaceIf.
func hasRequiresReplace(mods []interface {
	Description(context.Context) string
}) bool {
	for _, m := range mods {
		if strings.Contains(m.Description(context.Background()), "destroy and recreate the resource") {
			return true
		}
	}
	return false
}

func messageMappingSchema(t *testing.T) schema.Schema {
	t.Helper()
	r := NewMessageMappingResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp.Schema
}

func TestMessageMappingResource_SchemaRequiredOptionalComputed(t *testing.T) {
	s := messageMappingSchema(t)

	cases := []struct {
		name               string
		required, optional bool
		computed           bool
	}{
		{"id", false, false, true},
		{"package_id", true, false, false},
		{"mapping_id", true, false, false},
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

// TestMessageMappingResource_UpdateModelDiffersFromValueMapping locks in
// this phase's key design decision at the schema level: package_id and
// mapping_id replace the resource (immutable identity, same as every
// sibling resource), but name, content, and content_hash do not — unlike
// sapintegrationsuite_value_mapping, where all three are RequiresReplace
// because no confirmed in-place update exists for that entity set. See
// docs/sap-api-references.md for why message mapping's evidence points the
// other way.
func TestMessageMappingResource_UpdateModelDiffersFromValueMapping(t *testing.T) {
	s := messageMappingSchema(t)

	replaceExpected := map[string]bool{
		"package_id":   true,
		"mapping_id":   true,
		"name":         false,
		"content":      false,
		"content_hash": false,
	}

	for name, wantReplace := range replaceExpected {
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

		if got := hasRequiresReplace(mods); got != wantReplace {
			t.Errorf("%s: RequiresReplace present = %v, want %v", name, got, wantReplace)
		}
	}
}

func TestMessageMappingResource_ImportState(t *testing.T) {
	r := NewMessageMappingResource().(resource.ResourceWithImportState)

	t.Run("valid ID splits into package_id and mapping_id", func(t *testing.T) {
		resp := &resource.ImportStateResponse{State: newTestState(t, messageMappingSchema(t))}
		r.ImportState(context.Background(), resource.ImportStateRequest{ID: "UTILITIES/customer-mapping"}, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("ImportState() produced diagnostics: %v", resp.Diagnostics)
		}

		var packageID, mappingID types.String
		resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRoot("package_id"), &packageID)...)
		resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRoot("mapping_id"), &mappingID)...)
		if packageID.ValueString() != "UTILITIES" {
			t.Errorf("package_id = %q, want %q", packageID.ValueString(), "UTILITIES")
		}
		if mappingID.ValueString() != "customer-mapping" {
			t.Errorf("mapping_id = %q, want %q", mappingID.ValueString(), "customer-mapping")
		}
	})

	t.Run("invalid ID produces an error", func(t *testing.T) {
		resp := &resource.ImportStateResponse{State: newTestState(t, messageMappingSchema(t))}
		r.ImportState(context.Background(), resource.ImportStateRequest{ID: "not-a-composite-id"}, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected an error for an import ID without a slash")
		}
	})
}

func TestMessageMappingToModel_PreservesPreviousContentFields(t *testing.T) {
	previous := messageMappingModel{
		Content:     types.StringValue("${path.module}/message-mappings/customer-mapping.zip"),
		ContentHash: types.StringValue("deadbeef"),
	}
	mapping := &cloudintegration.MessageMapping{
		ID:      "customer-mapping",
		Name:    "Customer Mapping",
		Version: "1.0.1",
	}

	got := messageMappingToModel("UTILITIES", mapping, previous)

	if got.ID.ValueString() != "UTILITIES/customer-mapping" {
		t.Errorf("ID = %q, want %q", got.ID.ValueString(), "UTILITIES/customer-mapping")
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
