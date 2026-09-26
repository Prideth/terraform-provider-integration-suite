package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/partnerdirectory"
)

func partnerUserCredentialParameterSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := NewPartnerUserCredentialParameterResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

func TestPartnerUserCredentialParameterResource_SchemaRequiredComputed(t *testing.T) {
	s := partnerUserCredentialParameterSchema(t).Schema

	cases := []struct {
		name               string
		required, computed bool
	}{
		{"id", false, true},
		{"partner_id", true, false},
		{"parameter_id", true, false},
		{"user", true, false},
		{"password_wo", true, false},
		{"password_wo_version", true, false},
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

// TestPartnerUserCredentialParameterResource_PasswordIsWriteOnlyAndSensitive
// is the safety property this whole resource depends on: password_wo must
// never be persisted to plan or state.
func TestPartnerUserCredentialParameterResource_PasswordIsWriteOnlyAndSensitive(t *testing.T) {
	s := partnerUserCredentialParameterSchema(t).Schema

	attr, ok := s.Attributes["password_wo"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("password_wo is not a StringAttribute")
	}
	if !attr.IsWriteOnly() {
		t.Error("password_wo must be WriteOnly")
	}
	if !attr.IsSensitive() {
		t.Error("password_wo must be Sensitive")
	}
}

// Only the identity forces replacement. user and password_wo_version
// change in place, through the POST SAP documents as this entity's update.
func TestPartnerUserCredentialParameterResource_ReplacementFields(t *testing.T) {
	s := partnerUserCredentialParameterSchema(t).Schema

	for name, wantReplace := range map[string]bool{
		"partner_id":          true,
		"parameter_id":        true,
		"user":                false,
		"password_wo_version": false,
	} {
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
			t.Errorf("%s: RequiresReplace = %v, want %v", name, got, wantReplace)
		}
	}
}

// userCredentialTestRequest builds a plan and a configuration carrying the
// write-only password, the way Terraform hands them to Create and Update.
func userCredentialTestRequest(t *testing.T, s schema.Schema, m partnerUserCredentialParameterModel, password string) (tfsdk.Plan, tfsdk.Config) {
	t.Helper()
	ctx := context.Background()

	plan := tfsdk.Plan{Schema: s, Raw: newTestState(t, s).Raw}
	if diags := plan.Set(ctx, m); diags.HasError() {
		t.Fatalf("building plan: %v", diags)
	}

	m.PasswordWO = types.StringValue(password)
	cfg := newTestState(t, s)
	if diags := cfg.Set(ctx, m); diags.HasError() {
		t.Fatalf("building config: %v", diags)
	}
	return plan, tfsdk.Config{Schema: s, Raw: cfg.Raw}
}

func sampleUserCredentialParameter() partnerUserCredentialParameterModel {
	return partnerUserCredentialParameterModel{
		ID:                types.StringUnknown(),
		PartnerID:         types.StringValue("Receiver_1"),
		ParameterID:       types.StringValue("USER"),
		User:              types.StringValue("commuser2"),
		PasswordWO:        types.StringNull(),
		PasswordWOVersion: types.StringValue("2"),
		RuntimeLocationID: types.StringValue("edge1"),
	}
}

func TestPartnerUserCredentialParameterResource_Update_PostsUserAndPassword(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decoding body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"d":{"Pid":"Receiver_1","Id":"USER","User":"commuser2","Password":null}}`))
	}))
	defer server.Close()

	r := &partnerUserCredentialParameterResource{client: partnerdirectory.New(http.DefaultClient, server.URL)}
	s := partnerUserCredentialParameterSchema(t).Schema
	ctx := context.Background()

	plan, cfg := userCredentialTestRequest(t, s, sampleUserCredentialParameter(), "new-secret")
	resp := &resource.UpdateResponse{State: newTestState(t, s)}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, Config: cfg}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Update() produced diagnostics: %v", resp.Diagnostics)
	}

	if gotMethod != http.MethodPost || gotPath != "/location/edge1/api/v1/UserCredentialParameters" {
		t.Errorf("request = %s %s, want POST to the entity set", gotMethod, gotPath)
	}
	if gotBody["User"] != "commuser2" || gotBody["Password"] != "new-secret" {
		t.Errorf("body = %v, want the new user and password", gotBody)
	}

	var got partnerUserCredentialParameterModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &got)...)
	if !got.PasswordWO.IsNull() {
		t.Error("password_wo must never reach state")
	}
	if got.PasswordWOVersion.ValueString() != "2" || got.RuntimeLocationID.ValueString() != "edge1" {
		t.Errorf("state = version %q location %q, want 2 and edge1", got.PasswordWOVersion.ValueString(), got.RuntimeLocationID.ValueString())
	}
}

// SAP's POST overwrites an existing credential, so Create must stop when
// one is already there instead of replacing a password it does not own.
func TestPartnerUserCredentialParameterResource_Create_RefusesExisting(t *testing.T) {
	posted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			posted = true
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"d":{"Pid":"Receiver_1","Id":"USER","User":"someone","Password":null}}`))
	}))
	defer server.Close()

	r := &partnerUserCredentialParameterResource{client: partnerdirectory.New(http.DefaultClient, server.URL)}
	s := partnerUserCredentialParameterSchema(t).Schema

	plan, cfg := userCredentialTestRequest(t, s, sampleUserCredentialParameter(), "secret")
	resp := &resource.CreateResponse{State: newTestState(t, s)}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan, Config: cfg}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected Create to fail for an existing credential")
	}
	if posted {
		t.Error("Create sent a POST and overwrote the existing credential")
	}
}

func TestPartnerUserCredentialParameterResource_ImportState(t *testing.T) {
	r := NewPartnerUserCredentialParameterResource().(resource.ResourceWithImportState)

	resp := &resource.ImportStateResponse{State: newTestState(t, partnerUserCredentialParameterSchema(t).Schema)}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "Receiver_1/USER"}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ImportState() produced diagnostics: %v", resp.Diagnostics)
	}

	var partnerID, parameterID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRoot("partner_id"), &partnerID)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), pathRoot("parameter_id"), &parameterID)...)
	if partnerID.ValueString() != "Receiver_1" {
		t.Errorf("partner_id = %q, want Receiver_1", partnerID.ValueString())
	}
	if parameterID.ValueString() != "USER" {
		t.Errorf("parameter_id = %q, want USER", parameterID.ValueString())
	}
}
