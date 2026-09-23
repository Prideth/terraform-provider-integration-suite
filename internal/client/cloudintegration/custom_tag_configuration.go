package cloudintegration

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const (
	customTagConfigurationsEntitySet = "CustomTagConfigurations"

	// customTagConfigurationID is the fixed, tenant-wide-singleton key SAP
	// documents for this entity — confirmed directly from SAP's own "Get
	// Custom Tags Defined on the Tenant" example request:
	// GET /CustomTagConfigurations('CustomTags')/$value. There is exactly
	// one custom tag configuration per tenant, and this is its only
	// confirmed identity.
	customTagConfigurationID = "CustomTags"
)

// CustomTag is one entry in the tenant's custom tag configuration: a tag
// name every integration package can be classified with, whether it is
// mandatory, and (optionally) a closed list of permitted values. Field
// names and the wrapping "customTagsConfiguration" JSON shape are
// confirmed directly from SAP's own "Create New Custom Tags Configuration"
// and "Get Custom Tags Defined on the Tenant" documentation pages.
type CustomTag struct {
	Name            string   `json:"tagName"`
	Mandatory       bool     `json:"isMandatory"`
	PermittedValues []string `json:"permittedValues,omitempty"`
}

// customTagsConfigurationContent is the decoded JSON shape SAP documents
// for a custom tag configuration, confirmed verbatim from SAP's own
// example: {"customTagsConfiguration":[{"tagName":"Owner","isMandatory":true}]}.
type customTagsConfigurationContent struct {
	CustomTagsConfiguration []CustomTag `json:"customTagsConfiguration"`
}

// customTagConfigurationWriteRequest is the outer request body shape SAP's
// Create endpoint documents: a single property, CustomTagsConfigurationContent,
// holding the base64-encoded, already-JSON-encoded configuration above.
// Confirmed verbatim from SAP's own documented example request.
type customTagConfigurationWriteRequest struct {
	CustomTagsConfigurationContent string `json:"CustomTagsConfigurationContent"`
}

func customTagConfigurationValuePath() string {
	return v2.BuildPath(customTagConfigurationsEntitySet, v2.KeyPredicate(customTagConfigurationID), "") + "/$value"
}

// GetCustomTagConfiguration reads the tenant's current custom tag
// configuration. Confirmed directly from SAP's own documentation: `GET
// /api/v1/CustomTagConfigurations('CustomTags')/$value`. Unlike every
// other GET in this client, the response is not wrapped in the standard
// OData "d" envelope: $value is OData's raw-media-stream convention, so
// this decodes the response body directly as the documented
// customTagsConfiguration JSON shape, never through v2.DecodeEntity.
//
// A tenant that has never had a custom tag configuration created on it
// surfaces as a 404 *apierror.Error from this call — this project could
// not confirm that behavior against a live tenant, but it is the standard
// OData interpretation of requesting $value from an entity that does not
// exist, and callers (see resource_custom_tag_configuration.go and
// data_source_custom_tag_configuration.go) already handle 404 as "no
// configuration exists yet" rather than treating it as an unexpected
// error.
func (c *Client) GetCustomTagConfiguration(ctx context.Context) ([]CustomTag, error) {
	body, err := c.odata.Get(ctx, customTagConfigurationValuePath())
	if err != nil {
		return nil, err
	}

	var content customTagsConfigurationContent
	if err := json.Unmarshal(body, &content); err != nil {
		return nil, fmt.Errorf("cloudintegration: decoding custom tag configuration: %w", err)
	}
	return content.CustomTagsConfiguration, nil
}

// SetCustomTagConfiguration creates or wholesale-replaces the tenant's
// custom tag configuration. SAP documents a single write operation for
// this entity: `POST /api/v1/CustomTagConfigurations` with the complete
// desired configuration, base64-encoded, as CustomTagsConfigurationContent;
// SAP's documentation adds "If a custom tags configuration is already
// available on the tenant, add the following query parameter to the
// request: Overwrite=true". This client always sends Overwrite=true,
// on both the very first write and every subsequent one: SAP's own
// example never shows what a plain POST does when a configuration
// already exists (a 409 conflict is plausible but unconfirmed, mirroring
// the confirmed duplicate-ID-is-an-error behavior for Integration
// Adapters), while Overwrite=true is documented to work when one already
// exists and there is no stated reason it would behave differently, or
// worse, on a tenant with no prior configuration — so always sending it
// avoids depending on an unconfirmed distinction entirely.
//
// tags is sorted by name, and each tag's PermittedValues sorted
// alphabetically, before encoding, so the request body — and therefore
// what this method actually sends to SAP — is deterministic regardless
// of what order the caller's own data happened to be in (Terraform's
// nested-block-to-Go-slice conversion does not guarantee it matches
// configuration source order). tags itself is never mutated; sorting
// operates on a copy.
func (c *Client) SetCustomTagConfiguration(ctx context.Context, tags []CustomTag) error {
	sorted := sortedCustomTags(tags)

	decoded, err := json.Marshal(customTagsConfigurationContent{CustomTagsConfiguration: sorted})
	if err != nil {
		return fmt.Errorf("cloudintegration: encoding custom tag configuration: %w", err)
	}

	payload, err := json.Marshal(customTagConfigurationWriteRequest{
		CustomTagsConfigurationContent: base64.StdEncoding.EncodeToString(decoded),
	})
	if err != nil {
		return fmt.Errorf("cloudintegration: encoding custom tag configuration request: %w", err)
	}

	path := v2.BuildPath(customTagConfigurationsEntitySet, "", "Overwrite=true")
	_, err = c.odata.Post(ctx, path, payload)
	return err
}

// sortedCustomTags returns a copy of tags sorted by name, with each tag's
// own PermittedValues sorted alphabetically, so equivalent configurations
// always serialize identically regardless of input order. SAP's
// documentation does not state that tag or permitted-value order carries
// any meaning, and this project found nothing suggesting otherwise.
func sortedCustomTags(tags []CustomTag) []CustomTag {
	sorted := make([]CustomTag, len(tags))
	copy(sorted, tags)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	for i := range sorted {
		if len(sorted[i].PermittedValues) == 0 {
			continue
		}
		values := make([]string, len(sorted[i].PermittedValues))
		copy(values, sorted[i].PermittedValues)
		sort.Strings(values)
		sorted[i].PermittedValues = values
	}

	return sorted
}
