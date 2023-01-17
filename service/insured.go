package service

import (
	"context"
	"fmt"
	"log"
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

// TODO: delete this
func createEntity(record entity.Record, entityType string) (interface{}, error) {
	if entityType == "employees" {

	} else if entityType == "insured" {
		return entity.Insured{}, fmt.Errorf("'insured' has its own Create method. Cannot use this one.")
	}

	return entity.Insured{}, fmt.Errorf("Sorry")

}

// TODO: are we skipping Create* methods now?
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

func (s *SqliteRecordService) createInsured(ctx context.Context, timestamp time.Time, record entity.Record) (newRecord entity.Record, err error) {
	name := record.DataVal("name")
	fmt.Println("SqliteRecordService.CreateRecord name:", name)
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
	fmt.Println("here we are again")
	return newRecord, nil
}
func (s *SqliteRecordService) createAddress(ctx context.Context, timestamp time.Time, record entity.Record) (newRecord entity.Record, err error) {
	addressString := record.DataVal("address")
	if addressString == nil {
		return entity.Record{}, ErrServerError
	}
	fmt.Println("SqliteRecordService.CreateRecord name:", addressString)
	var address *entity.Address
	address = &entity.Address{}
	address.Address = *addressString
	address.RecordTimestamp = timestamp

	if ii := record.DataVal("insuredId"); ii != nil { // SET INSURED ID
		if address.InsuredId, err = strconv.Atoi(*ii); err != nil {
			fmt.Println("Error converting to int:", err)
			return newRecord, fmt.Errorf("Problem converting string to int")
		}
	} else {
		fmt.Println("ii", ii)
		return newRecord, fmt.Errorf("Insured ID required to create Address: %v", err)
	}
	insuredRecord, err := s.GetRecordById(ctx, "insured", address.InsuredId)	
	if err != nil {
		return newRecord, ErrNonexistentParentRecord
	}
	insured := entity.Insured{}
	insured.FromRecord(insuredRecord)
	addressCount, err := s.service.CountInsuredAddresses(ctx, insured)
	if err != nil {
		return newRecord, ErrServerError
	}	
	if addressCount > 0 {
		return newRecord, ErrRecordAlreadyExists
	}
	newRecord, err = s.service.CreateAddress(ctx, address)
	if err != nil {
		return entity.Record{}, err
	}
	return newRecord, err
}

func (s *SqliteRecordService) createEmployee(ctx context.Context, timestamp time.Time, record entity.Record) (newRecord entity.Record, err error) {
	log.Println("SqliteRecordService createEmployee record:", record)
	name := record.DataVal("name")
	log.Println("SqliteRecordService.CreateRecord name:", *name)
	insuredIdStr := record.DataVal("insuredId")
	if *insuredIdStr == "" {
		log.Println("Missing insuredId")
		return newRecord, ErrRecordIDInvalid
	}
	insuredId, err := strconv.Atoi(*insuredIdStr)
	if err != nil {
		log.Println("Invalid insuredId")
		return newRecord, ErrRecordIDInvalid
	}

	_, err = s.GetRecordById(ctx, "insured", insuredId)
	if err != nil {
		return newRecord, ErrNonexistentParentRecord
	}

	var employee *entity.Employee
	employee = &entity.Employee{}
	employee.Name = *name
	employee.RecordTimestamp = timestamp
	employee.InsuredId = insuredId

	count, err := s.service.CountEmployeeRecords(ctx, *employee)
	if err != nil {
		log.Println("*****LOG*****Error count", err)
		return newRecord, ErrServerError
	} else if count != 0 {
		log.Println("**********Count != 0", count)
		return newRecord, ErrRecordAlreadyExists
	}

	if ii := record.DataVal("insuredId"); *ii != "" {
		if employee.InsuredId, err = strconv.Atoi(*ii); err != nil { // SET INSURED ID
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

	newRecord, err = s.service.CreateEmployee(ctx, employee)
	if err != nil {
		return entity.Record{}, err
	}
	return newRecord, err
}
func (s *SqliteRecordService) updateEmployee(ctx context.Context, timestamp time.Time, record entity.Record) (newRecord entity.Record, err error) {
	name := record.DataVal("name")
	fmt.Println("SqliteRecordService.CreateRecord name:", name)

	startDate := record.DataVal("startDate")
	endDate := record.DataVal("endDate")
	insuredId := record.DataVal("insuredId")
	if insuredId == nil {
		fmt.Println("Bad id?, nil. Record:", record)
		return newRecord, ErrRecordIDInvalid
	}
	insuredIdVal := *insuredId
	insuredIdInt, err := strconv.Atoi(insuredIdVal)
	if err != nil {
		fmt.Println("Missing insuredId in record?", record)
		return newRecord, ErrInvalidRequest
	}

	timestampString := timestamp.Format("2006-01-02T15:04:05Z07:00")

	employee, err := entity.NewEmployee(*name, *startDate, *endDate, insuredIdInt, timestampString)

	/* var employee *entity.Employee
	employee = &entity.Employee{}
	employee.Name = *name
	employee.StartDate = start
	employee.RecordTimestamp = timestamp */
	if err != nil {
		fmt.Println("Error from entity.NewEmployee. Record:", record, "Employee:", employee)
		return newRecord, ErrServerError
	}

	count, err := s.service.CountEmployeeRecords(ctx, *employee)
	if err != nil {
		fmt.Println("Count: ", count)
		return newRecord, ErrServerError
	} else if count == 0 {
		return newRecord, ErrRecordDoesNotExist
	}
	fmt.Println("Count whoa: ", count)
	newRecord, err = s.service.CreateEmployee(ctx, employee) // add record to DB with employee update
	if err != nil {
		return entity.Record{}, err
	}
	return newRecord, nil
}

func (s *SqliteRecordService) updateAddress(ctx context.Context, timestamp time.Time, record entity.Record) (newRecord entity.Record, err error) {
	name := record.DataVal("name")
	fmt.Println("SqliteRecordService.updateAddress name:", name)

	var address *entity.Address
	address = &entity.Address{}
	address.Address = *record.DataVal("address")
	address.RecordTimestamp = timestamp

	if ii := record.DataVal("insuredId"); *ii != "" {
		if address.InsuredId, err = strconv.Atoi(*ii); err != nil { // SET INSURED ID
			fmt.Println("Error converting to int:", err)
			return newRecord, ErrServerError
		}
	} else {
		fmt.Println("ii", ii)
		return newRecord, fmt.Errorf("Insured ID required to create Address: %v", err)
	}

	insuredRecord, err := s.GetRecordById(ctx, "insured", address.InsuredId)
	if err != nil {
		return newRecord, ErrRecordIDInvalid
	}
	insured := entity.Insured{}
	insured.FromRecord(insuredRecord)
	addressCount, err := s.service.CountInsuredAddresses(ctx, insured)
	if err != nil {
		return newRecord, ErrServerError
	}
	if addressCount == 0 {
		return newRecord, ErrRecordDoesNotExist
	}

	newRecord, err = s.service.CreateAddress(ctx, address) // add record to DB indicating an address change
	if err != nil {
		return entity.Record{}, err
	}
	return newRecord, err
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
