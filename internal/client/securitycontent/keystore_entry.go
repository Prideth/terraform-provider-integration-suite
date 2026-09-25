package securitycontent

import (
	"context"
	"sort"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const keystoreEntriesEntitySet = "KeystoreEntries"

// KeystoreEntry is the read-only wire representation of a KeystoreEntries
// entity: one certificate, key pair or SSH key in the tenant keystore. The
// fields follow the tenant $metadata, where KeystoreEntry inherits from
// KeystoreEntryCertificatePart, KeystoreEntryCertificatePartBase and
// KeystoreEntryAlias. Date properties are Edm.DateTimeOffset and arrive as
// "/Date(<millis>+0000)/" literals; they are kept as strings here and
// converted by the caller.
type KeystoreEntry struct {
	Hexalias           string `json:"Hexalias"`
	Alias              string `json:"Alias"`
	KeyType            string `json:"KeyType"`
	KeySize            int    `json:"KeySize"`
	ValidNotBefore     string `json:"ValidNotBefore,omitempty"`
	ValidNotAfter      string `json:"ValidNotAfter,omitempty"`
	SerialNumber       string `json:"SerialNumber,omitempty"`
	SignatureAlgorithm string `json:"SignatureAlgorithm,omitempty"`
	EllipticCurve      string `json:"EllipticCurve,omitempty"`
	Validity           string `json:"Validity,omitempty"`
	SubjectDN          string `json:"SubjectDN,omitempty"`
	IssuerDN           string `json:"IssuerDN,omitempty"`
	Version            int    `json:"Version,omitempty"`
	FingerprintSha1    string `json:"FingerprintSha1,omitempty"`
	FingerprintSha256  string `json:"FingerprintSha256,omitempty"`
	FingerprintSha512  string `json:"FingerprintSha512,omitempty"`
	Type               string `json:"Type,omitempty"`
	Owner              string `json:"Owner,omitempty"`
	Status             string `json:"Status,omitempty"`
	CreatedBy          string `json:"CreatedBy,omitempty"`
	CreatedTime        string `json:"CreatedTime,omitempty"`
	LastModifiedBy     string `json:"LastModifiedBy,omitempty"`
	LastModifiedTime   string `json:"LastModifiedTime,omitempty"`
}

func keystoreEntryPath(hexAlias string) string {
	return v2.BuildPath(keystoreEntriesEntitySet, v2.KeyPredicate(hexAlias), "")
}

// GetKeystoreEntry reads a single keystore entry by its alias. alias is the
// plain (human-readable) alias; this client computes the hex-encoded OData
// key internally via v2.EncodeUTF8Hex, confirmed as SAP's exact encoding
// rule for this field (see internal/client/odata/v2/hexkey.go). Confirmed
// directly from SAP's own "Get Keystore Entry by Alias" documentation:
// GET /api/v1/KeystoreEntries('{Hexalias}').
func (c *Client) GetKeystoreEntry(ctx context.Context, alias string) (*KeystoreEntry, error) {
	hexAlias := v2.EncodeUTF8Hex(alias)
	body, err := c.odata.Get(ctx, keystoreEntryPath(hexAlias))
	if err != nil {
		return nil, err
	}

	var entry KeystoreEntry
	if err := v2.DecodeEntity(body, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// ListKeystoreEntries reads every entry in the tenant keystore, both
// tenant-administrator-owned and SAP-owned ones (see the Owner property).
// Confirmed directly from SAP's own "Get All Keystore Entries"
// documentation: GET /api/v1/KeystoreEntries. Follows server-driven paging
// (GetAllPages), matching every other collection this provider reads — a
// tenant keystore can hold thousands of entries (SAP documents up to
// around 6000 X.509 certificates within the 6 MB Cloud Foundry limit).
// Results are sorted deterministically by Alias before returning: SAP does
// not document a guaranteed response order, and sorting here means a
// Terraform plan never shows spurious churn purely because SAP happened to
// return the same entries in a different order.
func (c *Client) ListKeystoreEntries(ctx context.Context) ([]KeystoreEntry, error) {
	entries, err := v2.GetAllPages[KeystoreEntry](ctx, c.odata, keystoreEntriesEntitySet)
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Alias < entries[j].Alias })
	return entries, nil
}
