package service

import (
	"context"
	"strings"
	//"errors"

	"github.com/nickcoast/timetravel/sqlite"

	"github.com/temelpa/timetravel/entity"
)

// InMemoryRecordService is an in-memory implementation of RecordService.
type SqliteRecordService struct {
	data map[int]entity.Record
}

func NewSqliteRecordService() SqliteRecordService {
	return SqliteRecordService{
		data: map[int]entity.Record{},
	}
}

func (s *SqliteRecordService) GetRecordById(ctx context.Context, id int) (entity.Record, error) {
	record := s.data[id]
	if record.ID == 0 {
		return entity.Record{}, ErrRecordDoesNotExist
	}

	record = record.Copy() // copy is necessary so modifations to the record don't change the stored record
	return record, nil
}

func (s *SqliteRecordService) CreateRecord(ctx context.Context, record entity.Record) error {
	id := record.ID
	if id <= 0 {
		return ErrRecordIDInvalid
	}

	existingRecord := s.data[id] // don't need this if relying on auto-increment primary key
	if existingRecord.ID != 0 {
		return ErrRecordAlreadyExists
	}

	s.data[id] = record // creation
	return nil
}

func (s *SqliteRecordService) UpdateRecord(ctx context.Context, id int, updates map[string]*string) (entity.Record, error) {
	entry := s.data[id]
	if entry.ID == 0 {
		return entity.Record{}, ErrRecordDoesNotExist
	}

	for key, value := range updates {
		if value == nil { // deletion update
			delete(entry.Data, key)
		} else {
			entry.Data[key] = *value
		}
	}

	return entry.Copy(), nil
}

// findInsureds returns a list of insureds matching a filter. Also returns a count of
// total matching insureds which may differ if filter.Limit is set.
func findInsureds(ctx context.Context, tx *sqlite.Tx, filter entity.InsuredFilter) (_ []*entity.Insured, n int, err error) {
	// Build WHERE clause.
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := filter.Name; v != nil {
		where, args = append(where, "name = ?"), append(args, *v)
	}
	if v := filter.PolicyNumber; v != nil {
		where, args = append(where, "policy_number = ?"), append(args, *v)
	}
	if v := filter.RecordTimestamp; v != nil {
		where, args = append(where, "record_timestamp = ?"), append(args, *v)
	}

	// Execute query to fetch insured rows.
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
		`+sqlite.FormatLimitOffset(filter.Limit, filter.Offset),
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
			(*sqlite.NullTime)(&insured.RecordTimestamp), // TODO: get NullTime to work
			
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