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

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/securitycontent"
)

// secureParameterEntity is a GET response in the shape a tenant returned in
// September 2026: SecureParam null, DeployedOn as /Date(ms)/.
const secureParameterEntity = `{"d":{"Name":"MyAlias","Description":"d","SecureParam":null,"DeployedBy":"sb-client","DeployedOn":"\/Date(1790410073242)\/","Status":"DEPLOYED"}}`

func secureParameterSchema(t *testing.T) schema.Schema {
	t.Helper()
	var resp resource.SchemaResponse
	NewSecureParameterResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp.Schema
}

func TestSecureParameterResource_ValueIsWriteOnlyAndSensitive(t *testing.T) {
	attr, ok := secureParameterSchema(t).Attributes["secure_param_wo"].(schema.StringAttribute)
	if !ok {
		t.Fatal("secure_param_wo is not a StringAttribute")
	}
	if !attr.IsWriteOnly() || !attr.IsSensitive() {
		t.Error("secure_param_wo must be WriteOnly and Sensitive")
	}
}

// secureParameterRequest builds a plan and a configuration carrying the
// write-only value, the way Terraform passes them to Create and Update.
func secureParameterRequest(t *testing.T, s schema.Schema, value string) (tfsdk.Plan, tfsdk.Config) {
	t.Helper()
	ctx := context.Background()
	m := secureParameterModel{
		ID:                   types.StringValue("MyAlias"),
		Description:          types.StringValue("d"),
		SecureParamWO:        types.StringNull(),
		SecureParamWOVersion: types.StringValue("2"),
		Status:               types.StringUnknown(),
		DeployedBy:           types.StringUnknown(),
		DeployedOn:           types.StringUnknown(),
	}
	plan := tfsdk.Plan{Schema: s, Raw: newTestState(t, s).Raw}
	if diags := plan.Set(ctx, m); diags.HasError() {
		t.Fatalf("building plan: %v", diags)
	}
	m.SecureParamWO = types.StringValue(value)
	cfg := newTestState(t, s)
	if diags := cfg.Set(ctx, m); diags.HasError() {
		t.Fatalf("building config: %v", diags)
	}
	return plan, tfsdk.Config{Schema: s, Raw: cfg.Raw}
}

// SAP answers the create with 202 and no body; the resource must read the
// artifact back and never put the value into state.
func TestSecureParameterResource_Create(t *testing.T) {
	var posted map[string]string
	created := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost:
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Errorf("decoding body: %v", err)
			}
			created = true
			w.WriteHeader(http.StatusAccepted)
		case !created:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"code":"Not Found","message":{"lang":"en","value":"not found"}}}`))
		default:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(secureParameterEntity))
		}
	}))
	defer server.Close()

	r := &secureParameterResource{client: securitycontent.New(http.DefaultClient, server.URL)}
	s := secureParameterSchema(t)
	ctx := context.Background()

	plan, cfg := secureParameterRequest(t, s, "s3cr3t")
	resp := &resource.CreateResponse{State: newTestState(t, s)}
	r.Create(ctx, resource.CreateRequest{Plan: plan, Config: cfg}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Create() produced diagnostics: %v", resp.Diagnostics)
	}
	if posted["Name"] != "MyAlias" || posted["SecureParam"] != "s3cr3t" {
		t.Errorf("POST body = %v, want name and value", posted)
	}

	var got secureParameterModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &got)...)
	if !got.SecureParamWO.IsNull() {
		t.Error("secure_param_wo must never reach state")
	}
	if got.Status.ValueString() != "DEPLOYED" || got.DeployedOn.ValueString() == "" {
		t.Errorf("state = status %q deployed_on %q, want the values read back", got.Status.ValueString(), got.DeployedOn.ValueString())
	}
}

func TestSecureParameterResource_Create_RefusesExisting(t *testing.T) {
	posted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			posted = true
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(secureParameterEntity))
	}))
	defer server.Close()

	r := &secureParameterResource{client: securitycontent.New(http.DefaultClient, server.URL)}
	s := secureParameterSchema(t)

	plan, cfg := secureParameterRequest(t, s, "s3cr3t")
	resp := &resource.CreateResponse{State: newTestState(t, s)}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan, Config: cfg}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected Create to fail for an existing secure parameter")
	}
	if posted {
		t.Error("Create sent a POST for an existing name")
	}
}

func TestSecureParameterResource_Update_SendsValue(t *testing.T) {
	var method, path string
	var body map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			method, path = r.Method, r.URL.Path
			_ = json.NewDecoder(r.Body).Decode(&body)
			w.WriteHeader(http.StatusAccepted)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(secureParameterEntity))
	}))
	defer server.Close()

	r := &secureParameterResource{client: securitycontent.New(http.DefaultClient, server.URL)}
	s := secureParameterSchema(t)

	plan, cfg := secureParameterRequest(t, s, "rotated")
	resp := &resource.UpdateResponse{State: newTestState(t, s)}
	r.Update(context.Background(), resource.UpdateRequest{Plan: plan, Config: cfg}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Update() produced diagnostics: %v", resp.Diagnostics)
	}
	if method != http.MethodPut || path != "/api/v1/SecureParameters('MyAlias')" || body["SecureParam"] != "rotated" {
		t.Errorf("request = %s %s %v, want a PUT with the new value", method, path, body)
	}
}

func TestSecureParameterResource_Read_RemovesDeleted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"code":"Not Found","message":{"lang":"en","value":"not found"}}}`))
	}))
	defer server.Close()

	r := &secureParameterResource{client: securitycontent.New(http.DefaultClient, server.URL)}
	s := secureParameterSchema(t)
	ctx := context.Background()

	state := newTestState(t, s)
	resp0 := state.Set(ctx, secureParameterModel{
		ID: types.StringValue("MyAlias"), Description: types.StringNull(), SecureParamWO: types.StringNull(),
		SecureParamWOVersion: types.StringValue("1"), Status: types.StringNull(), DeployedBy: types.StringNull(),
		DeployedOn: types.StringNull(),
	})
	if resp0.HasError() {
		t.Fatalf("building state: %v", resp0)
	}

	resp := &resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Read() produced diagnostics: %v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Error("state should be removed when SAP reports 404")
	}
}
