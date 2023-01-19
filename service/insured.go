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
type InsuredRecordService struct {
	data      map[int]entity.Record
	dbService sqlite.InsuredDBService
}

func NewSqliteRecordService() InsuredRecordService {
	return InsuredRecordService{
		data: map[int]entity.Record{},
	}
}

func (s *InsuredRecordService) SetService(db *sqlite.DB) {
	s.dbService = *sqlite.NewInsuredService(db)
}

func (s *InsuredRecordService) CreateRecord(ctx context.Context, resource string, record entity.Record) (newRecord entity.Record, err error) {
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

func (s *InsuredRecordService) UpdateRecord(ctx context.Context, resource string, record entity.Record) (updateRecord entity.Record, err error) {
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

func (s *InsuredRecordService) GetRecordById(ctx context.Context, resource string, id int) (entity.Record, error) {
	if id == 0 {
		return entity.Record{}, ErrRecordDoesNotExist
	}
	if resource == "employee" { // TODO: DELETE
		resource = "employees"
	}
	id64 := int64(id)
	e, err := s.dbService.Db.GetById(ctx, resource, id64)
	if err != nil || e.ID == 0 {
		return entity.Record{}, ErrRecordDoesNotExist
	}
	ed := e.DataVal("end_date") // check "empty" end date
	if ed != "" && ed == "0001-01-01" {
		if ed == "0001-01-01" {
			e.Data["end_date"] = "None"
		}
	}
	fmt.Println("service.SqliteRecordService.GetById: Data", e.Data, "requested id", e.ID)

	return e, nil
}

func (s *InsuredRecordService) DeleteRecord(ctx context.Context, resource string, id int64) (record entity.Record, err error) {
	if id == 0 {
		return record, ErrRecordDoesNotExist
	}
	if resource == "employee" {
		resource = "employees"
	}
	if resource == "address" {
		resource = "insured_addresses"
	}
	record, err = s.dbService.Db.DeleteById(ctx, resource, id)
	if err != nil {
		return record, ErrRecordDoesNotExist
	}
	fmt.Print("Deleted ", resource, " with id:", id)
	return record, nil
}

func (s *InsuredRecordService) createInsured(ctx context.Context, timestamp time.Time, record entity.Record) (newRecord entity.Record, err error) {
	name := record.DataVal("name")
	fmt.Println("SqliteRecordService.CreateRecord name:", name)
	var insured *entity.Insured
	insured = &entity.Insured{}
	insured.Name = name
	insured.RecordTimestamp = timestamp

	newRecord, err = s.dbService.CreateInsured(ctx, insured)
	if err != nil {
		return entity.Record{}, err
	}
	fmt.Println("here we are again")
	return newRecord, nil
}

func (s *InsuredRecordService) GetInsuredByDate(ctx context.Context, insuredId int64, dateValid time.Time) (entity.Insured, error) {
	return s.dbService.Db.GetInsuredByDate(ctx, insuredId, dateValid)
}

// TODO: use map for natural key, or struct
func (s *InsuredRecordService) GetRecordByDate(ctx context.Context, resource string, naturalKey string, insuredId int64, dateValid time.Time) (entity.Record, error) {
	if insuredId == 0 {
		return entity.Record{}, ErrNonexistentParentRecord
	}
	if resource == "employee" {
		resource = "employees"
	}
	record, err := s.dbService.Db.GetByDate(ctx, resource, naturalKey, insuredId, dateValid)
	if err != nil {
		return entity.Record{}, ErrServerError
	}
	fmt.Println("service.SqliteRecordService.GetRecordByDate: records", record)
	for _, r := range record {
		return r, nil
	}
	return entity.Record{}, nil
}
