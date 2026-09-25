package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

type fakeConfigurationServer struct {
	mu     sync.Mutex
	values map[string]string
	types  map[string]string
	puts   []map[string]any
}

func newFakeConfigurationServer(t *testing.T, f *fakeConfigurationServer) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		base := "/api/v1/IntegrationDesigntimeArtifacts(Id='Order_Flow',Version='1.0.3')"
		switch {
		case r.Method == http.MethodGet && r.URL.Path == base+"/Configurations":
			var items []string
			for k, v := range f.values {
				items = append(items, `{"ParameterKey":"`+k+`","ParameterValue":"`+v+`","DataType":"`+f.types[k]+`"}`)
			}
			_, _ = w.Write([]byte(`{"d":{"results":[` + strings.Join(items, ",") + `]}}`))
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, base+"/$links/Configurations('"):
			key := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, base+"/$links/Configurations('"), "')")
			raw, _ := io.ReadAll(r.Body)
			var body map[string]any
			_ = json.Unmarshal(raw, &body)
			body["_key"] = key
			f.puts = append(f.puts, body)
			f.values[key] = body["ParameterValue"].(string)
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func configurationResourceWithClient(url string) *integrationFlowConfigurationResource {
	return &integrationFlowConfigurationResource{client: cloudintegration.New(http.DefaultClient, url)}
}

func configurationSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	var resp resource.SchemaResponse
	NewIntegrationFlowConfigurationResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

func configurationPlan(t *testing.T, params map[string]string) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	s := configurationSchema(t).Schema
	plan := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
	m, diags := types.MapValueFrom(ctx, types.StringType, params)
	if diags.HasError() {
		t.Fatalf("building map: %v", diags)
	}
	diags = plan.Set(ctx, integrationFlowConfigurationModel{
		ID:          types.StringUnknown(),
		FlowID:      types.StringValue("Order_Flow"),
		FlowVersion: types.StringValue("1.0.3"),
		Parameters:  m,
	})
	if diags.HasError() {
		t.Fatalf("setting plan: %v", diags)
	}
	return plan
}

func TestIntegrationFlowConfiguration_CreateWritesValuesAndKeepsDataType(t *testing.T) {
	fake := &fakeConfigurationServer{
		values: map[string]string{"receiver_host": "erp-dev.example.invalid", "batch_size": "100", "unmanaged": "keep"},
		types:  map[string]string{"receiver_host": "xsd:string", "batch_size": "xsd:integer", "unmanaged": "xsd:string"},
	}
	server := newFakeConfigurationServer(t, fake)
	defer server.Close()

	ctx := context.Background()
	r := configurationResourceWithClient(server.URL)
	s := configurationSchema(t).Schema
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}}
	r.Create(ctx, resource.CreateRequest{Plan: configurationPlan(t, map[string]string{
		"receiver_host": "erp-prod.example.invalid", "batch_size": "250",
	})}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Create() diagnostics: %v", resp.Diagnostics)
	}

	if len(fake.puts) != 2 {
		t.Fatalf("PUT count = %d, want 2 (only managed keys)", len(fake.puts))
	}
	for _, p := range fake.puts {
		if p["_key"] == "batch_size" && p["DataType"] != "xsd:integer" {
			t.Errorf("batch_size sent DataType %v, want the existing xsd:integer", p["DataType"])
		}
	}
	if fake.values["unmanaged"] != "keep" {
		t.Error("an unmanaged parameter was changed")
	}
}

func TestIntegrationFlowConfiguration_UnknownKeyFailsBeforeWriting(t *testing.T) {
	fake := &fakeConfigurationServer{
		values: map[string]string{"receiver_host": "a"},
		types:  map[string]string{"receiver_host": "xsd:string"},
	}
	server := newFakeConfigurationServer(t, fake)
	defer server.Close()

	ctx := context.Background()
	r := configurationResourceWithClient(server.URL)
	s := configurationSchema(t).Schema
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}}
	r.Create(ctx, resource.CreateRequest{Plan: configurationPlan(t, map[string]string{
		"receiver_host": "b", "receiver_port": "typo",
	})}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error for an unknown parameter key")
	}
	detail := resp.Diagnostics.Errors()[0].Detail()
	if !strings.Contains(detail, "receiver_port") || !strings.Contains(detail, "receiver_host") {
		t.Errorf("error %q should name the unknown key and list the available ones", detail)
	}
	if len(fake.puts) != 0 {
		t.Errorf("PUT count = %d, want 0: nothing may be written when a key is unknown", len(fake.puts))
	}
}

func TestIntegrationFlowConfiguration_ReadReportsOnlyManagedKeys(t *testing.T) {
	fake := &fakeConfigurationServer{
		values: map[string]string{"receiver_host": "changed-in-ui", "unmanaged": "x"},
		types:  map[string]string{"receiver_host": "xsd:string", "unmanaged": "xsd:string"},
	}
	server := newFakeConfigurationServer(t, fake)
	defer server.Close()

	ctx := context.Background()
	r := configurationResourceWithClient(server.URL)
	plan := configurationPlan(t, map[string]string{"receiver_host": "erp.example.invalid", "gone_key": "v"})
	state := tfsdk.State(plan)
	resp := &resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Read() diagnostics: %v", resp.Diagnostics)
	}

	var got integrationFlowConfigurationModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &got)...)
	var params map[string]string
	resp.Diagnostics.Append(got.Parameters.ElementsAs(ctx, &params, false)...)
	if len(params) != 1 || params["receiver_host"] != "changed-in-ui" {
		t.Errorf("parameters = %v, want only the managed, still existing key with its current value", params)
	}
}

func TestIntegrationFlowConfiguration_DeleteWarnsAndCallsNothing(t *testing.T) {
	var resp resource.DeleteResponse
	NewIntegrationFlowConfigurationResource().Delete(context.Background(), resource.DeleteRequest{}, &resp)
	if resp.Diagnostics.HasError() || resp.Diagnostics.WarningsCount() != 1 {
		t.Errorf("Delete() diagnostics = %v, want exactly one warning and no error", resp.Diagnostics)
	}
}
