package handlers

import (
	"encoding/json"
	"log"
	"mime/multipart"
	"moniqq/models"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Function of the endpoint for uploading QoS data
func UploadFile(w http.ResponseWriter, r *http.Request) {
	var csvFiles [4]multipart.File
	var qosDataset []models.DatasetQos
	var bandwidth float64

	w.Header().Set("Content-Type", "multipart/form-data")

	// Parse our multipart form, 10 << 20 specifies a maximum upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)

	// Forms for QoS detail information
	customerID := primitive.NewObjectID()
	isp := strings.ToLower(r.FormValue("isp"))
	service := r.FormValue("service")
	name := r.FormValue("name")
	city := r.FormValue("city")

	// In this capstone poject, the bandwidth is still in accordance with the name of the service
	// The name of the service is still hardcoded
	if service == "Internet 10Mbps" {
		bandwidth = 10
	} else if service == "Internet 20Mbps" {
		bandwidth = 20
	} else if service == "Internet 30Mbps" {
		bandwidth = 30
	} else if service == "Internet 40Mbps" {
		bandwidth = 40
	} else if service == "Internet 50Mbps" {
		bandwidth = 50
	} else if service == "Internet 100Mbps" {
		bandwidth = 100
	}

	// Check the value from all forms. If one of the form has no value then return error
	if isp == "" || service == "" || name == "" || city == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	// Put the value to customerISP
	customerISP := models.CustomerISP{
		ID:            customerID,
		ISP:           isp,
		Service:       service,
		Bandwidth:     bandwidth,
		Customer_Name: name,
		City:          city,
	}

	// Processing the file input
	// forloop is used because there is 4 files that will submitted
	for i, qosParam := range qosParameters {
		var err error
		var h *multipart.FileHeader

		csvFiles[i], h, err = r.FormFile(qosParam)
		// Check if the file is csv or not
		if h.Header.Get("Content-Type") != "text/csv" {
			http.Error(w, "Tipe file tidak sesuai", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Fatalln(err)
			return
		}

		// Put the information and values from csv file to qosDatasetOneParam
		qosDatasetOneParam, err := loadQosDatasetFromCSV(customerID, qosParam, csvFiles[i])
		if err != nil {
			http.Error(w, "Failed to parse string value to date-time or float64", http.StatusBadRequest)
			return
		}

		qosDataset = append(qosDataset, qosDatasetOneParam)
	}

	customerISP.Upload_Date = time.Now()

	// Insert data qos from one customer to Mongodb
	err := insertDataQosToMongoDB(customerISP, qosDataset)
	if err != nil {
		log.Println(err)
		http.Error(w, "ERROR", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("SUCCESS UPLOADING QOS DATA"))
}

// Function of the endpoint for obtaining QoS parameter analysis results from
// an ISP product in a certain city on a certain date
func GetRecapQosOneParamFilteredByDate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	payload, err := recapQosOneParamFilteredByDate(
		vars["qos_param"], vars["isp"], vars["city"],
		vars["service"], vars["from_date"], vars["to_date"],
	)
	if err != nil {
		log.Println(err)
		http.Error(w, "ERROR", http.StatusInternalServerError)
	}

	// If the payload is empty then return empty
	if payload.Index_Rating == 0 && payload.Category == "" {
		w.Write([]byte("empty"))
	} else {
		json.NewEncoder(w).Encode(payload)
	}
}

// Function of the endpoint for getting all list of customers ISP
func GetAllQosList(w http.ResponseWriter, r *http.Request) {
	var qosList []models.CustomerISP

	payload := getAllDocs(collCustomerISP, qosList)

	json.NewEncoder(w).Encode(payload)
}

// Function of the endpoint for getting the QoS analysis results of one customer
func GetRecapQosCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	payload, err := recapQosCustomer(vars["id"])
	if err != nil {
		log.Println(err)
		http.Error(w, "ERROR", http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(payload)
}

// Function of the endpoint for deleting QoS data from a single customer
func DeleteOneQosRecord(w http.ResponseWriter, r *http.Request) {
	parameter := mux.Vars(r)
	deleteOneQosRecord(parameter["id"])

	json.NewEncoder(w).Encode("SUCCESS: DELETED QOS RECORD WITH LIST ID " + parameter["id"])
}
