package sqlite

import (
	"context"
	"fmt"
	"log"


	"github.com/nickcoast/timetravel/entity"
)

func (s *InsuredService) CreateAddress(ctx context.Context, address *entity.Address) (record entity.Record, err error) {
	tx, err := s.Db.BeginTx(ctx, nil)
	if err != nil {
		return record, err
	}
	defer tx.Rollback()

	// Create a new address record
	record, err = createAddress(ctx, tx, address)
	if err != nil {
		return record, err
	}
	return record, tx.Commit()
}

// creates a new address for insured
func createAddress(ctx context.Context, tx *Tx, address *entity.Address) (newRecord entity.Record, err error) {
	// Perform basic field validation.
	if err := address.Validate(); err != nil {
		return newRecord, err
	}
	query := `
	INSERT INTO insured_addresses (` + "\n" +
		`	address,` + "\n" +
		`	insured_id,` + "\n" +
		`	record_timestamp` + "\n" +
		`)` + "\n" +
		`VALUES (?, ?, ?)`
	fmt.Println("createAddress: ", address.Address, "insured_id:", address.InsuredId, "timestamp:", address.RecordTimestamp.Unix(), "\nQuery:\n", query)
	result, err := tx.ExecContext(ctx, query,
		address.Address,
		address.InsuredId,
		address.RecordTimestamp.Unix(), // can use a Scan method here if necessary
	)
	if err != nil {
		return newRecord, FormatError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return newRecord, err
	}
	// TODO: try to set newRecord using DB.GetById
	address.ID = int(id)
	newRecord = address.ToRecord()

	return newRecord, nil
}

func (s *InsuredService) CountInsuredAddresses(ctx context.Context, insured entity.Insured) (count int, err error) {
	tx, err := s.Db.BeginTx(ctx, nil)
	if err != nil {
		return count, err
	}
	defer tx.Rollback()
	if err := insured.Validate(); err != nil {
		return count, err
	}
	query := `
	SELECT COUNT(*) FROM insured_addresses
	WHERE insured_id = ?
`
	result, err := tx.QueryContext(ctx, query, insured.ID)
	log.Println("GetAddressesForInsured", insured, "Query: ", query)
	if err != nil {
		return count, err
	}
	defer result.Close()
	for result.Next() {
		if err := result.Scan(&count); err != nil {
			return count, err
		}
	}
	return count, nil
}
