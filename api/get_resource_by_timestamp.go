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
// GET /{type}/getbytimestamp/{insuredId}/{date}
// Get unique records for this resource valid at this date
func (a *API) GetResourceByTimestamp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	insuredObject, err := a.NewInsuredObjectFromRequest(r)
	if err != nil {
		err := writeError(w, err.Error(), http.StatusBadRequest)
		logError(err)
		return
	}

	date := mux.Vars(r)["date"]
	naturalKey := "date" // might rename this to groupingKey.
	insuredId := mux.Vars(r)["insuredId"]
	idNumber, err := strconv.ParseInt(insuredId, 10, 32)

	// check if timestamp
	timestamp, err := strconv.ParseInt(date, 10, 64)
	if err != nil {
		err := writeError(w, fmt.Sprintf("Please submit date in timestamp (integer) format"), http.StatusBadRequest)
		logError(err)
		return
	}
	timestampDate := time.Unix(timestamp, 0)

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
