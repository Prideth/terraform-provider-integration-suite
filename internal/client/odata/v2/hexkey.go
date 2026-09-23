package v2

import (
	"encoding/hex"
	"fmt"
)

// EncodeUTF8Hex converts s to the lowercase hexadecimal encoding of its
// UTF-8 byte representation. This is the shared primitive behind every
// hex-encoded OData key this provider has confirmed against SAP's own
// documentation — SAP's Security Content API states it explicitly ("Hex
// representation of the alias does mean that a UTF-8-encoded byte array is
// built from the alias string. A hex string is then calculated from this
// byte array"), and the Partner Directory AlternativePartners entity's
// documented example (`"agency1"` -> `"6167656e637931"`) is consistent with
// the same rule, even though its own documentation never states it as
// explicitly. Using Go's native UTF-8 string-to-[]byte conversion (rather
// than an ASCII-only or byte-by-byte scheme) is what makes this total for
// every possible alias, including one containing spaces, punctuation, or
// non-ASCII Unicode characters.
func EncodeUTF8Hex(s string) string {
	return hex.EncodeToString([]byte(s))
}

// DecodeUTF8Hex reverses EncodeUTF8Hex, recovering the original string from
// its lowercase (or uppercase — encoding/hex accepts either) hex encoding.
// Used both to validate a hex string this client did not itself produce
// (for example one a practitioner typed into a Terraform import ID by
// hand) and to recover a plain identity component from an import ID built
// out of hex components.
func DecodeUTF8Hex(hexValue string) (string, error) {
	decoded, err := hex.DecodeString(hexValue)
	if err != nil {
		return "", fmt.Errorf("odata: %q is not a valid hex-encoded value: %w", hexValue, err)
	}
	return string(decoded), nil
}
