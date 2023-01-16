package sqlite_test

import (
	"context"
	"fmt"

	//"reflect"
	"testing"
	"time"

	//"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp"
	"github.com/nickcoast/timetravel/entity"
	"github.com/nickcoast/timetravel/sqlite"
)

func TestAddressService_CreateAddress(t *testing.T) {
	// Ensure address can be created.
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewInsuredService(db)
		ctx := context.Background()
		address1 := &entity.Address{
			Address:         "1600 Pennsylvania Avenue N.W.",
			InsuredId:       1,
			RecordTimestamp: time.Now().UTC(),
		}

		// Create new address & verify ID and timestamps are set.
		newRecord, err := s.CreateAddress(ctx, address1)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println("New address created")

		addressRec, err := s.Db.GetById(ctx, "insured_addresses", int64(newRecord.ID))
		gottenId := int64(addressRec.ID)

		if err != nil {
			fmt.Println("id", newRecord, "gottenId:", gottenId)
			t.Fatal(err)
		} else if got, want := int64(newRecord.ID), gottenId; got != want {
			t.Fatalf("ID=%v, want %v", got, want)
		} else if address1.RecordTimestamp.IsZero() {
			t.Fatal("expected created at")
		}

		address2 := &entity.Address{Address: "420 Wobegone Place", InsuredId: 2, RecordTimestamp: time.Now()}
		if _, err := s.CreateAddress(ctx, address2); err != nil {
			t.Fatal(err)
		} else if got, want := address2.ID, 6; got != want {
			t.Fatalf("ID=%v, want %v", got, want)
		}

		// Fetch address from database & compare.
		uRecord := address1.ToRecord()
		//u2Record := u2.ToRecord()
		if other, err := s.Db.GetById(ctx, "insured_addresses", 5); err != nil {
			t.Fatal(err)
		} else if !cmp.Equal(other, uRecord) {
			t.Fatal("Records are not equal")
		}
	})

	// Ensure an error is returned if address name is not set.
	t.Run("ErrNameRequired", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewInsuredService(db)
		if _, err := s.CreateAddress(context.Background(), &entity.Address{}); err == nil {
			t.Fatal("expected error")
		} else if entity.ErrorCode(err) != entity.EINVALID || entity.ErrorMessage(err) != `Address required.` {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

// Not allowing update of core data in "address" table for now
// Will induce TimeTravel instead
/* func TestAddressService_UpdateAddress(t *testing.T) {
	// Ensure address name & email can be updated by current user.
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewAddressService(db)
		user0, ctx0 := MustCreateAddress(t, context.Background(), db, &entity.Address{
			Name:  "susy",
			Email: "susy@gmail.com",
		})

		// Update user.
		newName, newEmail := "jill", "jill@gmail.com"
		uu, err := s.UpdateAddress(ctx0, user0.ID, entity.AddressUpdate{
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

		// Fetch address from database & compare.
		if other, err := s.FindAddressByID(context.Background(), 1); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(uu, other) {
			t.Fatalf("mismatch: %#v != %#v", uu, other)
		}
	})
} */

// TODO: uncomment and check
/* func TestAddressService_DeleteAddress(t *testing.T) {
	// Ensure address can delete self.
	t.Run("OK", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewAddressService(db)
		address0, ctx0 := MustCreateAddress(t, context.Background(), db, &entity.Address{Name: "Johnny Rotten", PolicyNumber: 333, RecordTimestamp: time.Now().UTC(), ID: 666})

		// Delete address & ensure it is actually gone.
		if err := s.DeleteAddress(ctx0, address0.ID); err != nil {
			t.Fatal(err)
		} else if _, err := s.FindAddressByID(ctx0, address0.ID); entity.ErrorCode(err) != entity.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	// Ensure an error is returned if deleting a non-existent address.
	t.Run("ErrNotFound", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewAddressService(db)
		if err := s.DeleteAddress(context.Background(), 777); entity.ErrorCode(err) != entity.ENOTFOUND {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}
*/
func TestAddressService_FindAddress(t *testing.T) {
	// Ensure an error is returned if fetching a non-existent address.
	t.Run("ErrNotFound", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		//s := sqlite.NewInsuredService(db)
		if _, err := db.GetById(context.Background(), "insured_addresses", 999); err == nil {
			//t.Fatalf("unexpected error: %#v", err)
			t.Fatalf("Should be an error.") // TODO: test for specific error
		}
	})
}

func TestAddressService_FindAddresss(t *testing.T) {
	// Ensure addresses can be fetched by email address.
	/* 	t.Run("PolicyNumber", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewInsuredService(db)

		ctx := context.Background()
		MustCreateAddress(t, ctx, db, &entity.Address{Name: "john", PolicyNumber: 500, RecordTimestamp: time.Now().UTC()})
		MustCreateAddress(t, ctx, db, &entity.Address{Name: "jane", PolicyNumber: 501, RecordTimestamp: time.Now().UTC()})
		MustCreateAddress(t, ctx, db, &entity.Address{Name: "frank", PolicyNumber: 502, RecordTimestamp: time.Now().UTC()})
		MustCreateAddress(t, ctx, db, &entity.Address{Name: "sue", PolicyNumber: 503, RecordTimestamp: time.Now().UTC()})

		policyNumber := 501
		if a, err := db.GetById(ctx, entity.AddressFilter{PolicyNumber: &policyNumber}); err != nil {
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

// MustCreateAddress creates a address in the database. Fatal on error
func MustCreateAddress(tb testing.TB, ctx context.Context, db *sqlite.DB, address *entity.Address) (newAddress entity.Address, c context.Context) {
	tb.Helper()
	record, err := sqlite.NewInsuredService(db).CreateAddress(ctx, address)
	if err != nil {
		tb.Fatal(err)
	}
	fmt.Println(record)
	newAddress.FromRecord(record)
	return newAddress, ctx
}
