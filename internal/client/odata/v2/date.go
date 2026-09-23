package v2

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FormatDateLiteral renders t as an OData V2 JSON Edm.DateTime literal,
// "/Date(<milliseconds-since-epoch>)/" — confirmed directly from SAP's own
// documented "Generate a Key Pair" request example
// (`"ValidNotBefore":"/Date(1469685029780)/"`). This is OData V2's standard
// JSON date convention, distinct from the ISO 8601 strings OData V4 uses.
func FormatDateLiteral(t time.Time) string {
	return fmt.Sprintf("/Date(%d)/", t.UnixMilli())
}

// ParseDateLiteral parses an OData V2 JSON Edm.DateTime literal
// ("/Date(<millis>)/") back into a time.Time in UTC.
func ParseDateLiteral(s string) (time.Time, error) {
	const prefix, suffix = "/Date(", ")/"
	if !strings.HasPrefix(s, prefix) || !strings.HasSuffix(s, suffix) {
		return time.Time{}, fmt.Errorf("odata: %q is not an OData V2 date literal", s)
	}
	millisStr := s[len(prefix) : len(s)-len(suffix)]
	millis, err := strconv.ParseInt(millisStr, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("odata: %q is not an OData V2 date literal: %w", s, err)
	}
	return time.UnixMilli(millis).UTC(), nil
}
