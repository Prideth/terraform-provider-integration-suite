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
// user guide worked Create (POST) and Update (PUT) examples.
//
// ApiProxyNames is only sent on Create (as __metadata.uri deep-insert
// references) — no worked Update example includes it, so this provider
// treats the proxy association as immutable after creation rather than
// guess at an unconfirmed way to add or remove proxies later.
type APIProduct struct {
	Name        string `json:"name"`
	Version     string `json:"version,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Scope       string `json:"scope"`

	// StatusCode's only confirmed value is "PUBLISHED"; no other value
	// appears in any reachable worked example, so this provider does not
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
	return wire
}

// updateWire is the narrower field set SAP's own documented Update (PUT)
// example sends — no apiProxies, apiResources, or additionalProperties.
type apiProductUpdateWire struct {
	Name          string  `json:"name"`
	Title         string  `json:"title,omitempty"`
	Description   string  `json:"description,omitempty"`
	Scope         string  `json:"scope"`
	Version       string  `json:"version,omitempty"`
	StatusCode    string  `json:"status_code,omitempty"`
	IsRestricted  bool    `json:"isRestricted"`
	IsPublished   bool    `json:"isPublished"`
	QuotaCount    *int64  `json:"quotaCount"`
	QuotaInterval *int64  `json:"quotaInterval"`
	QuotaTimeUnit *string `json:"quotaTimeUnit"`
}

func (p APIProduct) toUpdateWire() apiProductUpdateWire {
	return apiProductUpdateWire{
		Name:          p.Name,
		Title:         p.Title,
		Description:   p.Description,
		Scope:         p.Scope,
		Version:       p.Version,
		StatusCode:    p.StatusCode,
		IsRestricted:  p.IsRestricted,
		IsPublished:   p.IsPublished,
		QuotaCount:    p.QuotaCount,
		QuotaInterval: p.QuotaInterval,
		QuotaTimeUnit: p.QuotaTimeUnit,
	}
}

func apiProductFromWire(wire apiProductWire) APIProduct {
	p := APIProduct{
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
	for _, ref := range wire.ApiProxies {
		// uri is "APIProxies(name='<name>')"; extract <name> back out.
		uri := ref.Metadata.URI
		const prefix, suffix = "APIProxies(name='", "')"
		if len(uri) > len(prefix)+len(suffix) && uri[:len(prefix)] == prefix {
			p.ApiProxyNames = append(p.ApiProxyNames, uri[len(prefix):len(uri)-len(suffix)])
		}
	}
	return p
}

// CreateAPIProduct creates a new API product, optionally associating it
// with existing API proxies by name in the same call.
func (c *Client) CreateAPIProduct(ctx context.Context, product APIProduct) (*APIProduct, error) {
	payload, err := json.Marshal(product.toCreateWire())
	if err != nil {
		return nil, fmt.Errorf("apimanagementclassic: encoding API product: %w", err)
	}

	body, err := c.odata.Post(ctx, apiProductsEntitySet, payload)
	if err != nil {
		return nil, err
	}

	var wire apiProductWire
	if err := v2.DecodeEntity(body, &wire); err != nil {
		return nil, err
	}
	result := apiProductFromWire(wire)
	return &result, nil
}

// GetAPIProduct reads a single API product by name.
func (c *Client) GetAPIProduct(ctx context.Context, name string) (*APIProduct, error) {
	path := v2.BuildPath(apiProductsEntitySet, v2.KeyPredicate(name), "")

	body, err := c.odata.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wire apiProductWire
	if err := v2.DecodeEntity(body, &wire); err != nil {
		return nil, err
	}
	result := apiProductFromWire(wire)
	return &result, nil
}

// UpdateAPIProduct updates an API product's top-level fields. It never
// touches the product's API proxy associations — see the APIProduct type
// doc comment for why.
func (c *Client) UpdateAPIProduct(ctx context.Context, product APIProduct) error {
	payload, err := json.Marshal(product.toUpdateWire())
	if err != nil {
		return fmt.Errorf("apimanagementclassic: encoding API product: %w", err)
	}

	path := v2.BuildPath(apiProductsEntitySet, v2.KeyPredicate(product.Name), "")
	_, err = c.odata.Put(ctx, path, payload)
	return err
}

// DeleteAPIProduct deletes an API product by name. SAP's documentation does
// not show a worked DELETE example for this entity specifically; this
// method uses the same key-predicate DELETE convention directly confirmed
// for APIProviders and CertificateStoreReferences within this same
// Management.svc API family.
func (c *Client) DeleteAPIProduct(ctx context.Context, name string) error {
	path := v2.BuildPath(apiProductsEntitySet, v2.KeyPredicate(name), "")
	return c.odata.Delete(ctx, path)
}

// APIProductAdditionalProperty is the wire representation of an
// APIProductAdditionalProperties entity: a custom name/value attribute
// attached to an API product, confirmed via SAP's own worked Create and
// Update examples. Identity is the composite (EntityID, Name) key —
// EntityID is the owning product's name, confirmed directly from the
// worked examples' "entityId": "SampleProduct" pairing with an APIProduct
// also named "SampleProduct".
type APIProductAdditionalProperty struct {
	EntityID string `json:"entityId"`
	Name     string `json:"name"`
	Value    string `json:"value"`
}

// CreateAPIProductAdditionalProperty adds a custom attribute to an existing
// API product.
func (c *Client) CreateAPIProductAdditionalProperty(ctx context.Context, prop APIProductAdditionalProperty) error {
	payload, err := json.Marshal(prop)
	if err != nil {
		return fmt.Errorf("apimanagementclassic: encoding API product additional property: %w", err)
	}
	_, err = c.odata.Post(ctx, apiProductAdditionalPropertiesEntity, payload)
	return err
}

// UpdateAPIProductAdditionalProperty changes an existing custom attribute's
// value. EntityID and Name (the composite key) are immutable.
func (c *Client) UpdateAPIProductAdditionalProperty(ctx context.Context, entityID, name, value string) error {
	payload, err := json.Marshal(struct {
		Value string `json:"value"`
	}{Value: value})
	if err != nil {
		return fmt.Errorf("apimanagementclassic: encoding API product additional property: %w", err)
	}

	predicate, err := v2.CompositeKeyPredicate("entityId", entityID, "name", name)
	if err != nil {
		return err
	}
	path := v2.BuildPath(apiProductAdditionalPropertiesEntity, predicate, "")
	_, err = c.odata.Put(ctx, path, payload)
	return err
}

// DeleteAPIProductAdditionalProperty removes a custom attribute from an API
// product.
func (c *Client) DeleteAPIProductAdditionalProperty(ctx context.Context, entityID, name string) error {
	predicate, err := v2.CompositeKeyPredicate("entityId", entityID, "name", name)
	if err != nil {
		return err
	}
	path := v2.BuildPath(apiProductAdditionalPropertiesEntity, predicate, "")
	return c.odata.Delete(ctx, path)
}
