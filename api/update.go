package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nickcoast/timetravel/entity"
)

// API V2
// POST /{type}/new
// if the record exists, the record is updated.
// "insured" name and policy number cannot be updated.
// "employees" and "insuredAddress" can be updated.
// if the record doesn't exist, the record is updated.
func (a *API) Update(w http.ResponseWriter, r *http.Request) {
	requestType := mux.Vars(r)["type"]
	resource, err := resourceNameFromSynonym(requestType)
	if err != nil {
		err := writeError(w, err.Error(), http.StatusBadRequest)
		logError(err)
		return
	}

	fmt.Println("UpdateRecord")
	ctx := r.Context()

	var body map[string]*string
	err = json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		err := writeError(w, "invalid input; could not parse json", http.StatusBadRequest)
		logError(err)
		return
	}

	recordMap := map[string]string{}
	for key, value := range body {
		if value != nil {
			recordMap[key] = *value
		}
	}
	var requestRecord entity.Record
	requestRecord.Data = recordMap

	fmt.Println("api.UpdateInsured requestRecord:", requestRecord)
	newRecord, err := a.sqlite.UpdateRecord(ctx, resource, requestRecord)

	if err != nil && err.Error() == "Record does not exist. Use 'new' instead" {
		errInWriting := writeError(w, err.Error(), http.StatusBadRequest)
		logError(err)
		logError(errInWriting)
		return
	} else if err != nil {
		errInWriting := writeError(w, ErrInternal.Error(), http.StatusInternalServerError)
		logError(err)
		logError(errInWriting)
		return
	}

	fmt.Println("newRecord", newRecord)
	err = writeJSON(w, newRecord, http.StatusOK) //TODO: actually return new record
	logError(err)
}
