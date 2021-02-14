package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/IktaS/go-home/internal/auth"
	"github.com/IktaS/go-home/internal/device"
	"github.com/IktaS/go-home/internal/store"
	"github.com/IktaS/go-home/internal/store/postgres"
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
	HubCode 	: To authenticate that device is an authenticated device that user actually want to connect to hub
	Name		: Device Name
	Serv 		: An compressed text message of the device respective .serv definition
	Algorithm 	: Defines what algorithm they use to compress said Serv file
*/
type NewConnection struct {
	HubCode   string `json:"hub-code"`
	Name      string `json:"name"`
	Serv      []byte `json:"serv"`
	Algorithm string `json:"algorithm"`
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
}

func appHandlerWrapper(f func(http.ResponseWriter, *http.Request, *App), a *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(w, r, a)
	}
}

func main() {
	repo, err := postgres.NewPostgreSQLStore("")
	if err != nil {
		panic(err)
	}
	app := NewApp(repo)
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/connect", appHandlerWrapper(ConnectHandler, app))
}
