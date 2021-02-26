package handlers

import (
	"fmt"
	"net/http"

	"github.com/IktaS/go-home/internal/app"
	"github.com/gorilla/mux"
)

func getAllDeviceHandler(w http.ResponseWriter, r *http.Request, a *app.App) {
	devs, err := a.Devices.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonString := "["
	notFirst := false
	for _, dev := range devs {
		if notFirst {
			jsonString += ","
		}
		notFirst = true
		jsonString += dev.ToJSON()
	}
	jsonString += "]"
	fmt.Fprintf(w, jsonString)
}

// DeviceHandlers add routes to handle device operation, getting, updating and such
func DeviceHandlers(r *mux.Router, a *app.App) {
	s := r.PathPrefix("/device").Subrouter()
	s.HandleFunc("/", appHandlerWrapper(getAllDeviceHandler, a))
}
