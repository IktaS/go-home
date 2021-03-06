package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/IktaS/go-home/internal/app/store"
	"github.com/IktaS/go-home/internal/pkg/device"
	"github.com/IktaS/go-serv/pkg/serv"
	"github.com/gorilla/mux"
)

// DeviceHandlers is exported handlers for device
type DeviceHandlers struct{}

// HandleGetAllDevice handles getting all device
func (*DeviceHandlers) HandleGetAllDevice(repo store.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		devs, err := repo.GetAll()
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
}

// HandleGetDevice handles getting device
func (*DeviceHandlers) HandleGetDevice(repo store.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		val, ok := vars["id"]
		if !ok {
			http.Error(w, "No id", http.StatusBadRequest)
			return
		}
		dev, err := repo.Get(val)
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
}

// HandleGetDeviceService handles getting device service
func (*DeviceHandlers) HandleGetDeviceService(repo store.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		val, ok := vars["id"]
		if !ok {
			http.Error(w, "No id", http.StatusBadRequest)
			return
		}
		dev, err := repo.Get(val)
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
}

// HandleGetDeviceMessage handles getting device message
func (*DeviceHandlers) HandleGetDeviceMessage(repo store.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		val, ok := vars["id"]
		if !ok {
			http.Error(w, "No id", http.StatusBadRequest)
			return
		}
		dev, err := repo.Get(val)
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
}

// HandleDeviceServiceCall handles callign a device service
func (*DeviceHandlers) HandleDeviceServiceCall(repo store.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, ok := vars["id"]
		if !ok {
			http.Error(w, "No id", http.StatusBadRequest)
			return
		}
		dev, err := repo.Get(id)
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
		return fmt.Sprintf("{\"name\":\"%v\",\"request\":%v,\"Inbound\":%v,\"Outbound\":%v}",
			s.Name,
			typeArrayToJSON(s.Request),
			s.Inbound,
			s.Outbound,
		)
	}
	return fmt.Sprintf("{\"name\":\"%v\",\"Inbound\":%v,\"Outbound\":%v,\"response\":%v,\"request\":%v}",
		s.Name,
		s.Inbound,
		s.Outbound,
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
