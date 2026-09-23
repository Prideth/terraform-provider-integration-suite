package securitycontent

import (
	"context"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

func sshKeyValuePath(hexAlias string) string {
	return v2.BuildPath(keystoreEntriesEntitySet, v2.KeyPredicate(hexAlias), "") + "/Sshkey/$value"
}

// GetSSHPublicKey exports a keystore entry's public key in OpenSSH format
// ("ssh-rsa AAAA... <comment>"). Confirmed directly from SAP's own "Export
// Public Key in OpenSSH Format" documentation: GET
// /api/v1/KeystoreEntries('{Hexalias}')/Sshkey/$value. SAP documents this
// as supported only for RSA- and DSA-keyed entries ("Other algorithms, for
// example elliptic curve (EC), aren't supported") — this client does not
// pre-filter by key type itself and simply surfaces whatever error SAP
// returns for an unsupported entry, rather than guessing at which key
// types are safe to call this for.
//
// There is no independent "SSH Key" entity in SAP's Security Content API:
// the tenant keystore's own "Creating a Key Pair/SSH Key Pair" UI
// documentation uses the identical Key Pair attribute set (alias, key
// type, key size, signature algorithm, subject DN fields, validity) for
// both, and this OpenSSH export is simply an additional read operation on
// the same KeystoreEntries resource a Key Pair (or an imported Certificate
// entry) already is — see docs/guides/security-content.md for why this
// provider does not implement a separate sapintegrationsuite_ssh_key
// resource.
func (c *Client) GetSSHPublicKey(ctx context.Context, alias string) ([]byte, error) {
	hexAlias := v2.EncodeUTF8Hex(alias)
	return c.odata.Get(ctx, sshKeyValuePath(hexAlias))
}
