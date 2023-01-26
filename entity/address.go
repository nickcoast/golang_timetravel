package entity

import (
	"context"
	"encoding/json"
	"strconv"
	"time"
)

// Address represents a address in the system.
// addresses can also be created directly for testing.
type Address struct {
	ID int `json:"id"`

	Address string `json:"address"`

	InsuredId int `json:"insuredId"`

	// Timestamps for address creation & last update.
	RecordTimestamp time.Time `json:"recordTimestamp"`
}

var _ InsuredInterface = (*Address)(nil)

func (u *Address) GetId() int64 {
	return int64(u.ID)
}
func (u *Address) GetInsuredId() int64 {
	return int64(u.InsuredId)
}
func (u *Address) GetDataTableName() string {
	return "insured_addresses_records"
}
func (u *Address) GetIdentTableName() string {
	return "insured_addresses_records"
}
func (u *Address) GetInsertFields() map[string]string {
	return map[string]string{
		"address": u.Address,
	}
}

// Validate returns an error if the address contains invalid fields.
// This only performs basic validation.
func (u *Address) Validate() error {
	if u.Address == "" {
		return Errorf(EINVALID, "Address required.")
	}
	if u.InsuredId < 1 {
		return Errorf(EINVALID, "Address must have an insured_id")
	}
	return nil
}

// AddressService represents a service for managing addresses.
type AddressService interface {
	// Retrieves a address by ID
	// Returns ENOTFOUND if address does not exist.
	FindAddressByID(ctx context.Context, id int) (*Address, error)

	// Retrieves a list of addresses by filter. Also returns total count of matching
	// addresses which may differ from returned results if filter.Limit is specified.
	FindAddresses(ctx context.Context, filter AddressFilter) ([]*Address, int, error)

	// Creates a new address.
	CreateAddress(ctx context.Context, address *Address) (int64, int, error)

	// Updates a address object. Returns ENOTFOUND if address does not exist.
	// REMOVED from interface. Will not support updates to the core table for now
	/* UpdateAddress(ctx context.Context, id int, upd AddressUpdate) (*Address, error) */

	// Permanently deletes a address and all owned dials. Returns ENOTFOUND if
	// address does not exist.
	DeleteAddress(ctx context.Context, id int) error
}

// AddressFilter represents a filter passed to FindAddresses().
type AddressFilter struct {
	// Filtering fields.
	ID              *int    `json:"id"`
	Address         *string `json:"address"`
	RecordTimestamp *int    `json:"recordTimestamp"`

	// Restrict to subset of results.
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// AddressUpdate represents a set of fields to be updated via UpdateAddress().
type AddressUpdate struct {
	Address string `json:"address"`
}

func (e *Address) ToRecord() Record {
	idString := strconv.Itoa(e.ID)
	r := Record{
		ID: e.ID,
		Data: map[string]string{
			"id":               idString,
			"address":          e.Address,
			"insuredId":       strconv.Itoa(e.InsuredId),
			"recordTimestamp": strconv.Itoa(int(e.RecordTimestamp.Unix())),
		},
	}
	return r
}

func (e *Address) FromRecord(r Record) (err error) {
	e.ID = r.ID
	e.Address = r.Data["address"]
	e.InsuredId, err = strconv.Atoi(r.Data["insured_id"])
	timestampInt, err := strconv.Atoi(r.Data["recordTimestamp"])
	e.RecordTimestamp = time.Unix(int64(timestampInt), 0)
	return err
}

func AddressesFromRecords(records map[int]Record) (map[int]Address, error) {
	addresses := make(map[int]Address)
	for i, e := range records {
		id := i
		address := Address{}
		err := address.FromRecord(e)
		if err != nil {
			return map[int]Address{}, err
		}
		addresses[id] = address
	}
	return addresses, nil
}

// Returns Address map. Skips any non-addresss
func AddressesFromInsuredInterface(insuredIfaceObjs map[int]InsuredInterface) (map[int]Address, error) {
	addresss := make(map[int]Address)
	for i, obj := range insuredIfaceObjs {
		e, ok := obj.(*Address)
		if ok {
			addresss[i] = *e
		}
	}
	return addresss, nil
}

func (a Address) MarshalJSON() ([]byte, error) {
	if a.ID == 0 {
		return json.Marshal(&struct {
			ID string `json:"id"`
		}{
			ID: "",
		})
	}
	return json.Marshal(&struct {
		ID              string `json:"id"`
		Address         string `json:"address"`
		RecordTimestamp string `json:"recordTimestamp"`
		RecordDateTime  string `json:"recordDateTime"`
	}{
		ID:              strconv.Itoa(a.ID),
		Address:         a.Address,
		RecordTimestamp: strconv.Itoa(int(a.RecordTimestamp.Unix())),
		RecordDateTime:  a.RecordTimestamp.Format("Mon, 02 Jan 2006 15:04:05 MST"),
	})
}
