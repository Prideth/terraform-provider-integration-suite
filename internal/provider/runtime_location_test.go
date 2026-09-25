package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/securitycontent"
)

func TestSplitLocatedImportID(t *testing.T) {
	cases := []struct {
		id      string
		parts   int
		loc     string
		rest    []string
		wantErr bool
	}{
		{"UTILITIES/metering", 2, "", []string{"UTILITIES", "metering"}, false},
		{"myedge/UTILITIES/metering", 2, "myedge", []string{"UTILITIES", "metering"}, false},
		{"BACKEND_BASIC", 1, "", []string{"BACKEND_BASIC"}, false},
		{"plant-a/BACKEND_BASIC", 1, "plant-a", []string{"BACKEND_BASIC"}, false},
		{"a/b/c/d", 2, "", nil, true},
		{"../x/y", 2, "", nil, true},
		{"UTILITIES/", 2, "", nil, true},
	}
	for _, c := range cases {
		loc, rest, err := splitLocatedImportID(c.id, c.parts)
		if (err != nil) != c.wantErr {
			t.Errorf("%q: err = %v, wantErr %v", c.id, err, c.wantErr)
			continue
		}
		if !c.wantErr && (loc != c.loc || !reflect.DeepEqual(rest, c.rest)) {
			t.Errorf("%q: got %q %v, want %q %v", c.id, loc, rest, c.loc, c.rest)
		}
	}
}

func TestRuntimeLocationAttribute_OnEveryPerRuntimeResource(t *testing.T) {
	for _, r := range []resource.Resource{
		NewIntegrationFlowDeploymentResource(), NewMessageMappingDeploymentResource(),
		NewScriptCollectionDeploymentResource(), NewValueMappingDeploymentResource(),
		NewIntegrationAdapterDeploymentResource(), NewUserCredentialResource(),
		NewOAuth2ClientCredentialResource(),
	} {
		var resp resource.SchemaResponse
		r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
		attr, ok := resp.Schema.Attributes["runtime_location_id"].(schema.StringAttribute)
		if !ok || !attr.Optional || attr.Computed {
			t.Errorf("%T: runtime_location_id must be an Optional, non-computed string", r)
			continue
		}
		var mods []interface {
			Description(context.Context) string
		}
		for _, m := range attr.PlanModifiers {
			mods = append(mods, m)
		}
		if !hasRequiresReplace(mods) {
			t.Errorf("%T: runtime_location_id must force replacement", r)
		}
	}
}

func TestUserCredentialRead_RoutesToEdgeIntegrationCell(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"d": {"Name": "BACKEND_BASIC", "Kind": "default", "User": "svc"}}`))
	}))
	defer server.Close()

	ctx := context.Background()
	res := NewUserCredentialResource().(*userCredentialResource)
	res.client = securitycontent.New(http.DefaultClient, server.URL)

	var sresp resource.SchemaResponse
	res.Schema(ctx, resource.SchemaRequest{}, &sresp)
	state := newTestState(t, sresp.Schema)
	state.SetAttribute(ctx, pathRootID(), "BACKEND_BASIC")
	state.SetAttribute(ctx, pathRoot("runtime_location_id"), "myedge")

	resp := &resource.ReadResponse{State: state}
	res.Read(ctx, resource.ReadRequest{State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Read() diagnostics: %v", resp.Diagnostics)
	}
	if want := "/location/myedge/api/v1/UserCredentials('BACKEND_BASIC')"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
	var loc types.String
	resp.State.GetAttribute(ctx, pathRoot("runtime_location_id"), &loc)
	if loc.ValueString() != "myedge" {
		t.Errorf("runtime_location_id after read = %q, want it kept", loc.ValueString())
	}
}
