// Package v2 implements the OData V2 conventions used by SAP Cloud
// Integration's Integration Content, Security Content, and Partner
// Directory APIs: the "d"/"results" response envelope, $-query construction,
// key-predicate formatting, pagination, and the OData V2 error format.
package v2

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Query builds the $-prefixed query options for an OData V2 request.
type Query struct {
	Filter      string
	Select      []string
	Expand      []string
	Top         int
	Skip        int
	InlineCount bool
}

// Encode renders the query as URL-encoded $-options, in a stable order.
func (q Query) Encode() string {
	values := url.Values{}
	if q.Filter != "" {
		values.Set("$filter", q.Filter)
	}
	if len(q.Select) > 0 {
		values.Set("$select", strings.Join(q.Select, ","))
	}
	if len(q.Expand) > 0 {
		values.Set("$expand", strings.Join(q.Expand, ","))
	}
	if q.Top > 0 {
		values.Set("$top", strconv.Itoa(q.Top))
	}
	if q.Skip > 0 {
		values.Set("$skip", strconv.Itoa(q.Skip))
	}
	if q.InlineCount {
		values.Set("$inlinecount", "allpages")
	}
	return values.Encode()
}

// EscapeLiteral escapes a string literal for use inside an OData V2 $filter
// expression, following the protocol's single-quote-doubling rule.
func EscapeLiteral(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// FilterEquals builds a `<field> eq '<value>'` filter fragment with value
// safely escaped and quoted.
func FilterEquals(field, value string) string {
	return fmt.Sprintf("%s eq '%s'", field, EscapeLiteral(value))
}

// KeyPredicate renders a single-key OData key predicate, e.g. Id('foo').
func KeyPredicate(key string) string {
	return fmt.Sprintf("('%s')", EscapeLiteral(key))
}

// CompositeKeyPredicate renders a multi-key OData key predicate, e.g.
// (Id='foo',Version='1.0.0'), from key/value pairs supplied as consecutive
// arguments (k1, v1, k2, v2, ...). Values are quoted and escaped as string
// literals; callers needing a non-string key segment should format it
// themselves and pass the finished predicate through RawKeyPredicate.
func CompositeKeyPredicate(pairs ...string) (string, error) {
	if len(pairs)%2 != 0 {
		return "", fmt.Errorf("odata: CompositeKeyPredicate requires an even number of key/value arguments")
	}
	parts := make([]string, 0, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		parts = append(parts, fmt.Sprintf("%s='%s'", pairs[i], EscapeLiteral(pairs[i+1]))) //nolint:gosec // G602: the even-length check above guarantees i+1 < len(pairs) for every i this loop reaches
	}
	return "(" + strings.Join(parts, ",") + ")", nil
}

// BuildPath joins a base entity-set path with an optional key predicate and
// query string.
func BuildPath(entitySet, keyPredicate, query string) string {
	path := entitySet + keyPredicate
	if query == "" {
		return path
	}
	return path + "?" + query
}
