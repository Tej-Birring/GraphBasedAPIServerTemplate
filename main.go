package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/lestrrat-go/jwx/jwk"
	"log"
	"net/http"
	"os"
	"strconv"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// serve the API docs
	_, err := fmt.Fprint(w, "Welcome!\n")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func main() {
	// GET SERVER VARS
	err := godotenv.Load("creds/.env")
	if err != nil {
		log.Fatal("Error loading .env file.")
	}

	var port = 8080
	productionMode, _ := os.LookupEnv("RUN_IN_PRODUCTION")
	if productionMode == "true" {
		port = 80
	}
	portStr := strconv.Itoa(port)

	// LOAD SIGNATURE KEYS
	sigKeySetPrv, err := jwk.ReadFile("creds/.jwkSigPairSet.json")
	if err != nil {
		log.Fatal("Error loading JWK set from file.")
	}
	sigKeySetPub, err := jwk.PublicSetOf(sigKeySetPrv)
	if err != nil {
		log.Fatal("Error producing public JWK set from private one.")
	}

	// Initialize driver for DB communication
	dbDriver := InitializeDB()

	// CREATE SERVER
	mux := httprouter.New()
	// handle routes
	mux.GET("/", Index)
	HandleUsers(mux, &dbDriver, &sigKeySetPrv, &sigKeySetPub)

	// START SERVER
	log.Printf("Serving HTTP on port %s...\n", portStr)
	err = http.ListenAndServe(":"+portStr, mux)
	if err != nil {
		panic("Failed to run server!")
	}
}
