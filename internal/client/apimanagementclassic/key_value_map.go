package apimanagementclassic

import (
	"context"
	"encoding/json"
	"fmt"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const genericKeyMapEntriesEntitySet = "GenericKeyMapEntries"

// KeyValueMapEntry is one key/value pair inside a KeyValueMap, confirmed
// via SAP's own worked GenericKeyMapEntries Create example (nested under
// genericKeyMapEntryValues).
type KeyValueMapEntry struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// KeyValueMap is the wire representation of a GenericKeyMapEntries entity.
// Identity is the confirmed (Name, Scope, ScopeID) combination. SAP's own
// user guide documents a full Create/Read/Update(entries)/Delete UI
// lifecycle for Key Value Maps, but only Create's REST payload is shown
// verbatim; this provider's Terraform resource is scoped accordingly (see
// docs/sap-api-references.md).
//
// Encrypted maps (SAP's "Encrypted" checkbox, wire field IsEncrypted) are
// not supported by this client's Terraform resource: SAP's documentation
// does not confirm whether GET returns an encrypted entry's plaintext value
// back, a masked placeholder, or nothing at all, and getting this wrong
// would either leak a secret into Terraform state or produce permanent
// diffs. Only unencrypted maps are supported until that is confirmed.
type KeyValueMap struct {
	Name        string `json:"name"`
	Scope       string `json:"scope"`
	ScopeID     string `json:"scopeId"`
	IsEncrypted bool   `json:"isEncrypted"`

	Entries []KeyValueMapEntry `json:"-"`
}

type keyValueMapEntryValueWire struct {
	Name    string `json:"name"`
	MapName string `json:"mapName"`
	Value   string `json:"value"`
	ScopeID string `json:"scopeId"`
	Scope   string `json:"scope"`
}

type keyValueMapWire struct {
	Name                   string                      `json:"name"`
	ScopeID                string                      `json:"scopeId"`
	Scope                  string                      `json:"scope"`
	IsEncrypted            bool                        `json:"isEncrypted"`
	GenericKeyMapEntryVals []keyValueMapEntryValueWire `json:"genericKeyMapEntryValues"`
}

func (m KeyValueMap) toWire() keyValueMapWire {
	wire := keyValueMapWire{
		Name:        m.Name,
		ScopeID:     m.ScopeID,
		Scope:       m.Scope,
		IsEncrypted: m.IsEncrypted,
	}
	for _, entry := range m.Entries {
		wire.GenericKeyMapEntryVals = append(wire.GenericKeyMapEntryVals, keyValueMapEntryValueWire{
			Name:    entry.Name,
			MapName: m.Name,
			Value:   entry.Value,
			ScopeID: m.ScopeID,
			Scope:   m.Scope,
		})
	}
	return wire
}

// keyValueMapReadWire is the map as SAP returns it. genericKeyMapEntryValues
// comes back as a {"__deferred": {"uri": ...}} link, not a list (tenant
// test, September 2026), so it is left out here and the entries are read
// through the navigation property instead.
type keyValueMapReadWire struct {
	Name        string `json:"name"`
	ScopeID     string `json:"scopeId"`
	Scope       string `json:"scope"`
	IsEncrypted bool   `json:"isEncrypted"`
}

func keyValueMapFromWire(wire keyValueMapReadWire) KeyValueMap {
	return KeyValueMap{
		Name:        wire.Name,
		Scope:       wire.Scope,
		ScopeID:     wire.ScopeID,
		IsEncrypted: wire.IsEncrypted,
	}
}

// CreateKeyValueMap creates a new key value map together with its initial
// set of entries in a single call, matching SAP's own confirmed worked
// example. SAP answers 201 with the map but only a link to its entries, so
// the returned map carries the entries that were sent.
func (c *Client) CreateKeyValueMap(ctx context.Context, kvm KeyValueMap) (*KeyValueMap, error) {
	payload, err := json.Marshal(kvm.toWire())
	if err != nil {
		return nil, fmt.Errorf("apimanagementclassic: encoding key value map: %w", err)
	}

	body, err := c.odata.Post(ctx, genericKeyMapEntriesEntitySet, payload)
	if err != nil {
		return nil, err
	}

	var wire keyValueMapReadWire
	if err := v2.DecodeEntity(body, &wire); err != nil {
		return nil, err
	}
	result := keyValueMapFromWire(wire)
	result.Entries = kvm.Entries
	return &result, nil
}

// GetKeyValueMap reads a single key value map, identified by its composite
// (name, scope, scopeId) key, and then its entries through the
// genericKeyMapEntryValues navigation property.
func (c *Client) GetKeyValueMap(ctx context.Context, name, scope, scopeID string) (*KeyValueMap, error) {
	predicate, err := v2.CompositeKeyPredicate("name", name, "scope", scope, "scopeId", scopeID)
	if err != nil {
		return nil, err
	}
	path := v2.BuildPath(genericKeyMapEntriesEntitySet, predicate, "")

	body, err := c.odata.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var wire keyValueMapReadWire
	if err := v2.DecodeEntity(body, &wire); err != nil {
		return nil, err
	}
	result := keyValueMapFromWire(wire)

	body, err = c.odata.Get(ctx, path+"/genericKeyMapEntryValues")
	if err != nil {
		return nil, err
	}
	var entries []keyValueMapEntryValueWire
	if err := v2.DecodeCollection(body, &entries); err != nil {
		return nil, err
	}
	for _, entry := range entries {
		result.Entries = append(result.Entries, KeyValueMapEntry{Name: entry.Name, Value: entry.Value})
	}
	return &result, nil
}

// DeleteKeyValueMap deletes a key value map by its composite key.
func (c *Client) DeleteKeyValueMap(ctx context.Context, name, scope, scopeID string) error {
	predicate, err := v2.CompositeKeyPredicate("name", name, "scope", scope, "scopeId", scopeID)
	if err != nil {
		return err
	}
	path := v2.BuildPath(genericKeyMapEntriesEntitySet, predicate, "")
	return c.odata.Delete(ctx, path)
}
