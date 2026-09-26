package partnerdirectory

import (
	"context"
	"encoding/json"
	"fmt"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const stringParametersEntitySet = "StringParameters"

// StringParameter is the wire representation of a StringParameters entity:
// a single named string value scoped to a partner (Pid). Identity is the
// composite key (Pid, Id); SAP documents full CRUD for this entity set.
type StringParameter struct {
	Pid   string `json:"Pid"`
	Id    string `json:"Id"`
	Value string `json:"Value"`
}

func stringParameterKey(pid, id string) (string, error) {
	return v2.CompositeKeyPredicate("Pid", pid, "Id", id)
}

// GetStringParameter reads a single string parameter by its (Pid, Id) key.
func (c *Client) GetStringParameter(ctx context.Context, pid, id string) (*StringParameter, error) {
	key, err := stringParameterKey(pid, id)
	if err != nil {
		return nil, err
	}

	body, err := c.odata.Get(ctx, v2.BuildPath(stringParametersEntitySet, key, ""))
	if err != nil {
		return nil, err
	}

	var sp StringParameter
	if err := v2.DecodeEntity(body, &sp); err != nil {
		return nil, err
	}
	return &sp, nil
}

// CreateStringParameter creates a new string parameter. The Pid it names
// does not need to exist beforehand: SAP creates the partner implicitly if
// this is the first entity referencing it.
func (c *Client) CreateStringParameter(ctx context.Context, sp StringParameter) (*StringParameter, error) {
	payload, err := json.Marshal(sp)
	if err != nil {
		return nil, fmt.Errorf("partnerdirectory: encoding string parameter: %w", err)
	}

	body, err := c.odata.Post(ctx, stringParametersEntitySet, payload)
	if err != nil {
		return nil, err
	}
	// SAP may accept the write with 202 and no body; read the entry back.
	if v2.EmptyBody(body) {
		return c.GetStringParameter(ctx, sp.Pid, sp.Id)
	}

	var created StringParameter
	if err := v2.DecodeEntity(body, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateStringParameter replaces the value of an existing string parameter
// via PUT, addressed by its (Pid, Id) key. Pid and Id are the entity's key
// and are never sent in the body; SAP documents PUT (full replace) rather
// than PATCH for this entity set.
func (c *Client) UpdateStringParameter(ctx context.Context, pid, id, value string) error {
	key, err := stringParameterKey(pid, id)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(struct {
		Value string `json:"Value"`
	}{Value: value})
	if err != nil {
		return fmt.Errorf("partnerdirectory: encoding string parameter: %w", err)
	}

	_, err = c.odata.Put(ctx, v2.BuildPath(stringParametersEntitySet, key, ""), payload)
	return err
}

// DeleteStringParameter deletes a single string parameter by its (Pid, Id)
// key.
func (c *Client) DeleteStringParameter(ctx context.Context, pid, id string) error {
	key, err := stringParameterKey(pid, id)
	if err != nil {
		return err
	}
	return c.odata.Delete(ctx, v2.BuildPath(stringParametersEntitySet, key, ""))
}

// ListStringParameters returns every string parameter for a given partner,
// following SAP's server-driven paging until exhausted. Partner Directory
// tenants can have large numbers of string parameters, so this always
// fetches every page rather than assuming a single response covers them
// all.
func (c *Client) ListStringParameters(ctx context.Context, pid string) ([]StringParameter, error) {
	filter := v2.Query{Filter: v2.FilterEquals("Pid", pid)}
	path := v2.BuildPath(stringParametersEntitySet, "", filter.Encode())

	params, err := v2.GetAllPages[StringParameter](ctx, c.odata, path)
	if err != nil {
		return nil, fmt.Errorf("partnerdirectory: listing string parameters: %w", err)
	}
	return params, nil
}
