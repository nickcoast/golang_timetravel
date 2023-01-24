package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nickcoast/timetravel/entity"
)

// API V2
// GET /{type}/id/{id:[0-9]+}
// GetInsureds retrieves the record.
func (a *API) GetResourceById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	insuredObject, err := a.NewInsuredObjectFromRequest(r)
	if err != nil {
		err := writeError(w, err.Error(), http.StatusBadRequest)
		logError(err)
		return
	}

	id := mux.Vars(r)["id"]
	idNumber, err := strconv.ParseInt(id, 10, 32)
	if err != nil || idNumber <= 0 {
		err := writeError(w, "invalid id; id must be a positive number", http.StatusBadRequest)
		logError(err)
		return
	}

	record, err := a.sqlite.GetResourceById(
		ctx,
		insuredObject,
		int(idNumber), // TODO: get id from insuredObject
	)
	if err != nil {
		err := writeError(w, fmt.Sprintf("record of id %v does not exist", idNumber), http.StatusBadRequest)
		logError(err)
		return
	}
	insured, ok := record.(*entity.Insured)
	if ok { // trying to handle nil pointer
		if insured.Addresses == nil {
			var Addresses map[int]entity.Address
			//insured.Addresses = map[int]entity.Address{}
			insured.Addresses = &Addresses
		}
		if insured.Employees == nil {
			var Employees map[int]entity.Employee
			insured.Employees = &Employees
		}
	}

	err = writeJSON(w, record, http.StatusOK)
	logError(err)
}
