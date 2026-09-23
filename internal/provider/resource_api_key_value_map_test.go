package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apimanagementclassic"
)

func apiKeyValueMapSchema(t *testing.T) (resource.SchemaResponse, *apiKeyValueMapResource) {
	t.Helper()
	r := &apiKeyValueMapResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	return resp, r
}

func kvmPlanValue(t *testing.T, objType tftypes.Object, name, scope, scopeID string, encrypted bool, entries []tftypes.Value) tftypes.Value {
	t.Helper()
	entriesType := objType.AttributeTypes["entries"].(tftypes.List)
	return tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":        tftypes.NewValue(tftypes.String, nil),
		"name":      tftypes.NewValue(tftypes.String, name),
		"scope":     tftypes.NewValue(tftypes.String, scope),
		"scope_id":  tftypes.NewValue(tftypes.String, scopeID),
		"entries":   tftypes.NewValue(entriesType, entries),
		"encrypted": tftypes.NewValue(tftypes.Bool, encrypted),
	})
}

func TestAPIKeyValueMapResource_ValidateConfig_RejectsEncrypted(t *testing.T) {
	schemaResp, r := apiKeyValueMapSchema(t)
	s := schemaResp.Schema
	ctx := context.Background()
	objType := s.Type().TerraformType(ctx).(tftypes.Object)

	configRaw := kvmPlanValue(t, objType, "kvm1", "APIPROXY", "proxy1", true, []tftypes.Value{})

	req := resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: s, Raw: configRaw}}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a validation error for encrypted = true")
	}
}

func TestAPIKeyValueMapResource_ValidateConfig_AllowsUnencrypted(t *testing.T) {
	schemaResp, r := apiKeyValueMapSchema(t)
	s := schemaResp.Schema
	ctx := context.Background()
	objType := s.Type().TerraformType(ctx).(tftypes.Object)

	configRaw := kvmPlanValue(t, objType, "kvm1", "APIPROXY", "proxy1", false, []tftypes.Value{})

	req := resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: s, Raw: configRaw}}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics for encrypted = false: %v", resp.Diagnostics)
	}
}

func TestAPIKeyValueMapResource_Configure_RequiresAPIManagementClient(t *testing.T) {
	r := NewAPIKeyValueMapResource().(resource.ResourceWithConfigure)
	var resp resource.ConfigureResponse
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: &Data{}}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a configuration error when APIManagementClassicHTTPClient is nil")
	}
}

func TestAPIKeyValueMapResource_CreateReadDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"d": {"name": "kvm1", "scope": "APIPROXY", "scopeId": "proxy1", "isEncrypted": false, "genericKeyMapEntryValues": [{"name": "default", "value": "secret", "mapName": "kvm1", "scope": "APIPROXY", "scopeId": "proxy1"}]}}`))
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"d": {"name": "kvm1", "scope": "APIPROXY", "scopeId": "proxy1", "isEncrypted": false, "genericKeyMapEntryValues": [{"name": "default", "value": "secret", "mapName": "kvm1", "scope": "APIPROXY", "scopeId": "proxy1"}]}}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	r := &apiKeyValueMapResource{client: apimanagementclassic.New(http.DefaultClient, server.URL)}
	schemaResp, _ := apiKeyValueMapSchema(t)
	s := schemaResp.Schema
	ctx := context.Background()
	objType := s.Type().TerraformType(ctx).(tftypes.Object)

	entryObjType := objType.AttributeTypes["entries"].(tftypes.List).ElementType.(tftypes.Object)
	entry := tftypes.NewValue(entryObjType, map[string]tftypes.Value{
		"key":   tftypes.NewValue(tftypes.String, "default"),
		"value": tftypes.NewValue(tftypes.String, "secret"),
	})

	planRaw := kvmPlanValue(t, objType, "kvm1", "APIPROXY", "proxy1", false, []tftypes.Value{entry})

	createReq := resource.CreateRequest{Plan: tfsdk.Plan{Schema: s, Raw: planRaw}}
	createResp := &resource.CreateResponse{State: newTestState(t, s)}
	r.Create(ctx, createReq, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create() produced diagnostics: %v", createResp.Diagnostics)
	}

	var got apiKeyValueMapModel
	createResp.Diagnostics.Append(createResp.State.Get(ctx, &got)...)
	if got.ID.ValueString() != "kvm1/APIPROXY/proxy1" {
		t.Errorf("id = %q, want kvm1/APIPROXY/proxy1", got.ID.ValueString())
	}

	readReq := resource.ReadRequest{State: createResp.State}
	readResp := &resource.ReadResponse{State: createResp.State}
	r.Read(ctx, readReq, readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("Read() produced diagnostics: %v", readResp.Diagnostics)
	}

	deleteReq := resource.DeleteRequest{State: readResp.State}
	deleteResp := &resource.DeleteResponse{}
	r.Delete(ctx, deleteReq, deleteResp)
	if deleteResp.Diagnostics.HasError() {
		t.Fatalf("Delete() produced diagnostics: %v", deleteResp.Diagnostics)
	}
}

func TestAPIKeyValueMapResource_ImportState_ParsesCompositeID(t *testing.T) {
	r := &apiKeyValueMapResource{}
	schemaResp, _ := apiKeyValueMapSchema(t)
	s := schemaResp.Schema
	ctx := context.Background()

	req := resource.ImportStateRequest{ID: "kvm1/APIPROXY/proxy1"}
	resp := &resource.ImportStateResponse{State: newTestState(t, s)}
	r.ImportState(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ImportState() produced diagnostics: %v", resp.Diagnostics)
	}

	var got apiKeyValueMapModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &got)...)
	if got.Name.ValueString() != "kvm1" || got.Scope.ValueString() != "APIPROXY" || got.ScopeID.ValueString() != "proxy1" {
		t.Errorf("parsed = %+v", got)
	}
}

func TestAPIKeyValueMapResource_ImportState_RejectsMalformedID(t *testing.T) {
	r := &apiKeyValueMapResource{}
	schemaResp, _ := apiKeyValueMapSchema(t)
	s := schemaResp.Schema
	ctx := context.Background()

	req := resource.ImportStateRequest{ID: "not-composite"}
	resp := &resource.ImportStateResponse{State: newTestState(t, s)}
	r.ImportState(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic error for a malformed import ID")
	}
}
