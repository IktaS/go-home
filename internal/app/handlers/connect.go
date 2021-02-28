package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/IktaS/go-home/internal/app"
	"github.com/IktaS/go-home/internal/pkg/auth"
	"github.com/IktaS/go-home/internal/pkg/device"
	"github.com/gorilla/mux"
)

/*
newConnection defines a device connect JSON payload :
	HubCode 	`hub-code`	: To authenticate that device is an authenticated device that user actually want to connect to hub
	Name		`name`		: Device Name
	Serv 		`serv`		: An compressed text message of the device respective .serv definition
	Algorithm 	`algo`		: Defines what algorithm they use to compress said Serv file
*/
type newConnection struct {
	ID        interface{} `json:"id,omitempty"`
	HubCode   string      `json:"hub-code"`
	Name      string      `json:"name"`
	Serv      string      `json:"serv"`
	Algorithm string      `json:"algo"`
}

func connectHandler(w http.ResponseWriter, r *http.Request, a *app.App) {
	var newconn newConnection
	err := json.NewDecoder(r.Body).Decode(&newconn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !auth.Authenticate(newconn.HubCode) {
		http.Error(w, "Wrong Hub Code", http.StatusBadRequest)
		return
	}
	log.Println("New connection from :" + r.RemoteAddr)
	ipStr := strings.Split(r.RemoteAddr, ":")
	ip := net.ParseIP(ipStr[0])
	addr := &net.IPAddr{IP: ip, Zone: ""}
	if newconn.ID != nil {
		dev, err := a.Devices.Get(newconn.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		dev.Addr = addr
		a.Devices.Save(dev)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Device Reconnected to Hub!")
		return
	}
	var DecompServ []byte
	switch algo := newconn.Algorithm; algo {
	case "none":
		DecompServ = []byte(newconn.Serv)
	}
	dev, err := device.NewDevice(newconn.Name, addr, DecompServ)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = a.Devices.Save(dev)
	if err != nil {
		http.Error(w, "Error Saving New Device \n"+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, dev.ID.String())
}

/*
NewServiceCall defines a service call JSON payload :
	DeviceID `device-id` 	: The ID of the device you want to call a service on
	Service	 `service`		: The service name that they want to call
	Data 	 `data`			: A json object that has all the data you want to pass to the service, it's in "name" : "value" format
*/
type newServiceCall struct {
	DeviceID string                 `json:"device-id"`
	Service  string                 `json:"service"`
	Data     map[string]interface{} `json:"data"`
}

// ConnectHandlers add routes to handle connecting and calling endpoints
func ConnectHandlers(r *mux.Router, a *app.App) {
	s := r.NewRoute().Subrouter()
	s.HandleFunc("/connect", appHandlerWrapper(connectHandler, a))
}
