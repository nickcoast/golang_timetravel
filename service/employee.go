package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/nickcoast/timetravel/entity"
)

func (s *SqliteRecordService) createEmployee(ctx context.Context, timestamp time.Time, record entity.Record) (newRecord entity.Record, err error) {
	log.Println("SqliteRecordService createEmployee record:", record)
	name := record.DataVal("name")
	insuredIdStr := record.DataVal("insuredId")
	if insuredIdStr == "" {
		log.Println("Missing insuredId")
		return newRecord, ErrRecordIDInvalid
	}
	insuredId, err := strconv.Atoi(insuredIdStr)
	if err != nil {
		log.Println("Invalid insuredId")
		return newRecord, ErrRecordIDInvalid
	}

	_, err = s.GetResourceById(ctx, "insured", insuredId)
	if err != nil {
		return newRecord, ErrNonexistentParentRecord
	}

	var employee *entity.Employee
	employee = &entity.Employee{}
	employee.Name = name
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

	if ii := record.DataVal("insuredId"); ii != "" {
		if employee.InsuredId, err = strconv.Atoi(ii); err != nil { // SET INSURED ID
			fmt.Println("Error converting to int:", err)
			return newRecord, fmt.Errorf("Problem converting string to int")
		}
	} else {
		fmt.Println("ii", ii)
		return newRecord, fmt.Errorf("Insured ID required to create Employee: %v", err)
	}
	sd := record.DataVal("startDate")
	if sd == "" || len(sd) != 10 {
		return newRecord, fmt.Errorf("startDate required to create Employee: %v, %v, %v", err, "sd: ", sd)
	}
	employee.StartDate, err = time.Parse("2006-01-02", sd)
	if err != nil {
		fmt.Println("Start Date required")
		return newRecord, ErrInvalidRequest
	}
	ed := record.DataVal("endDate")
	if ed == "" || len(ed) != 10 {
		fmt.Println("Bad endDate or error. SKIPPING (not required). End date: ", ed)
	} else {
		t, err := time.Parse("2006-01-02", ed)
		if err != nil {
			fmt.Println("Bad endDate or error. SKIPPING (not required). End date: ", ed)
			return newRecord, err
		}
		employee.EndDate = t
	}

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
	if insuredId == "" {
		fmt.Println("Bad id?, nil. Record:", record)
		return newRecord, ErrRecordIDInvalid
	}
	insuredIdVal := insuredId
	insuredIdInt, err := strconv.Atoi(insuredIdVal)
	if err != nil {
		fmt.Println("Missing insuredId in record?", record)
		return newRecord, ErrInvalidRequest
	}

	timestampString := timestamp.Format("2006-01-02T15:04:05Z07:00")

	employee, err := entity.NewEmployee(name, startDate, endDate, insuredIdInt, timestampString)

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
	newRecord, err = s.service.CreateEmployee(ctx, employee) // add record to DB with employee update
	ed := newRecord.DataVal("end_date")
	if ed == "" || len(ed) != 10 || ed == "0001-01-01" {
		fmt.Println("Bad endDate or error. SKIPPING (not required). End date: ", ed)
		delete(newRecord.Data, "end_date")
	}
	if err != nil {
		return entity.Record{}, err
	}
	return newRecord, nil
}
