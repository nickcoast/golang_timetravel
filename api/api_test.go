package api_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/nickcoast/timetravel/api"
	"github.com/nickcoast/timetravel/service"
	"github.com/nickcoast/timetravel/sqlite"
)

type APItest struct {
	db         *sqlite.DB
	HTTPServer *http.Server
	tempDir    *string
}

/* func TestGetById(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/asdf", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	apitest.New(). // configuration
			HandlerFunc(handler).
			Get("/message"). // request
			Expect(t).       // expectations
			Body(`{"message": "hello"}`).
			Status(http.StatusOK).
			End()
} */

/* func executeRequest(req *http.Request) *httptest.ResponseRecorder {
    rr := httptest.NewRecorder()
    //a.Router.ServeHTTP(rr, req)
	http.NewServeMux().ServeHTTP(rr, req)
    return rr
} */

// TODO: can test a bunch of GET requests in one test. New, Update, Delete can be separate

func TestAPI_InvalidRequest_Get(t *testing.T) {
	_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
	defer MustCloseDB(t, db)

	t.Run("Path", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v2/bad_path/id/2", nil)
		expectedResponseCode := http.StatusBadRequest
		expectedResponseString := `{"error":"` + api.ErrInvalidEndpoint.Error() + "bad_path\"}\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})
	t.Run("Action", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v2/employee/asdf/2", nil)
		expectedResponseCode := http.StatusNotFound
		expectedResponseString := `404 page not found` + "\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})
	t.Run("API", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v999/address/id/1", nil)
		expectedResponseCode := http.StatusNotFound
		expectedResponseString := `404 page not found` + "\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})

}

func TestAPI_GetById(t *testing.T) {
	_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
	defer MustCloseDB(t, db)

	t.Run("Insured", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v2/insured/id/2", nil)
		expectedResponseCode := http.StatusOK
		expectedResponseString := `{"id":"2","name":"John Smith","policyNumber":"1001","recordTimestamp":"946684799","recordDateTime":"Fri, 31 Dec 1999 23:59:59 UTC","employees":null,"insuredAddresses":null}` + "\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})
	t.Run("Employee", func(t *testing.T) { // should get latest Mister Bungle record. lol
		req, _ := http.NewRequest("GET", "/api/v2/employee/id/2", nil)
		expectedResponseCode := http.StatusOK
		expectedResponseString := `{"id":"2","name":"Mister Bungle","startDate":"1984-11-10","endDate":"1996-06-01","insuredId":"1","recordTimestamp":"852206400","recordDateTime":"Thu, 02 Jan 1997 12:00:00 UTC"}` + "\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})
	t.Run("Employee_EmptyEndDate", func(t *testing.T) { // should get latest Mister Bungle record. lol
		req, _ := http.NewRequest("GET", "/api/v2/employee/id/1", nil)
		expectedResponseCode := http.StatusOK
		expectedResponseString := `{"id":"1","name":"Jimmy Temelpa","startDate":"1984-10-01","endDate":"","insuredId":"1","recordTimestamp":"468072000","recordDateTime":"Wed, 31 Oct 1984 12:00:00 UTC"}` + "\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})
	t.Run("Address", func(t *testing.T) { // should get 123 Fake Street, Springfield, Oregon
		req, _ := http.NewRequest("GET", "/api/v2/address/id/1", nil)
		expectedResponseCode := http.StatusOK
		expectedResponseString := `{"id":"1","address":"123 Fake Street, Springfield, Oregon","recordTimestamp":"468072000","recordDateTime":"Wed, 31 Oct 1984 12:00:00 UTC"}` + "\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})
}

func TestAPI_InvalidRequest_Put(t *testing.T) {

	t.Run("SQL_DROP_DATABASE", func(t *testing.T) {
		_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
		defer MustCloseDB(t, db)
		req, _ := http.NewRequest("PUT", "/api/v2/employee/update", nil)
		requestBody := map[string]string{
			"employeeId": "1",
			"name":       "DROP DATABASE;", // existing record
			"insuredId":  "1",
			"startDate":  "2006-01-02",
		}
		expectedResponseCode := http.StatusOK
		expectedResponseString := `{"id":1,"data":{"endDate":"","id":"1","insuredId":"1","name":"DROP DATABASE;","recordTimestamp":"","startDate":"2006-01-02"}}` + "\n"
		checkResponse(t, req, httpserver, requestBody, expectedResponseCode, expectedResponseString)
	})
	t.Run("SQL_DELETE_FROM_INSURED", func(t *testing.T) {
		_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
		defer MustCloseDB(t, db)
		req, _ := http.NewRequest("PUT", "/api/v2/employee/update", nil)
		requestBody := map[string]string{
			"employeeId": "1",
			"insuredId":  "1", // existing record
			"name":       "DELETE FROM insured;",
			"startDate":  "1000-01-01",
			"endDate":    "1420-04-20",
		}
		expectedResponseCode := http.StatusOK
		expectedResponseString := `{"id":1,"data":{"endDate":"1420-04-20","id":"1","insuredId":"1","name":"DELETE FROM insured;","recordTimestamp":"","startDate":"1000-01-01"}}` + "\n"
		checkResponse(t, req, httpserver, requestBody, expectedResponseCode, expectedResponseString)

		/* req, _ = http.NewRequest("GET", "/api/v2/insured/id/1", nil)
		expectedResponseCode = http.StatusOK
		checkResponse(t, req, httpserver, requestBody, expectedResponseCode, expectedResponseString) */
	})
}

func TestAPI_GetById_ShouldFail_NotFound(t *testing.T) {
	_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
	defer MustCloseDB(t, db)

	t.Run("TestAPI_GetById_Insured_ShouldFail_NotFound", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v2/insured/id/99", nil)
		expectedResponseCode := http.StatusNotFound
		expectedResponseString := `{"error":"record of id 99 does not exist"}` + "\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})
	t.Run("TestAPI_GetById_Employee_ShouldFail_NotFound", func(t *testing.T) { // should get latest Mister Bungle record. lol
		req, _ := http.NewRequest("GET", "/api/v2/employee/id/99", nil)
		expectedResponseCode := http.StatusNotFound
		expectedResponseString := `{"error":"record of id 99 does not exist"}` + "\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})
	t.Run("TestAPI_GetById_Address_ShouldFail_NotFound", func(t *testing.T) { // should get 123 Fake Street, Springfield, Oregon
		req, _ := http.NewRequest("GET", "/api/v2/address/id/99", nil)
		expectedResponseCode := http.StatusNotFound
		expectedResponseString := `{"error":"record of id 99 does not exist"}` + "\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})
}

// TestAPI_GetByTime
func TestAPI_GetByTime(t *testing.T) {
	_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
	defer MustCloseDB(t, db)
	t.Run("TestAPI_GetByTime_Timestamp", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v2/insured/getbytimestamp/2/954590400", nil) // 2000-04-01
		expectedResponseCode := http.StatusOK
		expectedResponseString := `{"id":"2","name":"John Smith","policyNumber":"1001","recordTimestamp":"946684799","recordDateTime":"Fri, 31 Dec 1999 23:59:59 UTC","employees":{"0":{"id":"3","name":"John Smith","startDate":"1985-05-15","endDate":"1999-12-25","insuredId":"2","recordTimestamp":"946684799","recordDateTime":"Fri, 31 Dec 1999 23:59:59 UTC"},"1":{"id":"4","name":"Jane Doe","startDate":"1985-05-15","endDate":"1999-12-25","insuredId":"2","recordTimestamp":"954590400","recordDateTime":"Sat, 01 Apr 2000 12:00:00 UTC"},"2":{"id":"5","name":"Grant Tombly","startDate":"1985-05-15","endDate":"1999-12-25","insuredId":"2","recordTimestamp":"954590400","recordDateTime":"Sat, 01 Apr 2000 12:00:00 UTC"}},"insuredAddresses":{}}` + "\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})
	t.Run("TestAPI_GetByTime_Date", func(t *testing.T) { // 2000-04-01
		req, _ := http.NewRequest("GET", "/api/v2/insured/getbydate/2/2000-04-01", nil)
		expectedResponseCode := http.StatusOK
		expectedResponseString := `{"id":"2","name":"John Smith","policyNumber":"1001","recordTimestamp":"946684799","recordDateTime":"Fri, 31 Dec 1999 23:59:59 UTC","employees":{"0":{"id":"3","name":"John Smith","startDate":"1985-05-15","endDate":"1999-12-25","insuredId":"2","recordTimestamp":"946684799","recordDateTime":"Fri, 31 Dec 1999 23:59:59 UTC"},"1":{"id":"4","name":"Jane Doe","startDate":"1985-05-15","endDate":"1999-12-25","insuredId":"2","recordTimestamp":"954590400","recordDateTime":"Sat, 01 Apr 2000 12:00:00 UTC"},"2":{"id":"5","name":"Grant Tombly","startDate":"1985-05-15","endDate":"1999-12-25","insuredId":"2","recordTimestamp":"954590400","recordDateTime":"Sat, 01 Apr 2000 12:00:00 UTC"}},"insuredAddresses":{}}` + "\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})
	t.Run("TestAPI_GetByTime_Date_NotFound", func(t *testing.T) { // non-existent insuredId. // 2000-04-01
		req, _ := http.NewRequest("GET", "/api/v2/insured/getbydate/99/2000-04-01", nil)
		expectedResponseCode := http.StatusNotFound
		expectedResponseString := `{"error":"No record for Insured 99 and date 2000-04-01 exist"}` + "\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})

	t.Run("TestAPI_GetByTime_Date_OldDate", func(t *testing.T) { // 1000 A.D. - will still return insured with date, but no employees, address
		req, _ := http.NewRequest("GET", "/api/v2/insured/getbydate/2/1000-04-01", nil)
		expectedResponseCode := http.StatusOK
		expectedResponseString := `{"id":"2","name":"John Smith","policyNumber":"1001","recordTimestamp":"946684799","recordDateTime":"Fri, 31 Dec 1999 23:59:59 UTC","employees":{},"insuredAddresses":{}}` + "\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})
	t.Run("TestAPI_GetByTime_Timestamp", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v2/insured/getbytimestamp/2/954590400", nil) // 2000-04-01
		expectedResponseCode := http.StatusOK
		expectedResponseString := `{"id":"2","name":"John Smith","policyNumber":"1001","recordTimestamp":"946684799","recordDateTime":"Fri, 31 Dec 1999 23:59:59 UTC","employees":{"0":{"id":"3","name":"John Smith","startDate":"1985-05-15","endDate":"1999-12-25","insuredId":"2","recordTimestamp":"946684799","recordDateTime":"Fri, 31 Dec 1999 23:59:59 UTC"},"1":{"id":"4","name":"Jane Doe","startDate":"1985-05-15","endDate":"1999-12-25","insuredId":"2","recordTimestamp":"954590400","recordDateTime":"Sat, 01 Apr 2000 12:00:00 UTC"},"2":{"id":"5","name":"Grant Tombly","startDate":"1985-05-15","endDate":"1999-12-25","insuredId":"2","recordTimestamp":"954590400","recordDateTime":"Sat, 01 Apr 2000 12:00:00 UTC"}},"insuredAddresses":{}}` + "\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})
}

func TestAPI_DeleteById(t *testing.T) {
	t.Run("TestAPI_DeleteById_Insured", func(t *testing.T) {
		_, httpserver, db := MustOpenDBAndSetUpRoutes(t) // Load new test DB for each sub-test that alters the DB
		defer MustCloseDB(t, db)

		// 1.)
		req, _ := http.NewRequest("DELETE", "/api/v2/employees/delete/4", nil)
		expectedResponseCode := http.StatusOK
		expectedResponseString := `{"id":"4","name":"Jane Doe","startDate":"1985-05-15","endDate":"1999-12-25","insuredId":"2","recordTimestamp":"954590400","recordDateTime":"Sat, 01 Apr 2000 12:00:00 UTC"}` + "\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)

		// 2.) CONFIRM DELETED. 2nd request should return 404
		expectedResponseCode = http.StatusNotFound
		expectedResponseString = "{\"error\":\"Cannot delete. Record does not exist.\"}\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)

	})
	t.Run("TestAPI_DeleteById_Employee", func(t *testing.T) {
		_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
		defer MustCloseDB(t, db)

		// 1.) DELETE
		req, _ := http.NewRequest("DELETE", "/api/v2/employees/delete/2", nil)
		expectedResponseCode := http.StatusOK
		expectedResponseString := `{"id":"2","name":"Mister Bungle","startDate":"1984-11-10","endDate":"1996-06-01","insuredId":"1","recordTimestamp":"852206400","recordDateTime":"Thu, 02 Jan 1997 12:00:00 UTC"}` + "\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)

		// 2.) CONFIRM DELETED. 2nd request should return 404
		expectedResponseCode = http.StatusNotFound
		expectedResponseString = "{\"error\":\"Cannot delete. Record does not exist.\"}\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})
	t.Run("TestAPI_DeleteById_Address", func(t *testing.T) {
		_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
		defer MustCloseDB(t, db)

		// 1.) DELETE
		req, _ := http.NewRequest("DELETE", "/api/v2/address/delete/2", nil)
		expectedResponseCode := http.StatusOK
		expectedResponseString := `{"id":"2","address":"123 REAL Street, Springfield, Oregon","recordTimestamp":"469368001","recordDateTime":"Thu, 15 Nov 1984 12:00:01 UTC"}` + "\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)

		// 2.) CONFIRM DELETED. 2nd request should return 404
		expectedResponseCode = http.StatusNotFound
		expectedResponseString = "{\"error\":\"Cannot delete. Record does not exist.\"}\n"
		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})
}

func TestAPI_DeleteById_NotFound(t *testing.T) {
	_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
	defer MustCloseDB(t, db)
	t.Run("TestAPI_DeleteById_NotFound_Insured", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/v2/employees/delete/99", nil)
		expectedResponseCode := http.StatusNotFound
		expectedResponseString := "{\"error\":\"Cannot delete. Record does not exist.\"}\n"

		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})
	t.Run("TestAPI_DeleteById_NotFound_Employee", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/v2/employees/delete/99", nil)
		expectedResponseCode := http.StatusNotFound
		expectedResponseString := "{\"error\":\"Cannot delete. Record does not exist.\"}\n"

		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})
	t.Run("TestAPI_DeleteById_NotFound_Address", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/v2/address/delete/99", nil)
		expectedResponseCode := http.StatusNotFound
		expectedResponseString := "{\"error\":\"Cannot delete. Record does not exist.\"}\n"

		checkResponse(t, req, httpserver, nil, expectedResponseCode, expectedResponseString)
	})
}

func TestAPI_Create(t *testing.T) {
	_, httpserver, db := MustOpenDBAndSetUpRoutes(t) // move this inside each sub-test if they affect each other
	defer MustCloseDB(t, db)

	t.Run("Insured", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v2/insured/new", nil)
		expectedResponseCode := http.StatusCreated
		expectedResponseString := `{"id":3,"data":{"id":"3","name":"Muppy","policyNumber":"1002","recordTimestamp":""}}` + "\n"
		requestBody := map[string]string{
			"name": "Muppy",
		}
		checkResponse(t, req, httpserver, requestBody, expectedResponseCode, expectedResponseString)
	})
	t.Run("Address", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v2/address/new", nil)
		expectedResponseCode := http.StatusCreated
		expectedResponseString := `{"id":5,"data":{"address":"911 Reno Street","id":"5","insuredId":"2","recordTimestamp":""}}` + "\n"
		requestBody := map[string]string{
			"address":   "911 Reno Street",
			"insuredId": "2",
		}
		// 1.) Create
		checkResponse(t, req, httpserver, requestBody, expectedResponseCode, expectedResponseString)

		// 2.) Duplicate
		expectedResponseCode = http.StatusConflict
		expectedResponseString = `{"error":"Record already exists. Use 'update' to update"}` + "\n"
		checkResponse(t, req, httpserver, requestBody, expectedResponseCode, expectedResponseString)
	})
	t.Run("Employee", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v2/employee/new", nil)
		expectedResponseCode := http.StatusCreated
		expectedResponseString := `{"id":7,"data":{"endDate":"1994-01-14","id":"7","insuredId":"2","name":"Charles Bronson","recordTimestamp":"","startDate":"1974-07-24"}}` + "\n"
		requestBody := map[string]string{
			"name":      "Charles Bronson",
			"startDate": "1974-07-24",
			"endDate":   "1994-01-14",
			"insuredId": "2",
		}
		checkResponse(t, req, httpserver, requestBody, expectedResponseCode, expectedResponseString)
	})
}

func TestAPI_Update_Insured(t *testing.T) {
	_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
	defer MustCloseDB(t, db)
	t.Run("TestAPI_Update_Insured", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/v2/insured/update", nil)
		expectedResponseCode := http.StatusConflict
		expectedResponseString := fmt.Sprintf(`{"error":"%s"}`, service.ErrRecordAlreadyExists) + "\n" // TODO: change to "cannot update insured core data without authorization"
		requestBody := map[string]string{
			"id":   "1",
			"name": "Jimmy Temelpa", // existing record
		}
		checkResponse(t, req, httpserver, requestBody, expectedResponseCode, expectedResponseString)
	})
}

func TestAPI_Update_Employee(t *testing.T) {
	t.Run("Fail_MissingEmployeeId", func(t *testing.T) {
		_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
		defer MustCloseDB(t, db)
		req, _ := http.NewRequest("PUT", "/api/v2/employee/update", nil)
		expectedResponseCode := http.StatusBadRequest
		expectedResponseString := fmt.Sprintf(`{"error":"%s"}`, service.ErrEntityIDInvalid) + "\n"
		requestBody := map[string]string{
			"name":       "Charles Bronson",
			"startDate":  "1974-07-24",
			"endDate":    "1999-01-14",
			"insuredId":  "2",
			"employeeId": "",
		}
		checkResponse(t, req, httpserver, requestBody, expectedResponseCode, expectedResponseString)
	})
	// Deleted "TestAPI_Update_Employee_Fail_RequireCreate"

	t.Run("Fail_NoChanges", func(t *testing.T) {
		_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
		defer MustCloseDB(t, db)
		req, _ := http.NewRequest("PUT", "/api/v2/employee/update", nil)
		expectedResponseCode := http.StatusConflict
		expectedResponseString := fmt.Sprintf(`{"error":"%s"}`, service.ErrRecordUpdateRequireChange) + "\n"
		requestBody := map[string]string{
			"name":       "Mister Bungle",
			"startDate":  "1984-11-10",
			"endDate":    "1996-06-01",
			"insuredId":  "1", // try with 2
			"employeeId": "2",
		}

		checkResponse(t, req, httpserver, requestBody, expectedResponseCode, expectedResponseString)
	})

	// TODO:
	/* t.Run("Fail_MalformedDate", func(t *testing.T) {
		_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
		defer MustCloseDB(t, db)
		req, _ := http.NewRequest("PUT", "/api/v2/employee/update", nil)
		expectedResponseCode := http.StatusOK
		expectedResponseString := `{"id":2,"data":{"endDate":"0001-01-01","id":"2","insuredId":"1","name":"Mister Bungle","recordTimestamp":"","startDate":"1974-07-24"}}` + "\n"
		requestBody := map[string]string{
			"name":       "Mister Bungle",
			"startDate":  "1974-07-24",
			"endDate":    "199-01-14",
			"insuredId":  "1",
			"employeeId": "2",
		}
		checkResponse(t, req, httpserver, requestBody, expectedResponseCode, expectedResponseString)
	}) */

	// TODO:
	/* t.Run("Fail_StartDateAfterEndDate", func(t *testing.T) {
		_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
		defer MustCloseDB(t, db)
		req, _ := http.NewRequest("PUT", "/api/v2/employee/update", nil)
		expectedResponseCode := http.StatusBadRequest
		expectedResponseString := `` + "\n"
		requestBody := map[string]string{
			"name":       "Mister Bungle",
			"startDate":  "1974-07-24",
			"endDate":    "1000-01-14",
			"insuredId":  "1",
			"employeeId": "2",
		}
		checkResponse(t, req, httpserver, requestBody, expectedResponseCode, expectedResponseString)
	}) */

	// TODO:
	/* t.Run("Fail_WrongInsuredId", func(t *testing.T) {
		_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
		defer MustCloseDB(t, db)
		req, _ := http.NewRequest("PUT", "/api/v2/employee/update", nil)
		expectedResponseCode := http.StatusConflict
		expectedResponseString := fmt.Sprintf(`{"error":"%s"}`, service.ErrRecordUpdateRequireChange) + "\n" // wrong error
		requestBody := map[string]string{
			"name":      "Mister Bungle",
			"startDate": "1984-11-10",
			"endDate":   "1996-06-01",
			"insuredId": "2", // actual id is 1
			"employeeId": "2",
		}

		checkResponse(t, req, httpserver, requestBody, expectedResponseCode, expectedResponseString)
	}) */

	t.Run("Fail_MissingInsuredId", func(t *testing.T) {
		_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
		defer MustCloseDB(t, db)
		req, _ := http.NewRequest("PUT", "/api/v2/employee/update", nil)
		expectedResponseCode := http.StatusConflict
		expectedResponseString := `{"error":"Cannot create record for non-existent insuredId"}` + "\n"
		requestBody := map[string]string{
			"name":       "Mister Bungle",
			"startDate":  "1974-07-24",
			"endDate":    "1999-01-14",
			"employeeId": "2",
			// "insuredId" REQUIRED
		}

		// should fail, doesn't exist
		checkResponse(t, req, httpserver, requestBody, expectedResponseCode, expectedResponseString)
	})

	// update
	t.Run("Succeed_FullData", func(t *testing.T) {
		_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
		defer MustCloseDB(t, db)
		req, _ := http.NewRequest("PUT", "/api/v2/employee/update", nil)
		expectedResponseCode := http.StatusOK
		expectedResponseString := `{"id":2,"data":{"endDate":"1999-01-14","id":"2","insuredId":"1","name":"Mister Bungle","recordTimestamp":"","startDate":"1974-07-24"}}` + "\n"
		requestBody := map[string]string{
			"name":       "Mister Bungle",
			"startDate":  "1974-07-24",
			"endDate":    "1999-01-14",
			"insuredId":  "1",
			"employeeId": "2",
		}
		checkResponse(t, req, httpserver, requestBody, expectedResponseCode, expectedResponseString)
	})

	t.Run("Succeed_EmptyEndDate", func(t *testing.T) {
		_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
		defer MustCloseDB(t, db)
		req, _ := http.NewRequest("PUT", "/api/v2/employee/update", nil)
		requestBody := map[string]string{
			"employeeId": "1",
			"name":       "Jimathy Trashleigh Moganstern III Esquire", // change name
			"insuredId":  "1",
			"startDate":  "2006-01-02", // change startDate
			// no end date - should not change existing end date
		}
		expectedResponseCode := http.StatusOK
		expectedResponseString := `{"id":1,"data":{"endDate":"","id":"1","insuredId":"1","name":"Jimathy Trashleigh Moganstern III Esquire","recordTimestamp":"","startDate":"2006-01-02"}}` + "\n"
		checkResponse(t, req, httpserver, requestBody, expectedResponseCode, expectedResponseString)
	})

	// TODO:
	/* t.Run("PartialUpdate_Succeed", func(t *testing.T) {
		_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
		defer MustCloseDB(t, db)
		req, _ := http.NewRequest("PUT", "/api/v2/employee/update", nil)
		requestBody := map[string]string{
			"employeeId": "1", // existing record
			"name":        "Jim-Jim",
			"insuredId":   "1",
			// omit "startDate" required for new records.
		}
		expectedResponseCode := http.StatusOK
		expectedResponseString := `{"error": "Cannot create record for non-existent insuredId"}` + "\n"
		checkResponse(t, req, httpserver, requestBody, expectedResponseCode, expectedResponseString)
	}) */
}

func checkResponse(t *testing.T, req *http.Request, httpserver *http.Server, requestBody map[string]string, expectedResponseCode int, expectedResponseString string) {
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		t.Errorf("Bad request")
	}
	ignoreTimestamp := false
	if req.Method == "POST" || req.Method == "PUT" {
		ignoreTimestamp = true
	}

	req.Body = io.NopCloser(strings.NewReader(string(requestJSON)))
	response := executeRequest(req, httpserver)
	checkResponseCode(t, expectedResponseCode, response.Code)
	checkResponseData(t, expectedResponseString, response.Body.String(), ignoreTimestamp)
}

func TestAPI_Update_Address(t *testing.T) {
	_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
	defer MustCloseDB(t, db)

	t.Run("Conflict", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/api/v2/address/update", nil)
		expectedResponseCode := http.StatusNotFound
		expectedResponseString := `{"error":"Record does not exist. Use 'new' to create."}` + "\n"
		requestBody := map[string]string{
			"address":   "911 Las Vegas Street",
			"insuredId": "2",
		}
		// 1.) test update on already-created - should fail
		checkResponse(t, req, httpserver, requestBody, expectedResponseCode, expectedResponseString)

		// 2.) change of address
		expectedResponseCode = http.StatusOK
		expectedResponseString = `{"id":5,"data":{"address":"911 Las Vegas Street","id":"5","insuredId":"1","recordTimestamp":""}}` + "\n"
		requestBody = map[string]string{
			"address":   "911 Las Vegas Street",
			"insuredId": "1",
		}
		checkResponse(t, req, httpserver, requestBody, expectedResponseCode, expectedResponseString)
	})

}

func executeRequest(req *http.Request, httpserver *http.Server) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	httpserver.Handler.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}
func checkResponseData(t *testing.T, expected, actual string, ignoreTimestamp bool) {
	//fmt.Println("Response string:\n", actual, "\nExpected:\n", expected)
	if ignoreTimestamp {
		var re = regexp.MustCompile(`(recordTimestamp":")([1-2][0-9]{9})"`)
		actual = re.ReplaceAllString(actual, `${1}"`)
	}
	if expected != actual {
		t.Errorf("Expected response string %s. Got %s\n", expected, actual)
	}
}

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
		dsn = filepath.Join(dir, "db")
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

/*
TODO: delete DB after tests that alter it
func MustCloseDBAndDelete(tb testing.TB, db *sqlite.DB, dir string) {
	MustCloseDB(tb, db)
	os.Remove(dir) // dangerous?
} */

//"*mux.Router, service.SqliteRecordService, "

// TODO: get this from server.go
func SetUpRoutes(db *sqlite.DB) (*api.API, service.SqliteRecordService, *http.Server) {
	router := mux.NewRouter()
	memoryService := service.NewInMemoryRecordService() // not testing this but need to avoid nil pointer ref
	sqliteService := service.NewSqliteRecordService()
	api := api.NewAPI(&memoryService, &sqliteService)

	apiRoute := router.PathPrefix("/api/v1").Subrouter()
	apiRoute.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		fmt.Println("Err setting up route:", err)
	})
	apiRouteV2 := router.PathPrefix("/api/v2").Subrouter()
	apiRouteV2.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		fmt.Println("Err setting up route:", err)
	})
	api.CreateRoutes(apiRoute, apiRouteV2)

	address := "127.0.0.1:8000"
	HTTPServer := &http.Server{
		Handler:      router,
		Addr:         address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	//db := sqlite.NewDB("file:main.db?cache=shared&mode=rwc&locking_mode=NORMAL&_fk=1&synchronous=2")
	// injects db into the sqlite service so it can interact with the database.
	sqliteService.SetService(db)
	return api, sqliteService, HTTPServer
}

func MustOpenDBAndSetUpRoutes(t testing.TB) (*api.API, *http.Server, *sqlite.DB) {
	fmt.Println("Test name opening DB:", t.Name())
	db := MustOpenDB(t)
	api, _, httpserver := SetUpRoutes(db)
	return api, httpserver, db
}
