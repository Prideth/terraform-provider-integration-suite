package securitycontent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const keystoreResourcesEntitySet = "KeystoreResources"

// systemKeystoreName is the fixed, confirmed key for the tenant's one
// keystore: SAP's own "Import/Back up a Keystore" documentation states
// "Currently, only `system` is allowed (SAP supports only one keystore)".
const systemKeystoreName = "system"

// deleteKeystoreEntriesRequest is the request body shape SAP documents for
// mass deletion: {"Aliases":"alias1;alias2;alias3"}.
type deleteKeystoreEntriesRequest struct {
	Aliases string `json:"Aliases"`
}

// EscapeKeystoreAlias escapes a single alias for inclusion in the
// semicolon-delimited Aliases string DeleteKeystoreEntries sends. Confirmed
// verbatim from SAP's own "Trigger Mass Deletion of Keystore Entries"
// documentation, which gives two worked examples:
//
//	"Company A; Company B" -> "Company A\; Company B"
//	Alias1="Company\A\", Alias2="CompanyB" -> "Company\\A\\;CompanyB"
//
// Backslashes are escaped first, then semicolons — escaping in the other
// order would double-escape the backslash this function itself inserts to
// escape a semicolon, producing the wrong result. This ordering is this
// project's own inference from the two examples above (SAP's documentation
// states both rules but does not spell out which is applied first); it is
// the only order that reproduces both documented examples exactly, which
// is asserted byte-for-byte in this package's tests.
func EscapeKeystoreAlias(alias string) string {
	escaped := strings.ReplaceAll(alias, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, ";", `\;`)
	return escaped
}

// joinKeystoreAliases builds the Aliases string DeleteKeystoreEntries
// sends: every alias individually escaped via EscapeKeystoreAlias, then
// joined with a plain (unescaped) ";" separator — the separator itself
// never needs escaping, since every semicolon that was part of an alias's
// own content was already escaped to "\;" by EscapeKeystoreAlias.
func joinKeystoreAliases(aliases []string) string {
	escaped := make([]string, len(aliases))
	for i, a := range aliases {
		escaped[i] = EscapeKeystoreAlias(a)
	}
	return strings.Join(escaped, ";")
}

// DeleteKeystoreEntries deletes the given plain aliases from the tenant
// keystore in a single call. Confirmed directly from SAP's own "Trigger
// Mass Deletion of Keystore Entries" documentation: PUT
// /api/v1/KeystoreResources('system')?deleteEntries=true, body
// {"Aliases":"alias1;alias2;alias3"}. Unlike KeystoreEntries' own key
// predicate, this endpoint's Aliases string uses the plain alias text, not
// its hex-encoded form — this client does not hex-encode aliases here.
//
// This is the only confirmed public delete operation for keystore entries
// (certificates and key pairs); there is no documented
// "DELETE /KeystoreEntries(...)" or similar per-entry REST verb, so
// sapintegrationsuite_certificate and sapintegrationsuite_key_pair both
// call this with exactly the one alias they own, never a caller-supplied
// list, so a destroy of one Terraform resource can never be constructed in
// a way that also submits an unrelated alias — see
// docs/guides/security-content.md.
//
// SAP-owned keystore entries are documented as protected against
// tenant-administrator deletion; this client does not attempt to detect
// that case in advance (no confirmed API field distinguishes ownership —
// see KeystoreEntry), and instead relies on and surfaces whatever error
// SAP's own server-side protection returns.
func (c *Client) DeleteKeystoreEntries(ctx context.Context, aliases []string) error {
	if len(aliases) == 0 {
		return fmt.Errorf("securitycontent: DeleteKeystoreEntries requires at least one alias")
	}

	payload, err := json.Marshal(deleteKeystoreEntriesRequest{Aliases: joinKeystoreAliases(aliases)})
	if err != nil {
		return fmt.Errorf("securitycontent: encoding keystore entry deletion request: %w", err)
	}

	path := v2.BuildPath(keystoreResourcesEntitySet, v2.KeyPredicate(systemKeystoreName), "deleteEntries=true")
	_, err = c.odata.Put(ctx, path, payload)
	return err
}
