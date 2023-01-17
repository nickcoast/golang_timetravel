package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// API V2
// GET /{type}/getbytimestamp/{insuredId}/{date}
// Get unique records for this resource valid at this date
func (a *API) GetResourceByTimestamp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestType := mux.Vars(r)["type"]

	resource, err := resourceNameFromSynonym(requestType)
	if err != nil {
		err := writeError(w, err.Error(), http.StatusBadRequest)
		logError(err)
		return
	}

	date := mux.Vars(r)["date"]
	naturalKey := "date" // might rename this to groupingKey.
	insuredId := mux.Vars(r)["insuredId"]
	idNumber, err := strconv.ParseInt(insuredId, 10, 32)

	fmt.Println("api.GetResourceByTimestamp date:", date)
	fmt.Println("api.GetResourceByTimestamp resource:", resource)
	// check if timestamp
	timestamp, err := strconv.ParseInt(date, 10, 64)
	if err != nil {
		err := writeError(w, fmt.Sprintf("Please submit date in timestamp (integer) format"), http.StatusBadRequest)
		logError(err)
		return
	}
	timestampDate := time.Unix(timestamp, 0)

	if resource == "insured" {
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

	record, err := a.sqlite.GetRecordByDate(
		ctx,
		resource,
		naturalKey,
		idNumber,
		timestampDate,
	)
	fmt.Println(record)
	if err != nil || record.ID == 0 {
		err := writeError(w, fmt.Sprintf("No record for %v and date %v exist", resource, date), http.StatusBadRequest)
		logError(err)
		return
	}

	err = writeJSON(w, record, http.StatusOK)
	logError(err)
}
