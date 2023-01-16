package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"

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

	tableNames         map[string]int
	allowedNaturalKeys map[string]int

	Now func() time.Time
}

func NewDB(dsn string) *DB {
	tn := make(map[string]int)
	tn["insured"] = 0
	tn["employees"] = 1
	tn["addresses"] = 2
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

func (db *DB) GetById(ctx context.Context, tableName string, id int64) (record entity.Record, err error) {
	if id == 0 {
		return record, ErrRecordDoesNotExist
	}
	tx, err := db.db.Begin()
	if err != nil {
		return record, err
	}
	defer tx.Rollback()

	fmt.Println("sqlite DB.GetById")

	if _, ok := db.tableNames[tableName]; !ok {
		fmt.Println("tableName", tableName)
		return record, fmt.Errorf("DB - table name doesn't exist.")
	}

	rows, err := tx.QueryContext(ctx, `SELECT * FROM `+tableName+` WHERE id = ?`, id)
	if err != nil {
		fmt.Println("bad query")
		return record, fmt.Errorf("Query failed")
	}

	// https://kylewbanks.com/blog/query-result-to-map-in-golang
	columnNames, err := rows.Columns()
	fmt.Println(columnNames)
	recordMap := map[string]string{}
	m := make(map[string]interface{})
	rowCount := 0
	for rows.Next() {
		columns := make([]interface{}, len(columnNames))
		columnPointers := make([]interface{}, len(columnNames))

		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return record, err
		}
		// Make map, get value for each column
		for i, colName := range columnNames {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}
		rowCount++
		fmt.Print("Incremented rowCount:", rowCount)
	}
	if rowCount == 0 {
		fmt.Println("rowCount:", rowCount, "m:", m, "table:", tableName)
		return record, ErrRecordDoesNotExist
	}
	fmt.Println(recordMap)
	if err := rows.Err(); err != nil {
		return record, fmt.Errorf("rowsErr: %v", err)
	}

	// convert to string map
	data := make(map[string]string)
	for key, value := range m {
		strKey := fmt.Sprintf("%v", key)
		strVal := fmt.Sprintf("%v", value)
		data[strKey] = strVal
	}

	record = entity.Record{
		ID:   int(id),
		Data: data,
	}
	fmt.Println("db:", record)
	tx.Commit()
	return record, err
}

// TODO: can remove naturalKey from signature?
func (db *DB) GetByDate(ctx context.Context, tableName string, naturalKey string, insuredId int64, date time.Time) (record entity.Record, err error) {
	id := insuredId
	if id == 0 {
		return record, ErrRecordDoesNotExist
	}
	if _, ok := db.tableNames[tableName]; !ok {
		fmt.Println("tableName", tableName)
		return record, fmt.Errorf("DB - table name doesn't exist.")
	}
	// TODO: re-enable this
	/* if _, ok := db.allowedNaturalKeys[naturalKey]; !ok {
		fmt.Println("naturalKey", naturalKey)
		return record, fmt.Errorf("DB - key not allowed.")
	} */

	timestamp := date.Unix()
	fmt.Println("DB.GetByDate timestamp", timestamp)
	tx, err := db.db.Begin()
	if err != nil {
		return record, err
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
	fmt.Println(query)
	rows, err := tx.QueryContext(ctx, query, id)
	if err != nil {
		fmt.Println("bad query")
		return record, fmt.Errorf("Query failed")
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
			return record, err
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
		return record, err
	}

	jsonData := map[string]string{}
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
		/* complexRecord.Data["employees"] = map[string][string]{
			"id":id
		} */

		//a := "{'employees' : { "+string(jsonData)+"}"
		recordId, err := strconv.Atoi(data["id"])
		if err != nil {
			return entity.Record{}, FormatError(ErrRecordIDInvalid)
		}
		record = entity.Record{
			ID:   recordId,
			Data: data,
		}
		jD, err := json.Marshal(record)
		if err != nil {
			return entity.Record{}, err
		}
		jDString := string(jD)
		jsonData[strconv.Itoa(record.ID)] = jDString
	}

	collectionJSON, err := json.Marshal(jsonData)
	if err != nil {
		return entity.Record{}, FormatError(err)
	}
	collectionJSONMap := map[string]string{
		tableName: string(collectionJSON),
	}
	record = entity.Record{
		ID:   int(id),
		Data: collectionJSONMap,
	}
	tx.Commit()
	return record, nil
}

func (db *DB) DeleteById(ctx context.Context, tableName string, id int64) (deletedRecord entity.Record, err error) {
	if id == 0 {
		return deletedRecord, ErrRecordIDInvalid
	}
	tx, err := db.db.Begin()
	if err != nil {
		return deletedRecord, err
	}
	defer tx.Rollback()

	fmt.Println("sqlite DB.DeleteById")

	if _, ok := db.tableNames[tableName]; !ok {
		fmt.Println("tableName", tableName)
		return deletedRecord, fmt.Errorf("DB - table name doesn't exist.")
	}
	deletedRecord, err = db.GetById(ctx, tableName, id)
	if err != nil {
		return deletedRecord, entity.Errorf("Failed to get record before deletion. Aborting delete. Error: %v", err.Error())
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM `+tableName+` WHERE id = ?`, id)
	if err != nil {
		fmt.Println("Failed to DELETE record", err)
		return deletedRecord, fmt.Errorf("Failed to DELETE record. Error: %v", err.Error())
	}

	tx.Commit()
	return deletedRecord, nil
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
