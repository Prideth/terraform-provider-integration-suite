package partnerdirectory

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const binaryParametersEntitySet = "BinaryParameters"

// MaxBinaryParameterValueBytes is the largest decoded (raw, not
// base64-encoded) Value a Binary Parameter can hold: the tenant $metadata
// declares BinaryParameter.Value as Edm.Binary with MaxLength 1572864
// (1.5 MiB), matching the "maximum size of 1,5 MB" on SAP's entity types
// page. Other SAP pages still state 260 KB or 262144 bytes; the service's
// own metadata is taken as authoritative, and SAP rejects anything its
// actual limit does not allow. Resources check this before uploading so an
// oversized file fails early with a clear message.
const MaxBinaryParameterValueBytes = 1572864

// DocumentedBinaryParameterContentTypes lists the Content-Type values SAP's
// own documentation names for Binary Parameters. This is documentation,
// not a validator allow-list: SAP's Content-Type field is intentionally
// more permissive than this fixed set (for example, encoding suffixes like
// "xml;encoding=UTF-8" are valid but would not appear in a short canonical
// list), so this provider's content_type schema attribute does not
// restrict values to this list.
var DocumentedBinaryParameterContentTypes = []string{
	"xml",
	"xsl",
	"xsd",
	"json",
	"text",
	"zip",
	"gz",
	"zlib",
	"crt",
}

// BinaryParameter is the wire representation of a BinaryParameters entity:
// a single named binary value scoped to a partner (Pid), transported as a
// base64-encoded string over JSON. Identity is the composite key (Pid, Id).
type BinaryParameter struct {
	Pid         string `json:"Pid"`
	Id          string `json:"Id"`
	ContentType string `json:"ContentType"`

	// Value is the base64-encoded content, exactly as the caller supplied
	// or SAP returned it. This client transports it opaquely and performs
	// no decoding of its own.
	Value string `json:"Value"`
}

func binaryParameterKey(pid, id string) (string, error) {
	return v2.CompositeKeyPredicate("Pid", pid, "Id", id)
}

// GetBinaryParameter reads a single binary parameter by its (Pid, Id) key.
func (c *Client) GetBinaryParameter(ctx context.Context, pid, id string) (*BinaryParameter, error) {
	key, err := binaryParameterKey(pid, id)
	if err != nil {
		return nil, err
	}

	body, err := c.odata.Get(ctx, v2.BuildPath(binaryParametersEntitySet, key, ""))
	if err != nil {
		return nil, err
	}

	var bp BinaryParameter
	if err := v2.DecodeEntity(body, &bp); err != nil {
		return nil, err
	}
	return &bp, nil
}

// CreateBinaryParameter creates a new binary parameter from raw (not yet
// base64-encoded) content. The Pid it names does not need to exist
// beforehand, the same implicit-creation semantics as StringParameter.
func (c *Client) CreateBinaryParameter(ctx context.Context, pid, id, contentType string, content []byte) (*BinaryParameter, error) {
	payload, err := json.Marshal(BinaryParameter{
		Pid:         pid,
		Id:          id,
		ContentType: contentType,
		Value:       base64.StdEncoding.EncodeToString(content),
	})
	if err != nil {
		return nil, fmt.Errorf("partnerdirectory: encoding binary parameter: %w", err)
	}

	body, err := c.odata.Post(ctx, binaryParametersEntitySet, payload)
	if err != nil {
		return nil, err
	}

	var created BinaryParameter
	if err := v2.DecodeEntity(body, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateBinaryParameter replaces the content type and value of an existing
// binary parameter via PUT, addressed by its (Pid, Id) key, from raw (not
// yet base64-encoded) content — the same full-replace semantics as
// UpdateStringParameter.
func (c *Client) UpdateBinaryParameter(ctx context.Context, pid, id, contentType string, content []byte) error {
	key, err := binaryParameterKey(pid, id)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(struct {
		ContentType string `json:"ContentType"`
		Value       string `json:"Value"`
	}{ContentType: contentType, Value: base64.StdEncoding.EncodeToString(content)})
	if err != nil {
		return fmt.Errorf("partnerdirectory: encoding binary parameter: %w", err)
	}

	_, err = c.odata.Put(ctx, v2.BuildPath(binaryParametersEntitySet, key, ""), payload)
	return err
}

// DeleteBinaryParameter deletes a single binary parameter by its (Pid, Id)
// key.
func (c *Client) DeleteBinaryParameter(ctx context.Context, pid, id string) error {
	key, err := binaryParameterKey(pid, id)
	if err != nil {
		return err
	}
	return c.odata.Delete(ctx, v2.BuildPath(binaryParametersEntitySet, key, ""))
}

// ListBinaryParameters returns every binary parameter for a given partner,
// following SAP's server-driven paging until exhausted.
func (c *Client) ListBinaryParameters(ctx context.Context, pid string) ([]BinaryParameter, error) {
	filter := v2.Query{Filter: v2.FilterEquals("Pid", pid)}
	path := v2.BuildPath(binaryParametersEntitySet, "", filter.Encode())

	params, err := v2.GetAllPages[BinaryParameter](ctx, c.odata, path)
	if err != nil {
		return nil, fmt.Errorf("partnerdirectory: listing binary parameters: %w", err)
	}
	return params, nil
}
