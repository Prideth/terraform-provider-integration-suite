package cloudintegration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_AtLocation_RoutesThroughLocationPrefix(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"d": {"Id": "Order_Flow", "Version": "1.0.3", "Status": "STARTED"}}`))
	}))
	defer server.Close()

	base := New(http.DefaultClient, server.URL)
	edge, err := base.AtLocation("myedge")
	if err != nil {
		t.Fatalf("AtLocation() error: %v", err)
	}
	if _, err := edge.GetRuntimeArtifact(context.Background(), "Order_Flow"); err != nil {
		t.Fatalf("GetRuntimeArtifact() error: %v", err)
	}
	if want := "/location/myedge/api/v1/IntegrationRuntimeArtifacts('Order_Flow')"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}

	if same, _ := base.AtLocation(""); same != base {
		t.Error("AtLocation(\"\") must return the cloud client unchanged")
	}
	if _, err := base.AtLocation("../x"); err == nil {
		t.Error("AtLocation must reject IDs that are not a single safe path segment")
	}
}
