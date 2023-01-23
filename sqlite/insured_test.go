package sqlite_test

import (
	"context"
	"fmt"
	"strconv"
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
			Name: "susy",
			/* PolicyNumber:    1002, */ // will be 1002 no matter what is set here because of DB trigger to automatically increment
			RecordTimestamp:             uTimestamp, //time.Now().UTC(),
		}

		// Create new insured & verify ID and timestamps are set.
		newRecord, err := s.CreateInsured(context.Background(), u)
		fmt.Println("New insured with id:", newRecord.ID, "and data (should include policy number):", newRecord.Data)
		if err != nil {
			t.Fatal(err)
		} else if got, want := newRecord.ID, 3; got != want {
			t.Fatalf("ID=%v, want %v", got, want)
		}
		newTimestamp, err := strconv.Atoi(newRecord.Data["record_timestamp"])
		if newTimestamp == 0 {
			t.Fatal("Invalid timestamp created: ", newTimestamp, " - Error:", err)
		}

		policyNumber, err := strconv.Atoi(newRecord.Data["policy_number"])
		expectedPolicyNumber := 1002
		if err != nil {
			t.Fatal(err)
		}
		if got, want := policyNumber, expectedPolicyNumber; got != want { // confirming these should always be the same
			t.Fatal("Unexpected policy_number created: ", policyNumber, " - Expected: ", expectedPolicyNumber)
		}

		// Create second insured with PolicyNumber.
		u2 := &entity.Insured{Name: "jane" /* PolicyNumber: 1003, */, RecordTimestamp: time.Now()}
		expectedPolicyNumber2 := 1003
		newRecord2, err := s.CreateInsured(context.Background(), u2)
		if err != nil {
			t.Fatal(err)
		} else if got, want := newRecord2.ID, 4; got != want {
			t.Fatalf("ID=%v, want %v", got, want)
		}
		policyNumber2, err := strconv.Atoi(newRecord2.Data["policy_number"])
		if err != nil {
			t.Fatal(err)
		}
		if got, want := policyNumber2, expectedPolicyNumber2; got != want { // confirming these should always be the same
			t.Fatal("Unexpected policy_number created: ", policyNumber2, " - Expected: ", expectedPolicyNumber2)
		}
		// Fetch insured from database & compare.
		if other, err := db.GetById(context.Background(), &entity.Insured{}, 3); err != nil {
			//if other, err := db.GetById(context.Background(), "insured", 3); err != nil {
			t.Fatal(err)
		} else if !cmp.Equal(newRecord, other) {
			t.Fatalf("mismatch: %#v != %#v", newRecord, other)
		}
	})

	// Ensure an error is returned if insured name is not set.
	t.Run("ErrNameRequired", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)
		s := sqlite.NewInsuredService(db)
		if _, err := s.CreateInsured(context.Background(), &entity.Insured{}); err == nil {
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
		if _, err := s.Db.GetById(context.Background(), &entity.Insured{}, 1111); err == nil { // TODO: entity.ErrorCode(err) != entity.ENOTFOUND
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
		MustCreateInsured(t, ctx, db, &entity.Insured{Name: "sue" /* PolicyNumber: 1005, */, RecordTimestamp: time.Now().UTC()}) // PolicyNumber 1005. id 6?

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

func MustCreateEmployees(tb testing.TB, ctx context.Context, db *sqlite.DB, insured entity.Insured) ([]*entity.Employee, map[int]time.Time, context.Context) {
	tb.Helper()
	timestampFirst, _ := time.Parse("2006-01-02 15:04:05", "2022-01-02 15:04:05")
	timestampSecond := timestampFirst.Add(time.Hour * 24 * 30)
	timestampThird := timestampSecond.Add(time.Hour * 24 * 30) // TODO: add "NULL" date for employee2 first record
	//timestampFourth := timestampThird.Add(time.Hour * 24 * 30)

	startDateOrig := timestampFirst
	startDateUpdate := timestampFirst.Add(time.Hour * 24 * 30) // update to 30 days later
	endDateOrig := startDateOrig.Add(time.Hour * 24 * 365)     //
	endDateUpdate := endDateOrig.Add(time.Hour * 24 * 30)      //update to 30 days later

	employeeName := insured.Name + " 1"
	employeeName2 := insured.Name + " 2"
	insuredId := insured.ID

	var employees = []*entity.Employee{
		{
			Name:            employeeName,
			StartDate:       startDateOrig,
			EndDate:         endDateOrig,
			InsuredId:       insuredId,
			RecordTimestamp: timestampFirst,
		},
		{
			Name:            employeeName,
			StartDate:       startDateUpdate, // UPDATE
			EndDate:         endDateOrig,     // SAME
			InsuredId:       insuredId,
			RecordTimestamp: timestampSecond,
		},
		{
			Name:            employeeName,
			StartDate:       startDateUpdate,
			EndDate:         endDateUpdate, // UPDATE
			InsuredId:       insuredId,
			RecordTimestamp: timestampThird,
		},
		{
			Name:            employeeName2,
			StartDate:       startDateOrig.Add(time.Hour * -24),
			EndDate:         endDateOrig,
			InsuredId:       insuredId,
			RecordTimestamp: timestampFirst.Add(time.Hour * -24),
		},
		{
			Name:            employeeName2,
			StartDate:       startDateUpdate, // UPDATE
			EndDate:         endDateOrig,     // SAME
			InsuredId:       insuredId,
			RecordTimestamp: timestampSecond,
		},
	}

	for _, e := range employees {
		MustCreateEmployee(tb, ctx, db, e)
	}

	timestamps := map[int]time.Time{ // for return
		0: timestampFirst,
		1: timestampSecond,
		2: timestampThird,
		3: timestampFirst.Add(time.Hour * -24),
	}
	return employees, timestamps, ctx
}

func MustCreateInsuredAddresses(tb testing.TB, ctx context.Context, db *sqlite.DB, insured entity.Insured) ([]*entity.Address, map[int]time.Time, context.Context) {
	tb.Helper()
	// 1st timestamp same as in employees. 2nd +1 and 3rd -1
	timestampFirst, _ := time.Parse("2006-01-02 15:04:05", "2022-01-02 15:04:05")
	timestampSecond := timestampFirst.Add(time.Hour * 24 * 30 +1)
	timestampThird := timestampSecond.Add(time.Hour * 24 * 30 -1) 
	//timestampFourth := timestampThird.Add(time.Hour * 24 * 30)

	addressOrig := "123 Fake Street, Springfield, Oregon"
	addressUpdate := "The Shining City On The Hill"
	addressThird := "Atlantis"
	addressFourth := "A Van Down By The River"
	insuredId := insured.ID
	
	var addresss = []*entity.Address{
		{
			Address:         addressOrig,
			InsuredId:       insuredId,
			RecordTimestamp: timestampFirst,
		},
		{			
			Address:      	 addressUpdate,			
			InsuredId:       insuredId,
			RecordTimestamp: timestampSecond,
		},
		{			
			Address:      	 addressThird,			
			InsuredId:       insuredId,
			RecordTimestamp: timestampThird,
		},
		{			
			Address:       	 addressFourth,			
			InsuredId:       insuredId,
			RecordTimestamp: timestampFirst.Add(time.Hour * -24),
		},
	
	}
	for _, e := range addresss {
		MustCreateAddress(tb, ctx, db, e)
	}

	timestamps := map[int]time.Time{ // for return
		0: timestampFirst,
		1: timestampSecond,
		2: timestampThird,
		3: timestampFirst.Add(time.Hour * -24),
	}
	return addresss, timestamps, ctx
}

// MustCreateInsured creates a insured in the database. Fatal on error.
// Returns from newly created Insured in DB
func MustCreateInsured(tb testing.TB, ctx context.Context, db *sqlite.DB, insured *entity.Insured) (newInsured entity.Insured, c context.Context) {
	tb.Helper()
	record, err := sqlite.NewInsuredService(db).CreateInsured(ctx, insured)
	if err != nil {
		tb.Fatal(err)
	}
	fmt.Println(record)
	newInsured.FromRecord(record)
	return newInsured, ctx
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

func TestInsuredService_GetInsuredByDate(tb *testing.T) {
	// Ensure Resource can be gotten by ID
	tb.Run("TestInsuredService_GetInsuredByDat", func(tb *testing.T) { // TODO: add employees, addresses tests
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

		insureds := map[int]entity.Insured{}
		for _, t := range timestamps {
			insured, err := db.GetInsuredByDate(ctx, int64(sue.ID), t.Add(time.Second*1)) // make this so each call should get ONE record?
			if err != nil {
				tb.Fatalf("Failed to GetInsuredByDate")
			}
			insureds[insured.ID] = insured
		}

		pastTimestampString := strconv.FormatInt(pastTimestamp.Unix(), 10)
		insuredID := 6
		m := map[string]string{"id": strconv.Itoa(insuredID), "name": "sue", "policy_number": "1005", "record_timestamp": pastTimestampString}

		wantRecord := entity.Record{
			ID:   int(insuredID),
			Data: m,
		}

		fmt.Println("wantRecord", wantRecord)
		/* if record, err := db.GetByDate(ctx, "employees", "name", int64(insuredID), now); err != nil {
			tb.Fatal(err)
		} else if got, want := record.ID, insuredID; got != want {
			tb.Fatalf("ID=%v, want %v", got, want)
		} else if got, want := record.ID, wantRecord.ID; !cmp.Equal(got, want) { // ?? why doesn't it pass if I compare the structs??
			tb.Fatalf("No match. got record: %v, want: %v", got, want)
		} else if got, want := record.Data, wantRecord.Data; !cmp.Equal(got, want) {
			tb.Fatalf("No match. got record: %v, want: %v", got, want)
		} */
	})
}
