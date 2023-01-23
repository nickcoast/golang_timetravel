package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
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
	tn["employees"] = 1
	tn["insured_addresses"] = 2
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
		fmt.Println("bad query")
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

	timestamp := date.Unix()
	fmt.Println("DB.GetByDate timestamp", timestamp)
	tx, err := db.db.Begin()
	if err != nil {
		return insured, err
	}
	defer tx.Rollback()

	fmt.Println("sqlite DB.GetInsuredByDate")

	employeeRecords, err := db.GetByDate(ctx, "employees", "naturalkey", insuredId, date)
	if err != nil {
		return entity.Insured{}, FormatError(err)
	}
	addressRecords, err := db.GetByDate(ctx, "insured_addresses", "naturalkey", insuredId, date)
	if err != nil {
		return entity.Insured{}, FormatError(err)
	}
	insuredIfaceObj, err := db.GetById(ctx, &entity.Insured{}, insuredId)
	if err != nil {
		return entity.Insured{}, FormatError(err)
	}
	employees, err := entity.EmployeesFromRecords(employeeRecords)
	addresses, err := entity.AddressesFromRecords(addressRecords)

	insuredObj, ok := insuredIfaceObj.(*entity.Insured)
	if !ok {
		return entity.Insured{}, fmt.Errorf("Internal Server Error")
	}
	insuredObj.Employees = &employees
	insuredObj.Addresses = &addresses

	tx.Commit()
	return insured, nil
}

// TODO: can remove naturalKey from signature?
func (db *DB) GetByDate(ctx context.Context, tableName string, naturalKey string, insuredId int64, date time.Time) (records map[int]entity.Record, err error) {
	id := insuredId
	if id == 0 {
		return map[int]entity.Record{}, ErrRecordDoesNotExist
	}
	if _, ok := db.tableNames[tableName]; !ok {
		fmt.Println("tableName", tableName)
		return map[int]entity.Record{}, fmt.Errorf("DB - table name doesn't exist.")
	}

	timestamp := date.Unix()
	fmt.Println("DB.GetByDate timestamp", timestamp)
	tx, err := db.db.Begin()
	if err != nil {
		return map[int]entity.Record{}, err
	}
	defer tx.Rollback()

	fmt.Println("sqlite DB.GetByDate")

	groupBy := ""
	if tableName == "employees" {
		groupBy = "t2.name"
	} else {
		groupBy = "t1.id"
	}
	query := `SELECT t2.*, MAX(t2.record_timestamp) as max_timestamp` + "\n" +
		`FROM insured t1` + "\n" +
		`JOIN ` + tableName + ` t2 ON t1.id = t2.insured_id` + "\n" +
		`WHERE t2.record_timestamp <= ` + strconv.Itoa(int(timestamp)) + "\n" +
		`AND t1.id = ?` + "\n" +
		`GROUP BY insured_id, ` + groupBy
	rows, err := tx.QueryContext(ctx, query, id)
	if err != nil {
		fmt.Println("bad query")
		return map[int]entity.Record{}, fmt.Errorf("Query failed")
	}
	// https://kylewbanks.com/blog/query-result-to-map-in-golang
	columnNames, err := rows.Columns()
	fmt.Println(columnNames)
	recordMap := map[string]string{}
	//m := make(map[string]interface{})
	dbRecords := make(map[int]map[string]interface{})
	i := 0
	for rows.Next() {
		columns := make([]interface{}, len(columnNames))
		columnPointers := make([]interface{}, len(columnNames))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		// Scan result into the pointers
		if err := rows.Scan(columnPointers...); err != nil {
			return map[int]entity.Record{}, err
		}
		// Make map, get value for each column
		m := make(map[string]interface{})
		for i, colName := range columnNames {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}
		dbRecords[i] = m
		i++
	}
	fmt.Println(recordMap)
	if err := rows.Err(); err != nil {
		return map[int]entity.Record{}, err
	}

	records = map[int]entity.Record{}
	//complexRecord := entity.Record{}
	for _, n := range dbRecords {
		// convert to string map
		data := make(map[string]string)
		for key, value := range n {
			strVal := fmt.Sprintf("%v", value)
			strKey := fmt.Sprintf("%v", key)
			if strVal != "0001-01-01" && strKey != "max_timestamp" { // skip our fake "NULL" date values and result of sql MAX()
				data[strKey] = strVal
			}
		}
		recordId, err := strconv.Atoi(data["id"])
		if err != nil {
			return map[int]entity.Record{}, FormatError(ErrRecordIDInvalid)
		}
		record := entity.Record{
			ID:   recordId,
			Data: data,
		}
		records[recordId] = record
	}

	fmt.Println(records)
	tx.Commit()
	return records, nil
}

func generateSelectByDate(resourceName string, date time.Time) (query string) {
	timestamp := date.Unix()
	query = ""
	if resourceName == "employee" {
		query = `SELECT t3.employee_id as id, t3.id AS record_id, t2.insured_id, t3.name, t3.employee_id, t3.start_date, t3.end_date, t3.record_timestamp, MAX(t3.record_timestamp) as max_timestamp` + "\n" +
			`FROM insured t1` + "\n" +
			`JOIN employees t2 ON t1.id = t2.insured_id` + "\n" +
			`JOIN employees_records t3 ON t2.id = t3.employee_id` + "\n" +
			`WHERE t3.record_timestamp <= ` + strconv.Itoa(int(timestamp)) + "\n" +
			`AND t1.id = ?` + "\n" +
			`GROUP BY insured_id, t2.id`
	} else if resourceName == "insured" {
		query = `SELECT t2.*, MAX(t2.record_timestamp) as max_timestamp` + "\n" +
			`FROM insured t1` + "\n" +
			`JOIN insured_addresses_records t2 ON t1.id = t2.insured_id` + "\n" +
			`WHERE t2.record_timestamp <=` + strconv.Itoa(int(timestamp)) + "\n" +
			`AND t1.id = ?` + "\n" +
			`GROUP BY insured_id`
	}
	return query
}

// generateSelectByIds can return multiple records
func generateSelectByIds(resourceName entity.InsuredInterface, ids []int64) (query string) {
	idString := idToString(ids)
	query = ""
	switch asdf := resourceName.(type) {
	case *entity.Employee:
		fmt.Println(asdf)
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
	fmt.Println(query)
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
	if val, ok := valtypes[valtype]; ok {
		if int64val, ok := value.(int64); ok {
			*(*time.Time)(n) = time.Unix(int64val, 0).UTC()
			fmt.Println("Int 64 to time")
			return nil
		} else if intval, ok := value.(int32); ok {
			int64val := int64(intval)
			*(*time.Time)(n) = time.Unix(int64val, 0).UTC()
			fmt.Println("Int 32 to time")
			return nil
		} else if intval, ok := value.(int); ok {
			int64val := int64(intval)
			*(*time.Time)(n) = time.Unix(int64val, 0).UTC()
			fmt.Println("int to time", val, valtype, value)
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
