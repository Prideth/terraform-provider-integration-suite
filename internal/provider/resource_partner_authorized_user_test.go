package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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

func TestLowercaseUserValidator(t *testing.T) {
	for value, wantErr := range map[string]bool{
		"commuser1":           false,
		"sb-abc|it!b123":      false,
		"MyUser":              true,
		"p123456@example.com": false,
		"müller":              false,
		"Müller":              true,
		"MÜLLER":              true,
	} {
		resp := &validator.StringResponse{}
		lowercaseUserValidator{}.ValidateString(context.Background(), validator.StringRequest{
			Path:        path.Root("user"),
			ConfigValue: types.StringValue(value),
		}, resp)
		if got := resp.Diagnostics.HasError(); got != wantErr {
			t.Errorf("%q: error = %v, want %v", value, got, wantErr)
		}
	}
}

// Update must keep runtime_location_id in state and send the PUT to the
// Edge Integration Cell's service root.
func TestPartnerAuthorizedUserResource_Update_KeepsRuntimeLocation(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	r := &partnerAuthorizedUserResource{client: partnerdirectory.New(http.DefaultClient, server.URL)}
	s := partnerAuthorizedUserSchema(t).Schema
	ctx := context.Background()

	planned := partnerAuthorizedUserModel{
		ID:                types.StringValue("commuser1"),
		User:              types.StringValue("commuser1"),
		PartnerID:         types.StringValue("PartnerZ"),
		RuntimeLocationID: types.StringValue("edge1"),
	}
	plan := tfsdk.Plan{Schema: s, Raw: newTestState(t, s).Raw}
	if diags := plan.Set(ctx, planned); diags.HasError() {
		t.Fatalf("building plan: %v", diags)
	}

	resp := &resource.UpdateResponse{State: newTestState(t, s)}
	r.Update(ctx, resource.UpdateRequest{Plan: plan}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Update() produced diagnostics: %v", resp.Diagnostics)
	}
	if want := "/location/edge1/api/v1/AuthorizedUsers('commuser1')"; gotPath != want {
		t.Errorf("PUT path = %q, want %q", gotPath, want)
	}

	var got partnerAuthorizedUserModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &got)...)
	if got.RuntimeLocationID.ValueString() != "edge1" {
		t.Errorf("runtime_location_id = %q after Update, want edge1", got.RuntimeLocationID.ValueString())
	}
}
