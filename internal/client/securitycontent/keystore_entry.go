package securitycontent

import (
	"context"
	"sort"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const keystoreEntriesEntitySet = "KeystoreEntries"

// KeystoreEntry is the read-only wire representation of a KeystoreEntries
// entity — a single entry (certificate, key pair, or SSH-capable key) in
// the tenant's keystore. Fields are confirmed verbatim from SAP's own
// documented example response (identical example body reused across "Get
// All Keystore Entries" and "Get Keystore Entry by Alias"):
//
//	<d:Hexalias>736D74702E6D61696C2E7961686F6F2E636F6D</d:Hexalias>
//	<d:Alias>smtp.mail.yahoo.com</d:Alias>
//	<d:KeyType>RSA</d:KeyType>
//	<d:KeySize>2048</d:KeySize>
//	<d:ValidNotBefore>2019-10-04T00:00:00Z</d:ValidNotBefore>
//	<d:ValidNotAfter>2020-04-01T12:00:00Z</d:ValidNotAfter>
//	....
//
// The trailing "...." in SAP's own example is a genuine, deliberate
// truncation, not an artifact of this project's research: SAP's prose
// elsewhere on the same overview page states a keystore entry also carries
// "Subject DN and Issuer DN, and administrative information such as the
// time when the entry was last modified" — but without a worked example or
// $metadata to confirm their exact OData property names and casing, this
// client does not guess at them as JSON fields here. Where subject/issuer/
// serial-number/fingerprint metadata is needed (see
// docs/guides/security-content.md), it is derived by parsing the
// certificate bytes this client can independently, confirmedly retrieve
// (GetCertificate/GetSSHPublicKey) with Go's own crypto/x509, not by
// guessing at an unconfirmed SAP field name.
type KeystoreEntry struct {
	Hexalias       string `json:"Hexalias"`
	Alias          string `json:"Alias"`
	KeyType        string `json:"KeyType"`
	KeySize        int    `json:"KeySize"`
	ValidNotBefore string `json:"ValidNotBefore,omitempty"`
	ValidNotAfter  string `json:"ValidNotAfter,omitempty"`
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

// ListKeystoreEntries reads every entry in the tenant keystore — both
// tenant-administrator-owned and SAP-owned entries; SAP's documentation
// does not expose a field distinguishing the two (see
// docs/guides/security-content.md for how this provider handles that gap).
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
