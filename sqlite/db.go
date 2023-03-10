package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"database/sql/driver"
	"embed"
	"fmt"
	"io/fs"
	"sort"

	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/nickcoast/timetravel/entity"
)

// see https://github.com/benbjohnson/wtf/blob/321f7917f4004f4365f826d3fae3d5777ecf54d8/sqlite/sqlite.go

//go:embed migration/*.sql
var migrationFS embed.FS

type DB struct {
	db     *sql.DB
	ctx    context.Context
	cancel func()
	DSN    string

	tableNames         map[string]int // TODO: DELETE?
	allowedNaturalKeys map[string]int // TODO: DELETE

	Now func() time.Time
}

func NewDB(dsn string) *DB {
	tn := make(map[string]int)
	tn["insured"] = 0
	tn["employee"] = 1
	tn["address"] = 2
	ank := make(map[string]int)
	ank["name"] = 0
	ank["address"] = 1
	db := &DB{
		DSN:                dsn,
		Now:                time.Now,
		tableNames:         tn,
		allowedNaturalKeys: ank,
	}
	db.ctx, db.cancel = context.WithCancel((context.Background())) // new context?
	return db
}

var ErrRecordDoesNotExist = errors.New("record with that id does not exist")
var ErrRecordIDInvalid = errors.New("record id must >= 0")
var ErrRecordAlreadyExists = errors.New("record already exists")
var ErrRecordMatchingCriteriaDoesNotExist = errors.New("no records matched your search")
var ErrUpdateMustChangeAValue = errors.New("update must modify at least one value")

func (db *DB) Open() (err error) { // need ctx here or not?

	if db.DSN == "" {
		return fmt.Errorf("dsn required")
	}

	if db.DSN != ":memory" {
		if err := os.MkdirAll(filepath.Dir(db.DSN), 0700); err != nil {
			return err
		}
	}

	if db.db, err = sql.Open("sqlite3", db.DSN); err != nil { // could hard-code DB DSN here instead
		return err
	}

	if _, err := db.db.Exec(`PRAGMA journal_mode = wal;`); err != nil {
		return fmt.Errorf("enable wal: %w", err)
	}

	if _, err := db.db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		return fmt.Errorf("foreign keys pragma: %w", err)
	}

	if err := db.migrate(); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	return nil
}

