package sqlite

import (
	"context"
	"time"

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
	table := address.GetDataTableName()
	query := `
	INSERT INTO ` + table + ` (` + "\n" +
		`	address,` + "\n" +
		`	insured_id,` + "\n" +
		`	record_timestamp` + "\n" +
		`)` + "\n" +
		`VALUES (?, ?, ?)`	
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
	//entity.Address.GetIdentTableName
	address := entity.Address{}
	table := address.GetIdentTableName()
	query := `
	SELECT COUNT(*) FROM ` + table + "\n" +
		`WHERE insured_id = ?
`
	result, err := tx.QueryContext(ctx, query, insured.ID)
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

func (s *InsuredService) UpdateAddress(ctx context.Context, address *entity.Address) (record entity.Record, err error) {
	tx, err := s.Db.BeginTx(ctx, nil)
	if err != nil {
		return record, err
	}
	defer tx.Rollback()

	insured := entity.Insured{}
	date := time.Now()
	insured, err = s.Db.GetInsuredByDate(ctx, int64(address.InsuredId), date)

	count, err := s.CountInsuredAddresses(ctx, insured)
	if count == 0 {
		return record, ErrRecordAlreadyExists
	}

	currentAddresses := *insured.Addresses
	currentAddress := currentAddresses[0]
	//existingAddress, err := s.Db.GetAddressById(ctx, *address, int64(address.ID))

	if address.Address == currentAddress.Address {
		return record, ErrUpdateMustChangeAValue
	}

	// Create a new address record
	record, err = createAddress(ctx, tx, address)
	if err != nil {
		return record, err
	}
	return record, tx.Commit()
}
