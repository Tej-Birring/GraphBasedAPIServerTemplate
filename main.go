package main

import (
	"GraphBasedServer/auth"
	"GraphBasedServer/configs"
	"GraphBasedServer/db"
	"GraphBasedServer/httpControllers"
	"GraphBasedServer/messaging"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
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
	// INITIALIZE SERVER
	configs.InitializeConfigs()
	auth.InitializeAuth()
	messaging.InitializeMessaging()
	dbController := db.InitializeDB()

	// HANDLE ROUTES
	mux := httprouter.New()
	mux.GET("/", handleIndex)
	httpControllers.HandleUserLogin(mux, &dbController)
	httpControllers.HandleUserEmailVerification(mux)
	httpControllers.HandleUserPhoneVerification(mux)

	// START SERVER
	port := configs.Configs.Port
	log.Printf("Serving HTTP on port %d...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
	if err != nil {
		panic("Failed to run server!")
	}
}
