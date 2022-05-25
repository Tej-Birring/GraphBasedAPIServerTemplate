package main

import (
	"HayabusaBackend/auth"
	"HayabusaBackend/db"
	"HayabusaBackend/httpControllers"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
	"strconv"
)

/*
	handleIndex
	Return information about the API and how to use it.
*/
func handleIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// serve the API docs
	_, err := fmt.Fprint(w, "Welcome!\n")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func main() {
	// get environment vars used by the server
	err := godotenv.Load("creds/.env")
	if err != nil {
		log.Fatal("Error loading .env file.")
	}
	// use port 80 in production mode, or 8080 in debug mode
	// if RUN_IN_PRODUCTION is set to "true", then this will be
	// production mode
	var port = 8080
	productionMode, _ := os.LookupEnv("RUN_IN_PRODUCTION")
	if productionMode == "true" {
		port = 80
	}
	portStr := strconv.Itoa(port)

	// INITIALIZE SERVER SUBSYSTEMS
	// initialize driver for DB communication
	auth.InitializeAuth()
	dbController := db.InitializeDB()

	// INITIALIZE SERVER
	mux := httprouter.New()
	// handle routes
	mux.GET("/", handleIndex)
	httpControllers.HandleUserAccount(mux, &dbController)

	// START SERVER
	log.Printf("Serving HTTP on port %s...\n", portStr)
	err = http.ListenAndServe(":"+portStr, mux)
	if err != nil {
		panic("Failed to run server!")
	}
}
