package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// GET /insured/id/{id}
// GetInsureds retrieves the record.
func (a *API) GetResourceById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resource := mux.Vars(r)["type"]	
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
