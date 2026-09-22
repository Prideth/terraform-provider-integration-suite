package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/partnerdirectory"
)

func partnerAuthorizedUserSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := NewPartnerAuthorizedUserResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

func TestPartnerAuthorizedUserResource_SchemaRequiredComputed(t *testing.T) {
	s := partnerAuthorizedUserSchema(t).Schema

	cases := []struct {
		name               string
		required, computed bool
	}{
		{"id", false, true},
		{"user", true, false},
		{"partner_id", true, false},
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

func TestPartnerAuthorizedUserResource_ImportState(t *testing.T) {
	r := NewPartnerAuthorizedUserResource().(resource.ResourceWithImportState)

	resp := &resource.ImportStateResponse{State: newTestState(t, partnerAuthorizedUserSchema(t).Schema)}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "commuser1"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ImportState() produced diagnostics: %v", resp.Diagnostics)
	}

	var user, id types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRoot("user"), &user)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRootID(), &id)...)
	if user.ValueString() != "commuser1" {
		t.Errorf("user = %q, want commuser1", user.ValueString())
	}
	if id.ValueString() != "commuser1" {
		t.Errorf("id = %q, want commuser1", id.ValueString())
	}
}

func TestAuthorizedUserToModel(t *testing.T) {
	au := &partnerdirectory.AuthorizedUser{User: "commuser1", Pid: "PartnerZ"}
	got := authorizedUserToModel(au)
	if got.ID.ValueString() != "commuser1" {
		t.Errorf("ID = %q, want commuser1", got.ID.ValueString())
	}
	if got.PartnerID.ValueString() != "PartnerZ" {
		t.Errorf("PartnerID = %q, want PartnerZ", got.PartnerID.ValueString())
	}
}
