package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nickcoast/timetravel/service"
)

// API V2
// DELETE /{type}/delete/{id:[0-9]+}
// if the record exists, the record is updated.
// if the record doesn't exist, the record is created.
// note "Record" is used here but its for the "sqlite" service that currently acts on Insureds only
func (a *API) Delete(w http.ResponseWriter, r *http.Request) {
	requestType := mux.Vars(r)["type"]
	fmt.Println("DELETE")
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	//name := mux.Vars(r)["name"]
	idNumber, err := strconv.ParseInt(id, 10, 32)

	resource, err := resourceNameFromSynonym(requestType)
	if err != nil {
		err := writeError(w, err.Error(), http.StatusBadRequest)
		logError(err)
		return
	}

	// first retrieve the record
	_, err = a.sqlite.GetRecordById(
		ctx,
		resource,
		int(idNumber),
	)

	if errors.Is(err, service.ErrRecordDoesNotExist) { // record exists
		err = writeError(w, "Cannot delete. Record does not exist.", http.StatusBadRequest)
		fmt.Println("Yikes")
		return
	}

	/* this API response can give the user a chance to "undo" by sumbitting this data as a new record */
	deletedRecord, err := a.sqlite.DeleteRecord(ctx, resource, idNumber)
	if err != nil {
		err := writeError(w, "Bad request or server error", http.StatusBadRequest)
		logError(err)
		fmt.Println("oh no")
		return
	}
	err = writeJSON(w, deletedRecord, http.StatusOK)
	logError(err)
	return
}
