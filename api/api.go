package api

import (
	"github.com/gorilla/mux"
	"github.com/nickcoast/timetravel/service"
)

type API struct {
	records service.RecordService
	sqlite  service.RecordService
}

func NewAPI(records service.RecordService, sqlite service.RecordService) *API {
	return &API{records, sqlite}
}

// generates all api routes
func (a *API) CreateRoutes(routes *mux.Router) {
	routes.Path("/records/{id}").HandlerFunc(a.GetRecords).Methods("GET")
	routes.Path("/records/{id}").HandlerFunc(a.PostRecords).Methods("POST")
	routes.Path("/insured/id/{id}").HandlerFunc(a.GetInsuredsById).Methods("GET")
	routes.Path("/employee/id/{id}").HandlerFunc(a.GetRecords).Methods("GET")	
}
