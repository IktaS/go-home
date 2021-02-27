package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/IktaS/go-home/internal/app"
	"github.com/IktaS/go-home/internal/app/handlers"
	"github.com/IktaS/go-home/internal/app/store/sqlite"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

//HomeHandler handles home
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Works")
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				if os.Getenv("PORT") == "" {
					return ipnet.IP.String()
				}
				return ipnet.IP.String() + ":" + os.Getenv("PORT")
			}
		}
	}
	return ""
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func loadEnv() {
	env := os.Getenv("ENV")
	if "" == env {
		env = "development"
	}

	godotenv.Load(".env." + env + ".local")
	if "test" != env {
		godotenv.Load(".env.local")
	}
	godotenv.Load(".env." + env)
	godotenv.Load() // The Original .env

	if os.Getenv("PORT") == "" {
		os.Setenv("APP_URL", os.Getenv("URL"))
	} else {
		os.Setenv("APP_URL", os.Getenv("URL")+":"+os.Getenv("PORT"))
	}
}

func main() {
	loadEnv()
	repo, err := sqlite.NewSQLiteStore("sqlite.db")
	if err != nil {
		panic(err)
	}
	a := app.NewApp(repo)
	r := mux.NewRouter().StrictSlash(true)
	handlers.ConnectHandlers(r, a)
	handlers.DeviceHandlers(r, a)
	r.Use(loggingMiddleware)
	srv := &http.Server{
		Handler: r,
		Addr:    os.Getenv("APP_URL"),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("App running in	:\t" + srv.Addr)
	log.Println("App local IP	:\t" + getLocalIP())
	log.Fatal(srv.ListenAndServe())
}
