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

func UploadFile2(w http.ResponseWriter, r *http.Request) {
	var csvFiles [4]multipart.File
	var collQDataset []models.QDatasetPerDay
	var bandwidth float64

	w.Header().Set("Content-Type", "multipart/form-data")

	// Parse our multipart form, 10 << 20 specifies a maximum upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)

	// Forms for QoS detail information
	idQos := primitive.NewObjectID()
	isp := strings.ToLower(r.FormValue("isp"))
	service := r.FormValue("service")
	name := r.FormValue("name")
	city := r.FormValue("city")

	// Sementara untuk bandwidth masih menyesuaikan nama service
	// Untuk nama service atau produk masih hardcoded
	//sb := r.FormValue("bandwidth")
	// bandwidth, err := strconv.ParseFloat(sb, 64)
	// if err != nil {
	// 	log.Println(err)
	// 	http.Error(w, "ERROR", http.StatusInternalServerError)
	// }

	if service == "Internet 10Mbps" {
		bandwidth = 10
	} else if service == "Internet 20Mbps" {
		bandwidth = 20
	} else if service == "Internet 30Mbps" {
		bandwidth = 30
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

	// Put the value to qosList
	qosList := models.QosList2{
		ID:            idQos,
		ISP:           isp,
		Service:       service,
		Bandwidth:     bandwidth,
		Customer_Name: name,
		City:          city,
	}

	// Processing the file input
	// forloop is used because there is 4 files that will submitted
	for i, param := range parameters {
		var err error
		var h *multipart.FileHeader

		csvFiles[i], h, err = r.FormFile(param)
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

		// Put the information and values from csv file to collQDatasetPerDay
		collQDatasetPerDay, datasetList, err := loadQdata(idQos, param, csvFiles[i])
		if err != nil {
			http.Error(w, "Failed to parse string value to date-time or float64", http.StatusBadRequest)
			return
		}

		qosList.Dataset_List = append(qosList.Dataset_List, datasetList...)
		collQDataset = append(collQDataset, collQDatasetPerDay...)
	}

	qosList.Upload_Date = time.Now()

	// Insert qosList and docsDataset to Mongodb
	insertQosToDB(qosList, collQDataset)
}

func GetAllQosList(w http.ResponseWriter, r *http.Request) {
	var qosList []models.QosList

	payload := getAllDocs(collQosList, qosList)

	// respond on client/browser
	json.NewEncoder(w).Encode(payload)
}

func GetRecapFilteredQos(w http.ResponseWriter, r *http.Request) {
	//var empty models.RecapFilteredQos
	vars := mux.Vars(r)
	payload, err := recapFilteredQos(
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

func GetRecapQosCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	payload, err := recapQosCustomer(vars["id"])
	if err != nil {
		log.Println(err)
		http.Error(w, "ERROR", http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(payload)
}

func GetOneQosRecord(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	parameter := mux.Vars(r)
	result := getOneQosRecord(parameter["id"])
	json.NewEncoder(w).Encode(result)
}

func DeleteOneQosRecord(w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	// w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	parameter := mux.Vars(r)
	deleteOneQosRecord(parameter["id"])
	//w.WriteHeader(http.StatusOK)
	//w.Write([]byte("DELETE ONE QOS LIST SUCCESS"))
	json.NewEncoder(w).Encode("SUCCESS: DELETED QOS RECORD WITH LIST ID " + parameter["id"])
}

// func GetAllIsp(w http.ResponseWriter, r *http.Request) {
// 	// get all isp
// 	var ispList []models.Isp

// 	result := getAllDocs(collIsp, ispList)

// 	json.NewEncoder(w).Encode(result)
// }
