package router

import (
	"moniqq/handlers"
	"net/http"

	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	router := mux.NewRouter()

	adminAPI := router.PathPrefix("/admin").Subrouter()
	adminAPI.HandleFunc("/qos", handlers.GetAllQosList).Methods(http.MethodGet)
	adminAPI.HandleFunc("/qos/upload", handlers.UploadFile).Methods(http.MethodPost)
	adminAPI.HandleFunc("/qos/delete/{id}", handlers.DeleteOneQosRecord).Methods("DELETE", "OPTIONS")
	adminAPI.HandleFunc("/qos/{id}", handlers.GetRecapQosCustomer).Methods(http.MethodGet)

	qosAPI := router.PathPrefix("/qos").Subrouter()
	qosAPI.HandleFunc("/{qos_param}/{isp}/{city}/{service}/{from_date}/{to_date}", handlers.GetRecapQosOneParamFilteredByDate).Methods(http.MethodGet)

	authAPI := router.PathPrefix("").Subrouter()
	authAPI.HandleFunc("/whoami", handlers.WhoAmI).Methods(http.MethodGet)
	authAPI.HandleFunc("/login", handlers.Login).Methods(http.MethodPost)
	authAPI.HandleFunc("/logout", handlers.Logout).Methods(http.MethodGet)

	return router
}
