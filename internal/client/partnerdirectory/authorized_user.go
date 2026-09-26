package partnerdirectory

import (
	"context"
	"encoding/json"
	"fmt"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const authorizedUsersEntitySet = "AuthorizedUsers"

// AuthorizedUser is the wire representation of an AuthorizedUsers entity:
// a mapping from a communication user to the Partner ID (Pid) that user is
// authorized to act as for inbound communication. SAP documents this as a
// many-to-one mapping: a communication user can be assigned to exactly one
// Pid, but a Pid can have several authorized users. The User field is the
// entity's key.
//
// SAP stores User lowercased (Locale.English): its example request creates
// "MyUser" and the response carries "myuser", and filters on User must use
// lowercase. The client passes User through unchanged; the provider rejects
// uppercase values before they reach SAP.
type AuthorizedUser struct {
	User string `json:"User"`
	Pid  string `json:"Pid"`
}

// GetAuthorizedUser reads a single authorized user mapping by its User key.
func (c *Client) GetAuthorizedUser(ctx context.Context, user string) (*AuthorizedUser, error) {
	path := v2.BuildPath(authorizedUsersEntitySet, v2.KeyPredicate(user), "")

	body, err := c.odata.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var au AuthorizedUser
	if err := v2.DecodeEntity(body, &au); err != nil {
		return nil, err
	}
	return &au, nil
}

// CreateAuthorizedUser creates a new authorized user mapping.
func (c *Client) CreateAuthorizedUser(ctx context.Context, au AuthorizedUser) (*AuthorizedUser, error) {
	payload, err := json.Marshal(au)
	if err != nil {
		return nil, fmt.Errorf("partnerdirectory: encoding authorized user: %w", err)
	}

	body, err := c.odata.Post(ctx, authorizedUsersEntitySet, payload)
	if err != nil {
		return nil, err
	}
	// SAP may accept the write with 202 and no body; read the entry back.
	if v2.EmptyBody(body) {
		return c.GetAuthorizedUser(ctx, au.User)
	}

	var created AuthorizedUser
	if err := v2.DecodeEntity(body, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateAuthorizedUser repoints an existing authorized user at a different
// Pid via PUT, addressed by the User key.
func (c *Client) UpdateAuthorizedUser(ctx context.Context, user, pid string) error {
	path := v2.BuildPath(authorizedUsersEntitySet, v2.KeyPredicate(user), "")

	payload, err := json.Marshal(struct {
		Pid string `json:"Pid"`
	}{Pid: pid})
	if err != nil {
		return fmt.Errorf("partnerdirectory: encoding authorized user: %w", err)
	}

	_, err = c.odata.Put(ctx, path, payload)
	return err
}

// DeleteAuthorizedUser deletes a single authorized user mapping by its User
// key.
func (c *Client) DeleteAuthorizedUser(ctx context.Context, user string) error {
	path := v2.BuildPath(authorizedUsersEntitySet, v2.KeyPredicate(user), "")
	return c.odata.Delete(ctx, path)
}

// ListAuthorizedUsers returns every authorized user mapped to a given Pid,
// following SAP's server-driven paging until exhausted.
func (c *Client) ListAuthorizedUsers(ctx context.Context, pid string) ([]AuthorizedUser, error) {
	filter := v2.Query{Filter: v2.FilterEquals("Pid", pid)}
	path := v2.BuildPath(authorizedUsersEntitySet, "", filter.Encode())

	users, err := v2.GetAllPages[AuthorizedUser](ctx, c.odata, path)
	if err != nil {
		return nil, fmt.Errorf("partnerdirectory: listing authorized users: %w", err)
	}
	return users, nil
}
