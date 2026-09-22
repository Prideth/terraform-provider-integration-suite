// Command gendocs regenerates docs/feature-support.md from the canonical
// feature catalog in internal/features. It is the single place that turns
// features.Catalog into prose and tables, so that catalog and
// documentation cannot drift apart the way independently hand-maintained
// copies would — see CONTRIBUTING.md for the rule that every feature
// change updates the catalog and regenerates this file in the same
// change.
//
// Usage: go run ./cmd/gendocs > docs/feature-support.md
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/features"
)

// reasonOrder fixes the order SupportReason groups appear in the
// "Unsupported and partially supported features" section, and gives each a
// short, human-readable heading.
var reasonOrder = []struct {
	reason  features.SupportReason
	heading string
}{
	{features.ReasonNotImplemented, "Public API exists but provider implementation is pending"},
	{features.ReasonPublicAPIIncomplete, "Public API details are not fully confirmed"},
	{features.ReasonUnsafeTerraformLifecycle, "Public lifecycle insufficient for safe Terraform management"},
	{features.ReasonNoPublicAPI, "No suitable public SAP API"},
	{features.ReasonResearchRequired, "Further research required"},
	{features.ReasonOutOfScope, "Out of provider scope"},
}

func main() {
	var b strings.Builder

	fmt.Fprintln(&b, "# Feature Support")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "Generated from `internal/features/catalog.go` by `go run ./cmd/gendocs`. Do not edit by hand — regenerate it instead, and see `CONTRIBUTING.md` for when this is required.\n\n")
	fmt.Fprintln(&b, "This is what this provider version implements, derived from the same catalog the "+
		"provider binary itself exposes through `sapintegrationsuite_provider_features` (whose "+
		"`provider_version` attribute reports the exact version at apply time). It is queryable "+
		"directly from Terraform, with no SAP tenant connection required:")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "```hcl")
	fmt.Fprintln(&b, `data "sapintegrationsuite_provider_features" "all" {}`)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, `output "supported_features" {`)
	fmt.Fprintln(&b, `  value = [`)
	fmt.Fprintln(&b, `    for f in data.sapintegrationsuite_provider_features.all.features :`)
	fmt.Fprintln(&b, `    f.key`)
	fmt.Fprintln(&b, `    if f.support_status == "supported"`)
	fmt.Fprintln(&b, `  ]`)
	fmt.Fprintln(&b, `}`)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, `data "sapintegrationsuite_provider_feature" "one" {`)
	fmt.Fprintln(&b, `  key = "cloud_integration.value_mapping"`)
	fmt.Fprintln(&b, `}`)
	fmt.Fprintln(&b, "```")
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "## Relationship to the other capability documents")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "This document, `docs/api-capability-matrix.md`, `docs/provisioning-capability-matrix.md`, "+
		"and a possible future tenant-capability data source answer three related but distinct questions:")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "| Document | Answers |")
	fmt.Fprintln(&b, "|---|---|")
	fmt.Fprintln(&b, "| `docs/api-capability-matrix.md` / `docs/provisioning-capability-matrix.md` | What SAP's public APIs expose, independent of this provider |")
	fmt.Fprintln(&b, "| `docs/feature-support.md` (this document) / `sapintegrationsuite_provider_features` | What *this provider version* implements — static provider metadata, no SAP tenant required |")
	fmt.Fprintln(&b, "| A possible future `sapintegrationsuite_tenant_capabilities` data source (not implemented) | Which Integration Suite capabilities are *active in a specific SAP tenant* — would require SAP credentials and a reliable public discovery API, neither of which this provider assumes here |")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Do not conflate these: a feature can be fully supported by this provider and still be "+
		"unusable in a given tenant because the underlying SAP capability was never activated there, and "+
		"vice versa a capability can be active in every tenant while this provider still does not implement "+
		"a resource for it.")
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "## All features")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "| Feature | Domain | Status | Public API | Create | Read | Update | Delete | Import | Deploy | Terraform |")
	fmt.Fprintln(&b, "|---|---|---|---|---|---|---|---|---|---|---|")

	sorted := make([]features.Feature, len(features.Catalog))
	copy(sorted, features.Catalog)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Key < sorted[j].Key })

	for _, f := range sorted {
		fmt.Fprintf(&b, "| `%s` | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			f.Key, f.Domain, statusCell(f), boolCell(f.PublicAPI),
			opCell(f.Operations.Create), opCell(f.Operations.Read), opCell(f.Operations.Update),
			opCell(f.Operations.Delete), opCell(f.Operations.Import), opCell(f.Operations.Deploy),
			terraformCell(f),
		)
	}
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "## Unsupported and partially supported features")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Grouped by why, not just that. A feature can be `partial` and reachable via one of these "+
		"reasons too — see docs/feature-support.md's per-feature `Limitations` (the `limitations` attribute "+
		"in Terraform) for exactly what is and is not covered.")
	fmt.Fprintln(&b)

	for _, group := range reasonOrder {
		var matches []features.Feature
		for _, f := range sorted {
			if f.SupportStatus != features.StatusSupported && f.SupportReason == group.reason {
				matches = append(matches, f)
			}
		}
		if len(matches) == 0 {
			continue
		}

		fmt.Fprintf(&b, "### %s\n\n", group.heading)
		for _, f := range matches {
			statusNote := ""
			if f.SupportStatus == features.StatusPartial {
				statusNote = " (partial support already implemented — see Limitations below)"
			}
			fmt.Fprintf(&b, "- **`%s`** — %s%s\n", f.Key, f.Description, statusNote)
			for _, l := range f.Limitations {
				fmt.Fprintf(&b, "  - %s\n", l)
			}
		}
		fmt.Fprintln(&b)
	}

	if _, err := os.Stdout.WriteString(b.String()); err != nil {
		fmt.Fprintln(os.Stderr, "gendocs:", err)
		os.Exit(1)
	}
}

func boolCell(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func opCell(b bool) string {
	if b {
		return "Yes"
	}
	return "—"
}

func statusCell(f features.Feature) string {
	if f.SupportReason == "" {
		return string(f.SupportStatus)
	}
	return fmt.Sprintf("%s (%s)", f.SupportStatus, f.SupportReason)
}

func terraformCell(f features.Feature) string {
	hasResource := len(f.ResourceTypes) > 0
	hasDataSource := len(f.DataSourceTypes) > 0
	switch {
	case hasResource && hasDataSource:
		return "Resource + Data Source"
	case hasResource:
		return "Resource"
	case hasDataSource:
		return "Data Source"
	default:
		return "—"
	}
}
