package sqlite_test

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/nickcoast/timetravel/entity"
	"github.com/nickcoast/timetravel/sqlite"
)

var dump = flag.Bool("dump", true, "save work data")

// Ensure the test database can open & close.
func TestDB(t *testing.T) {
	db := MustOpenDB(t)
	MustCloseDB(t, db)
}

// MustOpenDB returns a new, open DB. Fatal on error.
func MustOpenDB(tb testing.TB) *sqlite.DB {
	tb.Helper()

	// Write to an in-memory database by default.
	// If the -dump flag is set, generate a temp file for the database.
	//dsn := ":memory:"
	dsn := "file:test.db?cache=shared&mode=rwc&locking_mode=NORMAL&_fk=1&synchronous=2"
	if *dump {
		dir, err := ioutil.TempDir("", "")
		if err != nil {
			tb.Fatal(err)
		}
		dsn = filepath.Join(dir, "db") // TODO: this is dumb. Tosses out my entire DSN
		println("DUMP=" + dsn)
	}

	db := sqlite.NewDB(dsn)
	if err := db.Open(); err != nil {
		tb.Fatal(err)
	}
	return db
}

// MustCloseDB closes the DB. Fatal on error.
func MustCloseDB(tb testing.TB, db *sqlite.DB) {
	tb.Helper()
	if err := db.Close(); err != nil {
		tb.Fatal(err)
	}
}

func TestDB_GetResourceById(tb *testing.T) {
	ctx := context.Background()
	past, err := time.Parse("2006-01-02 15:04:05", "2006-01-02 15:04:05")
	pastTimestamp := past
	db := MustOpenDB(tb)
	defer MustCloseDB(tb, db)
	// Ensure Resource can be gotten by ID
	tb.Run("TestDB_GetResourceById_Insured", func(tb *testing.T) { // TODO: add employees, addresses tests

		//s := sqlite.NewInsuredService(db)

		if err != nil {
			tb.Fatalf("Failed to create past time")
		}

		now := time.Now()
		fmt.Println("pastTimestamp", pastTimestamp, "now", now)
		// Starts at id:3, policy_number 1002
		MustCreateInsured(tb, ctx, db, &entity.Insured{Name: "john" /* PolicyNumber: 500, */, RecordTimestamp: now}) // id: 3
		MustCreateInsured(tb, ctx, db, &entity.Insured{Name: "jane" /* PolicyNumber: 501, */, RecordTimestamp: now})
		MustCreateInsured(tb, ctx, db, &entity.Insured{Name: "frank" /* PolicyNumber: 502, */, RecordTimestamp: now})
		MustCreateInsured(tb, ctx, db, &entity.Insured{Name: "sue" /* PolicyNumber: 503, */, RecordTimestamp: pastTimestamp}) // id: 6, pn: 1005

		pastTimestampString := strconv.FormatInt(pastTimestamp.Unix(), 10)
		insuredID := 6
		m := map[string]string{"id": strconv.Itoa(insuredID), "name": "sue", "policy_number": "1005", "record_timestamp": pastTimestampString}

		wantRecord := entity.Record{
			ID:   int(insuredID),
			Data: m,
		}

		fmt.Println("wantRecord", wantRecord)
		if record, err := db.GetById(ctx, "insured", int64(insuredID)); err != nil {
			tb.Fatal(err)
		} else if got, want := record.ID, insuredID; got != want {
			tb.Fatalf("ID=%v, want %v", got, want)
		} else if got, want := record.ID, wantRecord.ID; !cmp.Equal(got, want) { // ?? why doesn't it pass if I compare the structs??
			tb.Fatalf("No match. got record: %v, want: %v", got, want)
		} else if got, want := record.Data, wantRecord.Data; !cmp.Equal(got, want) {
			tb.Fatalf("No match. got record: %v, want: %v", got, want)
		}

	})
	tb.Run("TestDB_GetResourceById_Insured", func(tb *testing.T) {
		/* db := MustOpenDB(tb)
		defer MustCloseDB(tb, db) */
		eStartDate, err := time.Parse("2006-01-02", "2006-01-02")
		eEndDate2, err := time.Parse("2006-01-02", "2007-07-04")
		emp1 := &entity.Employee{Name: "john", StartDate: eStartDate, InsuredId: 3, RecordTimestamp: pastTimestamp}
		emp2 := &entity.Employee{Name: "Jimmy G", StartDate: eStartDate, EndDate: eEndDate2, InsuredId: 3, RecordTimestamp: pastTimestamp}
		MustCreateEmployee(tb, ctx, db, emp1) // id: 7
		MustCreateEmployee(tb, ctx, db, emp2) // id: 8

		e2Record := emp2.ToRecord()
		record, err := db.GetById(ctx, "employees", int64(8))
		if err != nil {
			tb.Fatal(err)
		}
		if got, want := record, e2Record; !cmp.Equal(got, want) {
			tb.Fatalf("Employees not equal. record: %v e2Record: %v", record, e2Record)
		}
	})
}

