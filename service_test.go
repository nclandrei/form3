package form3

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Form3TestSuite defines the integration testing that we want
// to perform against the fake accounts API inside the container.
type Form3TestSuite struct {
	suite.Suite
	client            *Client
	testOrganisations []OrganisationAccount
}

// This method is run before each test and it will
// create all the organisations insidde valid_organisations JSON file.
func (s *Form3TestSuite) SetupTest() {
	s.client = NewClient(os.Getenv("API_BASE_URL"))

	f, err := os.Open(os.Getenv("TESTDATA_ORGANISATIONS_FILE_PATH"))
	if err != nil {
		panic(err)
	}

	var organisations []OrganisationAccount
	err = json.NewDecoder(f).Decode(&organisations)
	if err != nil {
		panic(err)
	}

	s.testOrganisations = organisations

	for _, org := range s.testOrganisations {
		_, err := s.client.Create(org)
		if err != nil {
			panic(err)
		}
	}
}

// This method is run after each test  and it will
// clear up all organisations that were created.
func (s *Form3TestSuite) TearDownTest() {
	orgs, err := s.client.List()
	if err != nil {
		panic(err)
	}

	for _, org := range orgs {
		err := s.client.Delete(org.ID, org.Version)
		if err != nil {
			panic(err)
		}
	}
}

func (s *Form3TestSuite) TestFetch() {
	testCases := []struct {
		name               string
		id                 uuid.UUID
		shouldReturnErr    bool
		expectedErrMessage string
	}{
		{
			name: "OK - existing org",
			id:   uuid.MustParse("a9e3b971-a241-4930-a09f-a7c04bf394fe"),
		},
		{
			name:               "Not Found - non-existing org",
			id:                 uuid.MustParse("3cbbba27-3b51-42f4-88a7-729fa42a4a68"),
			shouldReturnErr:    true,
			expectedErrMessage: "does not exist",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			org, err := s.client.Fetch(tc.id)

			if tc.shouldReturnErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.id, org.ID)
			}
		})
	}
}

func (s *Form3TestSuite) TestList() {
	testCases := []struct {
		name         string
		paging       bool
		pageNumber   int
		pageSize     int
		expectedOrgs []OrganisationAccount
	}{
		{
			name:         "OK - all organisations",
			expectedOrgs: s.testOrganisations,
		},
		{
			name:         "OK - first 2 organisations",
			paging:       true,
			pageNumber:   0,
			pageSize:     2,
			expectedOrgs: s.testOrganisations[:2],
		},
		{
			name:         "OK - last 2 organisations",
			paging:       true,
			pageNumber:   1,
			pageSize:     3,
			expectedOrgs: s.testOrganisations[3:],
		},
		{
			name:         "OK - no organisations",
			paging:       true,
			pageNumber:   5,
			pageSize:     5,
			expectedOrgs: nil,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			var options []ListOption
			if tc.paging {
				options = append(
					options,
					PageNumberListOption(tc.pageNumber),
					PageSizeListOption(tc.pageSize),
				)
			}

			orgs, err := s.client.List(options...)

			assert.NoError(t, err)
			assert.Equal(t, len(tc.expectedOrgs), len(orgs))
			assert.ElementsMatch(t, tc.expectedOrgs, orgs)
		})
	}
}

