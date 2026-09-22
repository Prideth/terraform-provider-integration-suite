package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestPartnerDataSource_Schema(t *testing.T) {
	d := NewPartnerDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	attr, ok := resp.Schema.Attributes["pid"]
	if !ok {
		t.Fatal("missing attribute \"pid\"")
	}
	if !attr.IsRequired() {
		t.Error("pid should be Required")
	}
}

func TestPartnersDataSource_Schema(t *testing.T) {
	d := NewPartnersDataSource()
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() produced diagnostics: %v", resp.Diagnostics)
	}
	attr, ok := resp.Schema.Attributes["pids"]
	if !ok {
		t.Fatal("missing attribute \"pids\"")
	}
	if !attr.IsComputed() {
		t.Error("pids should be Computed")
	}
}
