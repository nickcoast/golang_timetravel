package sqlite

import (
	"context"
	"database/sql"
	"reflect"

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

	Now func() time.Time
}

func NewDB(dsn string) *DB {
	db := &DB{
		DSN: dsn,
		Now: time.Now,
		/*
			EventService: wtf.NopEventService(), */
	}
	db.ctx, db.cancel = context.WithCancel((context.Background())) // new context?
	return db
}

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
