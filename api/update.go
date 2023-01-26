package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nickcoast/timetravel/entity"
	"github.com/nickcoast/timetravel/service"
)

// API V2
// POST /{type}/new
// if the record exists, the record is updated.
// "insured" name and policy number cannot be updated.
// "employees" and "insuredAddress" can be updated.
// if the record doesn't exist, the record is updated.
func (a *API) Update(w http.ResponseWriter, r *http.Request) {
	requestType := mux.Vars(r)["type"]
	resource, err := resourceNameFromSynonym(requestType)
	if err != nil {
		err := writeError(w, err.Error(), http.StatusBadRequest)
		logError(err)
		return
	}
	ctx := r.Context()

	var body map[string]*string
	err = json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		err := writeError(w, "invalid input; could not parse json", http.StatusBadRequest)
		logError(err)
		return
	}

	recordMap := map[string]string{}
	for key, value := range body {
		if value != nil {
			recordMap[key] = *value
		}
	}
	var requestRecord entity.Record
	requestRecord.Data = recordMap
	newRecord, err := a.sqlite.UpdateResource(ctx, resource, requestRecord)

	if err != nil {
		var status int
		if err == service.ErrRecordDoesNotExist {
			status = http.StatusNotFound
			/* errInWriting := writeError(w, err.Error(), http.StatusNotFound)
			logError(err)
			logError(errInWriting)
			return */
		} else if err == service.ErrNonexistentParentRecord {
			status = http.StatusConflict
		} else if err == service.ErrInvalidRequest || err == service.ErrEntityIDInvalid {
			status = http.StatusBadRequest
		} else if err == service.ErrRecordAlreadyExists || err == service.ErrRecordUpdateRequireChange { // test
			status = http.StatusConflict
		} else if err != nil {
			status = http.StatusInternalServerError
		}
		errInWriting := writeError(w, err.Error(), status)
		logError(err)
		logError(errInWriting)
		return
	}
	err = writeJSON(w, newRecord, http.StatusOK) //TODO: actually return new record
	logError(err)
}

/* if err != nil {
	var status int
	if err == service.ErrRecordDoesNotExist {
		status = http.StatusNotFound
	} else if err == service.ErrNonexistentParentRecord {
		status = http.StatusConflict
	} else if err == service.ErrInvalidRequest {
		status = http.StatusBadRequest
	} else if err != service.ErrRecordAlreadyExists { // test
		status = http.StatusConflict
	} else if err != nil {
		status = http.StatusInternalServerError
	}
	errInWriting := writeError(w, ErrInternal.Error(), status)
	logError(err)
	logError(errInWriting)
	return
} */
