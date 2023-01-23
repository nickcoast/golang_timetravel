package sqlite

// TODO: should this be in package "service" stead "sqlite"??

import (
	"context"
	"fmt"

	/* "database/sql" */

	"github.com/nickcoast/timetravel/entity"
)

// Create new employee *record*. Used for creating new employee, and for updating
func (s *InsuredService) CreateEmployee(ctx context.Context, employee *entity.Employee) (record entity.Record, err error) {
	tx, err := s.Db.BeginTx(ctx, nil)
	if err != nil {
		return record, err
	}
	defer tx.Rollback()

	fmt.Println("InsuredService.CreateEmployee employee:", employee)
	// Create a new employee record
	record, err = createEmployee(ctx, tx, employee)
	fmt.Println("InsuredService.CreateEmployee record:", record)
	if err != nil && err.Error() == "UNIQUE constraint failed: employees.insured_id, employees.name, employees.start_date, employees.end_date" {
		fmt.Println("Duplicate key error. Insert failed.")
		return record, ErrRecordAlreadyExists
	}
	if err != nil { // any other error
		return record, err
	}
	if err = tx.Commit(); err != nil {
		fmt.Println("jkl")
		return record, err
	}
	return record, nil
}

// createEmployee creates a new employee.
func createEmployee(ctx context.Context, tx *Tx, employee *entity.Employee) (record entity.Record, err error) {
	// Perform basic field validation.
	if err := employee.Validate(); err != nil {
		return record, err
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO employees (
			name,
			start_date,
			end_date,
			insured_id,		
			record_timestamp		
		)
		VALUES (?, ?, ?, ?, ?)
	`,
		employee.Name,
		employee.StartDate.Format("2006-01-02"),
		employee.EndDate.Format("2006-01-02"),
		employee.InsuredId,
		employee.RecordTimestamp.Unix(), // can use a Scan method here if necessary
	)
	if err != nil {
		return record, FormatError(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return record, err
	}
	employee.ID = int(id)
	record = employee.ToRecord()
	return record, nil
}

func (s *InsuredService) UpdateEmployee(ctx context.Context, employee *entity.Employee) (record entity.Record, err error) {
	tx, err := s.Db.BeginTx(ctx, nil)
	if err != nil {
		return record, err
	}
	defer tx.Rollback()

	count, err := s.CountEmployeeRecords(ctx, *employee)
	if err != nil {
		return entity.Record{}, err
	}
	if count != 0 {
		return entity.Record{}, fmt.Errorf("Employee '%v' for Insured ID '%v' does not exist. Use 'new' to update it.", employee.Name, employee.InsuredId)
	}

	// Update an employee record
	record, err = updateEmployee(ctx, tx, employee)
	fmt.Println("InsuredService.UpdateEmployee record:", record)
	if err != nil && err.Error() == "UNIQUE constraint failed: employees.insured_id, employees.name, employees.start_date, employees.end_date" {
		fmt.Println("Duplicate key error. Insert failed.")
		return record, ErrRecordAlreadyExists
	}
	if err != nil { // any other error
		return record, err
	}
	if err = tx.Commit(); err != nil {
		fmt.Println("jkl")
		return record, err
	}
	return record, nil
}

func updateEmployee(ctx context.Context, tx *Tx, employee *entity.Employee) (record entity.Record, err error) {
	// Perform basic field validation.
	if err := employee.Validate(); err != nil {
		return record, err
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO employees (
			name,
			start_date,
			end_date,
			insured_id,		
			record_timestamp		
		)
		VALUES (?, ?, ?, ?, ?)
	`,
		employee.Name,
		employee.StartDate.Format("2006-01-02"),
		employee.EndDate.Format("2006-01-02"),
		employee.InsuredId,
		employee.RecordTimestamp.Unix(), // can use a Scan method here if necessary
	)
	if err != nil {
		return record, FormatError(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return record, err
	}
	employee.ID = int(id)
	record = employee.ToRecord()
	return record, nil
}

// CountEmployeeRecords checks exists employee (regardless of time-travelable attributes)
// if exists, then API consumer should be submitting "UPDATE"
//
// TODO: check if can use general solution in *DB
func (s *InsuredService) CountEmployeeRecords(ctx context.Context, employee entity.Employee) (count int, err error) {
	tx, err := s.Db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if err := employee.Validate(); err != nil {
		return 0, err
	}
	query := `
	SELECT COUNT(*) FROM employees
	WHERE id = ?	
`
	result, err := tx.QueryContext(ctx, query,
		employee.ID,
	)

	if err != nil {
		return 0, err
	}
	defer result.Close()
	for result.Next() {
		if err := result.Scan(&count); err != nil {
			return 0, err
		}
	}
	return count, nil
}

// NOT USED BY API. Bypassing Service for generalized DB methods for Delete and Get
func (s *InsuredService) DeleteEmployee(ctx context.Context, id int) (record entity.Record, err error) {
	return record, nil
}