func TestDB_GetResourceByDate(tb *testing.T) {
	// Ensure Resource can be gotten by ID
	tb.Run("TestDB_GetResourceByDate_Insured", func(tb *testing.T) { // TODO: add employees, addresses tests
		db := MustOpenDB(tb)
		defer MustCloseDB(tb, db)
		//s := sqlite.NewInsuredService(db)

		ctx := context.Background()
		past, err := time.Parse("2006-01-02 15:04:05", "2006-01-02 15:04:05")
		if err != nil {
			tb.Fatalf("Failed to create past time")
		}
		pastTimestamp := past
		now := time.Now()
		fmt.Println("pastTimestamp", pastTimestamp, "now", now)
		// Starts at id:3, policy_number 1002
		MustCreateInsured(tb, ctx, db, &entity.Insured{Name: "john" /* PolicyNumber: 500, */, RecordTimestamp: now}) // id: 3
		MustCreateInsured(tb, ctx, db, &entity.Insured{Name: "jane" /* PolicyNumber: 501, */, RecordTimestamp: now})
		MustCreateInsured(tb, ctx, db, &entity.Insured{Name: "frank" /* PolicyNumber: 502, */, RecordTimestamp: now})
		sue, _ := MustCreateInsured(tb, ctx, db, &entity.Insured{Name: "sue" /* PolicyNumber: 503, */, RecordTimestamp: pastTimestamp}) // id: 6, pn: 1005
		fmt.Print("SUUUUUUUUUUE", sue)

		/* MustCreateEmployee(tb, ctx, db, employee{Name: "Sue 1,"}) */
		employees, timestamps, ctx := MustCreateEmployees(tb, ctx, db, sue)
		fmt.Println("employees", employees)

		for _, t := range timestamps {
			db.GetByDate(ctx, "employees", "name", int64(sue.ID), t.Add(time.Second*1))
		}

		pastTimestampString := strconv.FormatInt(pastTimestamp.Unix(), 10)
		insuredID := 6
		m := map[string]string{"id": strconv.Itoa(insuredID), "name": "sue", "policy_number": "1005", "record_timestamp": pastTimestampString}

		wantRecord := entity.Record{
			ID:   int(insuredID),
			Data: m,
		}

		fmt.Println("wantRecord", wantRecord)
		if record, err := db.GetByDate(ctx, "employees", "name", int64(insuredID), now); err != nil {
			tb.Fatal(err)
		} else if got, want := record.ID, insuredID; got != want {
			tb.Fatalf("ID=%v, want %v", got, want)
		} else if got, want := record.ID, wantRecord.ID; !cmp.Equal(got, want) { // ?? why doesn't it pass if I compare the structs??
			tb.Fatalf("No match. got record: %v, want: %v", got, want)
		} else if got, want := record.Data, wantRecord.Data; !cmp.Equal(got, want) {
			tb.Fatalf("No match. got record: %v, want: %v", got, want)
		}
	})
}
