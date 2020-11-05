package form3

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/google/uuid"
)

var (
	// retriableStatusCodes contains the status codes for which the client should retry
	// the operation using an exponential back-off algorithm.
	retriableStatusCodes = map[int]struct{}{
		http.StatusTooManyRequests:     {},
		http.StatusInternalServerError: {},
		http.StatusServiceUnavailable:  {},
		http.StatusGatewayTimeout:      {},
	}

	// default timeout for the inner HTTP server of the client
	// in production would be much bigger

	timeout = 10 * time.Second
	// max elapsed time for the rate limiter to retry requests
	// in production would be much bigger
	backoffMaxElapsedTime = 10 * time.Second
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
			Timeout: timeout,
		},
	}
}

// Fetch returns an organisation account given its accountID in the form of
// an UUID V4.
func (c *Client) Fetch(accountID uuid.UUID) (OrganisationAccount, error) {
	resp, err := c.performRequest(
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
	defer resp.Body.Close()

	err = c.checkErrorMessage(resp)
	if err != nil {
		return OrganisationAccount{}, err
	}

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
func (c *Client) List(loo ...ListOption) ([]OrganisationAccount, error) {
	options := listOptions{}
	for _, lo := range loo {
		lo(&options)
	}

	url, err := url.Parse(fmt.Sprintf("%s/v1/organisation/accounts", c.baseURL))
	if err != nil {
		return nil, err
	}

	urlQuery := url.Query()

	if options.pageNumber != 0 {
		urlQuery.Set("page[number]", strconv.Itoa(options.pageNumber))
	}

	if options.pageSize != 0 {
		urlQuery.Set("page[size]", strconv.Itoa(options.pageSize))
	}

	url.RawQuery = urlQuery.Encode()

	resp, err := c.performRequest(
		http.MethodGet,
		url.String(),
		nil,
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = c.checkErrorMessage(resp)
	if err != nil {
		return nil, err
	}

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
	resp, err := c.performRequest(
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
	defer resp.Body.Close()

	return c.checkErrorMessage(resp)
}

// Create will create a new organisation account.
func (c *Client) Create(organisationAccount OrganisationAccount) (OrganisationAccount, error) {
	body := struct {
		Data OrganisationAccount `json:"data"`
	}{
		Data: organisationAccount,
	}

	bodyBytes := new(bytes.Buffer)

	err := json.NewEncoder(bodyBytes).Encode(&body)
	if err != nil {
		return OrganisationAccount{}, err
	}

	resp, err := c.performRequest(
		http.MethodPost,
		fmt.Sprintf("%s/v1/organisation/accounts", c.baseURL),
		bodyBytes,
	)
	if err != nil {
		return OrganisationAccount{}, err
	}
	defer resp.Body.Close()

	err = c.checkErrorMessage(resp)
	if err != nil {
		return OrganisationAccount{}, err
	}

	var data struct {
		Data OrganisationAccount `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return OrganisationAccount{}, err
	}

	return data.Data, nil
}

// performRequest is the general method called by all exported methods of the client library
// to perform a request against the Form3 API.
//
// It uses an exponential back-off algorithm so that it can retry certain operations given
// a certain set of status codes (situated inside retriableStatusCodes at the top).
func (c *Client) performRequest(method string, url string, body io.Reader) (*http.Response, error) {
	expBackOff := backoff.NewExponentialBackOff()
	expBackOff.MaxElapsedTime = backoffMaxElapsedTime
	ticker := backoff.NewTicker(expBackOff)

	var req *http.Request
	var resp *http.Response
	var err error

	for range ticker.C {
		req, err = http.NewRequest(
			method,
			url,
			body,
		)
		if err != nil {
			ticker.Stop()
			break
		}

		resp, err = c.httpClient.Do(req)
		if err != nil {
			ticker.Stop()
			break
		}

		if _, ok := retriableStatusCodes[resp.StatusCode]; ok {
			continue
		}

		ticker.Stop()
		break
	}

	return resp, err
}

// checkErrorMessage verifies if we made a bad request, in which case
// we parse the error message and return it to the caller
func (c *Client) checkErrorMessage(resp *http.Response) error {
	if resp.StatusCode >= 300 {
		var data struct {
			ErrorMessage string `json:"error_message"`
		}
		err := json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return err
		}

		return errors.New(data.ErrorMessage)
	}

	return nil
}
