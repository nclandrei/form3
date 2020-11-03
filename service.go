package form3

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var (
	GeneralErr        = errors.New("could not perform operation")
	DeleteNotFoundErr = errors.New("specified resource does not exist")
	DeleteConflictErr = errors.New("specified version incorrect")
)

// Client is the service that interacts with the Form3 API. It can perform
// the following actions on Organisation Accounts: create, fetch, list and delete.
type Client struct {
	baseURL    string
	httpClient http.Client
}

// NewClient returns a new instance of the client service that
// interacts with the Form3 API.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Fetch returns an organisation account given its accountID in the form of
// an UUID V4.
func (c *Client) Fetch(accountID uuid.UUID) (OrganisationAccount, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf(
			"%s/v1/organisation/accounts/%s",
			c.baseURL,
			accountID.String(),
		),
		nil,
	)
	if err != nil {
		return OrganisationAccount{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return OrganisationAccount{}, err
	}
	defer resp.Body.Close()

	var organisationAccount struct {
		Data OrganisationAccount `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&organisationAccount)
	if err != nil {
		return OrganisationAccount{}, err
	}

	return organisationAccount.Data, nil
}

// List returns a list of organisation accounts. It can support paging,
// which implies that the caller of the method should provide a page
// number and its size.
func (c *Client) List() ([]OrganisationAccount, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/v1/organisation/accounts", c.baseURL),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var organisationAccounts struct {
		Data []OrganisationAccount `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&organisationAccounts)
	if err != nil {
		return nil, err
	}

	return organisationAccounts.Data, nil
}

// Delete will remove an organisation account given its account ID and version.
func (c *Client) Delete(accountID uuid.UUID, version int) error {
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf(
			"%s/v1/organisation/accounts/%s?version=%d",
			c.baseURL,
			accountID.String(),
			version,
		),
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return DeleteNotFoundErr
	case http.StatusConflict:
		return DeleteConflictErr
	case http.StatusNoContent:
		return nil
	default:
		return GeneralErr
	}
}

// Create will create a new organisation account given all the fields are correct.
func (c *Client) Create(organisationAccount OrganisationAccount) error {
	data := struct {
		Data OrganisationAccount `json:"data"`
	}{
		Data: organisationAccount,
	}

	var body *bytes.Buffer

	err := json.NewEncoder(body).Encode(&data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/v1/organisation/accounts", c.baseURL),
		body,
	)
	if err != nil {
		return err
	}

	_, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}

	return err
}
