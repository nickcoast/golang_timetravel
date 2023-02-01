package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/nickcoast/timetravel/entity"
)

// API V2
// GET /{type}/gethistory/{id}
// Get all updates for this resource
func (a *API) GetHistoryById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	insuredObject, err := a.NewInsuredObjectFromRequest(r)
	if err != nil {
		err := writeError(w, err.Error(), http.StatusBadRequest)
		logError(err)
		return
	}

	_, ok := insuredObject.(*entity.Insured)
	if ok {
		insured, err := a.sqlite.GetInsuredByDate(ctx, idNumber, timestampDate)
		if err != nil || insured.ID == 0 {
			err := writeError(w, fmt.Sprintf("No record for Insured %v and date %v exist", idNumber, date), http.StatusBadRequest)
			logError(err)
			return
		}
		err = writeJSON(w, insured, http.StatusOK)
		logError(err)
		return
	}

	record, err := a.sqlite.GetResourceByDate(
		ctx,
		insuredObject,
		naturalKey,
		idNumber,
		timestampDate,
	)
	fmt.Println(record)
	if err != nil || record.GetId() == 0 {
		err := writeError(w, fmt.Sprintf("No record for this date (%v) exists", date), http.StatusBadRequest)
		logError(err)
		return
	}

	err = writeJSON(w, record, http.StatusOK)
	logError(err)
}
