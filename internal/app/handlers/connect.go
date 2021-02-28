package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/IktaS/go-home/internal/app/store"
	"github.com/IktaS/go-home/internal/pkg/auth"
	"github.com/IktaS/go-home/internal/pkg/device"
)

//ConnectionHandlers is handlers for connection
type ConnectionHandlers struct{}

/*
newConnection defines a device connect JSON payload :
	HubCode 	`hub-code`	: To authenticate that device is an authenticated device that user actually want to connect to hub
	Name		`name`		: Device Name
	Serv 		`serv`		: An compressed text message of the device respective .serv definition
	Algorithm 	`algo`		: Defines what algorithm they use to compress said Serv file
*/
type newConnection struct {
	ID        interface{} `json:"id,omitempty"`
	Addr      string      `json:"addr,omitempty"`
	HubCode   string      `json:"hub-code"`
	Name      string      `json:"name"`
	Serv      string      `json:"serv"`
	Algorithm string      `json:"algo"`
}

// HandleConnect handles connecting a device to the hub
func (*ConnectionHandlers) HandleConnect(repo store.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		log.Println("New connection from :\t" + r.RemoteAddr)
		var addr net.Addr
		if newconn.Addr == "" {
			ipStr := strings.Split(r.RemoteAddr, ":")
			ip := net.ParseIP(ipStr[0])
			addr = &net.IPAddr{IP: ip, Zone: ""}
		} else {
			ipStr := strings.Split(newconn.Addr, ":")
			ip := net.ParseIP(ipStr[0])
			addr = &net.IPAddr{IP: ip, Zone: ""}
		}
		if newconn.ID != nil {
			dev, err := repo.Get(newconn.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			dev.Addr = addr
			repo.Save(dev)
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
		err = repo.Save(dev)
		if err != nil {
			http.Error(w, "Error Saving New Device \n"+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, dev.ID.String())
	}
}
