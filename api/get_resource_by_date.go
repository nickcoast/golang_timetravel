package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// API V2
// GET /{type}/getbydate/{insuredId}/{date}
// Get unique records for this resource valid at this date
func (a *API) GetResourceByDate(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println("api.GetResourceByDate date:", date)
	fmt.Println("api.GetResourceByDate resource:", resource)
	dateTime, err := time.Parse("2006-01-02", date)

	if err != nil {
		err := writeError(w, fmt.Sprintf("Please submit date in format: 2006-01-02. Or submit timestamp"), http.StatusBadRequest)
		logError(err)
		return
	}

	if resource == "insured" {
		insured, err := a.sqlite.GetInsuredByDate(ctx, idNumber, dateTime)
		if err != nil {
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
		resource,
		naturalKey,
		idNumber,
		dateTime,
	)
	fmt.Println(record)
	if err != nil || record.ID == 0 {
		err := writeError(w, fmt.Sprintf("No record for Insured %v and date %v exist", idNumber, date), http.StatusBadRequest)
		logError(err)
		return
	}

	err = writeJSON(w, record, http.StatusOK)
	logError(err)
}
