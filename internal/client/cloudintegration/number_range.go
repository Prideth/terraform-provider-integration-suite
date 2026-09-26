package cloudintegration

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const numberRangesEntitySet = "NumberRanges"

// NumberRange is the static, desired-state configuration of a Number Ranges
// object, confirmed verbatim from SAP's own "Add a Number Ranges Object" and
// "Update a Number Ranges Object" documentation pages. CurrentValue is
// intentionally not a plain field here: the counter advances at run time,
// so it is only ever sent when the caller explicitly opts in (see
// CurrentValue field and UpdateNumberRange), never silently resent from a
// remembered value.
type NumberRange struct {
	Name        string
	MinValue    string
	MaxValue    string
	Description string
	Rotate      bool
	FieldLength string

	// CurrentValue, when non-nil, is sent as the wire "CurrentValue"
	// property. On Create this is always required (SAP's documented
	// example always includes it). On Update, leave this nil to omit the
	// property from the request body entirely rather than resend a value
	// this client cannot confirm is still current — see
	// docs/guides/runtime-stores-and-number-ranges.md for why, and for the
	// unconfirmed-behavior risk that omission itself carries.
	CurrentValue *string
}

// numberRangeWireModel is the exact JSON shape SAP documents, field for
// field, for both Add and Update: {"CurrentValue":"0","Name":"...",
// "MinValue":"0","MaxValue":"9999","Description":"...","Rotate":"true",
// "FieldLength":"4"}. Every numeric-looking field is a JSON string on the
// wire, matching SAP's own example exactly — this client never assumes a
// narrower Go numeric type is safe (SAP documents values up to 14 digits,
// beyond what some numeric encodings preserve exactly).
type numberRangeWireModel struct {
	CurrentValue *string `json:"CurrentValue,omitempty"`
	Name         string  `json:"Name"`
	MinValue     string  `json:"MinValue"`
	MaxValue     string  `json:"MaxValue"`
	Description  string  `json:"Description"`
	Rotate       string  `json:"Rotate"`
	FieldLength  string  `json:"FieldLength"`
}

func numberRangeToWire(nr NumberRange) numberRangeWireModel {
	return numberRangeWireModel{
		CurrentValue: nr.CurrentValue,
		Name:         nr.Name,
		MinValue:     nr.MinValue,
		MaxValue:     nr.MaxValue,
		Description:  nr.Description,
		Rotate:       strconv.FormatBool(nr.Rotate),
		FieldLength:  nr.FieldLength,
	}
}

// CreateNumberRange adds a new Number Ranges object to the tenant. Confirmed
// directly from SAP's own documentation: `POST /api/v1/NumberRanges`.
// nr.CurrentValue must be non-nil: SAP's documented example always includes
// CurrentValue on creation, and this project found nothing suggesting it is
// optional.
func (c *Client) CreateNumberRange(ctx context.Context, nr NumberRange) error {
	if nr.CurrentValue == nil {
		return fmt.Errorf("cloudintegration: CreateNumberRange requires a CurrentValue; SAP's documented Add example always includes it")
	}

	payload, err := json.Marshal(numberRangeToWire(nr))
	if err != nil {
		return fmt.Errorf("cloudintegration: encoding number range create request: %w", err)
	}

	_, err = c.odata.Post(ctx, v2.BuildPath(numberRangesEntitySet, "", ""), payload)
	return err
}

// NumberRangeState is a Number Ranges object as GET returns it. SAP
// documents no GET for this entity, but a tenant answered
// GET NumberRanges('<name>') with 200, echoing every field exactly as it was
// sent (numbers and Rotate as strings), and the tenant $metadata confirms
// the properties (September 2026).
type NumberRangeState struct {
	Name         string `json:"Name"`
	Description  string `json:"Description"`
	MinValue     string `json:"MinValue"`
	MaxValue     string `json:"MaxValue"`
	Rotate       string `json:"Rotate"`
	CurrentValue string `json:"CurrentValue"`
	FieldLength  string `json:"FieldLength"`
	DeployedBy   string `json:"DeployedBy"`
	DeployedOn   string `json:"DeployedOn"`
}

// GetNumberRange reads a Number Ranges object by name. Like the other
// Message Store entity sets, it is requested without query options, which
// the service rejects.
func (c *Client) GetNumberRange(ctx context.Context, name string) (*NumberRangeState, error) {
	body, err := c.odata.Get(ctx, v2.BuildPath(numberRangesEntitySet, v2.KeyPredicate(name), ""))
	if err != nil {
		return nil, err
	}
	var nr NumberRangeState
	if err := v2.DecodeEntity(body, &nr); err != nil {
		return nil, err
	}
	return &nr, nil
}

// DeleteNumberRange removes a Number Ranges object. SAP documents no DELETE
// for this entity; a tenant answered DELETE NumberRanges('<name>') with 202
// and the object was gone afterwards (September 2026).
func (c *Client) DeleteNumberRange(ctx context.Context, name string) error {
	return c.odata.Delete(ctx, v2.BuildPath(numberRangesEntitySet, v2.KeyPredicate(name), ""))
}

// UpdateNumberRange updates an existing Number Ranges object's static
// configuration. Confirmed directly from SAP's own documentation:
// `PUT /api/v1/NumberRanges('{objectName}')`.
//
// nr.CurrentValue controls whether the request body includes the
// CurrentValue property at all:
//   - nil (the ordinary case: only static fields like description/min/max/
//     rotate/field_length changed): CurrentValue is omitted from the JSON
//     body. SAP's documentation does not confirm whether an omitted field is
//     preserved unchanged (a partial-merge PUT) or reset to a default (a
//     literal full-replace PUT). Omission is the more conservative choice:
//     it never transmits a value this client knows to be stale, whereas
//     resending a remembered value goes stale the moment SAP's runtime
//     consumes even one number.
//   - non-nil (the caller explicitly wants to (re)set the counter, mirroring
//     this provider's write-only-secret-rotation pattern): CurrentValue is
//     sent with the given value.
func (c *Client) UpdateNumberRange(ctx context.Context, nr NumberRange) error {
	payload, err := json.Marshal(numberRangeToWire(nr))
	if err != nil {
		return fmt.Errorf("cloudintegration: encoding number range update request: %w", err)
	}

	path := v2.BuildPath(numberRangesEntitySet, v2.KeyPredicate(nr.Name), "")
	_, err = c.odata.Put(ctx, path, payload)
	return err
}
