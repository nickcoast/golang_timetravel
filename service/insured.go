package service

import (
	"context"
	"log"
	"time"

	"github.com/nickcoast/timetravel/entity"
	"github.com/nickcoast/timetravel/sqlite"
)

// ObjectResourceService - new interface to disentangle RDBMS from Record service interface
//
// returns "Insured" objects (Insured, Employee, Address, and collections thereof)
//
// TODO: change each return type to entity.InsuredInterface
type ObjectResourceService interface {

	// GetResource will retrieve an resource.
	GetResourceById(ctx context.Context, resource entity.InsuredInterface, id int) (entity.InsuredInterface, error) // TODO: change resource from string to Struct
	//GetResourceById(ctx context.Context, insuredType entity.InsuredInterface, id int64) (entity.InsuredInterface, error) // TODO: change resource from string to Struct

	// CreateResource will insert a new resource.
	//
	// If it a resource with that id already exists it will fail.
	CreateResource(ctx context.Context, resource string, record entity.Record) (entity.Record, error)
	//CreateResource(ctx context.Context, insuredType entity.InsuredInterface) (entity.InsuredInterface, error)

	// UpdateResource will change the internal `Map` values of the resource if they exist.
	// if the update[key] is null it will delete that key from the resource's Map.
	//
	// UpdateResource will error if id <= 0 or the resource does not exist with that id.

	UpdateResource(ctx context.Context, resource string, record entity.Record) (entity.Record, error)
	//UpdateResource(ctx context.Context, insuredType entity.InsuredInterface) ( entity.InsuredInterface, error)

	//DeleteResource(ctx context.Context, resource string, id int64) (entity.Record, error)
	DeleteResource(ctx context.Context, insuredType entity.InsuredInterface, id int64) (entity.InsuredInterface, error)

	//GetResourceByDate(ctx context.Context, resource string, naturalKey string, insuredId int64, date time.Time) (records entity.Record, err error)
	// TODO: remove natural key. Maybe insuredId
	GetResourceByDate(ctx context.Context, insuredIfaceObj entity.InsuredInterface, naturalKey string, insuredId int64, date time.Time) (entity.InsuredInterface, error)

	GetInsuredByDate(ctx context.Context, insuredId int64, date time.Time) (insured entity.Insured, err error)
	//GetInsuredByDate(ctx context.Context, insuredType entity.InsuredInterface, date time.Time) (entity.InsuredInterface, error)
}

var _ ObjectResourceService = (*SqliteRecordService)(nil)

// InMemoryRecordService is an in-memory implementation of RecordService.
type SqliteRecordService struct {
	data    map[int]entity.Record
	service sqlite.InsuredService
}

//var _ ObjectResourceService = (*SqliteRecordService)(nil)

func NewSqliteRecordService() SqliteRecordService {
	return SqliteRecordService{
		data: map[int]entity.Record{},
	}
}

func (s *SqliteRecordService) SetService(db *sqlite.DB) {
	s.service = *sqlite.NewInsuredService(db)
}

func (s *SqliteRecordService) CreateResource(ctx context.Context, resource string, record entity.Record) (newRecord entity.Record, err error) {
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
	if err != nil {
		return newRecord, err
		// TODO: May want to use this here later
		//return ErrRecordAlreadyExists
	}
	return newRecord, nil
}

func (s *SqliteRecordService) UpdateResource(ctx context.Context, resource string, record entity.Record) (updateRecord entity.Record, err error) {
	/* id := updateRecord.ID
	if id != 0 {
		return updateRecord, ErrRecordIDInvalid
	} */
	timestamp := time.Now() // for all new record creation
	log.Println("SqliteRecordServce.updateRecord record:", updateRecord)
	if resource == "insured" {
		return record, ErrRecordAlreadyExists // cannot update insured (name, policy id). Address and address data are updateable
	} else if resource == "address" || resource == "addresses" || resource == "insured_addresses" || resource == "insured_address" {
		return s.updateAddress(ctx, timestamp, record)
	} else if resource == "employee" || resource == "employees" {
		updateRecord, err := s.updateEmployee(ctx, timestamp, record)
		if err == sqlite.ErrUpdateMustChangeAValue {
			err = ErrRecordUpdateRequireChange
		}
		return updateRecord, err
	}
	return updateRecord, ErrRecordAlreadyExists
}

func (s *SqliteRecordService) GetResourceById(ctx context.Context, resource entity.InsuredInterface, id int) (entity.InsuredInterface, error) {
	if id == 0 {
		return nil, ErrRecordDoesNotExist
	}
	e, err := s.service.Db.GetById(ctx, resource, int64(id))
	if err != nil || e.GetId() == 0 {
		return nil, ErrRecordDoesNotExist
	}
	// TODO: make sure Marshal handles this if needed
	/* ed := e.DataVal("end_date") // check "empty" end date
	if ed != "" && ed == "0001-01-01" {
		if ed == "0001-01-01" {
			e.Data["end_date"] = "None"
		}
	} */

	return e, nil
}

func (s *SqliteRecordService) DeleteResource(ctx context.Context, insuredObj entity.InsuredInterface, id int64) (record entity.InsuredInterface, err error) {
	if id == 0 {
		return record, ErrRecordDoesNotExist
	}
	record, err = s.service.Db.DeleteById(ctx, insuredObj, id)
	if err != nil {
		return record, ErrRecordDoesNotExist
	}
	return record, nil
}

func (s *SqliteRecordService) createInsured(ctx context.Context, timestamp time.Time, record entity.Record) (newRecord entity.Record, err error) {
	name := record.DataVal("name")
	var insured *entity.Insured
	insured = &entity.Insured{}
	insured.Name = name
	insured.RecordTimestamp = timestamp

	newRecord, err = s.service.CreateInsured(ctx, insured)
	if err != nil {
		return entity.Record{}, err
	}
	return newRecord, nil
}

func (s *SqliteRecordService) GetInsuredByDate(ctx context.Context, insuredId int64, dateValid time.Time) (entity.Insured, error) {
	return s.service.Db.GetInsuredByDate(ctx, insuredId, dateValid)
}

// TODO: use map for natural key, or struct
func (s *SqliteRecordService) GetResourceByDate(ctx context.Context, insuredIfaceObj entity.InsuredInterface, naturalKey string, insuredId int64, dateValid time.Time) (entity.InsuredInterface, error) {
	if insuredId == 0 {
		return nil, ErrNonexistentParentRecord
	}

	record, err := s.service.Db.GetByDate(ctx, insuredIfaceObj, naturalKey, insuredId, dateValid)
	if err != nil {
		return nil, ErrServerError
	}
	for _, r := range record {
		return r, nil
	}
	return nil, ErrRecordDoesNotExist
}
