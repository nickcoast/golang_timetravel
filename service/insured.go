package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	//"strings"
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

func (s *SqliteRecordService) GetRecordById(ctx context.Context, resource string, id int) (entity.Record, error) {
	if id == 0 {
		return entity.Record{}, ErrRecordDoesNotExist
	}
	if resource == "employee" {
		resource = "employees"
	}
	id64 := int64(id)
	e, err := s.service.Db.GetById(ctx, resource, id64)
	if err != nil {
		return entity.Record{}, ErrRecordDoesNotExist
	}
	//fmt.Println("service.SqliteRecordService.GetRecordById: Name", e.Name, "id", e.ID, "pn", e.PolicyNumber, "rt", e.RecordTimestamp)
	fmt.Println("service.SqliteRecordService.GetById: Data", e.Data, "requested id", e.ID)

	return e, nil
}

func createEntity(record entity.Record, entityType string) (interface{}, error) {
	if entityType == "employees" {

	} else if entityType == "insured" {
		return entity.Insured{}, nil
	}

	return entity.Insured{}, fmt.Errorf("Sorry")

}

func (s *SqliteRecordService) CreateRecord(ctx context.Context, resource string, record entity.Record) (newRecord entity.Record, err error) {
	id := record.ID
	if id != 0 {
		return newRecord, ErrRecordIDInvalid
	}

	name := record.DataVal("name")
	fmt.Println("SqliteRecordService.CreateRecord name:", name)
	if name == nil {
		return newRecord, errors.New("Name is required")
	}
	timestamp := time.Now()

	//resource, err := record.DataVal("type")
	if err != nil {
		return newRecord, err
	}
	if resource == "insured" {
		var insured *entity.Insured
		insured = &entity.Insured{}
		insured.Name = *name
		insured.RecordTimestamp = timestamp
		newRecord, err = s.service.CreateInsured(ctx, insured)
		fmt.Println("service.SqliteRecordService in insured.go created new policy number: ",
			newRecord.Data["policy_number"], "with id:", newRecord.ID, "for:", name)
		if err != nil {
			return entity.Record{}, err
			// May want to use this later
			//return ErrRecordAlreadyExists
		}
	} else if resource == "employee" || resource == "employees" { // TODO: this is gross. gotta break it out into other functions
		var employee *entity.Employee
		employee = &entity.Employee{}
		employee.Name = *name
		employee.RecordTimestamp = timestamp

		fmt.Println("The Record (employee), so far:", newRecord)
		//newRecord, err = s.service.CreateEmployee(ctx, insured)
		if ii := record.DataVal("insuredId"); ii != nil { // SET INSURED ID
			if employee.InsuredId, err = strconv.Atoi(*ii); err != nil {
				fmt.Println("Error converting to int:", err)
				return newRecord, fmt.Errorf("Problem converting string to int")
			}
		} else {
			fmt.Println("ii", ii)
			return newRecord, fmt.Errorf("Insured ID required to create Employee: %v", err)
		}
		sd := record.DataVal("startDate")
		if sd == nil || len(*sd) != 10 {
			return newRecord, fmt.Errorf("startDate required to create Employee: %v, %v, %v", err, "sd: ", *sd)
		}
		employee.StartDate, err = time.Parse("2006-01-02", *sd)
		fmt.Println("employee SO FAR AGAIN!!!!!!!!: ", employee)
		ed := record.DataVal("endDate")
		if ed == nil || len(*ed) != 10 {
			fmt.Println("Bad endDate or error. End date: ", ed)
			return newRecord, fmt.Errorf("Bad endDate or error. End date: %v", ed)
		}
		t, err := time.Parse("2006-01-02", *ed)
		if err != nil {
			return newRecord, err
		}
		employee.EndDate = t

		fmt.Print("employee.*******************:", employee)
		newRecord, err := s.service.CreateEmployee(ctx, employee)
		if err != nil {
			return entity.Record{}, err
		}
		return newRecord, err
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
	fmt.Println("service.SqliteRecordService.GetRecordByDate: Data", e.Data, "requested id", e.ID)

	return e, nil
}

func (s *SqliteRecordService) DeleteRecord(ctx context.Context, resource string, id int64) (record entity.Record, err error) {
	if id == 0 {
		return record, ErrRecordDoesNotExist
	}
	if resource == "employee" {
		resource = "employees"
	}
	record, err = s.service.Db.DeleteById(ctx, resource, id)
	if err != nil {
		return record, ErrRecordDoesNotExist
	}
	fmt.Print("Deleted ", resource, " with id:", id)
	return record, nil
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
