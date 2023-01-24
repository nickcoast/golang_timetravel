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

	insuredObject, err := a.NewInsuredObjectFromRequest(r)
	if err != nil {
		err := writeError(w, err.Error(), http.StatusBadRequest)
		logError(err)
		return
	}
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	idNumber, err := strconv.ParseInt(id, 10, 32)
	// first retrieve the record
	record, err := a.sqlite.GetResourceById( // This is also done by DB with deleted, err := db.GetById(ctx, tableName, id)
		ctx,
		insuredObject, // TODO: change to InsuredInterface
		int(idNumber),
	)

	if errors.Is(err, service.ErrRecordDoesNotExist) { // record exists
		err = writeError(w, "Cannot delete. Record does not exist.", http.StatusNotFound)
		fmt.Println("Yikes")
		return
	}

	deletedRecord, err := a.sqlite.DeleteResource(ctx, record, idNumber)
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
