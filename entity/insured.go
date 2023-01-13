package entity

import (
	"context"
	"time"
	
)

// Insured represents a insured in the system.
// insureds can also be created directly for testing.
type Insured struct {
	ID int `json:"id"`

	// Insured's preferred name & email.
	Name string `json:"name"`

	PolicyNumber int `json:"policyNumber"`

	// Randomly generated API key for use with the CLI.
	/* APIKey string `json:"-"` */

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
	// Retrieves a insured by ID along with their associated auth objects.
	// Returns ENOTFOUND if insured does not exist.
	FindInsuredByID(ctx context.Context, id int) (*Insured, error)

	// Retrieves a list of insureds by filter. Also returns total count of matching
	// insureds which may differ from returned results if filter.Limit is specified.
	FindInsureds(ctx context.Context, filter InsuredFilter) ([]*Insured, int, error)

	// Creates a new insured.
	CreateInsured(ctx context.Context, insured *Insured) (int64, int, error)

	// Updates a insured object. Returns ENOTFOUND if insured does not exist.
	// REMOVED from interface. Will not support updates to the core table for now
	/* UpdateInsured(ctx context.Context, id int, upd InsuredUpdate) (*Insured, error) */

	// Permanently deletes a insured and all owned dials. Returns EUNAUTHORIZED
	// if current insured is not the insured being deleted. Returns ENOTFOUND if
	// insured does not exist.
	DeleteInsured(ctx context.Context, id int) error
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
