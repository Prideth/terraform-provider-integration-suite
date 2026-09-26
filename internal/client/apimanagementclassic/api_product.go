package apimanagementclassic

import (
	"context"
	"encoding/json"
	"fmt"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const (
	apiProductsEntitySet                 = "APIProducts"
	apiProductAdditionalPropertiesEntity = "APIProductAdditionalProperties"
)

// apiProxyRef is a deep-insert reference to an already-existing API proxy,
// confirmed via SAP's own worked example: {"__metadata": {"uri":
// "APIProxies(name='SampleAPI')"}}.
type apiProxyRef struct {
	Metadata struct {
		URI string `json:"uri"`
	} `json:"__metadata"`
}

// APIProduct is the wire representation of an APIProducts entity, confirmed
// field-for-field against SAP's own "SAP API Management Standalone Service"
// user guide worked Create (POST) example and a tenant test in September
// 2026.
//
// A product cannot be changed after it is created: the same tenant test
// answered PUT, PATCH and MERGE on an existing product with 405 "UPDATE
// operation not supported on APIProduct entity". This client therefore
// offers create, read and delete only.
//
// ApiProxyNames is only sent on Create (as __metadata.uri deep-insert
// references). SAP rejects a product without one ("At least one API Proxy
// should be linked to an API Product"). A read returns the association as a
// __deferred link, not the names, so GetAPIProduct leaves it empty.
type APIProduct struct {
	Name        string `json:"name"`
	Version     string `json:"version,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Scope       string `json:"scope"`

	// StatusCode is required on create: without it SAP failed with a
	// NullPointerException on getStatus_code(). "PUBLISHED" is confirmed on
	// a tenant; no other value is confirmed, so this client does not
	// validate it against a closed enum.
	StatusCode string `json:"status_code,omitempty"`

	IsPublished  bool `json:"isPublished"`
	IsRestricted bool `json:"isRestricted"`

	// QuotaCount/QuotaInterval/QuotaTimeUnit are confirmed nullable
	// (SAP's own example sends -99/-99/null for "no quota" and
	// null/null/null in another). This provider mirrors that: a zero value
	// here is sent as JSON null, not 0, via MarshalJSON below.
	QuotaCount    *int64  `json:"-"`
	QuotaInterval *int64  `json:"-"`
	QuotaTimeUnit *string `json:"-"`

	// ApiProxyNames is only meaningful on Create; see the type doc comment.
	ApiProxyNames []string `json:"-"`

	// AdditionalProperties is only sent on Create, inside the product
	// (deep insert); see APIProductAdditionalProperty.
	AdditionalProperties []APIProductAdditionalProperty `json:"-"`
}

type apiProductWire struct {
	Name          string        `json:"name"`
	Version       string        `json:"version,omitempty"`
	Title         string        `json:"title,omitempty"`
	Description   string        `json:"description,omitempty"`
	Scope         string        `json:"scope"`
	StatusCode    string        `json:"status_code,omitempty"`
	IsPublished   bool          `json:"isPublished"`
	IsRestricted  bool          `json:"isRestricted"`
	QuotaCount    *int64        `json:"quotaCount"`
	QuotaInterval *int64        `json:"quotaInterval"`
	QuotaTimeUnit *string       `json:"quotaTimeUnit"`
	ApiProxies    []apiProxyRef `json:"apiProxies,omitempty"`
	ApiResources  []struct{}    `json:"apiResources"`

	AdditionalProperties []APIProductAdditionalProperty `json:"additionalProperties,omitempty"`
}

func (p APIProduct) toCreateWire() apiProductWire {
	wire := apiProductWire{
		Name:          p.Name,
		Version:       p.Version,
		Title:         p.Title,
		Description:   p.Description,
		Scope:         p.Scope,
		StatusCode:    p.StatusCode,
		IsPublished:   p.IsPublished,
		IsRestricted:  p.IsRestricted,
		QuotaCount:    p.QuotaCount,
		QuotaInterval: p.QuotaInterval,
		QuotaTimeUnit: p.QuotaTimeUnit,
		ApiResources:  []struct{}{},
	}
	for _, name := range p.ApiProxyNames {
		ref := apiProxyRef{}
		ref.Metadata.URI = fmt.Sprintf("APIProxies(name='%s')", v2.EscapeLiteral(name))
		wire.ApiProxies = append(wire.ApiProxies, ref)
	}
	// SAP rejects a deep-inserted property without the owning product's
	// name: 400 ADDITIONAL_PROPERTY_ENTITY_ID_MATCH_ERROR, "Entity Id in
	// Additional Property does not match".
	for _, prop := range p.AdditionalProperties {
		prop.EntityID = p.Name
		wire.AdditionalProperties = append(wire.AdditionalProperties, prop)
	}
	return wire
}

// apiProductReadWire is the entity as SAP returns it. Navigation properties
// (apiProxies, apiResources, additionalProperties and others) come back as
// {"__deferred": {"uri": ...}} objects, not arrays, so they are left out
// here; decoding them into the create shape's slices would fail.
type apiProductReadWire struct {
	Name          string  `json:"name"`
	Version       string  `json:"version"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Scope         string  `json:"scope"`
	StatusCode    string  `json:"status_code"`
	IsPublished   bool    `json:"isPublished"`
	IsRestricted  bool    `json:"isRestricted"`
	QuotaCount    *int64  `json:"quotaCount"`
	QuotaInterval *int64  `json:"quotaInterval"`
	QuotaTimeUnit *string `json:"quotaTimeUnit"`
}

func apiProductFromWire(wire apiProductReadWire) APIProduct {
	return APIProduct{
		Name:          wire.Name,
		Version:       wire.Version,
		Title:         wire.Title,
		Description:   wire.Description,
		Scope:         wire.Scope,
		StatusCode:    wire.StatusCode,
		IsPublished:   wire.IsPublished,
		IsRestricted:  wire.IsRestricted,
		QuotaCount:    wire.QuotaCount,
		QuotaInterval: wire.QuotaInterval,
		QuotaTimeUnit: wire.QuotaTimeUnit,
	}
}

// CreateAPIProduct creates a new API product linked to existing API proxies
// by name. SAP answers 201 with the product, whose proxy association is only
// a __deferred link, so the returned product carries the names that were
// sent.
func (c *Client) CreateAPIProduct(ctx context.Context, product APIProduct) (*APIProduct, error) {
	payload, err := json.Marshal(product.toCreateWire())
	if err != nil {
		return nil, fmt.Errorf("apimanagementclassic: encoding API product: %w", err)
	}

	body, err := c.odata.Post(ctx, apiProductsEntitySet, payload)
	if err != nil {
		return nil, err
	}

	var wire apiProductReadWire
	if err := v2.DecodeEntity(body, &wire); err != nil {
		return nil, err
	}
	result := apiProductFromWire(wire)
	result.ApiProxyNames = product.ApiProxyNames
	result.AdditionalProperties = product.AdditionalProperties
	return &result, nil
}

// GetAPIProduct reads a single API product by name. ApiProxyNames stays
// empty; see the APIProduct type doc comment.
func (c *Client) GetAPIProduct(ctx context.Context, name string) (*APIProduct, error) {
	path := v2.BuildPath(apiProductsEntitySet, v2.KeyPredicate(name), "")

	body, err := c.odata.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wire apiProductReadWire
	if err := v2.DecodeEntity(body, &wire); err != nil {
		return nil, err
	}
	result := apiProductFromWire(wire)
	return &result, nil
}

// GetAPIProductProxyNames lists the names of the API proxies linked to a
// product, through the apiProxies navigation property. A tenant test in
// September 2026 answered with the proxies as a plain results list.
func (c *Client) GetAPIProductProxyNames(ctx context.Context, name string) ([]string, error) {
	path := v2.BuildPath(apiProductsEntitySet, v2.KeyPredicate(name), "") + "/apiProxies"

	body, err := c.odata.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var proxies []struct {
		Name string `json:"name"`
	}
	if err := v2.DecodeCollection(body, &proxies); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(proxies))
	for _, p := range proxies {
		names = append(names, p.Name)
	}
	return names, nil
}

