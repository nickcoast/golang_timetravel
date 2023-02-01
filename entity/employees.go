package entity

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Employee represents a employee in the system.
// employees can also be created directly for testing.
type Employee struct {
	ID int `json:"id"`

	// Employee's preferred
	Name string `json:"name"`

	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`

	InsuredId int `json:"insuredId"`

	// Timestamps for employee creation & last update.
	RecordTimestamp time.Time `json:"recordTimestamp"`
}


type EmployeeCollection struct {
	EmployeeList *list.List
}

var _ InsuredInterface = (*Employee)(nil)

func (u *Employee) GetId() int64 {
	return int64(u.ID)
}
func (u *Employee) GetInsuredId() int64 {
	return int64(u.InsuredId)
}
func (u *Employee) GetDataTableName() string {
	return "employees_records"
}
func (u *Employee) GetIdentTableName() string {
	return "employees"
}
func (u *Employee) GetInsertFields() map[string]string {
	return map[string]string{
		"name":       u.Name,
		"start_date": u.StartDate.Format("2006-01-02"),
		"end_date":   u.EndDate.Format("2006-01-02"),
	}
}

// Validate returns an error if the employee contains invalid fields.
// This only performs basic validation.
func (u *Employee) Validate() error {
	if u.Name == "" {
		return Errorf(EINVALID, "Employee name required.")
	}
	if u.InsuredId < 1 {
		return Errorf(EINVALID, "Employee must have an insured_id")
	}
	return nil
}

// EmployeeService represents a service for managing employees.
type EmployeeService interface {
	// Retrieves a employee by ID
	// Returns ENOTFOUND if employee does not exist.
	FindEmployeeByID(ctx context.Context, id int) (*Employee, error)

	// Retrieves a list of employees by filter. Also returns total count of matching
	// employees which may differ from returned results if filter.Limit is specified.
	FindEmployees(ctx context.Context, filter EmployeeFilter) ([]*Employee, int, error)

	// Creates a new employee.
	CreateEmployee(ctx context.Context, employee *Employee) (int64, int, error)

	// Updates a employee object. Returns ENOTFOUND if employee does not exist.
	// REMOVED from interface. Will not support updates to the core table for now
	/* UpdateEmployee(ctx context.Context, id int, upd EmployeeUpdate) (*Employee, error) */

	// Permanently deletes a employee and all owned dials. Returns ENOTFOUND if
	// employee does not exist.
	DeleteEmployee(ctx context.Context, id int) error
}

// EmployeeFilter represents a filter passed to FindEmployees().
type EmployeeFilter struct {
	// Filtering fields.
	ID              *int       `json:"id"`
	Name            *string    `json:"name"`
	StartDate       *time.Time `json:"startDate"`
	EndDate         *time.Time `json:"endDate"`
	RecordTimestamp *int       `json:"recordTimestamp"`

	// Restrict to subset of results.
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// EmployeeUpdate represents a set of fields to be updated via UpdateEmployee().
type EmployeeUpdate struct {
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
}

// EndDate is optional
func NewEmployee(employeeID int, name string, startDate string, endDate string, insuredId int, recordTimestamp string) (employee *Employee, err error) {
	if len(name) == 0 {
		return employee, fmt.Errorf("Name must be at least 1 character long")
	}
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return employee, fmt.Errorf("Start/EndDate must be in format '2006-01-02'. EndDate may be empty.")
	}
	end, err := time.Parse("2006-01-02", endDate)
	/* if err == nil {
		employee.EndDate = end // leave empty on error
	} */

	//timestamp, err := time.Parse("2006-01-02 15:04:05", recordTimestamp) // wrong timezone. Need timezone part
	timestamp, err := time.Parse("2006-01-02T15:04:05Z07:00", recordTimestamp)
	if err != nil {
		return employee, err
	}
	employee = &Employee{
		ID:              employeeID,
		Name:            name,
		StartDate:       start,
		EndDate:         end,
		InsuredId:       insuredId,
		RecordTimestamp: timestamp,
	}

	return employee, nil
}

func (e *Employee) ToRecord() Record {
	endDateString := e.EndDate.Format("2006-01-02")
	if endDateString == "0001-01-01" {
		endDateString = ""
	}
	idString := strconv.Itoa(e.ID)
	r := Record{
		ID: e.ID,
		Data: map[string]string{
			"id":              idString,
			"name":            e.Name,
			"startDate":       e.StartDate.Format("2006-01-02"),
			"endDate":         endDateString,
			"insuredId":       strconv.Itoa(e.InsuredId),
			"recordTimestamp": strconv.Itoa(int(e.RecordTimestamp.Unix())),
		},
	}
	return r
}

func (e *Employee) FromRecord(r Record) (err error) {
	e.ID = r.ID
	e.Name = r.Data["name"]
	e.StartDate, err = time.Parse("2006-01-02", r.Data["start_date"])
	e.EndDate, _ = time.Parse("2006-01-02", r.Data["end_date"])
	e.InsuredId, err = strconv.Atoi(r.Data["insured_id"])
	timestampInt, err := strconv.Atoi(r.Data["record_timestamp"])
	e.RecordTimestamp = time.Unix(int64(timestampInt), 0)
	return err
}

func EmployeesFromRecords(records map[int]Record) (map[int]Employee, error) {
	employeees := make(map[int]Employee)
	for i, e := range records {
		id := i
		employee := Employee{}
		err := employee.FromRecord(e)
		if err != nil {
			return map[int]Employee{}, err
		}
		employeees[id] = employee
	}
	return employeees, nil
}

// Returns Employee map. Skips any non-employees
func EmployeesFromInsuredInterface(insuredIfaceObjs map[int]InsuredInterface) (map[int]Employee, error) {
	employees := make(map[int]Employee)
	for i, obj := range insuredIfaceObjs {
		e, ok := obj.(*Employee)
		if ok {
			employees[i] = *e
		}
	}
	return employees, nil
}

func (e Employee) MarshalJSON() ([]byte, error) {
	if e.ID == 0 {
		return json.Marshal(&struct {
			ID string `json:"id"`
		}{
			ID: "",
		})
	}
	endDate := e.EndDate.Format("2006-01-02")
	if endDate == "0001-01-01" {
		endDate = ""
	}
	return json.Marshal(&struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		StartDate       string `json:"startDate"`
		EndDate         string `json:"endDate"`
		InsuredId       string `json:"insuredId"`
		RecordTimestamp string `json:"recordTimestamp"`
		RecordDateTime  string `json:"recordDateTime"`
	}{
		ID:              strconv.Itoa(e.ID),
		Name:            e.Name,
		StartDate:       e.StartDate.Format("2006-01-02"),
		EndDate:         endDate,
		InsuredId:       strconv.Itoa(e.InsuredId),
		RecordTimestamp: strconv.Itoa(int(e.RecordTimestamp.Unix())),
		RecordDateTime:  e.RecordTimestamp.Format("Mon, 02 Jan 2006 15:04:05 MST"),
	})
}
