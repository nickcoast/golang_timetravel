package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nickcoast/timetravel/entity"
)

var ErrRecordDoesNotExist = errors.New("record with that id does not exist")
var ErrRecordIDInvalid = errors.New("record id must >= 0")
var ErrRecordAlreadyExists = errors.New("record already exists")

// Implements method to get, create, and update record data.
type RecordService interface {

	// GetRecord will retrieve an record.
	GetRecordById(ctx context.Context, resource string, id int) (entity.Record, error) // TODO: change resource from string to Struct

	// CreateRecord will insert a new record.
	//
	// If it a record with that id already exists it will fail.
	CreateRecord(ctx context.Context, resource string, record entity.Record) (entity.Record, error)

	// UpdateRecord will change the internal `Map` values of the record if they exist.
	// if the update[key] is null it will delete that key from the record's Map.
	//
	// UpdateRecord will error if id <= 0 or the record does not exist with that id.
	UpdateRecord(ctx context.Context, resource string, record entity.Record) (entity.Record, error)

	DeleteRecord(ctx context.Context, resource string, id int64) (entity.Record, error)

	GetRecordByDate(ctx context.Context, resource string, naturalKey string, insuredId int64, date time.Time) (records entity.Record, err error)
	GetInsuredByDate(ctx context.Context, insuredId int64, date time.Time) (insured entity.Insured, err error)
}

// InMemoryRecordService is an in-memory implementation of RecordService.
type InMemoryRecordService struct {
	data map[int]entity.Record
}

func NewInMemoryRecordService() InMemoryRecordService {
	return InMemoryRecordService{
		data: map[int]entity.Record{},
	}
}

// no API route will lead here.
func (s *InMemoryRecordService) GetRecordByDate(ctx context.Context, resource string, naturalKey string, insuredId int64, date time.Time) (records entity.Record, err error) {
	return entity.Record{}, fmt.Errorf("Cannot currently get memory resource by date")
}

// no API route will lead here.
func (s *InMemoryRecordService) GetInsuredByDate(ctx context.Context, insuredId int64, date time.Time) (insured entity.Insured, err error) {
	return entity.Insured{}, fmt.Errorf("Cannot currently get memory resource by date")
}

func (s *InMemoryRecordService) DeleteRecord(ctx context.Context, resource string, id int64) (record entity.Record, err error) {
	record = s.data[int(id)]
	delete(s.data, int(id))
	// return deleted record, give user chance to "undo" if deleted by mistake
	return record, nil
}

func (s *InMemoryRecordService) GetRecordById(ctx context.Context, resource string, id int) (entity.Record, error) { // TODO: maybe change resource to Struct
	record := s.data[id]
	if record.ID == 0 {
		return entity.Record{}, ErrRecordDoesNotExist
	}

	record = record.Copy() // copy is necessary so modifations to the record don't change the stored record
	return record, nil
}

func (s *InMemoryRecordService) CreateRecord(ctx context.Context, resource string, record entity.Record) (newRecord entity.Record, err error) {
	id := record.ID
	if id <= 0 {
		return record, ErrRecordIDInvalid
	}

	existingRecord := s.data[id]
	if existingRecord.ID != 0 {
		return record, ErrRecordAlreadyExists
	}

	s.data[id] = record // creation
	newRecord = s.data[id]
	return newRecord, nil
}

func (s *InMemoryRecordService) UpdateRecord(ctx context.Context, resource string, record entity.Record) (entity.Record, error) {
	id := record.ID
	entry := s.data[id]
	if entry.ID == 0 {
		return entity.Record{}, ErrRecordDoesNotExist
	}

	for key, value := range record.Data {
		if value == "" { // deletion update
			delete(entry.Data, key)
		} else {
			entry.Data[key] = value
		}
	}
	return entry.Copy(), nil
}

type SqlRecordService struct { // TODO: delete
	data map[int]entity.Record
}

// TODO: delete
func (s *SqlRecordService) GetRecordById(ctx context.Context, resource string, id int) (entity.Record, error) { // ignore resource
	record := s.data[id]
	if record.ID == 0 {
		return entity.Record{}, ErrRecordDoesNotExist
	}
	fmt.Println("service.InMemeryRecordService just kidding SqlRecordService")
	record = record.Copy() // copy is necessary so modifations to the record don't change the stored record
	return record, nil
}
