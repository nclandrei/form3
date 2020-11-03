package form3

import "github.com/google/uuid"

// OrganisationAccount represents a bank account that is registered with Form3.
// It is used to validate and allocate inbound payments.
type OrganisationAccount struct {
	ID             uuid.UUID
	Type           string
	OrganisationID uuid.UUID
	Version        int
	Attributes     OrganisationAccountAttributes
}

// OrganisationAccountAttributes represent various attributes that can be included
// inside the organisation account entity.
type OrganisationAccountAttributes struct {
	Country                 string
	BaseCurrency            string
	AccountNumber           string
	BankID                  string
	BankIDCode              string
	BIC                     string
	IBAN                    string
	Name                    []string
	AlternativeNames        []string
	AccountClasification    string
	JointAccount            bool
	AccountMatchingOptOut   bool
	SecondaryIdentification string
	Switched                bool
}
