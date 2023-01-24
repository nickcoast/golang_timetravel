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
// if the record doesn't exist, the record is created.
func (a *API) Create(w http.ResponseWriter, r *http.Request) {
	requestType := mux.Vars(r)["type"]
	resource, err := resourceNameFromSynonym(requestType)
	if err != nil {
		err := writeError(w, err.Error(), http.StatusBadRequest)
		logError(err)
		return
	}

	fmt.Println("CreateRecord")
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

	fmt.Println("api.CreateInsured requestRecord", requestRecord)
	newRecord, err := a.sqlite.CreateResource(ctx, resource, requestRecord)

	if err != nil {
		if err.Error() == entity.ErrRecordAlreadyExists {
			errInWriting := writeError(w, err.Error(), http.StatusConflict)
			logError(err)
			logError(errInWriting)
			return
		}
		errInWriting := writeError(w, err.Error(), http.StatusInternalServerError)
		logError(err)
		logError(errInWriting)
		return
	}

	fmt.Println("newRecord", newRecord)
	err = writeJSON(w, newRecord, http.StatusCreated)
	logError(err)
}
