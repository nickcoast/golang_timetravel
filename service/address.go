package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/nickcoast/timetravel/entity"
)

func (s *InsuredRecordService) createAddress(ctx context.Context, timestamp time.Time, record entity.Record) (newRecord entity.Record, err error) {
	addressString := record.DataVal("address")
	if addressString == "" {
		return entity.Record{}, ErrServerError
	}
	fmt.Println("SqliteRecordService.CreateRecord name:", addressString)
	var address *entity.Address
	address = &entity.Address{}
	address.Address = addressString
	address.RecordTimestamp = timestamp

	if ii := record.DataVal("insuredId"); ii != "" { // SET INSURED ID
		if address.InsuredId, err = strconv.Atoi(ii); err != nil {
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
	addressCount, err := s.dbService.CountInsuredAddresses(ctx, insured)
	if err != nil {
		return newRecord, ErrServerError
	}
	if addressCount > 0 {
		return newRecord, ErrRecordAlreadyExists
	}
	newRecord, err = s.dbService.CreateAddress(ctx, address)
	if err != nil {
		return entity.Record{}, err
	}
	return newRecord, err
}

func (s *InsuredRecordService) updateAddress(ctx context.Context, timestamp time.Time, record entity.Record) (newRecord entity.Record, err error) {
	name := record.DataVal("name")
	fmt.Println("SqliteRecordService.updateAddress name:", name)

	var address *entity.Address
	address = &entity.Address{}
	address.Address = record.DataVal("address")
	address.RecordTimestamp = timestamp

	if ii := record.DataVal("insuredId"); ii != "" {
		if address.InsuredId, err = strconv.Atoi(ii); err != nil { // SET INSURED ID
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
	addressCount, err := s.dbService.CountInsuredAddresses(ctx, insured)
	if err != nil {
		return newRecord, ErrServerError
	}
	if addressCount == 0 {
		return newRecord, ErrRecordDoesNotExist
	}

	newRecord, err = s.dbService.CreateAddress(ctx, address) // add record to DB indicating an address change
	if err != nil {
		return entity.Record{}, err
	}
	return newRecord, err
}
