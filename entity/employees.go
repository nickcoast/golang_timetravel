package entity

import (
	"context"
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

func (e *Employee) ToRecord() Record {
	idString := strconv.Itoa(e.ID)
	r := Record{
		ID: e.ID,
		Data: map[string]string{
			"id":               idString,
			"name":             e.Name,
			"start_date":       e.StartDate.Format("2006-01-02"),
			"end_date":         e.EndDate.Format("2006-01-02"),
			"insured_id":       strconv.Itoa(e.InsuredId),
			"record_timestamp": strconv.Itoa(int(e.RecordTimestamp.Unix())),
		},
	}
	return r
}

func (e *Employee) FromRecord(r Record) (err error) {
	e.ID = r.ID
	e.Name = r.Data["name"]
	e.StartDate, err = time.Parse("2006-01-02", r.Data["start_date"])
	//e.EndDate, _ = time.Parse
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

//func EmployeesFromData(data map[string]string ) 