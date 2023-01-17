package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nickcoast/timetravel/entity"
	"github.com/nickcoast/timetravel/sqlite"
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

func (s *SqliteRecordService) CreateRecord(ctx context.Context, resource string, record entity.Record) (newRecord entity.Record, err error) {
	id := record.ID
	if id != 0 {
		return newRecord, ErrRecordIDInvalid
	}
	timestamp := time.Now() // for all new record creation
	log.Println("SqliteRecordServce.CreateRecord record:", record)
	if resource == "insured" {
		return s.createInsured(ctx, timestamp, record)
	} else if resource == "employee" || resource == "employees" {
		return s.createEmployee(ctx, timestamp, record)
	} else if resource == "address" || resource == "insured_addresses" || resource == "addresses" {
		return s.createAddress(ctx, timestamp, record)
	}
	fmt.Println("You are here")

	if err != nil {
		return newRecord, err
		// TODO: May want to use this here later
		//return ErrRecordAlreadyExists
	}
	fmt.Println("here we are again")
	return newRecord, nil
}

func (s *SqliteRecordService) UpdateRecord(ctx context.Context, resource string, record entity.Record) (updateRecord entity.Record, err error) {
	id := updateRecord.ID
	if id != 0 {
		return updateRecord, ErrRecordIDInvalid
	}
	timestamp := time.Now() // for all new record creation
	log.Println("SqliteRecordServce.updateRecord record:", updateRecord)
	if resource == "insured" {
		return record, ErrRecordAlreadyExists // cannot update insured (name, policy id). Address and address data are updateable
	} else if resource == "address" || resource == "addresses" || resource == "insured_addresses" || resource == "insured_address" {
		return s.updateAddress(ctx, timestamp, record)
	} else if resource == "employee" || resource == "employees" {
		return s.updateEmployee(ctx, timestamp, record)
	}
	fmt.Println("here we are again")
	return updateRecord, ErrRecordAlreadyExists
}

func (s *SqliteRecordService) GetRecordById(ctx context.Context, resource string, id int) (entity.Record, error) {
	if id == 0 {
		return entity.Record{}, ErrRecordDoesNotExist
	}
	if resource == "employee" { // TODO: DELETE
		resource = "employees"
	}
	id64 := int64(id)
	e, err := s.service.Db.GetById(ctx, resource, id64)
	if err != nil || e.ID == 0 {
		return entity.Record{}, ErrRecordDoesNotExist
	}
	//fmt.Println("service.SqliteRecordService.GetRecordById: Name", e.Name, "id", e.ID, "pn", e.PolicyNumber, "rt", e.RecordTimestamp)
	fmt.Println("service.SqliteRecordService.GetById: Data", e.Data, "requested id", e.ID)

	return e, nil
}

func (s *SqliteRecordService) DeleteRecord(ctx context.Context, resource string, id int64) (record entity.Record, err error) {
	if id == 0 {
		return record, ErrRecordDoesNotExist
	}
	if resource == "employee" {
		resource = "employees"
	}
	if resource == "address" {
		resource = "insured_addresses"
	}
	record, err = s.service.Db.DeleteById(ctx, resource, id)
	if err != nil {
		return record, ErrRecordDoesNotExist
	}
	fmt.Print("Deleted ", resource, " with id:", id)
	return record, nil
}

func (s *SqliteRecordService) createInsured(ctx context.Context, timestamp time.Time, record entity.Record) (newRecord entity.Record, err error) {
	name := record.DataVal("name")
	fmt.Println("SqliteRecordService.CreateRecord name:", name)
	var insured *entity.Insured
	insured = &entity.Insured{}
	insured.Name = *name
	insured.RecordTimestamp = timestamp

	newRecord, err = s.service.CreateInsured(ctx, insured)
	if err != nil {
		return entity.Record{}, err
	}
	fmt.Println("here we are again")
	return newRecord, nil
}

func (s *SqliteRecordService) GetInsuredByDate(ctx context.Context, insuredId int64, dateValid time.Time) (entity.Insured, error) {
	return s.service.Db.GetInsuredByDate(ctx, insuredId, dateValid)
}

// TODO: use map for natural key, or struct
func (s *SqliteRecordService) GetRecordByDate(ctx context.Context, resource string, naturalKey string, insuredId int64, dateValid time.Time) (entity.Record, error) {
	if insuredId == 0 {
		return entity.Record{}, ErrRecordDoesNotExist
	}
	if resource == "employee" {
		resource = "employees"
	}
	e, err := s.service.Db.GetByDate(ctx, resource, naturalKey, insuredId, dateValid)
	if err != nil {
		return entity.Record{}, ErrRecordDoesNotExist
	}
	fmt.Println("service.SqliteRecordService.GetRecordByDate: records", e)
	for _, r := range e {
		return r, nil
	}
	return entity.Record{}, nil
}
