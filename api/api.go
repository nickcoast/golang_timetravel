package api

import (
	"fmt"

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
	i.Path("/{type}/delete/{id:[0-9]+}").HandlerFunc(a.Delete).Methods("DELETE")
	i.Path("/{type}/getbydate/{insuredId}/{date}").HandlerFunc(a.GetResourceByDate).Methods("GET")

	ad := routes.PathPrefix("/address").Subrouter()
	ad.Path("/id/{id:[0-9]+}").HandlerFunc(a.GetRecords).Methods("GET")
}
