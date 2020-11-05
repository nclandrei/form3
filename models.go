package form3

import (
	"github.com/google/uuid"
)

var (
	// PageNumber is a List call option to set the page number.
	PageNumber = func(pageNumber int) func(*listOptions) {
		return func(lo *listOptions) {
			lo.pageNumber = pageNumber
		}
	}

	// PageSize is a List call option to set the page size.
	PageSize = func(pageSize int) func(*listOptions) {
		return func(lo *listOptions) {
			lo.pageSize = pageSize
		}
	}
)

type listOptions struct {
	pageNumber int
	pageSize   int
}

// ListOption is a function that can determine whether the List call
// to the Form3 API should have any paging settings.
type ListOption = func(*listOptions)

// OrganisationAccount represents a bank account that is registered with Form3.
// It is used to validate and allocate inbound payments.
type OrganisationAccount struct {
	ID             uuid.UUID                     `json:"id"`
	Type           string                        `json:"type"`
	OrganisationID uuid.UUID                     `json:"organisation_id"`
	Version        int                           `json:"version"`
	Attributes     OrganisationAccountAttributes `json:"attributes"`
}

// OrganisationAccountAttributes represent various attributes that can be included
// inside the organisation account entity.
type OrganisationAccountAttributes struct {
	Country                 string   `json:"country"`
	BaseCurrency            string   `json:"base_currency"`
	AccountNumber           string   `json:"account_number"`
	BankID                  string   `json:"bank_id"`
	BankIDCode              string   `json:"bank_id_code"`
	BIC                     string   `json:"bic"`
	IBAN                    string   `json:"iban"`
	Name                    []string `json:"name"`
	AlternativeNames        []string `json:"alternative_names"`
	AccountClassification   string   `json:"account_classification"`
	JointAccount            bool     `json:"joint_account"`
	AccountMatchingOptOut   bool     `json:"account_matching_opt_out"`
	SecondaryIdentification string   `json:"secondary_identification"`
	Switched                bool     `json:"switched"`
}
