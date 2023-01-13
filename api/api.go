package api

import (
	"github.com/gorilla/mux"
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
	a.CreateV1Routes(routesV1)
	a.CreateV2Routes(routesV2)
}

func (a *API) CreateV1Routes(routes *mux.Router) {
	routes.Path("/records/{id}").HandlerFunc(a.GetRecords).Methods("GET")
	routes.Path("/records/{id}").HandlerFunc(a.PostRecords).Methods("POST")
}
func (a *API) CreateV2Routes(routes *mux.Router) {	
	routes.Path("/insured/id/{id}").HandlerFunc(a.GetInsuredsById).Methods("GET")

	routes.Path("/insured/new").HandlerFunc(a.CreateInsured).Methods("POST")
	routes.Path("/insured/delete/{id}").HandlerFunc(a.DeleteInsured).Methods("DELETE")

	routes.Path("/employee/id/{id}").HandlerFunc(a.GetRecords).Methods("GET")
}
