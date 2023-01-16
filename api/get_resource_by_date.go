package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// GET /insured/id/{id}
// Get this unique records for this resources valid at this date
func (a *API) GetResourceByDate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resource := mux.Vars(r)["type"]
	date := mux.Vars(r)["date"]
	naturalKey := "date" // might rename this to groupingKey.
	insuredId := mux.Vars(r)["insuredId"]
	idNumber, err := strconv.ParseInt(insuredId, 10, 32)

	fmt.Println("api.GetResourceByDate date:", date)
	fmt.Println("api.GetResourceByDate resource:", resource)
	dateTime, err := time.Parse("2006-01-02", date)
	if err != nil {
		err := writeError(w, fmt.Sprintf("Please submit date in format: 2006-01-02"), http.StatusBadRequest)
		logError(err)
	}

	record, err := a.sqlite.GetRecordByDate(
		ctx,
		resource,
		naturalKey,
		idNumber,
		dateTime,
	)
	fmt.Println(record)
	if err != nil {
		err := writeError(w, fmt.Sprintf("No record for Insured %v and date %v exist", idNumber, date), http.StatusBadRequest)
		logError(err)
		return
	}

	err = writeJSON(w, record, http.StatusOK)
	logError(err)
}
