package apimanagementclassic

import (
	"context"
	"encoding/json"
	"fmt"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const certificateStoreReferencesEntitySet = "CertificateStoreReferences"

// CertificateStoreReference is the wire representation of a
// CertificateStoreReferences entity: a named pointer to an already-existing
// keystore or truststore, used so a virtual host's TLS configuration can be
// repointed at a new store (for certificate rotation) without editing the
// virtual host itself. It does not manage the certificate or keystore
// content — that remains a UI-only upload, confirmed to have no
// accompanying REST API — only the reference/pointer object, which SAP
// documents with a full Create/Read/Update/Delete lifecycle via service
// calls. This is a distinct remote object from Cloud Integration's own
// sapintegrationsuite_certificate/sapintegrationsuite_key_pair, which
// manage actual keystore entries under a completely different API and
// service (see docs/sap-api-references.md); it is not reused here.
type CertificateStoreReference struct {
	Name                 string `json:"name,omitempty"`
	CertificateStoreName string `json:"certificateStoreName"`

	// StoreType is returned by SAP (confirmed value: "TRUSTSTORE"), never
	// set by the caller.
	StoreType string `json:"storeType,omitempty"`
}

// CreateCertificateStoreReference creates a new certificate store
// reference.
func (c *Client) CreateCertificateStoreReference(ctx context.Context, ref CertificateStoreReference) (*CertificateStoreReference, error) {
	payload, err := json.Marshal(struct {
		Name                 string `json:"name"`
		CertificateStoreName string `json:"certificateStoreName"`
	}{Name: ref.Name, CertificateStoreName: ref.CertificateStoreName})
	if err != nil {
		return nil, fmt.Errorf("apimanagementclassic: encoding certificate store reference: %w", err)
	}

	body, err := c.odata.Post(ctx, certificateStoreReferencesEntitySet, payload)
	if err != nil {
		return nil, err
	}

	var created CertificateStoreReference
	if err := v2.DecodeEntity(body, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// GetCertificateStoreReference reads a single certificate store reference
// by name.
func (c *Client) GetCertificateStoreReference(ctx context.Context, name string) (*CertificateStoreReference, error) {
	path := v2.BuildPath(certificateStoreReferencesEntitySet, v2.KeyPredicate(name), "")

	body, err := c.odata.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var ref CertificateStoreReference
	if err := v2.DecodeEntity(body, &ref); err != nil {
		return nil, err
	}
	return &ref, nil
}

// UpdateCertificateStoreReference repoints an existing reference at a
// different certificate store. name (the key) cannot be changed this way —
// confirmed by SAP's documented PUT payload, which carries only
// certificateStoreName.
func (c *Client) UpdateCertificateStoreReference(ctx context.Context, name, certificateStoreName string) error {
	payload, err := json.Marshal(struct {
		CertificateStoreName string `json:"certificateStoreName"`
	}{CertificateStoreName: certificateStoreName})
	if err != nil {
		return fmt.Errorf("apimanagementclassic: encoding certificate store reference: %w", err)
	}

	path := v2.BuildPath(certificateStoreReferencesEntitySet, v2.KeyPredicate(name), "")
	_, err = c.odata.Put(ctx, path, payload)
	return err
}

// DeleteCertificateStoreReference deletes a certificate store reference by
// name.
func (c *Client) DeleteCertificateStoreReference(ctx context.Context, name string) error {
	path := v2.BuildPath(certificateStoreReferencesEntitySet, v2.KeyPredicate(name), "")
	return c.odata.Delete(ctx, path)
}
