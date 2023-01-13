package sqlite

import (
	"context"
	/* "database/sql" */
	"strings"

	"github.com/nickcoast/timetravel/entity"
)

// Ensure service implements interface.
var _ entity.InsuredService = (*InsuredService)(nil)

// InsuredService represents a service for managing insureds.
type InsuredService struct {
	db *DB
}

// NewInsuredService returns a new instance of InsuredService.
func NewInsuredService(db *DB) *InsuredService {
	return &InsuredService{db: db}
}

// FindInsuredByID retrieves a insured by ID along with their associated auth objects.
// Returns ENOTFOUND if insured does not exist.
func (s *InsuredService) FindInsuredByID(ctx context.Context, id int) (*entity.Insured, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Fetch insured and their associated OAuth objects.
	insured, err := findInsuredByID(ctx, tx, id)
	if err != nil {
		return insured, err
	}
	return insured, nil
}

// FindInsureds retrieves a list of insureds by filter. Also returns total count of
// matching insureds which may differ from returned results if filter.Limit is specified.
func (s *InsuredService) FindInsureds(ctx context.Context, filter entity.InsuredFilter) ([]*entity.Insured, int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()
	return findInsureds(ctx, tx, filter)
}

// CreateInsured creates a new insured. This is only used for testing since insureds are
// typically created during the OAuth creation process in AuthService.CreateAuth().
func (s *InsuredService) CreateInsured(ctx context.Context, insured *entity.Insured) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create a new insured object and attach associated OAuth objects.
	if err := createInsured(ctx, tx, insured); err != nil {
		return err
	}
	return tx.Commit()
}

// UpdateInsured updates a insured object. Returns EUNAUTHORIZED if current insured is
// not the insured that is being updated. Returns ENOTFOUND if insured does not exist.
/* func (s *InsuredService) UpdateInsured(ctx context.Context, id int, upd entity.InsuredUpdate) (*entity.Insured, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Update insured & attach associated OAuth objects.
	insured, err := updateInsured(ctx, tx, id, upd)
	if err != nil {
		return insured, err
	} else if err := tx.Commit(); err != nil {
		return insured, err
	}
	return insured, nil
} */

// DeleteInsured permanently deletes a insured and all owned dials.
// Returns EUNAUTHORIZED if current insured is not the insured being deleted.
// Returns ENOTFOUND if insured does not exist.
func (s *InsuredService) DeleteInsured(ctx context.Context, id int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := deleteInsured(ctx, tx, id); err != nil {
		return err
	}
	return tx.Commit()
}

// findInsuredByID is a helper function to fetch a insured by ID.
// Returns ENOTFOUND if insured does not exist.
func findInsuredByID(ctx context.Context, tx *Tx, id int) (*entity.Insured, error) {
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
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}

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
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return insureds, n, nil
}

// createInsured creates a new insured. Sets the new database ID to insured.ID and sets
// the timestamps to the current time.
func createInsured(ctx context.Context, tx *Tx, insured *entity.Insured) error {
	// Set timestamps to the current time.
	insured.RecordTimestamp = tx.now

	// Perform basic field validation.
	if err := insured.Validate(); err != nil {
		return err
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
		insured.PolicyNumber,
		insured.RecordTimestamp.Unix(), // can use a Scan method here if necessary
	) // alternatively could try this on last line of INSERT. Don't know if deepEqual checks unset values: VALUES (?, ?, STRFTIME('%s'))
	if err != nil {
		return FormatError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	insured.ID = int(id)

	return nil
}

// updateInsured updates fields on a insured object. Returns EUNAUTHORIZED if current
// insured is not the insured being updated.
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

// deleteInsured permanently removes a insured by ID. Returns EUNAUTHORIZED if current
// insured is not the one being deleted.
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
