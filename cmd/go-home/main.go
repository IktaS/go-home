package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/IktaS/go-home/internal/app"
	"github.com/IktaS/go-home/internal/app/handlers"
	"github.com/IktaS/go-home/internal/app/store/sqlite"
	"github.com/gorilla/mux"
)

//HomeHandler handles home
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Works")
}

func getLocalIP() net.IP {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP
			}
		}
	}
	return nil
}

func main() {
	repo, err := sqlite.NewSQLiteStore("sqlite.db")
	if err != nil {
		panic(err)
	}
	a := app.NewApp(repo)
	r := mux.NewRouter()
	handlers.ConnectHandlers(r, a)
	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:5575",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("App running in : " + getLocalIP().String())
	log.Fatal(srv.ListenAndServe())
}
