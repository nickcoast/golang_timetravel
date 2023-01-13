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
	i := routes.PathPrefix("/insured").Subrouter()
	i.Path("/id/{id}").HandlerFunc(a.GetInsuredsById).Methods("GET")
	i.Path("/new").HandlerFunc(a.CreateInsured).Methods("POST")
	i.Path("/delete/{id}").HandlerFunc(a.DeleteInsured).Methods("DELETE")

	e := routes.PathPrefix("/employee").Subrouter()
	e.Path("/id/{id}").HandlerFunc(a.GetRecords).Methods("GET")

	ad := routes.PathPrefix("/address").Subrouter()
	ad.Path("/id/{id}").HandlerFunc(a.GetRecords).Methods("GET")
}
