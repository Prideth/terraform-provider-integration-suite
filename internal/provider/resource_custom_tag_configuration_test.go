package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func customTagConfigurationSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := NewCustomTagConfigurationResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

func TestCustomTagConfigurationResource_SchemaRequiredComputed(t *testing.T) {
	s := customTagConfigurationSchema(t).Schema

	cases := []struct {
		name               string
		required, computed bool
	}{
		{"id", false, true},
		{"tags", true, false},
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

	tagsAttr, ok := s.Attributes["tags"].(schema.SetNestedAttribute)
	if !ok {
		t.Fatalf("tags is not a SetNestedAttribute (order must not be significant)")
	}

	nameAttr, ok := tagsAttr.NestedObject.Attributes["name"].(schema.StringAttribute)
	if !ok || !nameAttr.Required {
		t.Error("tags.name must be a required StringAttribute")
	}
	mandatoryAttr, ok := tagsAttr.NestedObject.Attributes["mandatory"].(schema.BoolAttribute)
	if !ok || !mandatoryAttr.Required {
		t.Error("tags.mandatory must be a required BoolAttribute")
	}
	permittedAttr, ok := tagsAttr.NestedObject.Attributes["permitted_values"].(schema.SetAttribute)
	if !ok || !permittedAttr.Optional {
		t.Error("tags.permitted_values must be an optional SetAttribute (order must not be significant)")
	}
}

func TestCustomTagConfigurationResource_Delete_AlwaysErrors(t *testing.T) {
	r := NewCustomTagConfigurationResource()

	var resp resource.DeleteResponse
	r.Delete(context.Background(), resource.DeleteRequest{}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Error("Delete() did not produce a diagnostic; SAP documents no delete/clear operation for this entity")
	}
}

func TestCustomTagConfigurationResource_ImportState_RejectsWrongID(t *testing.T) {
	r := NewCustomTagConfigurationResource().(resource.ResourceWithImportState)

	resp := &resource.ImportStateResponse{State: newTestState(t, customTagConfigurationSchema(t).Schema)}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "SomethingElse"}, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("ImportState() with a non-\"CustomTags\" ID did not produce a diagnostic")
	}
}

func TestCustomTagConfigurationResource_ImportState_AcceptsFixedID(t *testing.T) {
	r := NewCustomTagConfigurationResource().(resource.ResourceWithImportState)

	resp := &resource.ImportStateResponse{State: newTestState(t, customTagConfigurationSchema(t).Schema)}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "CustomTags"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ImportState() produced diagnostics: %v", resp.Diagnostics)
	}

	var id types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRootID(), &id)...)
	if id.ValueString() != "CustomTags" {
		t.Errorf("id = %q, want CustomTags", id.ValueString())
	}
}

// TestCustomTagConfigurationResource_ValidateConfig_RejectsDuplicateNames
// exercises ValidateConfig directly, since two tags sharing a name is
// otherwise representable in a Set (each object, taken as a whole, can
// still be a distinct set element even with the same name if other fields
// differ) and would not be caught by Terraform's own type system.
func TestCustomTagConfigurationResource_ValidateConfig_RejectsDuplicateNames(t *testing.T) {
	r := NewCustomTagConfigurationResource().(resource.ResourceWithValidateConfig)
	s := customTagConfigurationSchema(t).Schema
	ctx := context.Background()

	objType, ok := s.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("schema type is not a tftypes.Object")
	}
	tagsType := objType.AttributeTypes["tags"].(tftypes.Set)
	tagObjType := tagsType.ElementType.(tftypes.Object)

	tag := func(name string, mandatory bool) tftypes.Value {
		return tftypes.NewValue(tagObjType, map[string]tftypes.Value{
			"name":             tftypes.NewValue(tftypes.String, name),
			"mandatory":        tftypes.NewValue(tftypes.Bool, mandatory),
			"permitted_values": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
		})
	}

	tagsValue := tftypes.NewValue(tagsType, []tftypes.Value{
		tag("Owner", true),
		tag("Owner", false),
	})

	raw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":   tftypes.NewValue(tftypes.String, nil),
		"tags": tagsValue,
	})

	req := resource.ValidateConfigRequest{
		Config: tfsdk.Config{Schema: s, Raw: raw},
	}
	var resp resource.ValidateConfigResponse
	r.ValidateConfig(ctx, req, &resp)

	if !resp.Diagnostics.HasError() {
		t.Error("ValidateConfig() did not reject two tags sharing the name \"Owner\"")
	}
}
