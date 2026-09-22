package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/features"
)

// registeredResourceTypeNames returns the Terraform type name of every
// resource this provider registers, using the same ProviderTypeName the
// real provider server supplies.
func registeredResourceTypeNames(t *testing.T) []string {
	t.Helper()
	p := New("test")().(interface {
		Resources(context.Context) []func() resource.Resource
	})

	var names []string
	for _, factory := range p.Resources(context.Background()) {
		var metaResp resource.MetadataResponse
		factory().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "sapintegrationsuite"}, &metaResp)
		names = append(names, metaResp.TypeName)
	}
	return names
}

// registeredDataSourceTypeNames returns the Terraform type name of every
// data source this provider registers.
func registeredDataSourceTypeNames(t *testing.T) []string {
	t.Helper()
	p := New("test")().(interface {
		DataSources(context.Context) []func() datasource.DataSource
	})

	var names []string
	for _, factory := range p.DataSources(context.Background()) {
		var metaResp datasource.MetadataResponse
		factory().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "sapintegrationsuite"}, &metaResp)
		names = append(names, metaResp.TypeName)
	}
	return names
}

// catalogResourceTypes and catalogDataSourceTypes flatten every
// ResourceTypes/DataSourceTypes list across features.Catalog into a single
// set, for both directions of the consistency check below.
func catalogResourceTypes() map[string]string {
	out := map[string]string{}
	for _, f := range features.Catalog {
		for _, rt := range f.ResourceTypes {
			out[rt] = f.Key
		}
	}
	return out
}

func catalogDataSourceTypes() map[string]string {
	out := map[string]string{}
	for _, f := range features.Catalog {
		for _, dt := range f.DataSourceTypes {
			out[dt] = f.Key
		}
	}
	return out
}

// featureCatalogDataSourceTypeNames are the two data sources that describe
// the catalog itself; they are not, and never will be, entries in the
// catalog they expose, so the forward consistency check below exempts
// them explicitly rather than requiring a self-referential catalog entry.
func featureCatalogDataSourceTypeNames() map[string]bool {
	return map[string]bool{
		"sapintegrationsuite_provider_features": true,
		"sapintegrationsuite_provider_feature":  true,
	}
}

// TestFeatureCatalog_EveryRegisteredResourceIsCataloged is the check
// instructed by this phase: a developer who adds NewSomethingResource to
// the provider but forgets to add (or update) its features.Catalog entry
// should see this test fail, not discover the gap later in documentation
// review.
func TestFeatureCatalog_EveryRegisteredResourceIsCataloged(t *testing.T) {
	cataloged := catalogResourceTypes()

	for _, name := range registeredResourceTypeNames(t) {
		if _, ok := cataloged[name]; !ok {
			t.Errorf("resource %q is registered with the provider but does not appear in any features.Catalog entry's ResourceTypes", name)
		}
	}
}

// TestFeatureCatalog_EveryRegisteredDataSourceIsCataloged mirrors the
// resource check above for data sources, exempting the two catalog data
// sources themselves.
func TestFeatureCatalog_EveryRegisteredDataSourceIsCataloged(t *testing.T) {
	cataloged := catalogDataSourceTypes()
	exempt := featureCatalogDataSourceTypeNames()

	for _, name := range registeredDataSourceTypeNames(t) {
		if exempt[name] {
			continue
		}
		if _, ok := cataloged[name]; !ok {
			t.Errorf("data source %q is registered with the provider but does not appear in any features.Catalog entry's DataSourceTypes", name)
		}
	}
}

// TestFeatureCatalog_EveryCatalogedResourceIsRegistered is the reverse
// check: a features.Catalog entry must not claim a resource type that
// does not actually exist, which would otherwise let documentation
// generated from the catalog assert support for something unbuildable.
func TestFeatureCatalog_EveryCatalogedResourceIsRegistered(t *testing.T) {
	registered := make(map[string]bool)
	for _, name := range registeredResourceTypeNames(t) {
		registered[name] = true
	}

	for _, f := range features.Catalog {
		for _, rt := range f.ResourceTypes {
			if !registered[rt] {
				t.Errorf("feature %q claims resource type %q, but no such resource is registered with the provider", f.Key, rt)
			}
		}
	}
}

// TestFeatureCatalog_EveryCatalogedDataSourceIsRegistered mirrors the
// resource check above for data sources.
func TestFeatureCatalog_EveryCatalogedDataSourceIsRegistered(t *testing.T) {
	registered := make(map[string]bool)
	for _, name := range registeredDataSourceTypeNames(t) {
		registered[name] = true
	}

	for _, f := range features.Catalog {
		for _, dt := range f.DataSourceTypes {
			if !registered[dt] {
				t.Errorf("feature %q claims data source type %q, but no such data source is registered with the provider", f.Key, dt)
			}
		}
	}
}
