package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

// SAP rejects a package without ShortText ("Property 'ShortText' cannot be
// empty", tenant test September 2026), so the attribute is required.
func TestIntegrationPackageResource_ShortTextIsRequired(t *testing.T) {
	var resp resource.SchemaResponse
	NewIntegrationPackageResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	attr, ok := resp.Schema.Attributes["short_text"].(schema.StringAttribute)
	if !ok || !attr.Required {
		t.Fatal("short_text must be a required string attribute")
	}
}

// SAP returns a plain description wrapped in <p>...</p>; the model must hold
// the value as configured, or every refresh would show drift.
func TestPackageToModel_UnwrapsDescription(t *testing.T) {
	m := packageToModel(&cloudintegration.Package{
		ID: "P", Name: "P", Description: "<p>tf-acc probe</p>", ShortText: "short",
	})
	if m.Description.ValueString() != "tf-acc probe" {
		t.Errorf("description = %q, want the unwrapped text", m.Description.ValueString())
	}
	if m.ShortText.ValueString() != "short" {
		t.Errorf("short_text = %q, want short", m.ShortText.ValueString())
	}

	empty := packageToModel(&cloudintegration.Package{ID: "P", Name: "P", Description: "<p></p>", ShortText: "s"})
	if !empty.Description.IsNull() {
		t.Errorf("description = %q, want null for SAP's empty <p></p>", empty.Description.ValueString())
	}
}
