package entity

import (	
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Insured represents a insured in the system.
// insureds can also be created directly for testing.
type Insured struct {
	ID              int
	Name            string
	PolicyNumber    int
	RecordTimestamp time.Time // insured CREATION time
	Employees       *map[int]Employee
	Addresses       *map[int]Address
}
var _ InsuredInterface = (*Insured)(nil)

// InsuredInterface for methods related to Insured objects (Insured, Employee, Address, and collections thereof)
type InsuredInterface interface {
	/* New() InsuredInterface // TODO: */
	// TODO: check why naming this "GetId() int" caused error "type has no field or method GetId"
	GetId() int64
	GetInsuredId() int64
	//DeleteId()
	Validate() error
	//ToRecord() Record
	//FromRecord(r Record) (err error)
	//MultipleFromRecords(records map[int]Record) (map[int]InsuredInterface, error)
	MarshalJSON() ([]byte, error)
	GetIdentTableName() string // table name with identity column
	GetDataTableName() string // table name with values (may be the same as identity)
	GetInsertFields() map[string]string
}

// Validate returns an error if the insured contains invalid fields.
// This only performs basic validation.
func (u *Insured) Validate() error {
	if u.Name == "" {
		return Errorf(EINVALID, "Insured name required.")
	}
	return nil
}
func (u *Insured) GetId() int64 {
	return int64(u.ID)
}
func (u *Insured) GetInsuredId() int64 {
	return int64(u.ID)
}
func (u *Insured) GetDataTableName() string {
	return "insured"
}
func (u *Insured) GetIdentTableName() string {
	return "insured"
}
func (u *Insured) GetInsertFields() map[string]string {
	return map[string]string{
		"name": u.Name,
		"policy_number":   strconv.Itoa(u.PolicyNumber),
	}
}



// InsuredFilter represents a filter passed to FindInsureds().
type InsuredFilter struct {
	// Filtering fields.
	ID              *int    `json:"id"`
	Name            *string `json:"name"`
	PolicyNumber    *int    `json:"policyNumber"`
	RecordTimestamp *int    `json:"recordTimestamp"`

	// Restrict to subset of results.
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// InsuredUpdate represents a set of fields to be updated via UpdateInsured().
type InsuredUpdate struct {
	Name         *string `json:"name"`
	PolicyNumber *int    `json:"policyNumber"`
}

func GetEntity(entityType string) (InsuredInterface, error) {
	// https://refactoring.guru/design-patterns/factory-method/go/example
	if entityType == "insured" || entityType == "Insured" {
		return &Insured{}, nil
	} else if entityType == "employee" || entityType == "Employee" {
		return &Employee{}, nil
	} else if entityType == "address" || entityType == "Address" {
		return &Address{}, nil
	}
	return nil, fmt.Errorf("Non-existent entity type %v", entityType)
}

func (e *Insured) ToRecord() Record {
	idString := strconv.Itoa(e.ID)
	r := Record{
		ID: e.ID,
		Data: map[string]string{
			"id":               idString,
			"name":             e.Name,
			"policy_number":    strconv.Itoa(e.PolicyNumber),
			"record_timestamp": strconv.Itoa(int(e.RecordTimestamp.Unix())),
		},
	}
	return r
}

// Fill in Insured fields from Record. Does not fill in Employees or Addresses.
func (e *Insured) FromRecord(r Record) (err error) {
	e.ID = r.ID
	e.Name = r.Data["name"]
	pn, err := strconv.Atoi(r.Data["policy_number"])
	if err != nil {
		return err
	}
	e.PolicyNumber = pn
	timestampInt, err := strconv.Atoi(r.Data["record_timestamp"])
	e.RecordTimestamp = time.Unix(int64(timestampInt), 0)
	return err
}

func InsuredsFromRecords(records map[int]Record) (map[int]Insured, error) {
	insuredes := make(map[int]Insured)
	for i, e := range records {
		id := i
		insured := Insured{}
		err := insured.FromRecord(e)
		if err != nil {
			return map[int]Insured{}, err
		}
		insuredes[id] = insured
	}
	return insuredes, nil
}

func (i Insured) MarshalJSON() ([]byte, error) {
	fmt.Println("MARSHHHHHHHHHHHHHHHHHHH")	
	if i.Employees == nil {
		e := make(map[int]Employee)
		e[0] = Employee{}
		i.Employees = &e
	}
	if i.Addresses == nil {
		a := make(map[int]Address)
		a[0] = Address{}
		i.Addresses = &a
	}
	return json.Marshal(&struct {
		ID              string           `json:"id"`
		Name            string           `json:"name"`
		PolicyNumber    string           `json:"policy_number"`
		RecordTimestamp string           `json:"recordTimestamp"`
		RecordDateTime  string           `json:"recordDateTime"`
		Employees       map[int]Employee `json:"employees"`
		Addresses       map[int]Address  `json:"insuredAddresses"`
	}{
		ID:              strconv.Itoa(i.ID),
		Name:            i.Name,
		PolicyNumber:    strconv.Itoa(i.PolicyNumber),
		RecordTimestamp: strconv.Itoa(int(i.RecordTimestamp.Unix())),
		RecordDateTime:  i.RecordTimestamp.Format("Mon, 02 Jan 2006 15:04:05 MST"),
		Employees:       *i.Employees,
		Addresses:       *i.Addresses,
	})
}
