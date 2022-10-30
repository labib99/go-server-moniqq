package main

import (
	"log"
	"moniqq/handlers"
	"moniqq/router"
	"net/http"
	"os"

	gohandlers "github.com/gorilla/handlers"
)

func main() {
	r := router.Router()

	// CORS
	corsHandler := gohandlers.CORS(
		gohandlers.AllowedOrigins([]string{os.Getenv("CORS")}),
		gohandlers.AllowedMethods([]string{"OPTIONS", "GET", "POST", "DELETE"}),
		gohandlers.AllowCredentials(),
	)

	h := handlers.SessionManager.LoadAndSave(handlers.Authorizer(r))

	server := http.Server{
		Addr:    ":9000",
		Handler: corsHandler(h),
	}

	log.Println("Starting go-server on the port 9000...")
	log.Fatal(server.ListenAndServe())
}
