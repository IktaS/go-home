package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/IktaS/go-home/internal/app"
	"github.com/IktaS/go-home/internal/pkg/device"
	"github.com/IktaS/go-serv/pkg/serv"
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
		jsonString += DeviceToJSON(dev)
	}
	jsonString += "]"
	fmt.Fprintf(w, jsonString)
}

func getDeviceHandler(w http.ResponseWriter, r *http.Request, a *app.App) {
	vars := mux.Vars(r)
	val, ok := vars["id"]
	if !ok {
		http.Error(w, "No id", http.StatusBadRequest)
		return
	}
	dev, err := a.Devices.Get(val)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, DeviceToJSON(dev))
}

func getDeviceServiceHandler(w http.ResponseWriter, r *http.Request, a *app.App) {
	vars := mux.Vars(r)
	val, ok := vars["id"]
	if !ok {
		http.Error(w, "No id", http.StatusBadRequest)
		return
	}
	dev, err := a.Devices.Get(val)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, ServiceToJSON(dev))
}

func getDeviceMessageHandler(w http.ResponseWriter, r *http.Request, a *app.App) {
	vars := mux.Vars(r)
	val, ok := vars["id"]
	if !ok {
		http.Error(w, "No id", http.StatusBadRequest)
		return
	}
	dev, err := a.Devices.Get(val)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, MessagesToJSON(dev))
}

func deviceServiceCallHandler(w http.ResponseWriter, r *http.Request, a *app.App) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		http.Error(w, "No id", http.StatusBadRequest)
		return
	}
	dev, err := a.Devices.Get(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	service, ok := vars["service"]
	if !ok {
		http.Error(w, "No service", http.StatusBadRequest)
		return
	}
	body, err := dev.Call(service, r.URL.RawQuery)
	if err != nil {
		http.Error(w, "Cannot Call Device", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, string(body))
}

// DeviceHandlers add routes to handle device operation, getting, updating and such
func DeviceHandlers(r *mux.Router, a *app.App) {
	s := r.PathPrefix("/device").Subrouter()
	s.HandleFunc("/", appHandlerWrapper(getAllDeviceHandler, a)).Methods("GET")
	s.HandleFunc("/{id}", appHandlerWrapper(getDeviceHandler, a)).Methods("GET")
	s.HandleFunc("/{id}/services", appHandlerWrapper(getDeviceServiceHandler, a)).Methods("GET")
	s.HandleFunc("/{id}/services/{service}", appHandlerWrapper(deviceServiceCallHandler, a)).Methods("GET")
	s.HandleFunc("/{id}/messages", appHandlerWrapper(getDeviceMessageHandler, a)).Methods("GET")
}

//DeviceToJSON returns a json string that represent the device
func DeviceToJSON(d *device.Device) string {
	ret := fmt.Sprintf("{\"id\":\"%v\",\"addr\":\"%v\",\"name\":\"%v\",\"services\":\"%v\",\"messages\":\"%v\"}",
		d.ID.String(),
		d.Addr.String(),
		d.Name,
		fmt.Sprintf("%v/device/%v/services", os.Getenv("APP_URL"), d.ID.String()),
		fmt.Sprintf("%v/device/%v/messages", os.Getenv("APP_URL"), d.ID.String()),
	)
	return ret
}

//ServiceToJSON returns a json string that represent the device
func ServiceToJSON(d *device.Device) string {
	return serviceArrayToJSON(d.Services)
}

//MessagesToJSON returns a json string that represent the device
func MessagesToJSON(d *device.Device) string {
	return messageArrayToJSON(d.Messages)
}

func messageArrayToJSON(messages []*serv.Message) string {
	ret := "["
	notfirst := false
	for _, m := range messages {
		if notfirst {
			ret += ","
		}
		notfirst = true
		ret += messageToJSON(m)
	}
	ret += "]"
	return ret
}

func messageToJSON(m *serv.Message) string {
	ret := fmt.Sprintf("{\"name\":\"%v\",\"definitions\":%v}",
		m.Name,
		mesDefinitionArrayToJSON(m.Definitions),
	)
	return ret
}

func mesDefinitionArrayToJSON(mesDefs []*serv.MessageDefinition) string {
	ret := "["
	notfirst := false
	for _, m := range mesDefs {
		if notfirst {
			ret += ","
		}
		notfirst = true
		ret += mesDefToJSON(m)
	}
	ret += "]"
	return ret
}

func mesDefToJSON(m *serv.MessageDefinition) string {
	if m.Field != nil {
		f := m.Field
		isOptional := f.Optional && !f.Required
		return fmt.Sprintf("{\"name\":\"%v\",\"isOptional\":\"%v\",\"value\":%v}",
			f.Name,
			isOptional,
			typeToJSON(f.Type),
		)
	}
	return "\"None\""
}

func serviceArrayToJSON(services []*serv.Service) string {
	ret := "["
	notfirst := false
	for _, s := range services {
		if notfirst {
			ret += ","
		}
		notfirst = true
		ret += serviceToJSON(s)
	}
	ret += "]"
	return ret
}

func serviceToJSON(s *serv.Service) string {
	if s.Response == nil {
		return fmt.Sprintf("{\"name\":\"%v\",\"request\":%v}",
			s.Name,
			typeArrayToJSON(s.Request),
		)
	}
	return fmt.Sprintf("{\"name\":\"%v\",\"response\":%v,\"request\":%v}",
		s.Name,
		typeToJSON(s.Response),
		typeArrayToJSON(s.Request),
	)
}

func typeArrayToJSON(types []*serv.Type) string {
	ret := "["
	notfirst := false
	for _, t := range types {
		if notfirst {
			ret += ","
		}
		notfirst = true
		ret += typeToJSON(t)
	}
	ret += "]"
	return ret
}

func typeToJSON(t *serv.Type) string {
	if t.Reference == "" {
		return fmt.Sprintf("{\"isScalar\":\"true\",\"value\":\"%v\"}", t.Scalar.String())
	}
	return fmt.Sprintf("{\"isScalar\":\"false\",\"value\":\"%v\"}", t.Reference)
}
