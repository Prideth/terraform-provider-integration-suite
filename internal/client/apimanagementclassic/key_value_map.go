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

func keyValueMapFromWire(wire keyValueMapWire) KeyValueMap {
	m := KeyValueMap{
		Name:        wire.Name,
		Scope:       wire.Scope,
		ScopeID:     wire.ScopeID,
		IsEncrypted: wire.IsEncrypted,
	}
	for _, entry := range wire.GenericKeyMapEntryVals {
		m.Entries = append(m.Entries, KeyValueMapEntry{Name: entry.Name, Value: entry.Value})
	}
	return m
}

// CreateKeyValueMap creates a new key value map together with its initial
// set of entries in a single call, matching SAP's own confirmed worked
// example.
func (c *Client) CreateKeyValueMap(ctx context.Context, kvm KeyValueMap) (*KeyValueMap, error) {
	payload, err := json.Marshal(kvm.toWire())
	if err != nil {
		return nil, fmt.Errorf("apimanagementclassic: encoding key value map: %w", err)
	}

	body, err := c.odata.Post(ctx, genericKeyMapEntriesEntitySet, payload)
	if err != nil {
		return nil, err
	}

	var wire keyValueMapWire
	if err := v2.DecodeEntity(body, &wire); err != nil {
		return nil, err
	}
	result := keyValueMapFromWire(wire)
	return &result, nil
}

// GetKeyValueMap reads a single key value map, identified by its composite
// (name, scope, scopeId) key, by the same convention this API family uses
// consistently elsewhere (single-quoted key predicates on Create's identity
// fields).
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

	var wire keyValueMapWire
	if err := v2.DecodeEntity(body, &wire); err != nil {
		return nil, err
	}
	result := keyValueMapFromWire(wire)
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