func (s *Form3TestSuite) TestDelete() {
	testCases := []struct {
		name               string
		id                 uuid.UUID
		version            int
		expectedErr        bool
		expectedErrMessage string
		expectedOrgs       []OrganisationAccount
	}{
		{
			name:         "OK - organisation removed",
			id:           uuid.MustParse("a9e3b971-a241-4930-a09f-a7c04bf394fe"),
			version:      0,
			expectedErr:  false,
			expectedOrgs: s.testOrganisations[1:],
		},
		// the test below should actually return an error as per the official documentation
		// of the API (404 Not Found), but in the fake API you guys implemented, it returns 204 if a resource
		// you're trying to delete does not exist, so I adapted the test to run as per the
		// fake API implementation
		{
			name:         "OK - organisation does not exist in Form3, but no error",
			id:           uuid.MustParse("a0607d72-11d2-4a3a-87f8-186e53b07811"),
			version:      0,
			expectedErr:  false,
			expectedOrgs: s.testOrganisations[1:],
		},
		{
			name:               "Conflict - incorrect version used when removing an existing organisation",
			id:                 uuid.MustParse("3c76048a-2024-4917-b911-1b3e88fccfb3"),
			version:            5,
			expectedErr:        true,
			expectedErrMessage: "invalid version",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			err := s.client.Delete(tc.id, tc.version)

			if tc.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMessage)
			} else {
				assert.NoError(t, err)

				orgs, err := s.client.List()

				assert.NoError(t, err)
				assert.Equal(t, len(tc.expectedOrgs), len(orgs))
				assert.ElementsMatch(t, tc.expectedOrgs, orgs)
			}

		})
	}
}

func (s *Form3TestSuite) TestCreate() {
	f, err := os.Open(os.Getenv("TESTDATA_CREATE_ORGANISATIONS_FILE_PATH"))
	if err != nil {
		panic(err)
	}

	var createOrgs []OrganisationAccount

	err = json.NewDecoder(f).Decode(&createOrgs)
	if err != nil {
		panic(err)
	}

	testCases := []struct {
		name               string
		orgToCreate        OrganisationAccount
		expectedErr        bool
		expectedErrMessage string
		expectedOrgs       []OrganisationAccount
	}{
		{
			name:         "OK - organisation created",
			orgToCreate:  createOrgs[0],
			expectedErr:  false,
			expectedOrgs: append(s.testOrganisations, createOrgs[0]),
		},
		{
			name:               "Bad Request - organisation already exists",
			orgToCreate:        createOrgs[1],
			expectedErr:        true,
			expectedErrMessage: "violates a duplicate constraint",
		},
		{
			name:               "Bad Request - organisation has invalid fields",
			orgToCreate:        createOrgs[2],
			expectedErr:        true,
			expectedErrMessage: "validation failure list",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			org, err := s.client.Create(tc.orgToCreate)

			if tc.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.orgToCreate, org)

				orgs, err := s.client.List()

				assert.NoError(t, err)
				assert.Equal(t, len(tc.expectedOrgs), len(orgs))
				assert.ElementsMatch(t, tc.expectedOrgs, orgs)
			}

		})
	}
}

func (s *Form3TestSuite) TestRateLimiter() {
	testCases := []struct {
		name        string
		expectErr   bool
		handlerFunc func() http.HandlerFunc
	}{
		{
			name:      "OK - rate limiter retries till service is available",
			expectErr: false,
			handlerFunc: func() http.HandlerFunc {
				var serviceUnavailableCounter int

				return func(w http.ResponseWriter, r *http.Request) {
					if serviceUnavailableCounter == 3 {
						data := struct {
							Data []struct{} `json:"data"`
						}{
							Data: []struct{}{},
						}

						w.WriteHeader(http.StatusOK)

						_ = json.NewEncoder(w).Encode(&data)

						return
					}

					w.WriteHeader(http.StatusTooManyRequests)

					message := struct {
						ErrorMessage string `json:"error_message"`
					}{
						ErrorMessage: "Service Unavailable",
					}

					_ = json.NewEncoder(w).Encode(&message)

					serviceUnavailableCounter++
				}
			},
		},
		{
			name:      "Not OK - rate limiter sees 429 on every call, returns error",
			expectErr: true,
			handlerFunc: func() http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusTooManyRequests)

					message := struct {
						ErrorMessage string `json:"error_message"`
					}{
						ErrorMessage: "Service Unavailable",
					}

					_ = json.NewEncoder(w).Encode(&message)
				}
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.Handle("/v1/organisation/accounts", tc.handlerFunc())

			ts := httptest.NewServer(mux)
			defer ts.Close()

			client := NewClient(ts.URL)

			_, err := client.List()

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestForm3TestSuite(t *testing.T) {
	suite.Run(t, new(Form3TestSuite))
}
