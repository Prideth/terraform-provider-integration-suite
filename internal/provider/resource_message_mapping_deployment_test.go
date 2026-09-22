package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func messageMappingDeploymentSchema(t *testing.T) schema.Schema {
	t.Helper()
	r := NewMessageMappingDeploymentResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp.Schema
}

func TestMessageMappingDeploymentResource_SchemaRequiredOptionalComputed(t *testing.T) {
	s := messageMappingDeploymentSchema(t)

	cases := []struct {
		name               string
		required, computed bool
	}{
		{"id", false, true},
		{"package_id", true, false},
		{"mapping_id", true, false},
		{"mapping_version", true, false},
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

	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("expected a \"timeouts\" block")
	}
}

// TestMessageMappingDeploymentResource_IdentityIsImmutable confirms
// package_id and mapping_id replace the deployment resource, while
// mapping_version (the attribute a redeploy is driven by) does not — a
// version change should update the existing deployment in place, not
// replace it.
func TestMessageMappingDeploymentResource_IdentityIsImmutable(t *testing.T) {
	s := messageMappingDeploymentSchema(t)

	replaceExpected := map[string]bool{
		"package_id":      true,
		"mapping_id":      true,
		"mapping_version": false,
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

func TestMessageMappingDeploymentResource_ImportState(t *testing.T) {
	r := NewMessageMappingDeploymentResource().(resource.ResourceWithImportState)

	resp := &resource.ImportStateResponse{State: newTestState(t, messageMappingDeploymentSchema(t))}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "UTILITIES/customer-mapping"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ImportState() produced diagnostics: %v", resp.Diagnostics)
	}
}
