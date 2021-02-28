package main

import (
	"github.com/IktaS/go-home/internal/app/handlers"
	"github.com/gorilla/mux"
)

func (s *Server) routes() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	//Device Handler
	deviceHandlers := &handlers.DeviceHandlers{}
	subrouter := r.PathPrefix("/device").Subrouter()
	subrouter.HandleFunc("/", deviceHandlers.HandleGetAllDevice(s.store)).Methods("GET")
	subrouter.HandleFunc("/{id}", deviceHandlers.HandleGetDevice(s.store)).Methods("GET")
	subrouter.HandleFunc("/{id}/service", deviceHandlers.HandleGetDeviceService(s.store)).Methods("GET")
	subrouter.HandleFunc("/{id}/service/{service}", deviceHandlers.HandleDeviceServiceCall(s.store)).Methods("GET")
	subrouter.HandleFunc("/{id}/message", deviceHandlers.HandleGetDeviceMessage(s.store)).Methods("GET")

	//Connect Handler
	connectHandlers := &handlers.ConnectionHandlers{}
	r.HandleFunc("/connect", connectHandlers.HandleConnect(s.store)).Methods("POST")

	return r
}
