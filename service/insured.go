package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	//"errors"

	"github.com/nickcoast/timetravel/sqlite"

	"github.com/nickcoast/timetravel/entity"
)

// InMemoryRecordService is an in-memory implementation of RecordService.
type SqliteRecordService struct {
	data    map[int]entity.Record
	service sqlite.InsuredService
}

func NewSqliteRecordService() SqliteRecordService {
	return SqliteRecordService{
		data: map[int]entity.Record{},
	}
}

func (s *SqliteRecordService) SetService(db *sqlite.DB) {
	s.service = *sqlite.NewInsuredService(db)
}

func (s *SqliteRecordService) GetRecordById(ctx context.Context, id int) (entity.Record, error) {
	e, err := s.service.FindInsuredByID(ctx, id)
	if err != nil {
		return entity.Record{}, ErrRecordDoesNotExist
	}
	fmt.Println("service.SqliteRecordService.GetRecordById: Name", e.Name, "id", e.ID, "pn", e.PolicyNumber, "rt", e.RecordTimestamp)

	// exclude the delete updates
	recordMap := map[string]string{}
	recordMap["id"] = fmt.Sprintf("%d", e.ID)
	recordMap["name"] = e.Name
	recordMap["policyNumber"] = fmt.Sprintf("%d", e.PolicyNumber)
	recordMap["recordTimestamp"] = fmt.Sprintf("%s", e.RecordTimestamp)

	record := entity.Record{
		ID:   int(e.ID),
		Data: recordMap,
	}

	if record.ID == 0 {
		return entity.Record{}, ErrRecordDoesNotExist
	}

	return record, nil
}

func (s *SqliteRecordService) CreateRecord(ctx context.Context, record entity.Record) (err error) {
	id := record.ID
	if id != 0 {
		return ErrRecordIDInvalid
	}

	name, err := record.DataVal("name")
	fmt.Println("SqliteRecordService.CreateRecord name:", name)
	if err != nil {
		return err
	}

	var insured *entity.Insured
	insured = &entity.Insured{}
	insured.Name = name
	insured.RecordTimestamp = time.Now().UTC()
	fmt.Println("You are here")
	newId, policyNumber, err := s.service.CreateInsured(ctx, insured)
	fmt.Println("service.SqliteRecordService in insured.go created new policy number: ", policyNumber, "with id:", newId, "for:", name)
	if err != nil {
		return err
		// May want to use this later
		//return ErrRecordAlreadyExists
	}

	s.data[id] = record // creation
	return nil
}

func (s *SqliteRecordService) DeleteRecord(ctx context.Context, id int) error {
	return s.service.DeleteInsured(ctx, id)
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
	fmt.Println("service.SqliteRecordService findInsureds")
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
