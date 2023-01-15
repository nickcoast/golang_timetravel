package sqlite_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	//"github.com/google/go-cmp/cmp"
	"github.com/nickcoast/timetravel/entity"
	"github.com/nickcoast/timetravel/sqlite"
)

// TODO: delete all employee at end of each test?

func TestEmployeeService_CreateEmployee(t *testing.T) {
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
		u := &entity.Employee{
			Name:            "susy",
			StartDate:       start,
			EndDate:         end,
			InsuredId:       1,
			RecordTimestamp: time.Now().UTC(),
		}

		// Create new employee & verify ID and timestamps are set.
		id, err := s.CreateEmployee(ctx, u)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println("New employee created")

		employeeRec, err := s.Db.GetById(ctx, "employees", int64(id))
		gottenId := int64(employeeRec.ID)

		if err != nil {
			fmt.Println("id", id, "gottenId:", gottenId)
			t.Fatal(err)
		} else if got, want := id, gottenId; got != int64(want) {
			t.Fatalf("ID=%v, want %v", got, want)
		} else if u.RecordTimestamp.IsZero() {
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
		u2 := &entity.Employee{Name: "jane", StartDate: start2, EndDate: end2, InsuredId: 2, RecordTimestamp: time.Now()}
		if _, err := s.CreateEmployee(ctx, u2); err != nil {
			t.Fatal(err)
		} else if got, want := u2.ID, 8; got != want {
			t.Fatalf("ID=%v, want %v", got, want)
		}

		// Fetch employee from database & compare.
		uRecord := u.ToRecord()
		//u2Record := u2.ToRecord()
		if other, err := s.Db.GetById(ctx, "employees", 7); err != nil {
			t.Fatal(err)
			/* else if !cmp.Equal(other, uRecord) {
				t.Fatal("Records are not equal")
			} */
		} else if !reflect.DeepEqual(uRecord, other) {
			t.Fatalf("mismatch: %#v != %#v", uRecord, other)
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
func TestEmployeeService_FindEmployee(t *testing.T) {
	// Ensure an error is returned if fetching a non-existent employee.
	t.Run("ErrNotFound", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		//s := sqlite.NewInsuredService(db)
		if _, err := db.GetById(context.Background(), "employees", 999); err == nil {
			//t.Fatalf("unexpected error: %#v", err)
			t.Fatalf("Should be an error.") // TODO: test for specific error
		}
	})
}

func TestEmployeeService_FindEmployees(t *testing.T) {
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

// MustCreateEmployee creates a employee in the database. Fatal on error.
func MustCreateEmployee(tb testing.TB, ctx context.Context, db *sqlite.DB, employee *entity.Employee) (*entity.Employee, context.Context) {
	tb.Helper()
	if _, err := sqlite.NewInsuredService(db).CreateEmployee(ctx, employee); err != nil {
		tb.Fatal(err)
	}
	return employee, ctx
}

/* func TestEmployeeService_GetMaxPolicyNumber(t *testing.T) {
	// Ensure employees can be fetched by email address.
	t.Run("MaxPolicyNumber", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewEmployeeService(db)

		ctx := context.Background()
		MustCreateEmployee(t, ctx, db, &entity.Employee{Name: "john", PolicyNumber: 500, RecordTimestamp: time.Now().UTC()})
		MustCreateEmployee(t, ctx, db, &entity.Employee{Name: "jane", PolicyNumber: 501, RecordTimestamp: time.Now().UTC()})
		MustCreateEmployee(t, ctx, db, &entity.Employee{Name: "frank", PolicyNumber: 502, RecordTimestamp: time.Now().UTC()})
		MustCreateEmployee(t, ctx, db, &entity.Employee{Name: "sue", PolicyNumber: 1002, RecordTimestamp: time.Now().UTC()})

		tx, err := db.BeginTx(ctx, nil)
		defer tx.Commit()
		if err != nil {
			t.Fatalf("BeginTx failed %v", err)
		}

		maxPolicyNumber := 1002
		mp, err := s.GetMaxPolicyNumber(ctx, tx)

		if got, want := mp, maxPolicyNumber; got != want {
			t.Fatalf("maxp=%v, want %v", got, want)
		}
	})
} */
