package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nickcoast/timetravel/entity"
	"github.com/nickcoast/timetravel/service"
)

// POST /sqlite/{id}
// if the record exists, the record is updated.
// if the record doesn't exist, the record is created.
// note "Record" is used here but its for the "sqlite" service that currently acts on Insureds only
func (a *API) CreateInsured(w http.ResponseWriter, r *http.Request) {
	fmt.Println("CREATE")
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	//name := mux.Vars(r)["name"]

	idNumber, err := strconv.ParseInt(id, 10, 32)

	var body map[string]*string
	err = json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		err := writeError(w, "invalid input; could not parse json", http.StatusBadRequest)
		logError(err)
		return
	}

	// first retrieve the record
	_, err = a.sqlite.GetRecordById(
		ctx,
		int(idNumber),
	)

	if !errors.Is(err, service.ErrRecordDoesNotExist) { // record exists
		return
	}
	recordMap := map[string]string{}
	for key, value := range body {
		if value != nil {
			recordMap[key] = *value
		}
	}
	var newRecord entity.Record
	newRecord.Data = recordMap

	err = a.sqlite.CreateRecord(ctx, newRecord)

	if err != nil {
		errInWriting := writeError(w, ErrInternal.Error(), http.StatusInternalServerError)
		logError(err)
		logError(errInWriting)
		return
	}

	fmt.Println("newRecord", newRecord)
	err = writeJSON(w, newRecord, http.StatusOK)
	logError(err)
}
