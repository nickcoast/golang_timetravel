package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// API V2
// GET /{type}/id/{id:[0-9]+}
// GetInsureds retrieves the record.
func (a *API) GetResourceById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestType := mux.Vars(r)["type"]

	resource, err := resourceNameFromSynonym(requestType)
	if err != nil {
		err := writeError(w, err.Error(), http.StatusBadRequest)
		logError(err)
		return
	}

	id := mux.Vars(r)["id"]
	fmt.Println("api.GetInsuredsById id:", id)
	fmt.Println("api.GetInsuredsById resource:", resource)
	idNumber, err := strconv.ParseInt(id, 10, 32)
	if err != nil || idNumber <= 0 {
		err := writeError(w, "invalid id; id must be a positive number", http.StatusBadRequest)
		logError(err)
		return
	}

	record, err := a.sqlite.GetRecordById(
		ctx,
		resource,
		int(idNumber),
	)
	if err != nil {
		err := writeError(w, fmt.Sprintf("record of id %v does not exist", idNumber), http.StatusBadRequest)
		logError(err)
		return
	}

	err = writeJSON(w, record, http.StatusOK)
	logError(err)
}