func (db *DB) migrate() error {
	// Ensure the 'migrations' table exists so we don't duplicate migrations.
	if _, err := db.db.Exec(`CREATE TABLE IF NOT EXISTS migrations (name TEXT PRIMARY KEY);`); err != nil {
		return fmt.Errorf("cannot create migrations table: %w", err)
	}

	// Read migration files from our embedded file system.
	// This uses Go 1.16's 'embed' package.
	names, err := fs.Glob(migrationFS, "migration/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(names)

	// Loop over all migration files and execute them in order.
	for _, name := range names {
		if err := db.migrateFile(name); err != nil {
			return fmt.Errorf("migration error: name=%q err=%w", name, err)
		}
	}
	return nil
}

func (db *DB) migrateFile(name string) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Ensure migration has not already been run.
	var n int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM migrations WHERE name = ?`, name).Scan(&n); err != nil {
		return err
	} else if n != 0 {
		return nil // already run migration, skip
	}

	// Read and execute migration file.
	if buf, err := fs.ReadFile(migrationFS, name); err != nil {
		return err
	} else if _, err := tx.Exec(string(buf)); err != nil {
		return err
	}

	// Insert record into migrations to prevent re-running migration.
	if _, err := tx.Exec(`INSERT INTO migrations (name) VALUES (?)`, name); err != nil {
		return err
	}

	return tx.Commit()
}

// TODO: change tableName to entity.InsuredInterface
func (db *DB) GetById(ctx context.Context, insuredObj entity.InsuredInterface, id int64) (record entity.InsuredInterface, err error) {
	if id == 0 {
		return record, ErrRecordDoesNotExist
	}
	tx, err := db.db.Begin()
	defer tx.Commit()
	if err != nil {
		return record, err
	}
	defer tx.Rollback()

	switch objType := insuredObj.(type) {
	case *entity.Insured:
		return db.GetInsuredById(ctx, *objType, id)
	case *entity.Employee:
		return db.GetEmployeeById(ctx, *objType, id)
	case *entity.Address:
		return db.GetAddressById(ctx, *objType, id)
	}
	return nil, err
}

// GetInsuredById returns the insured record for this Id
func (db *DB) GetInsuredById(ctx context.Context, insured entity.Insured, id int64) (*entity.Insured, error) {
	if id == 0 {
		return &entity.Insured{}, ErrRecordDoesNotExist
	}
	tx, err := db.db.Begin()
	if err != nil {
		return &entity.Insured{}, err
	}
	defer tx.Rollback()

	ids := []int64{id}
	query := generateSelectByIds(&insured, ids)

	rows, err := tx.QueryContext(ctx, query) // id(s) are inserted in generateSelectByIds
	if err != nil {
		fmt.Println("bad query: ", query)
		return &entity.Insured{}, fmt.Errorf("Query failed")
	}

	for rows.Next() {
		if err := rows.Scan(&insured.ID,
			&insured.Name,
			&insured.PolicyNumber,
			(*NullTime)(&insured.RecordTimestamp),
		); err != nil {
			return &entity.Insured{}, err
		}
	}
	if err := rows.Err(); err != nil {
		return &entity.Insured{}, fmt.Errorf("rowsErr: %v", err)
	}
	if insured.ID == 0 {
		return &entity.Insured{}, ErrRecordDoesNotExist
	}
	tx.Commit()
	return &insured, err
}

// GetEmployeeById returns the most recent employee record for this Id
func (db *DB) GetEmployeeById(ctx context.Context, employee entity.Employee, id int64) (*entity.Employee, error) {
	if id == 0 {
		return &entity.Employee{}, ErrRecordDoesNotExist
	}
	tx, err := db.db.Begin()
	if err != nil {
		return &entity.Employee{}, err
	}
	defer tx.Rollback()

	ids := []int64{id}
	query := generateSelectByIds(&employee, ids)

	rows, err := tx.QueryContext(ctx, query) // id(s) are inserted in generateSelectByIds
	if err != nil {
		fmt.Println("bad query")
		return &entity.Employee{}, fmt.Errorf("Query failed")
	}
	var recordId int
	var garbage int
	for rows.Next() {
		if err := rows.Scan(
			&employee.ID,
			&recordId, // not implemented in Employee yet
			&employee.InsuredId,
			&employee.Name,
			(*ShortTime)(&employee.StartDate),
			(*ShortTime)(&employee.EndDate),
			(*NullTime)(&employee.RecordTimestamp),
			&garbage, // same as RecordTimestamp

		); err != nil {
			return &entity.Employee{}, err
		}
	}
	if err := rows.Err(); err != nil {
		return &entity.Employee{}, fmt.Errorf("rowsErr: %v", err)
	}

	tx.Commit()
	return &employee, err
}

// scanRows. Note: cannot handle empty result set
func scanRows(ctx context.Context, insuredIfaceObj entity.InsuredInterface, rows *sql.Rows) (map[int]entity.InsuredInterface, error) {
	insuredIfaceMap := make(map[int]entity.InsuredInterface)
	switch insuredIfaceObj.(type) {
	case *entity.Employee:
		var recordId int
		var garbage int
		i := 0
		for rows.Next() {
			employee := entity.Employee{}
			if err := rows.Scan( // will this overwrite with each loop??
				&employee.ID,
				&recordId, // not implemented in Employee yet
				&employee.InsuredId,
				&employee.Name,
				(*ShortTime)(&employee.StartDate),
				(*ShortTime)(&employee.EndDate),
				(*NullTime)(&employee.RecordTimestamp),
				&garbage, // same as RecordTimestamp

			); err != nil {
				return nil, err
			}
			insuredIfaceMap[i] = &employee
			i++
		}
	case *entity.Address:
		var garbage int
		i := 0
		for rows.Next() {
			address := entity.Address{}
			if err := rows.Scan(
				&address.ID,
				&address.Address,
				&address.InsuredId,
				(*NullTime)(&address.RecordTimestamp),
				&garbage, // same as record_timestamp
			); err != nil {
				return nil, err
			}
			insuredIfaceMap[i] = &address
			i++
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("rowsErr: %v", err)
		}
	case *entity.Insured:
		//var garbage int
		i := 0
		for rows.Next() {
			insured := entity.Insured{}
			if err := rows.Scan(&insured.ID,
				&insured.Name,
				&insured.PolicyNumber,
				(*NullTime)(&insured.RecordTimestamp),
				//&garbage, // same as record_timestamp
			); err != nil {
				return nil, err
			}
			insuredIfaceMap[i] = &insured
			i++
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("rowsErr: %v", err)
		}
	}
	return insuredIfaceMap, nil
}

// GetAddressById exactly the same as GetEmployeeById but for Address
func (db *DB) GetAddressById(ctx context.Context, address entity.Address, id int64) (*entity.Address, error) {
	if id == 0 {
		return &entity.Address{}, ErrRecordDoesNotExist
	}
	tx, err := db.db.Begin()
	if err != nil {
		return &entity.Address{}, err
	}
	defer tx.Rollback()

	ids := []int64{id}
	query := generateSelectByIds(&address, ids)

	rows, err := tx.QueryContext(ctx, query) // id(s) are inserted in generateSelectByIds
	if err != nil {
		fmt.Println("bad query")
		return &entity.Address{}, fmt.Errorf("Query failed")
	}
	var garbage int
	for rows.Next() {
		if err := rows.Scan(
			&address.ID,
			&address.Address,
			&address.InsuredId,
			(*NullTime)(&address.RecordTimestamp),
			&garbage, // same as record_timestamp
		); err != nil {
			return &entity.Address{}, err
		}
	}
	if err := rows.Err(); err != nil {
		return &entity.Address{}, fmt.Errorf("rowsErr: %v", err)
	}

	tx.Commit()
	return &address, err
}

// Get Insured entity with component employees and addresses valid at a particular date.
func (db *DB) GetInsuredByDate(ctx context.Context, insuredId int64, date time.Time) (insured entity.Insured, err error) {
	if insuredId == 0 {
		return insured, ErrRecordDoesNotExist
	}

	tx, err := db.db.Begin()
	if err != nil {
		return insured, err
	}
	defer tx.Rollback()

	insuredIfaceObj, err := db.GetById(ctx, &entity.Insured{}, insuredId)
	if err != nil {
		return entity.Insured{}, FormatError(err)
	}
	insuredObj, ok := insuredIfaceObj.(*entity.Insured)
	if !ok {
		return entity.Insured{}, fmt.Errorf("Internal Server Error")
	}
	if insuredObj.ID == 0 {
		return entity.Insured{}, ErrRecordDoesNotExist
	}
	employeeRecords, err := db.GetByDate(ctx, &entity.Employee{}, "naturalkey", insuredId, date)
	if err != nil {
		return entity.Insured{}, FormatError(err)
	}
	addressRecords, err := db.GetByDate(ctx, &entity.Address{}, "naturalkey", insuredId, date)
	if err != nil {
		return entity.Insured{}, FormatError(err)
	}

	employees, err := entity.EmployeesFromInsuredInterface(employeeRecords)
	addresses, err := entity.AddressesFromInsuredInterface(addressRecords)

	insuredObj.Employees = &employees
	insuredObj.Addresses = &addresses

	tx.Commit()
	return *insuredObj, nil
}

func (db *DB) GetAll(ctx context.Context, entityType entity.InsuredInterface) (records map[int]entity.InsuredInterface, err error) {
	query := generateSelectAll(entityType)
	tx, err := db.db.Begin()
	if err != nil {
		return records, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return records, err
	}
	return scanRows(ctx, entityType, rows)
}

func (db *DB) GetAllByEntityId(ctx context.Context, entityType entity.InsuredInterface, entityId int64) (records map[int]entity.InsuredInterface, err error) {
	query := generateSelectAllRecordsByEntityId(entityType, entityId)
	tx, err := db.db.Begin()
	if err != nil {
		return records, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, query, entityId)
	if err != nil {
		return records, err
	}
	return scanRows(ctx, entityType, rows)
}

// TODO: can remove naturalKey from signature?
func (db *DB) GetByDate(ctx context.Context, insuredIfaceObj entity.InsuredInterface, naturalKey string, insuredId int64, date time.Time) (records map[int]entity.InsuredInterface, err error) {
	id := insuredId
	if id == 0 {
		return records, ErrRecordDoesNotExist
	}
	tx, err := db.db.Begin()
	if err != nil {
		return records, err
	}
	defer tx.Rollback()
	count, err := db.CountInsuredRecordsAtDate(ctx, tx, insuredIfaceObj, insuredId, date)
	if err != nil {
		return records, fmt.Errorf("Server Error")
	}
	if count == 0 {
		return records, nil
	}

	query := generateSelectByDate(insuredIfaceObj, date)
	rows, err := tx.QueryContext(ctx, query, id)
	if err != nil {
		fmt.Println("bad query")
		return records, fmt.Errorf("Query failed")
	}

	records, err = scanRows(ctx, insuredIfaceObj, rows)
	if err != nil {
		return nil, fmt.Errorf("Server Error")
	}
	tx.Commit()
	return records, nil
}

func (db *DB) CountInsuredRecordsAtDate(ctx context.Context, tx *sql.Tx, insuredIfaceObj entity.InsuredInterface, insuredId int64, date time.Time) (int, error) {
	count := 0
	query := generateSelectByDate(insuredIfaceObj, date)
	var re = regexp.MustCompile(`^(SELECT )(.*as max_timestamp)`)
	query = re.ReplaceAllString(query, `${1}count(*) as count`)
	re = regexp.MustCompile(`(?m:GROUP BY insured_id, t2\.id)$`)
	query = re.ReplaceAllString(query, ``)
	err := tx.QueryRowContext(ctx, query, insuredId).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func generateSelectAll(entityType entity.InsuredInterface) (query string) {
	query = ""
	switch entityType.(type) {
	case *entity.Employee:
		query = `SELECT t3.employee_id as id, t3.id AS record_id, t2.insured_id, t3.name, t3.start_date, t3.end_date, t3.record_timestamp, MAX(t3.record_timestamp) as max_timestamp` + "\n" +
			`FROM employees t2` + "\n" +
			`JOIN employees_records t3 ON t2.id = t3.employee_id` + "\n" +
			`GROUP BY t2.id`
	case *entity.Insured:
		query = `SELECT t1.*` + "\n" +
			`FROM insured t1`
	case *entity.Address:
		query = `SELECT t2.*, t2.record_timestamp as max_timestamp` + "\n" +
			`FROM insured_addresses_records t2`
	}
	return query
}
func generateSelectAllRecordsByEntityId(entityType entity.InsuredInterface, entityId int64) (query string) {
	query = ""
	switch entityType.(type) {
	case *entity.Employee:
		query = `SELECT t3.employee_id as id, t3.id AS record_id, t2.insured_id, t3.name, t3.start_date, t3.end_date, t3.record_timestamp, t3.record_timestamp as max_timestamp` + "\n" + // garb
			`FROM employees t2` + "\n" +
			`JOIN employees_records t3 ON t2.id = t3.employee_id` + "\n" +
			`WHERE t3.employee_id = ?`
	case *entity.Insured:
		query = generateSelectAll(&entity.Insured{}) + "\n" +
			`WHERE t1.id = ?`
	case *entity.Address:
		query = generateSelectAll(&entity.Address{}) + "\n" +
			`WHERE t2.id = ?`
	}
	return query
}

func generateSelectByDate(insuredIfaceObj entity.InsuredInterface, date time.Time) (query string) {
	timestamp := date.Unix()
	query = ""
	switch insuredIfaceObj.(type) {
	case *entity.Employee:
		query = `SELECT t3.employee_id as id, t3.id AS record_id, t2.insured_id, t3.name, t3.start_date, t3.end_date, t3.record_timestamp, MAX(t3.record_timestamp) as max_timestamp` + "\n" +
			`FROM insured t1` + "\n" +
			`JOIN employees t2 ON t1.id = t2.insured_id` + "\n" +
			`JOIN employees_records t3 ON t2.id = t3.employee_id` + "\n" +
			`WHERE t3.record_timestamp <= ` + strconv.Itoa(int(timestamp)) + "\n" +
			`AND t1.id = ?` + "\n" +
			`GROUP BY insured_id, t2.id`
	case *entity.Insured:
		query = `SELECT t2.*, MAX(t2.record_timestamp) as max_timestamp` + "\n" +
			`FROM insured t1` + "\n" +
			`JOIN insured_addresses_records t2 ON t1.id = t2.insured_id` + "\n" +
			`WHERE t2.record_timestamp <= ` + strconv.Itoa(int(timestamp)) + "\n" +
			`AND t1.id = ?` + "\n" +
			`GROUP BY insured_id`
	case *entity.Address:
		query = `SELECT t2.*, MAX(t2.record_timestamp) as max_timestamp` + "\n" +
			`FROM insured_addresses_records t2` + "\n" +
			`WHERE t2.record_timestamp <= ` + strconv.Itoa(int(timestamp)) + "\n" +
			`AND t2.insured_id = ?`
	}
	return query
}

// generateSelectByIds can return multiple records
func generateSelectByIds(resourceName entity.InsuredInterface, ids []int64) (query string) {
	idString := idToString(ids)
	query = ""
	switch resourceName.(type) {
	case *entity.Employee:
		// MAX(record_timestamp) + GROUP BY max_record_timestamp gets us the most recent record in Sqlite.
		// This kind of trick does not work in MySQL and probably not in Postgresql.
		query = `SELECT t3.employee_id as id, t3.id AS record_id, t2.insured_id, t3.name, t3.start_date, t3.end_date, t3.record_timestamp, MAX(record_timestamp) AS max_record_timestamp` + "\n" +
			`FROM employees t2` + "\n" +
			`JOIN employees_records t3 ON t2.id = t3.employee_id` + "\n" +
			`WHERE t2.id IN (` + idString + `)` + "\n" +
			`GROUP BY t3.employee_id`
	case *entity.Address:
		query = `SELECT t2.*, MAX(t2.record_timestamp) as max_timestamp` + "\n" +
			`FROM insured_addresses_records t2` + "\n" +
			`WHERE t2.id IN (` + idString + `)`
	case *entity.Insured:
		query = `SELECT * FROM insured WHERE id IN (` + idString + ")"
	default:
		query = ""
	}
	return query
}
func idToString(ids []int64) string {
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ids)), ","), "[]")
}

func (db *DB) DeleteById(ctx context.Context, insuredObj entity.InsuredInterface, id int64) (deletedRecord entity.InsuredInterface, err error) {
	if id == 0 {
		return deletedRecord, ErrRecordIDInvalid
	}
	tx, err := db.db.Begin()
	if err != nil {
		return deletedRecord, err
	}
	defer tx.Rollback()

	tableName := insuredObj.GetIdentTableName()

	query := `DELETE FROM ` + tableName + ` WHERE id = ?`
	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return insuredObj, fmt.Errorf("Server error.")
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return insuredObj, fmt.Errorf("Server error.")
	}
	if rows == 0 {
		return insuredObj, ErrRecordDoesNotExist
	}
	tx.Commit()
	return insuredObj, nil // TODO: fill in insuredObj
}

/*
// Currently handled by Entity
func (db *DB) CreateRecord(ctx context.Context, tableName string, Record) */

// Close closes the database connection.
func (db *DB) Close() error {
	// Cancel background context.
	db.cancel()

	// Close database.
	if db.db != nil {
		return db.db.Close()
	}
	return nil
}

// BeginTx starts a transaction and returns a wrapper Tx type. This type
// provides a reference to the database and a fixed timestamp at the start of
// the transaction. The timestamp allows us to mock time during tests as well.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Return wrapper Tx that includes the transaction start time.
	return &Tx{
		Tx:  tx,
		db:  db,
		now: db.Now().UTC().Truncate(time.Second),
	}, nil
}

// Tx wraps the SQL Tx object to provide a timestamp at the start of the transaction.
type Tx struct {
	*sql.Tx
	db  *DB
	now time.Time
}

// lastInsertID is a helper function for reading the last inserted ID as an int.
func lastInsertID(result sql.Result) (int, error) {
	id, err := result.LastInsertId()
	return int(id), err
}

type ShortTime time.Time

// convert short string time ("2006-01-02") to time.Time
func (n *ShortTime) Scan(value interface{}) error {

	valtypes := map[string]int{"string": 0}
	valtype := reflect.TypeOf(value).String()
	if _, ok := valtypes[valtype]; ok {
		if strval, ok := value.(string); ok {
			newTime, err := time.Parse("2006-01-02", strval)
			*(*time.Time)(n) = newTime
			fmt.Println(fmt.Errorf("Error converting string time to time.Time: %v", err))
			return nil
		} else {
			fmt.Println("not an string")
		}
	}

	if value == nil {
		*(*time.Time)(n) = time.Time{}
		return nil
	} else if value, ok := value.(string); ok {
		*(*time.Time)(n), _ = time.Parse(time.RFC3339, value)
		return nil
	}
	return fmt.Errorf("NullTime: cannot scan to time.Time: %T", value)
}

// NullTime represents a helper wrapper for time.Time. It automatically converts
// time fields to/from RFC 3339 format. Also supports NULL for zero time.
type NullTime time.Time

// Scan reads a time value from the database.
// Maybe better way
func (n *NullTime) Scan(value interface{}) error {
	valtypes := map[string]int{"int": 0, "int32": 1, "int64": 2}
	valtype := reflect.TypeOf(value).String()
	if _, ok := valtypes[valtype]; ok {
		if int64val, ok := value.(int64); ok {
			*(*time.Time)(n) = time.Unix(int64val, 0).UTC()
			return nil
		} else if intval, ok := value.(int32); ok {
			int64val := int64(intval)
			*(*time.Time)(n) = time.Unix(int64val, 0).UTC()
			return nil
		} else if intval, ok := value.(int); ok {
			int64val := int64(intval)
			*(*time.Time)(n) = time.Unix(int64val, 0).UTC()
			return nil
		} else {
			fmt.Println("not an integer type")
		}
	}

	if value == nil {
		*(*time.Time)(n) = time.Time{}
		return nil
	} else if value, ok := value.(string); ok {
		*(*time.Time)(n), _ = time.Parse(time.RFC3339, value)
		return nil
	}
	return fmt.Errorf("NullTime: cannot scan to time.Time: %T", value)
}

// Value formats a time value for the database.
func (n *NullTime) Value() (driver.Value, error) {
	if n == nil || (*time.Time)(n).IsZero() {
		return nil, nil
	}
	return (*time.Time)(n).UTC().Format(time.RFC3339), nil
}

// FormatLimitOffset returns a SQL string for a given limit & offset.
// Clauses are only added if limit and/or offset are greater than zero.
func FormatLimitOffset(limit, offset int) string {
	if limit > 0 && offset > 0 {
		return fmt.Sprintf(`LIMIT %d OFFSET %d`, limit, offset)
	} else if limit > 0 {
		return fmt.Sprintf(`LIMIT %d`, limit)
	} else if offset > 0 {
		return fmt.Sprintf(`OFFSET %d`, offset)
	}
	return ""
}

// FormatError returns err as a WTF error, if possible.
// Otherwise returns the original error.
func FormatError(err error) error {
	if err == nil {
		return nil
	}

	switch err.Error() {
	case "UNIQUE constraint failed: dial_memberships.dial_id, dial_memberships.user_id":
		return entity.Errorf(entity.ECONFLICT, "Dial membership already exists.")
	default:
		return err
	}
}

// logstr is a helper function for printing and returning a string.
// It can be useful for printing out query text.
func logstr(s string) string {
	println(s)
	return s
}
