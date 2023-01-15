package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nickcoast/timetravel/entity"
)

// POST /sqlite/{id}
// if the record exists, the record is updated.
// if the record doesn't exist, the record is created.
// note "Record" is used here but its for the "sqlite" service that currently acts on Insureds only
func (a *API) Create(w http.ResponseWriter, r *http.Request) {
	resource := mux.Vars(r)["type"]
	fmt.Println("CreateRecord")
	ctx := r.Context()

	var body map[string]*string
	err := json.NewDecoder(r.Body).Decode(&body)

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
	var newRecord entity.Record
	newRecord.Data = recordMap

	fmt.Println("api.CreateInsured newRecord", newRecord)
	err = a.sqlite.CreateRecord(ctx, resource, newRecord)

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
