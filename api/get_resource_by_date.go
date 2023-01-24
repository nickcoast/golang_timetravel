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
// GET /{type}/getbydate/{insuredId}/{date}
// Get unique records for this resource valid at this date
func (a *API) GetResourceByDate(w http.ResponseWriter, r *http.Request) {
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

	t := time.Now()
	zone, offset := t.Zone()
	fmt.Println(zone, offset)

	dateTime, err := time.Parse("2006-01-02", date)
	dateTime = dateTime.Add(time.Hour*24 - time.Second*time.Duration(offset)) // Set to midnight of system timezone
	/* timestamp := dateTime.Unix()
	fmt.Println(timestamp) */
	if err != nil {
		err := writeError(w, fmt.Sprintf("Please submit date in format: 2006-01-02. Or submit timestamp"), http.StatusBadRequest)
		logError(err)
		return
	}

	_, ok := insuredObject.(*entity.Insured)
	if ok {
		insured, err := a.sqlite.GetInsuredByDate(ctx, idNumber, dateTime)
		if err != nil {
			err := writeError(w, fmt.Sprintf("No record for Insured %v and date %v exist", idNumber, date), http.StatusNotFound)
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
		dateTime,
	)
	fmt.Println(record)
	if err != nil || record.GetId() == 0 {
		err := writeError(w, fmt.Sprintf("No record for Insured %v and date %v exist", idNumber, date), http.StatusBadRequest)
		logError(err)
		return
	}

	err = writeJSON(w, record, http.StatusOK)
	logError(err)
}
