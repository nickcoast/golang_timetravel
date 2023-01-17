package sqlite

// TODO: should this be in package "service" stead "sqlite"??

import (
	"context"
	"fmt"

	/* "database/sql" */
	"strings"

	"github.com/nickcoast/timetravel/entity"
)

// Ensure service implements interface.
var _ entity.InsuredService = (*InsuredService)(nil)

// InsuredService represents a service for managing insureds.
type InsuredService struct {
	Db *DB
}

// NewInsuredService returns a new instance of InsuredService.
func NewInsuredService(db *DB) *InsuredService {
	return &InsuredService{Db: db}
}

// FindInsuredByID retrieves a insured by ID
// Returns ENOTFOUND if insured does not exist.
func (s *InsuredService) FindInsuredByID(ctx context.Context, id int) (insured *entity.Insured, err error) {
	fmt.Println("sqlite.InsuredService.FindInsuredById")
	tx, err := s.Db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	fmt.Println("InsuredService.FindInsuredByID id:", id)

	// Fetch insured
	record, err := s.Db.GetById(ctx, "insured", int64(id))
	if err != nil {
		insured.FromRecord(record)
		return insured, err
	}
	return insured, nil
}

// FindInsureds retrieves a list of insureds by filter. Also returns total count of
// matching insureds which may differ from returned results if filter.Limit is specified.
func (s *InsuredService) FindInsureds(ctx context.Context, filter entity.InsuredFilter) ([]*entity.Insured, int, error) {
	tx, err := s.Db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()
	return findInsureds(ctx, tx, filter)
}

// CreateInsured creates a new insured.
// Used by CreateRecord inside conditional
func (s *InsuredService) CreateInsured(ctx context.Context, insured *entity.Insured) (record entity.Record, err error) {
	tx, err := s.Db.BeginTx(ctx, nil)
	if err != nil {
		return record, err
	}
	defer tx.Rollback()

	// Create a new insured object
	record, err = createInsured(ctx, tx, insured)
	if err != nil {
		return record, err
	}
	return record, tx.Commit()
}



// findInsureds returns a list of insureds matching a filter. Also returns a count of
// total matching insureds which may differ if filter.Limit is set.
func findInsureds(ctx context.Context, tx *Tx, filter entity.InsuredFilter) (_ []*entity.Insured, n int, err error) {
	// Build WHERE clause.
	where, args := []string{"1 = 1"}, []interface{}{}
	// TODO: can we consolidate this?
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := filter.PolicyNumber; v != nil {
		where, args = append(where, "policy_number = ?"), append(args, *v)
	}
	if v := filter.RecordTimestamp; v != nil {
		where, args = append(where, "record_timestamp < ?"), append(args, *v)
	}
	if v := filter.Name; v != nil {
		where, args = append(where, "name = ?"), append(args, *v)
	}
	fmt.Println("sqlite.InsuredService findInsureds")
	// Execute query to fetch insured rows.
	// integer timestamp, or even date string, cannot be stored in Go type time.Time
	// because sqlite has no DATETIME type.
	// doesn't work: datetime(record_timestamp, 'unixepoch' /*, 'localtime' */) as record_timestamp,
	// only solution seems to be to switch from time.Time to integer and then convert to datetime in Go
	rows, err := tx.QueryContext(ctx, `
		SELECT 
		    id,
		    name,
		  	policy_number,			
			record_timestamp,
		    COUNT(*) OVER()
		FROM insured
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY id ASC
		`+FormatLimitOffset(filter.Limit, filter.Offset),
		args...,
	)
	if err != nil {
		return nil, n, err
	}
	defer rows.Close()

	// Deserialize rows into Insured objects.
	insureds := make([]*entity.Insured, 0)
	i := 0
	for rows.Next() {
		var insured entity.Insured
		if err := rows.Scan(
			&insured.ID,
			&insured.Name,
			&insured.PolicyNumber,
			(*NullTime)(&insured.RecordTimestamp), // TODO: check this
			&n,
		); err != nil {
			return nil, 0, err
		}

		insureds = append(insureds, &insured)
		i++
	}
	if i == 0 {
		return nil, 0, ErrRecordMatchingCriteriaDoesNotExist
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return insureds, n, nil
}

// creates a new insured. Sets the new record ID to insured.ID and retrieves new policyNumber
func createInsured(ctx context.Context, tx *Tx, insured *entity.Insured) (newRecord entity.Record, err error) {
	// Perform basic field validation.
	if err := insured.Validate(); err != nil {
		return newRecord, err
	}
	policyNumber, err := getMaxPolicyNumber(ctx, tx)
	policyNumber++ // safe if table is locked in transaction. else need trigger in DB
	if err != nil {
		return newRecord, FormatError(err)
	}
	insured.PolicyNumber = policyNumber
	result, err := tx.ExecContext(ctx, `
		INSERT INTO insured (
			name,
			policy_number,			
			record_timestamp		
		)
		VALUES (?, ?, ?)
	`,
		insured.Name,
		policyNumber,
		insured.RecordTimestamp.Unix(), // can use a Scan method here if necessary
	)
	if err != nil {
		return newRecord, FormatError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return newRecord, err
	}
	// TODO: try to set newRecord using DB.GetById
	insured.ID = int(id)
	newRecord = insured.ToRecord()

	return newRecord, nil
}

// private helper to help insert policy numbers in order
func getMaxPolicyNumber(ctx context.Context, tx *Tx) (max int, err error) {
	// coalesce ensures '1000' is returned if no data exists in table
	tx.QueryRowContext(ctx, `
		SELECT coalesce(MAX(policy_number), 1000) AS max_policy_number 		
		FROM insured		
		ORDER BY id ASC`,
	).Scan(&max)

	if max == 0 {
		return 0, fmt.Errorf("Failed to retrieve max policy number")
	}
	return max, nil

}

