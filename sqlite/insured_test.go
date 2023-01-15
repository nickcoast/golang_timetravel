package sqlite_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/nickcoast/timetravel/entity"
	"github.com/nickcoast/timetravel/sqlite"
)

// TODO: delete all insured at end of each test?

func TestInsuredService_CreateInsured(t *testing.T) {
	// Ensure insured can be created.
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewInsuredService(db)

		uTimestamp, _ := time.Parse("2006-01-02 15:04:05", "2023-01-13 12:15:00")
		u := &entity.Insured{
			Name:            "susy",
			PolicyNumber:    1002,       // will be 1002 whatever is set here because of DB trigger to automatically increment
			RecordTimestamp: uTimestamp, //time.Now().UTC(),
		}

		// Create new insured & verify ID and timestamps are set.
		id, policyNumber, err := s.CreateInsured(context.Background(), u)
		fmt.Println("New insured with id:", id, "and policy number:", policyNumber)
		if err != nil {
			t.Fatal(err)
		} else if got, want := u.ID, 3; got != want {
			t.Fatalf("ID=%v, want %v", got, want)
		} else if u.RecordTimestamp.IsZero() {
			t.Fatal("expected created at")
		}

		uid64 := int64(u.ID)
		if got, want := id, uid64; got == want { // confirming these should always be the same
			fmt.Println("id and u.ID are the same:", id)
		}

		// Create second insured with PolicyNumber.
		u2 := &entity.Insured{Name: "jane", PolicyNumber: 1500, RecordTimestamp: time.Now()}
		if _, _, err := s.CreateInsured(context.Background(), u2); err != nil {
			t.Fatal(err)
		} else if got, want := u2.ID, 4; got != want {
			t.Fatalf("ID=%v, want %v", got, want)
		}

		// Fetch insured from database & compare.
		if other, err := s.FindInsuredByID(context.Background(), 3); err != nil {
			t.Fatal(err)
		} else if !cmp.Equal(u, other) {
			t.Fatalf("mismatch: %#v != %#v", u, other)
		}
		/* else if !reflect.DeepEqual(u, other) {
			t.Fatalf("mismatch: %#v != %#v", u, other)
		} */
	})

	// Ensure an error is returned if insured name is not set.
	t.Run("ErrNameRequired", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewInsuredService(db)
		if _, _, err := s.CreateInsured(context.Background(), &entity.Insured{}); err == nil {
			t.Fatal("expected error")
		} else if entity.ErrorCode(err) != entity.EINVALID || entity.ErrorMessage(err) != `Insured name required.` {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

// Not allowing update of core data in "insured" table for now
/* func TestInsuredService_UpdateInsured(t *testing.T) {
	// Ensure insured name & email can be updated by current user.
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewInsuredService(db)
		user0, ctx0 := MustCreateInsured(t, context.Background(), db, &entity.Insured{
			Name:  "susy",
			Email: "susy@gmail.com",
		})

		// Update user.
		newName, newEmail := "jill", "jill@gmail.com"
		uu, err := s.UpdateInsured(ctx0, user0.ID, entity.InsuredUpdate{
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

		// Fetch insured from database & compare.
		if other, err := s.FindInsuredByID(context.Background(), 1); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(uu, other) {
			t.Fatalf("mismatch: %#v != %#v", uu, other)
		}
	})
} */

// TODO: uncomment and check
/* func TestInsuredService_DeleteInsured(t *testing.T) {
	// Ensure insured can delete self.
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewInsuredService(db)
		insured0, ctx0 := MustCreateInsured(t, context.Background(), db, &entity.Insured{Name: "Johnny Rotten", PolicyNumber: 333, RecordTimestamp: time.Now().UTC(), ID: 666})

		// Delete insured & ensure it is actually gone.
		if err := s.DeleteInsured(ctx0, insured0.ID); err != nil {
			t.Fatal(err)
		} else if _, err := s.FindInsuredByID(ctx0, insured0.ID); entity.ErrorCode(err) != entity.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	// Ensure an error is returned if deleting a non-existent insured.
	t.Run("ErrNotFound", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewInsuredService(db)
		if err := s.DeleteInsured(context.Background(), 777); entity.ErrorCode(err) != entity.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}
*/
func TestInsuredService_FindInsured(t *testing.T) {
	// Ensure an error is returned if fetching a non-existent insured.
	t.Run("ErrNotFound", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewInsuredService(db)
		if _, err := s.FindInsuredByID(context.Background(), 1111); err == nil { // TODO: entity.ErrorCode(err) != entity.ENOTFOUND
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func TestInsuredService_FindInsureds(t *testing.T) {
	// Ensure insureds can be fetched by email address.
	t.Run("PolicyNumber", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewInsuredService(db)

		ctx := context.Background()
		// PolicyNumbers created automatically by Sqlite trigger
		MustCreateInsured(t, ctx, db, &entity.Insured{Name: "john" /* PolicyNumber: 1002, */, RecordTimestamp: time.Now().UTC()})
		MustCreateInsured(t, ctx, db, &entity.Insured{Name: "jane" /* PolicyNumber: 1003, */, RecordTimestamp: time.Now().UTC()})
		MustCreateInsured(t, ctx, db, &entity.Insured{Name: "frank" /* PolicyNumber: 1004, */, RecordTimestamp: time.Now().UTC()})
		MustCreateInsured(t, ctx, db, &entity.Insured{Name: "sue" /* PolicyNumber: 1005, */, RecordTimestamp: time.Now().UTC()}) // PolicyNumber 1005

		policyNumber := 1003
		if a, n, err := s.FindInsureds(ctx, entity.InsuredFilter{PolicyNumber: &policyNumber}); err != nil {
			t.Fatal(err)
		} else if got, want := len(a), 1; got != want {
			t.Fatalf("len=%v, want %v", got, want)
		} else if got, want := a[0].Name, "jane"; got != want {
			t.Fatalf("name=%v, want %v", got, want)
		} else if got, want := n, 1; got != want {
			t.Fatalf("n=%v, want %v", got, want)
		}
	})
}

// MustCreateInsured creates a insured in the database. Fatal on error.
func MustCreateInsured(tb testing.TB, ctx context.Context, db *sqlite.DB, insured *entity.Insured) (*entity.Insured, context.Context) {
	tb.Helper()
	if _, _, err := sqlite.NewInsuredService(db).CreateInsured(ctx, insured); err != nil {
		tb.Fatal(err)
	}
	return insured, ctx
}

/* func TestInsuredService_GetMaxPolicyNumber(t *testing.T) {
	// Ensure insureds can be fetched by email address.
	t.Run("MaxPolicyNumber", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewInsuredService(db)

		ctx := context.Background()
		MustCreateInsured(t, ctx, db, &entity.Insured{Name: "john", PolicyNumber: 500, RecordTimestamp: time.Now().UTC()})
		MustCreateInsured(t, ctx, db, &entity.Insured{Name: "jane", PolicyNumber: 501, RecordTimestamp: time.Now().UTC()})
		MustCreateInsured(t, ctx, db, &entity.Insured{Name: "frank", PolicyNumber: 502, RecordTimestamp: time.Now().UTC()})
		MustCreateInsured(t, ctx, db, &entity.Insured{Name: "sue", PolicyNumber: 1002, RecordTimestamp: time.Now().UTC()})

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
