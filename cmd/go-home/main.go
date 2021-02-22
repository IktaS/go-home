package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/IktaS/go-home/internal/auth"
	"github.com/IktaS/go-home/internal/device"
	"github.com/IktaS/go-home/internal/store"
	"github.com/IktaS/go-home/internal/store/sqlite"
	"github.com/gorilla/mux"
)

//HomeHandler handles home
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Works")
}

// App devices what the app have
type App struct {
	Devices store.Repo
}

// NewApp creates a new app
func NewApp(r store.Repo) *App {
	return &App{
		Devices: r,
	}
}

/*
NewConnection defines a device connect JSON payload :
	HubCode 	`hub-code`	: To authenticate that device is an authenticated device that user actually want to connect to hub
	Name		`name`		: Device Name
	Serv 		`serv`		: An compressed text message of the device respective .serv definition
	Algorithm 	`algo`		: Defines what algorithm they use to compress said Serv file
*/
type NewConnection struct {
	HubCode   string `json:"hub-code"`
	Name      string `json:"name"`
	Serv      []byte `json:"serv"`
	Algorithm string `json:"algo"`
}

//ConnectHandler handles a device connecting to Hub
func ConnectHandler(w http.ResponseWriter, r *http.Request, a *App) {
	var newconn NewConnection
	err := json.NewDecoder(r.Body).Decode(&newconn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !auth.Authenticate(newconn.HubCode) {
		http.Error(w, "Wrong Hub Code", http.StatusBadRequest)
		return
	}
	var DecompServ []byte
	switch algo := newconn.Algorithm; algo {
	case "none":
		DecompServ = newconn.Serv
	}
	ip := net.ParseIP(r.RemoteAddr)
	addr := &net.IPAddr{IP: ip, Zone: ""}
	dev, err := device.NewDevice(newconn.Name, addr, DecompServ)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = a.Devices.Save(dev)
	if err != nil {
		http.Error(w, "Error Saving New Device", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "Device Connected to Hub!")
}

/*
NewServiceCall defines a service call JSON payload :
	DeviceID `device-id` 	: The ID of the device you want to call a service on
	Service	 `service`		: The service name that they want to call
	Data 	 `data`			: A json object that has all the data you want to pass to the service, it's in "name" : "value" format
*/
type NewServiceCall struct {
	DeviceID string                 `json:"device-id"`
	Service  string                 `json:"service"`
	Data     map[string]interface{} `json:"data"`
}

//ServiceHandler handles a service call
func ServiceHandler(w http.ResponseWriter, r *http.Request, a *App) {
	var newcall NewServiceCall
	err := json.NewDecoder(r.Body).Decode(&newcall)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dev, err := a.Devices.Get(newcall.DeviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res, err := dev.Call(newcall.Service, newcall.Data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, string(res))
}

func appHandlerWrapper(f func(http.ResponseWriter, *http.Request, *App), a *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(w, r, a)
	}
}

func main() {
	repo, err := sqlite.NewSQLiteStore("sqlite.db")
	if err != nil {
		panic(err)
	}
	app := NewApp(repo)
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/connect", appHandlerWrapper(ConnectHandler, app))
	r.HandleFunc("/service", appHandlerWrapper(ServiceHandler, app))
}
