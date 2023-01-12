package entity

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"	
	"io"
	"strings"

	//"github.com/benbjohnson/wtf"
)



// Ensure service implements interface.
//var _ wtf.InsuredService = (*InsuredService)(nil)

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
func (s *InsuredService) FindInsuredByID(ctx context.Context, id int) (*wtf.Insured, error) {
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
func (s *InsuredService) FindInsureds(ctx context.Context, filter wtf.InsuredFilter) ([]*wtf.Insured, int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()
	return findInsureds(ctx, tx, filter)
}

// CreateInsured creates a new insured. This is only used for testing since insureds are
// typically created during the OAuth creation process in AuthService.CreateAuth().
func (s *InsuredService) CreateInsured(ctx context.Context, insured *wtf.Insured) error {
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
func (s *InsuredService) UpdateInsured(ctx context.Context, id int, upd wtf.InsuredUpdate) (*wtf.Insured, error) {
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
}

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
func findInsuredByID(ctx context.Context, tx *Tx, id int) (*wtf.Insured, error) {
	a, _, err := findInsureds(ctx, tx, wtf.InsuredFilter{ID: &id})
	if err != nil {
		return nil, err
	} else if len(a) == 0 {
		return nil, &wtf.Error{Code: wtf.ENOTFOUND, Message: "Insured not found."}
	}
	return a[0], nil
}



// findInsureds returns a list of insureds matching a filter. Also returns a count of
// total matching insureds which may differ if filter.Limit is set.
func findInsureds(ctx context.Context, tx *Tx, filter wtf.InsuredFilter) (_ []*wtf.Insured, n int, err error) {
	// Build WHERE clause.
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := filter.Email; v != nil {
		where, args = append(where, "email = ?"), append(args, *v)
	}
	if v := filter.APIKey; v != nil {
		where, args = append(where, "api_key = ?"), append(args, *v)
	}

	// Execute query to fetch insured rows.
	rows, err := tx.QueryContext(ctx, `
		SELECT 
		    id,
		    name,
		    email,
		    api_key,
		    created_at,
		    updated_at,
		    COUNT(*) OVER()
		FROM insureds
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
	insureds := make([]*wtf.Insured, 0)
	for rows.Next() {
		var email sql.NullString
		var insured wtf.Insured
		if err := rows.Scan(
			&insured.ID,
			&insured.Name,
			&email,
			&insured.APIKey,
			(*NullTime)(&insured.CreatedAt),
			(*NullTime)(&insured.UpdatedAt),
			&n,
		); err != nil {
			return nil, 0, err
		}

		if email.Valid {
			insured.Email = email.String
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
func createInsured(ctx context.Context, tx *Tx, insured *wtf.Insured) error {
	// Set timestamps to the current time.
	insured.CreatedAt = tx.now
	insured.UpdatedAt = insured.CreatedAt

	// Perform basic field validation.
	if err := insured.Validate(); err != nil {
		return err
	}

	// Email is nullable and has a UNIQUE constraint so ensure we store blank
	// fields as NULLs.
	var email *string
	if insured.Email != "" {
		email = &insured.Email
	}

	// Generate random API key.
	apiKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, apiKey); err != nil {
		return err
	}
	insured.APIKey = hex.EncodeToString(apiKey)

	// Execute insertion query.
	result, err := tx.ExecContext(ctx, `
		INSERT INTO insureds (
			name,
			email,
			api_key,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?)
	`,
		insured.Name,
		email,
		insured.APIKey,
		(*NullTime)(&insured.CreatedAt),
		(*NullTime)(&insured.UpdatedAt),
	)
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
func updateInsured(ctx context.Context, tx *Tx, id int, upd wtf.InsuredUpdate) (*wtf.Insured, error) {
	// Fetch current object state.
	insured, err := findInsuredByID(ctx, tx, id)
	if err != nil {
		return insured, err
	} else if insured.ID != wtf.InsuredIDFromContext(ctx) {
		return nil, wtf.Errorf(wtf.EUNAUTHORIZED, "You are not allowed to update this insured.")
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
}

// deleteInsured permanently removes a insured by ID. Returns EUNAUTHORIZED if current
// insured is not the one being deleted.
func deleteInsured(ctx context.Context, tx *Tx, id int) error {
	// Verify object exists.
	if insured, err := findInsuredByID(ctx, tx, id); err != nil {
		return err
	} else if insured.ID != wtf.InsuredIDFromContext(ctx) {
		return wtf.Errorf(wtf.EUNAUTHORIZED, "You are not allowed to delete this insured.")
	}

	// Remove row from database.
	if _, err := tx.ExecContext(ctx, `DELETE FROM insureds WHERE id = ?`, id); err != nil {
		return FormatError(err)
	}
	return nil
}