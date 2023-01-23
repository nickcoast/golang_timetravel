package api_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"

	//"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/nickcoast/timetravel/api"
	"github.com/nickcoast/timetravel/service"
	"github.com/nickcoast/timetravel/sqlite"
	//"github.com/gorilla/mux"
	//"github.com/steinfletcher/apitest"
)

type APItest struct {
	db         *sqlite.DB
	HTTPServer *http.Server
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

func TestAPI_GetById(t *testing.T) {
	_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
	defer MustCloseDB(t, db)

	t.Run("TestAPI_GetById_Insured", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v2/insured/id/2", nil)
		response := executeRequest(req, httpserver)
		checkResponseCode(t, http.StatusOK, response.Code)

		responseString := "{\"id\":\"2\",\"name\":\"John Smith\",\"policy_number\":\"1001\",\"recordTimestamp\":\"946684799\",\"recordDateTime\":\"Fri, 31 Dec 1999 23:59:59 UTC\",\"employees\":null,\"insuredAddresses\":null}\n"
		fmt.Println("Response string:\n", response.Body.String(), "\nExpected:", responseString)
		checkResponseData(t, responseString, response.Body.String(), false)
	})
	t.Run("TestAPI_GetById_Employee", func(t *testing.T) { // should get latest Mister Bungle record. lol
		req, _ := http.NewRequest("GET", "/api/v2/employee/id/2", nil)
		response := executeRequest(req, httpserver)
		checkResponseCode(t, http.StatusOK, response.Code)

		responseString := "{\"id\":\"2\",\"name\":\"Mister Bungle\",\"startDate\":\"1984-11-10\",\"endDate\":\"1996-06-01\",\"insuredId\":\"1\",\"recordTimestamp\":\"852206400\",\"recordDateTime\":\"Thu, 02 Jan 1997 12:00:00 UTC\"}\n"
		fmt.Println("Response string:\n", response.Body.String(), "\nExpected:", responseString)
		checkResponseData(t, responseString, response.Body.String(), false)
	})
	t.Run("TestAPI_GetById_Address", func(t *testing.T) { // should get 123 Fake Street, Springfield, Oregon
		req, _ := http.NewRequest("GET", "/api/v2/address/id/1", nil)
		response := executeRequest(req, httpserver)
		checkResponseCode(t, http.StatusOK, response.Code)

		responseString := "{\"id\":\"1\",\"address\":\"123 Fake Street, Springfield, Oregon\",\"recordTimestamp\":\"468072000\",\"recordDateTime\":\"Wed, 31 Oct 1984 12:00:00 UTC\"}\n"
		fmt.Println("Response string:\n", response.Body.String(), "\nExpected:", responseString)
		checkResponseData(t, responseString, response.Body.String(), false)
	})

}

func TestAPI_GetResourceByTimestamp(t *testing.T) {
	_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
	defer MustCloseDB(t, db)

	req, _ := http.NewRequest("GET", "/api/v2/insured/getbytimestamp/2/954590400", nil)
	response := executeRequest(req, httpserver)
	checkResponseCode(t, http.StatusOK, response.Code)

	responseString := "{\"id\":2,\"data\":{\"id\":\"2\",\"name\":\"John Smith\",\"policy_number\":\"1001\",\"record_timestamp\":\"946684799\"}}\n"
	fmt.Println("Response string:\n", response.Body.String(), "\nExpected:", responseString)
	checkResponseData(t, responseString, response.Body.String(), false)
}

func TestAPI_DeleteResourceById(t *testing.T) {
	t.Run("TestAPI_DeleteResourceById_Insured", func(t *testing.T) {
		_, httpserver, db := MustOpenDBAndSetUpRoutes(t) // TODO: make so only do this once per set of tests.
		defer MustCloseDB(t, db)

		req, _ := http.NewRequest("DELETE", "/api/v2/employees/delete/4", nil)
		response := executeRequest(req, httpserver)
		checkResponseCode(t, http.StatusOK, response.Code)

		responseString := "{\"id\":\"4\",\"name\":\"Jane Doe\",\"startDate\":\"1985-05-15\",\"endDate\":\"1999-12-25\",\"insuredId\":\"2\",\"recordTimestamp\":\"954590400\",\"recordDateTime\":\"Sat, 01 Apr 2000 12:00:00 UTC\"}\n"
		fmt.Println("Response string:\n", response.Body.String(), "\nExpected:", responseString)
		checkResponseData(t, responseString, response.Body.String(), false)

		// 2nd request should return 404
		response = executeRequest(req, httpserver)
		checkResponseCode(t, http.StatusNotFound, response.Code)

		// TODO: return original record with ID set to 0 instead
		responseString = "{\"error\":\"Cannot delete. Record does not exist.\"}\n"
		fmt.Println("Response string:\n", response.Body.String(), "\nExpected:", responseString)
		checkResponseData(t, responseString, response.Body.String(), false)

	})
}

func TestCreateResource(t *testing.T) {
	_, httpserver, db := MustOpenDBAndSetUpRoutes(t)
	defer MustCloseDB(t, db)

	req, _ := http.NewRequest("POST", "/api/v2/insured/new", nil)
	requestBody, err := json.Marshal(map[string]string{
		"name": "Muppy",
	})
	if err != nil {
		t.Errorf("Bad request")
	}
	//req.Body = requestBody // io.ReadCloser
	//req.Body =io.NopCloser(strings.NewReader("{\"name\":\"Muppy\"}"))
	req.Body = io.NopCloser(strings.NewReader(string(requestBody)))

	response := executeRequest(req, httpserver)
	checkResponseCode(t, http.StatusOK, response.Code)

	responseString := "{\"id\":3,\"data\":{\"id\":\"3\",\"name\":\"Muppy\",\"policy_number\":\"1002\",\"record_timestamp\":\"\"}}\n"
	fmt.Println("Response string:\n", response.Body.String(), "\nExpected:", responseString)
	checkResponseData(t, responseString, response.Body.String(), true)
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
	if ignoreTimestamp {
		var re = regexp.MustCompile(`(record_timestamp":")([1-2][0-9]{9})"`)
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

//"*mux.Router, service.SqliteRecordService, "

func SetUpRoutes(db *sqlite.DB) (*api.API, service.SqliteRecordService, *http.Server) {
	router := mux.NewRouter()
	memoryService := service.NewInMemoryRecordService() // not testing this but need to avoid nil pointer ref
	sqliteService := service.NewSqliteRecordService()
	api := api.NewAPI(&memoryService, &sqliteService)

	apiRoute := router.PathPrefix("/api/v1").Subrouter()
	apiRoute.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		fmt.Println(err)
	})
	apiRouteV2 := router.PathPrefix("/api/v2").Subrouter()
	apiRouteV2.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		fmt.Println(err)
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
	db := MustOpenDB(t)
	api, _, httpserver := SetUpRoutes(db)
	return api, httpserver, db
}
