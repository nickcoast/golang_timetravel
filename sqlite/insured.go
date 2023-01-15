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
func (s *InsuredService) FindInsuredByID(ctx context.Context, id int) (*entity.Insured, error) {
	fmt.Println("sqlite.InsuredService.FindInsuredById")
	tx, err := s.Db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	fmt.Println("InsuredService.FindInsuredByID id:", id)

	// Fetch insured
	insured, err := findInsuredByID(ctx, tx, id)
	if err != nil {
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
func (s *InsuredService) CreateInsured(ctx context.Context, insured *entity.Insured) (id int64, policyNumber int, err error) {
	tx, err := s.Db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	// Create a new insured object
	id, policyNumber, err = createInsured(ctx, tx, insured)
	if err != nil {
		return 0, 0, err
	}
	return id, policyNumber, tx.Commit()
}

func (s *InsuredService) CreateEmployee(ctx context.Context, employee *entity.Employee) (id int64, err error) {
	tx, err := s.Db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Create a new insured object
	id, err = createEmployee(ctx, tx, employee)
	fmt.Println("InsuredService.CreateEmployee id:", id)
	if err != nil {
		fmt.Println("asdf")
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		fmt.Println("jkl")
		return 0, err
	}
	return id, nil

}

// UpdateInsured updates a insured object. Returns ENOTFOUND if insured does not exist.
/* func (s *InsuredService) UpdateInsured(ctx context.Context, id int, upd entity.InsuredUpdate) (*entity.Insured, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Update insured
	insured, err := updateInsured(ctx, tx, id, upd)
	if err != nil {
		return insured, err
	} else if err := tx.Commit(); err != nil {
		return insured, err
	}
	return insured, nil
} */

// DeleteInsured permanently deletes a insured and all owned dials.
// Returns ENOTFOUND if insured does not exist.
/* func (s *InsuredService) DeleteInsured(ctx context.Context, id int) error {
	tx, err := s.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := deleteInsured(ctx, tx, id); err != nil {
		return err
	}
	return tx.Commit()
} */

func (s *InsuredService) DeleteEmployee(ctx context.Context, id int) error {
	return nil
}

// findInsuredByID is a helper function to fetch a insured by ID.
// Returns ENOTFOUND if insured does not exist.
func findInsuredByID(ctx context.Context, tx *Tx, id int) (*entity.Insured, error) {
	fmt.Println("sqlite.InsuredService findInsuredById")
	a, _, err := findInsureds(ctx, tx, entity.InsuredFilter{ID: &id})
	if err != nil {
		return nil, err
	} else if len(a) == 0 {
		return nil, &entity.Error{Code: entity.ENOTFOUND, Message: "Insured not found."}
	}
	return a[0], nil
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

// createInsured creates a new insured. Sets the new database ID to insured.ID and sets
// the timestamps to the current time.
func createInsured(ctx context.Context, tx *Tx, insured *entity.Insured) (id int64, policyNumber int, err error) {
	// Set timestamps to the current time.

	// Perform basic field validation.
	if err := insured.Validate(); err != nil {
		return 0, 0, err
	}
	policyNumber, err = getMaxPolicyNumber(ctx, tx)
	policyNumber++ // safe if table is locked in transaction. else need trigger in DB
	if err != nil {
		return 0, 0, FormatError(err)
	}

	// Execute insertion query. // TODO: implement "auto increment" for policy_number
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
	) // alternatively could try this on last line of INSERT. Don't know if deepEqual checks unset values: VALUES (?, ?, STRFTIME('%s'))
	if err != nil {
		return 0, 0, FormatError(err)
	}

	id, err = result.LastInsertId()
	if err != nil {
		return 0, 0, err
	}
	insured.ID = int(id)

	return id, policyNumber, nil
}

// createEmployee creates a new employee. Sets the new database ID to insured.ID and sets
// the timestamps to the current time.
func createEmployee(ctx context.Context, tx *Tx, employee *entity.Employee) (id int64, err error) {
	// Perform basic field validation.
	if err := employee.Validate(); err != nil {
		return 0, err
	}

	// Execute insertion query. // TODO: implement "auto increment" for policy_number
	//dateVal = (employee.EndDate < time.Parse("2006-01-02","1971-01-01") ? nil : employee.EndDate.Format("2006-01-02"))

	result, err := tx.ExecContext(ctx, `
		INSERT INTO employees (
			name,
			start_date,
			end_date,
			insured_id,		
			record_timestamp		
		)
		VALUES (?, ?, ?, ?, ?)
	`,
		employee.Name,
		employee.StartDate.Format("2006-01-02"),
		employee.EndDate.Format("2006-01-02"), //.Format("2006-01-02"),
		employee.InsuredId,
		employee.RecordTimestamp.Unix(), // can use a Scan method here if necessary
	) // alternatively could try this on last line of INSERT. Don't know if deepEqual checks unset values: VALUES (?, ?, STRFTIME('%s'))
	if err != nil {
		return 0, FormatError(err)
	}

	id, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}
	employee.ID = int(id)

	return id, nil
}

// updateInsured updates fields on a insured object.
/* func updateInsured(ctx context.Context, tx *Tx, id int, upd entity.InsuredUpdate) (*entity.Insured, error) {
	// Fetch current object state.
	insured, err := findInsuredByID(ctx, tx, id)
	if err != nil {
		return insured, err
	}

	// Update fields.
	if v := upd.Name; v != nil {
		insured.Name = *v
	}
	if v := upd.Email; v != nil {
		insured.Email = *v
	}

	// Set last updated date to current time.
	insured.UpdatedAt = tx.now

	// Perform basic field validation.
	if err := insured.Validate(); err != nil {
		return insured, err
	}

	// Email is nullable and has a UNIQUE constraint so ensure we store blank
	// fields as NULLs.
	var email *string
	if insured.Email != "" {
		email = &insured.Email
	}

	// Execute update query.
	if _, err := tx.ExecContext(ctx, `
		UPDATE insureds
		SET name = ?,
		    email = ?,
		    updated_at = ?
		WHERE id = ?
	`,
		insured.Name,
		email,
		(*NullTime)(&insured.UpdatedAt),
		id,
	); err != nil {
		return insured, FormatError(err)
	}

	return insured, nil
} */

// deleteInsured permanently removes a insured by ID.
func deleteInsured(ctx context.Context, tx *Tx, id int) error {
	// Verify object exists.
	if _, err := findInsuredByID(ctx, tx, id); err != nil {
		return err
	}

	// Remove row from database.
	if _, err := tx.ExecContext(ctx, `DELETE FROM insured WHERE id = ?`, id); err != nil {
		return FormatError(err)
	}
	return nil
}

// private helper to help insert policy numbers in order
func getMaxPolicyNumber(ctx context.Context, tx *Tx) (max int, err error) {
	tx.QueryRowContext(ctx, `
		SELECT MAX(policy_number) AS max_policy_number 		
		FROM insured		
		ORDER BY id ASC`,
	).Scan(&max)

	if max == 0 {
		return 0, fmt.Errorf("Failed to retrieve max policy number")
	}
	return max, nil

}
