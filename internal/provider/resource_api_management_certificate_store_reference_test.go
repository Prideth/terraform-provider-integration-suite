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

func TestAPIManagementCertificateStoreReferenceResource_Configure_RequiresAPIManagementClient(t *testing.T) {
	r := NewAPIManagementCertificateStoreReferenceResource().(resource.ResourceWithConfigure)
	var resp resource.ConfigureResponse
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: &Data{}}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a configuration error when APIManagementClassicHTTPClient is nil")
	}
}

func TestAPIManagementCertificateStoreReferenceResource_CreateReadUpdateDelete(t *testing.T) {
	storeName := "ts"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"d": {"name": "ref1", "certificateStoreName": "` + storeName + `", "storeType": "TRUSTSTORE"}}`))
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"d": {"name": "ref1", "certificateStoreName": "` + storeName + `", "storeType": "TRUSTSTORE"}}`))
		case http.MethodPut, http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	r := &apiManagementCertificateStoreReferenceResource{client: apimanagementclassic.New(http.DefaultClient, server.URL)}

	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	s := schemaResp.Schema
	ctx := context.Background()
	objType := s.Type().TerraformType(ctx).(tftypes.Object)

	planRaw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                     tftypes.NewValue(tftypes.String, nil),
		"name":                   tftypes.NewValue(tftypes.String, "ref1"),
		"certificate_store_name": tftypes.NewValue(tftypes.String, storeName),
		"store_type":             tftypes.NewValue(tftypes.String, nil),
	})

	createReq := resource.CreateRequest{Plan: tfsdk.Plan{Schema: s, Raw: planRaw}}
	createResp := &resource.CreateResponse{State: newTestState(t, s)}
	r.Create(ctx, createReq, createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("Create() produced diagnostics: %v", createResp.Diagnostics)
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
