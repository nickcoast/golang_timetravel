package api

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"golang.org/x/exp/maps"
)

// API V2
// GET /{type}
// Get all current records (1 record for each entity)
func (a *API) GetResourceRecords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	insuredObject, err := a.NewInsuredObjectFromRequest(r)
	if err != nil {
		err := writeError(w, err.Error(), http.StatusBadRequest)
		logError(err)
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		err := writeError(w, err.Error(), http.StatusBadRequest)
		logError(err)
		return
	}

	entities, err := a.sqlite.GetAllByEntityId(ctx, insuredObject, int64(id))
	if err != nil {
		err := writeError(w, err.Error(), http.StatusInternalServerError)
		logError(err)
		return
	}
	m := maps.Values(entities) // convert map to slice for json array
	err = writeJSON(w, m, http.StatusOK)
}
