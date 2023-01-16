package entity

import (
	"context"
	"strconv"
	"time"
)

// Insured represents a insured in the system.
// insureds can also be created directly for testing.
type Insured struct {
	ID int `json:"id"`

	// Insured's preferred name & email.
	Name string `json:"name"`

	PolicyNumber int `json:"policyNumber"`

	// Timestamps for insured creation & last update.
	RecordTimestamp time.Time `json:"recordTimestamp"`
}

// Validate returns an error if the insured contains invalid fields.
// This only performs basic validation.
func (u *Insured) Validate() error {
	if u.Name == "" {
		return Errorf(EINVALID, "Insured name required.")
	}
	return nil
}

// InsuredService represents a service for managing insureds.
type InsuredService interface {
	// Retrieves a insured by ID
	// Returns ENOTFOUND if insured does not exist.
	FindInsuredByID(ctx context.Context, id int) (*Insured, error)

	// Retrieves a list of insureds by filter. Also returns total count of matching
	// insureds which may differ from returned results if filter.Limit is specified.
	FindInsureds(ctx context.Context, filter InsuredFilter) ([]*Insured, int, error)

	// Creates a new insured.
	CreateInsured(ctx context.Context, insured *Insured) (Record, error)

	// Updates a insured object. Returns ENOTFOUND if insured does not exist.
	// REMOVED from interface. Will not support updates to the core table for now
	/* UpdateInsured(ctx context.Context, id int, upd InsuredUpdate) (*Insured, error) */

	// Permanently deletes a insured and all owned dials. Returns ENOTFOUND if
	// insured does not exist.
	// removed in favor of DB method
	//DeleteInsured(ctx context.Context, id int) error
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
