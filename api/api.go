package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nickcoast/timetravel/entity"
	"github.com/nickcoast/timetravel/service"
)

type API struct {
	records service.RecordService // memory
	sqlite  service.RecordService // sqlite

}

func NewAPI(records service.RecordService, sqlite service.RecordService) *API {
	return &API{records, sqlite}
}

// generates all api routes
func (a *API) CreateRoutes(routesV1 *mux.Router, routesV2 *mux.Router) {
	fmt.Println("help")
	a.CreateV1Routes(routesV1)
	a.CreateV2Routes(routesV2)
}

func (a *API) CreateV1Routes(routes *mux.Router) {
	routes.Path("/records/{id:[0-9]+}").HandlerFunc(a.GetRecords).Methods("GET")
	routes.Path("/records/{id:[0-9]+}").HandlerFunc(a.PostRecords).Methods("POST")
}
func (a *API) CreateV2Routes(routes *mux.Router) {
	i := routes
	i.Path("/{type}/id/{id:[0-9]+}").HandlerFunc(a.GetResourceById).Methods("GET")
	i.Path("/{type}/new").HandlerFunc(a.Create).Methods("POST")
	i.Path("/{type}/update").HandlerFunc(a.Update).Methods("PUT")

	// Permanently deletes record (insured, employee, or insured address)
	// Should be allowed to supervisors in case of erroneous data or FBI investigations
	// TODO: force consumer to confirm before allowing permanent deletion.
	i.Path("/{type}/delete/{id:[0-9]+}").HandlerFunc(a.Delete).Methods("DELETE")

	// !!!TIME TRAVEL!!! - use getbydate and getbytimestamp to get records valid at a particular time
	//i.Path("/{type}/confirmdelete/{id:[0-9]+}").HandlerFunc(a.Delete).Methods("DELETE")
	i.Path("/{type}/getbydate/{insuredId}/{date}").HandlerFunc(a.GetResourceByDate).Methods("GET")
	// same as above, but using integer timestamp for exact times
	i.Path("/{type}/getbytimestamp/{insuredId}/{date}").HandlerFunc(a.GetResourceByTimestamp).Methods("GET")

	ad := routes.PathPrefix("/address").Subrouter()
	ad.Path("/id/{id:[0-9]+}").HandlerFunc(a.GetRecords).Methods("GET")
}

type updateCreate int

const (
	update = iota
	create
)

func (a *API) updateOrCreate(w http.ResponseWriter, r *http.Request, uOrC updateCreate) {
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

	// TODO: any way to DRY this?
	if uOrC == update {
		record, err := a.sqlite.UpdateRecord(ctx, resource, requestRecord)
		if err.Error() == "Record does not exist. Use 'new' instead" {
			errInWriting := writeError(w, err.Error(), http.StatusBadRequest)
			logError(err)
			logError(errInWriting)
			return
		} else if err != nil {
			errInWriting := writeError(w, ErrInternal.Error(), http.StatusInternalServerError)
			logError(err)
			logError(errInWriting)
			return
		}

		fmt.Println("newRecord", record)
		err = writeJSON(w, record, http.StatusOK) //TODO: actually return new record
		logError(err)
	} else if uOrC == create {
		record, err := a.sqlite.CreateRecord(ctx, resource, requestRecord)
		if err.Error() == "Record already exists. Use 'update' to update" {
			errInWriting := writeError(w, err.Error(), http.StatusBadRequest)
			logError(err)
			logError(errInWriting)
			return
		} else if err != nil {
			errInWriting := writeError(w, ErrInternal.Error(), http.StatusInternalServerError)
			logError(err)
			logError(errInWriting)
			return
		}

		fmt.Println("newRecord", record)
		err = writeJSON(w, record, http.StatusOK) //TODO: actually return new record
		logError(err)
	} else {
		errInWriting := writeError(w, ErrInternal.Error(), http.StatusInternalServerError)
		logError(err)
		logError(errInWriting)
		return
	}

}
