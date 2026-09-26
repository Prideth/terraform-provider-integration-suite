package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

// TestWaitForRuntimeArtifact_Started covers the ordinary success path used
// by both sapintegrationsuite_integration_flow_deployment and
// sapintegrationsuite_value_mapping_deployment: SAP eventually reports
// STARTED for the deployed artifact.
func TestWaitForRuntimeArtifact_Started(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"Id": "company-codes", "Version": "1.0.0", "Status": "STARTED"}}`))
	}))
	defer server.Close()

	client := cloudintegration.New(http.DefaultClient, server.URL)

	artifact, err := waitForRuntimeArtifact(context.Background(), client, "company-codes")
	if err != nil {
		t.Fatalf("waitForRuntimeArtifact() error: %v", err)
	}
	if artifact.Status != cloudintegration.StatusStarted {
		t.Errorf("Status = %q, want %q", artifact.Status, cloudintegration.StatusStarted)
	}
}

// TestWaitForRuntimeArtifact_Error covers a deployment that SAP reports as
// failed: waitForRuntimeArtifact must return an error carrying SAP's error
// text rather than treating ERROR as a transient state to keep polling
// through. The tenant $metadata declares ErrorInformation as a navigation
// property to a media entity, so the status read carries only a __deferred
// link and the text comes from ErrorInformation/$value.
func TestWaitForRuntimeArtifact_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if strings.HasSuffix(r.URL.Path, "/ErrorInformation/$value") {
			_, _ = w.Write([]byte("invalid mapping schema\n"))
			return
		}
		_, _ = w.Write([]byte(`{"d": {"Id": "company-codes", "Version": "1.0.0", "Status": "ERROR", "ErrorInformation": {"__deferred": {"uri": "https://host/api/v1/IntegrationRuntimeArtifacts('company-codes')/ErrorInformation"}}}}`))
	}))
	defer server.Close()

	client := cloudintegration.New(http.DefaultClient, server.URL)

	_, err := waitForRuntimeArtifact(context.Background(), client, "company-codes")
	if err == nil {
		t.Fatal("expected an error for a deployment reported as ERROR")
	}
	if !strings.Contains(err.Error(), "invalid mapping schema") {
		t.Errorf("error = %q, want it to contain SAP's error information", err.Error())
	}
}

// When the error text cannot be read, the deployment still fails with a
// generic reason instead of hiding the failure behind the read error.
func TestWaitForRuntimeArtifact_ErrorWithoutDetail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/ErrorInformation/$value") {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error": {"code": "NOT_FOUND", "message": {"value": "no error information"}}}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"Id": "company-codes", "Version": "1.0.0", "Status": "ERROR"}}`))
	}))
	defer server.Close()

	client := cloudintegration.New(http.DefaultClient, server.URL)

	_, err := waitForRuntimeArtifact(context.Background(), client, "company-codes")
	if err == nil || !strings.Contains(err.Error(), "without further detail") {
		t.Fatalf("error = %v, want the generic deployment failure", err)
	}
}

// TestWaitForRuntimeArtifact_Timeout covers a deployment that never reaches
// a terminal state before the caller's context expires (the resource's
// configured timeout): waitForRuntimeArtifact must stop polling and return
// promptly once ctx is done, never block past it.
func TestWaitForRuntimeArtifact_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"Id": "company-codes", "Version": "1.0.0", "Status": "STARTING"}}`))
	}))
	defer server.Close()

	client := cloudintegration.New(http.DefaultClient, server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := waitForRuntimeArtifact(ctx, client, "company-codes")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected a timeout error, got nil")
	}
	if elapsed > 5*time.Second {
		t.Errorf("waitForRuntimeArtifact() took %s after context expiry, want it to return promptly", elapsed)
	}
}

// TestWaitForRuntimeArtifact_NotFoundThenStarted covers the documented
// not-found-yet window right after a deploy is accepted, before SAP's
// asynchronous deployment pipeline has created the runtime artifact: a 404
// must be treated as "keep polling", not as a terminal failure.
func TestWaitForRuntimeArtifact_NotFoundThenStarted(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error": {"code": "NOT_FOUND", "message": {"value": "not yet deployed"}}}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"d": {"Id": "company-codes", "Version": "1.0.0", "Status": "STARTED"}}`))
	}))
	defer server.Close()

	client := cloudintegration.New(http.DefaultClient, server.URL)

	artifact, err := waitForRuntimeArtifact(context.Background(), client, "company-codes")
	if err != nil {
		t.Fatalf("waitForRuntimeArtifact() error: %v", err)
	}
	if artifact.Status != cloudintegration.StatusStarted {
		t.Errorf("Status = %q, want %q", artifact.Status, cloudintegration.StatusStarted)
	}
	if calls < 2 {
		t.Errorf("calls = %d, want at least 2 (a 404 followed by a successful poll)", calls)
	}
}