// GetAPIProductAdditionalProperties lists a product's custom attributes,
// through the additionalProperties navigation property.
func (c *Client) GetAPIProductAdditionalProperties(ctx context.Context, name string) ([]APIProductAdditionalProperty, error) {
	path := v2.BuildPath(apiProductsEntitySet, v2.KeyPredicate(name), "") + "/additionalProperties"

	body, err := c.odata.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var props []APIProductAdditionalProperty
	if err := v2.DecodeCollection(body, &props); err != nil {
		return nil, err
	}
	return props, nil
}

// DeleteAPIProduct deletes an API product by name. A tenant test in
// September 2026 answered 204.
func (c *Client) DeleteAPIProduct(ctx context.Context, name string) error {
	path := v2.BuildPath(apiProductsEntitySet, v2.KeyPredicate(name), "")
	return c.odata.Delete(ctx, path)
}

// APIProductAdditionalProperty is the wire representation of an
// APIProductAdditionalProperties entity: a custom name/value attribute
// attached to an API product. Identity is the composite (EntityID, Name)
// key, where EntityID is the owning product's name.
//
// SAP does not accept a separate create: a tenant test in September 2026
// answered POST APIProductAdditionalProperties with 405 "CREATE operation
// not supported on APIProductAdditionalProperty entity". The product's
// properties are read through APIProducts('<name>')/additionalProperties.
type APIProductAdditionalProperty struct {
	EntityID string `json:"entityId"`
	Name     string `json:"name"`
	Value    string `json:"value"`
}
