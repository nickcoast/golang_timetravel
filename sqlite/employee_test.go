package sqlite_test

import (
	"context"
	"fmt"
	"log"

	//"reflect"
	"testing"
	"time"

	//"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp"
	"github.com/nickcoast/timetravel/entity"
	"github.com/nickcoast/timetravel/sqlite"
)

// TODO: delete all employee at end of each test?

// Test DB record creation
func TestInsuredService_CreateEmployee(t *testing.T) {
	// Ensure employee can be created.
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewInsuredService(db)
		ctx := context.Background()
		start, err := time.Parse("2006-01-02", "2006-01-02")
		if err != nil {
			t.Fatal(err)
		}
		end, err := time.Parse("2006-01-02", "2008-01-02")
		if err != nil {
			t.Fatal(err)
		}
		employee1 := &entity.Employee{
			Name:            "susy",
			StartDate:       start,
			EndDate:         end,
			InsuredId:       1,
			RecordTimestamp: time.Now().UTC(),
		}

		// Create new employee & verify ID and timestamps are set.
		newRecord, err := s.CreateEmployee(ctx, employee1)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println("New employee created")

		employeeRec, err := s.Db.GetById(ctx, &entity.Employee{}, int64(newRecord.ID))
		gottenId := employeeRec.GetId()

		if err != nil {
			fmt.Println("id", newRecord, "gottenId:", gottenId)
			t.Fatal(err)
		} else if got, want := int64(newRecord.ID), gottenId; got != want { // testing reference
			t.Fatalf("ID=%v, want %v", got, want)
		} else if employee1.RecordTimestamp.IsZero() {
			t.Fatal("expected created at")
		}

		// Create second employee
		start2, err := time.Parse("2006-01-02", "2004-01-02")
		if err != nil {
			t.Fatal("time.Parse() error")
		}
		end2, err := time.Parse("2006-01-02", "2009-10-31")
		if err != nil {
			t.Fatal("time.Parse() error")
		}
		employee2 := &entity.Employee{Name: "jane", StartDate: start2, EndDate: end2, InsuredId: 2, RecordTimestamp: time.Now()}
		if _, err := s.CreateEmployee(ctx, employee2); err != nil {
			t.Fatal(err)
		} else if got, want := employee2.ID, 9; got != want {
			t.Fatalf("ID=%v, want %v", got, want)
		}

		// Fetch employee from database & compare.
		uRecord := employee1.ToRecord()
		//u2Record := u2.ToRecord()
		if other, err := s.Db.GetById(ctx, &entity.Employee{}, 8); err != nil {
			t.Fatal(err)
		} else if !cmp.Equal(other, uRecord) {
			fmt.Println()
			t.Fatal("Records are not equal. Go record:", uRecord, " DB record retrieved:", other)
		}
	})

	// Ensure an error is returned if employee name is not set.
	t.Run("ErrNameRequired", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewInsuredService(db)
		if _, err := s.CreateEmployee(context.Background(), &entity.Employee{}); err == nil {
			t.Fatal("expected error")
		} else if entity.ErrorCode(err) != entity.EINVALID || entity.ErrorMessage(err) != `Employee name required.` {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

// Not allowing update of core data in "employee" table for now
// Will induce TimeTravel instead
/* func TestEmployeeService_UpdateEmployee(t *testing.T) {
	// Ensure employee name & email can be updated by current user.
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewEmployeeService(db)
		user0, ctx0 := MustCreateEmployee(t, context.Background(), db, &entity.Employee{
			Name:  "susy",
			Email: "susy@gmail.com",
		})

		// Update user.
		newName, newEmail := "jill", "jill@gmail.com"
		uu, err := s.UpdateEmployee(ctx0, user0.ID, entity.EmployeeUpdate{
			Name:  &newName,
			Email: &newEmail,
		})
		if err != nil {
			t.Fatal(err)
		} else if got, want := uu.Name, "jill"; got != want {
			t.Fatalf("Name=%v, want %v", got, want)
		} else if got, want := uu.Email, "jill@gmail.com"; got != want {
			t.Fatalf("Email=%v, want %v", got, want)
		}

		// Fetch employee from database & compare.
		if other, err := s.FindEmployeeByID(context.Background(), 1); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(uu, other) {
			t.Fatalf("mismatch: %#v != %#v", uu, other)
		}
	})
} */

// TODO: uncomment and check
/* func TestEmployeeService_DeleteEmployee(t *testing.T) {
	// Ensure employee can delete self.
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewEmployeeService(db)
		employee0, ctx0 := MustCreateEmployee(t, context.Background(), db, &entity.Employee{Name: "Johnny Rotten", PolicyNumber: 333, RecordTimestamp: time.Now().UTC(), ID: 666})

		// Delete employee & ensure it is actually gone.
		if err := s.DeleteEmployee(ctx0, employee0.ID); err != nil {
			t.Fatal(err)
		} else if _, err := s.FindEmployeeByID(ctx0, employee0.ID); entity.ErrorCode(err) != entity.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	// Ensure an error is returned if deleting a non-existent employee.
	t.Run("ErrNotFound", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewEmployeeService(db)
		if err := s.DeleteEmployee(context.Background(), 777); entity.ErrorCode(err) != entity.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}
*/
func TestInsuredService_FindEmployee(t *testing.T) {
	// Ensure an error is returned if fetching a non-existent employee.
	t.Run("ErrNotFound", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		//s := sqlite.NewInsuredService(db)
		if _, err := db.GetById(context.Background(), &entity.Employee{}, 999); err == nil {
			//t.Fatalf("unexpected error: %#v", err)
			t.Fatalf("Should be an error.") // TODO: test for specific error
		}
	})
}

func TestInsuredService_FindEmployees(t *testing.T) {
	// Ensure employees can be fetched by email address.
	/* 	t.Run("PolicyNumber", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewInsuredService(db)

		ctx := context.Background()
		MustCreateEmployee(t, ctx, db, &entity.Employee{Name: "john", PolicyNumber: 500, RecordTimestamp: time.Now().UTC()})
		MustCreateEmployee(t, ctx, db, &entity.Employee{Name: "jane", PolicyNumber: 501, RecordTimestamp: time.Now().UTC()})
		MustCreateEmployee(t, ctx, db, &entity.Employee{Name: "frank", PolicyNumber: 502, RecordTimestamp: time.Now().UTC()})
		MustCreateEmployee(t, ctx, db, &entity.Employee{Name: "sue", PolicyNumber: 503, RecordTimestamp: time.Now().UTC()})

		policyNumber := 501
		if a, err := db.GetById(ctx, entity.EmployeeFilter{PolicyNumber: &policyNumber}); err != nil {
			t.Fatal(err)
		} else if got, want := len(a), 1; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		} else if got, want := a[0].Name, "jane"; got != want {
			t.Fatalf("name=%v, want %v", got, want)
		} else if got, want := n, 1; got != want {
			t.Fatalf("n=%v, want %v", got, want)
		}
	}) */
}

func TestInsuredService_CountEmployees(t *testing.T) {
	ctx := context.Background()
	db := MustOpenDB(t)
	defer MustCloseDB(t, db)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	s := sqlite.NewInsuredService(db)

	uTimestamp, _ := time.Parse("2006-01-02 15:04:05", "2023-01-13 12:15:00")
	insured := &entity.Insured{
		Name: "susy",
		/* PolicyNumber:    1002, */ // will be 1002 no matter what is set here because of DB trigger to automatically increment
		RecordTimestamp:             uTimestamp, //time.Now().UTC(),
	}
	MustCreateInsured(t, ctx, db, insured)

	employee1, err := entity.NewEmployee("Billy", "2001-01-01", "2002-01-01", insured.ID, "2006-01-02 12:00:00")
	if err != nil {
		t.Fatalf("Failed to create employee. %v - Error: %v", employee1, err)
	}
	employee2, err := entity.NewEmployee("Billy", "2001-01-01", "2002-06-01", insured.ID, "2006-06-02 12:00:00")
	if err != nil {
		t.Fatalf("Failed to create employee. %v - Error: %v", employee2, err)
	}

	log.Println("TestInsuredService_CountEmployees employee1:", employee1)
	log.Println("TestInsuredService_CountEmployees employee2:", employee2)

	MustCreateEmployee(t, ctx, db, employee1)
	MustCreateEmployee(t, ctx, db, employee2)

	if count, err := s.CountEmployeeRecords(ctx, *employee1); err != nil {
		t.Fatal(err)
	} else if got, want := count, 2; got != want { // should get count of 2 because two records for this employee
		t.Fatalf("Count from DB COUNT(*)=%v, want %v", got, want)
	}

	t.Run("Fail if exists", func(t *testing.T) {

		// exact duplicate
		employee3, err := entity.NewEmployee("Billy", "2001-01-01", "2002-01-01", insured.ID, "2006-01-02 12:00:00")
		MustCreateEmployee(t, ctx, db, employee3)
		if err != nil {
			t.Fatalf("Failed to create employee. %v - Error: %v", employee1, err)
		}
	})
}

// MustCreateEmployee creates a employee in the database. Fatal on error.
func MustCreateEmployee(tb testing.TB, ctx context.Context, db *sqlite.DB, employee *entity.Employee) (*entity.Employee, context.Context) {
	tb.Helper()
	if _, err := sqlite.NewInsuredService(db).CreateEmployee(ctx, employee); err != nil {
		tb.Fatal(err)
	}
	return employee, ctx
}
