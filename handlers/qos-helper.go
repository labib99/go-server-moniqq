package handlers

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"moniqq/models"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func loadQdata(id_qos primitive.ObjectID, param string, r io.Reader) ([]models.QDatasetPerDay, []models.ListDataset, error) {
	var num int
	var total, min, max float64
	var collQDatasetPerDay []models.QDatasetPerDay
	var dataset []models.QosData
	var qDatasetPerDay models.QDatasetPerDay
	var datasetList []models.ListDataset

	// Read file csv
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalln(err)
	}

	if len(records) <= 1 {
		return collQDatasetPerDay, datasetList, errors.New("data <1")
	}

	for row, values := range records {
		if row > 0 { // Omit header line
			var qData models.QosData
			var listDataset models.ListDataset
			for column, value := range values {
				if column == 0 { // DateTime
					var dt time.Time
					for k, layout := range dtLayout {
						dt, err = time.ParseInLocation(layout, value, timeLoc)
						if err != nil {
							k++
							if len(dtLayout) == k {
								log.Println(err)
								return collQDatasetPerDay, datasetList, err
							}
						} else {
							break
						}
					}
					qData.Date_Time = dt
				} else if column == 1 { // Value
					qData.Value, err = strconv.ParseFloat(value, 64)
					if err != nil {
						log.Println(err)
						return collQDatasetPerDay, datasetList, err
					}
					total += qData.Value
				}
			}
			if row == 1 {
				max = qData.Value
				min = qData.Value
			}

			if max < qData.Value {
				max = qData.Value
			}
			if min > qData.Value {
				min = qData.Value
			}
			year, month, day := qData.Date_Time.Date()
			date := time.Date(year, month, day, 0, 0, 0, 0, timeLoc)

			// When qosDataset was first created,
			// the value of qosDataset.Date equals to 0001-01-01 00:00:00 +0000 UTC
			if qDatasetPerDay.Date.Year() == 1 {
				qDatasetPerDay.Date = date
			}
			if qDatasetPerDay.Date == date {
				dataset = append(dataset, qData)
				num++
			}
			if qDatasetPerDay.Date != date || row == len(records)-1 {
				qDatasetPerDay = models.QDatasetPerDay{
					ID:            primitive.NewObjectID(),
					ID_Qos:        id_qos,
					Qos_Parameter: param,
					Date:          qDatasetPerDay.Date,
					Dataset:       dataset,
				}
				collQDatasetPerDay = append(collQDatasetPerDay, qDatasetPerDay)
				listDataset = models.ListDataset{
					Qos_Parameter: qDatasetPerDay.Qos_Parameter,
					ID_Dataset:    qDatasetPerDay.ID,
					Date:          qDatasetPerDay.Date,
					Num_Data:      num,
					Total_Value:   total,
					Min_Value:     min,
					Max_Value:     max,
				}
				datasetList = append(datasetList, listDataset)
				dataset = nil
				qDatasetPerDay.Date = date
				dataset = append(dataset, qData)
				num = 1
				total = 0
				min = qData.Value
				max = qData.Value
			}
		}
	}
	return collQDatasetPerDay, datasetList, nil
}

// Rate the qos parameter using TIPHON standard
func rating(param string, value, bandwidth float64) (index float32, category string) {
	if param == parameters[0] {
		// percent = (Throughput/bandwidth) x 100
		// bandwidth and throughput in Mbps
		percent := (value / bandwidth) * 100
		if percent > 75 {
			index = 4
			category = "Sangat Bagus"
		} else if percent > 50 && percent <= 75 {
			index = 3
			category = "Bagus"
		} else if percent >= 25 && percent <= 50 {
			index = 2
			category = "Sedang"
		} else if percent < 25 {
			index = 1
			category = "Buruk"
		}
	} else if param == parameters[1] {
		if value < 3 {
			index = 4
			category = "Sangat Bagus"
		} else if value >= 3 && value < 15 {
			index = 3
			category = "Bagus"
		} else if value >= 15 && value <= 25 {
			index = 2
			category = "Sedang"
		} else if value > 25 {
			index = 1
			category = "Buruk"
		}
	} else if param == parameters[2] {
		if value < 150 {
			index = 4
			category = "Sangat Bagus"
		} else if value >= 150 && value < 300 {
			index = 3
			category = "Bagus"
		} else if value >= 300 && value <= 450 {
			index = 2
			category = "Sedang"
		} else if value > 450 {
			index = 1
			category = "Buruk"
		}
	} else if param == parameters[3] {
		if value == 0 {
			index = 4
			category = "Sangat Bagus"
		} else if value > 0 && value <= 75 {
			index = 3
			category = "Bagus"
		} else if value >= 76 && value <= 125 {
			index = 2
			category = "Sedang"
		} else if value > 125 {
			index = 1
			category = "Buruk"
		}
	} else if param == "qos" {
		if 3.8 <= value && value <= 4 {
			index = float32(value)
			category = "Sangat Bagus"
		} else if 3 <= value && value <= 3.79 {
			index = float32(value)
			category = "Bagus"
		} else if 2 <= value && value <= 2.99 {
			index = float32(value)
			category = "Sedang"
		} else if 1 <= value && value <= 1.99 {
			index = float32(value)
			category = "Buruk"
		}
	}
	return index, category
}

// Insert one QosList and n QosDataset in the DB
func insertQosToDB(ql models.QosList2, qd []models.QDatasetPerDay) {
	result, err := collQosList.InsertOne(context.Background(), ql)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Inserted a Single QosList ", result.InsertedID)

	var docs []interface{}
	for _, d := range qd {
		docs = append(docs, d)
	}

	result2, err2 := collQosDataset.InsertMany(context.TODO(), docs)
	if err2 != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(qd); i++ {
		log.Println("Inserted QosDataset ", result2.InsertedIDs[i])
	}

}

// Get All docs from a collection in Mongodb
func getAllDocs(coll *mongo.Collection, results any) any {
	cursor, err := coll.Find(context.Background(), bson.D{{}})
	if err != nil {
		log.Fatalln(err)
	}

	if err = cursor.All(context.Background(), &results); err != nil {
		log.Fatalln(err)
	}

	return results
}

// Get qos recap based on qos_parameter and date
// If you want get  qos recap based on just qos_parameter,
// then the value of fromDateString and toDateString must be "no_date"
func recapFilteredQos(qosParam, isp, city, service, fromDateString, toDateString string) (models.RecapFilteredQos, error) {
	var bandwidth, totalAvg, overallMax, overallMin float64
	var unit string
	// var totalIndex float32
	var filteredQos models.RecapFilteredQosPerCustomer
	var collFilteredQos []models.RecapFilteredQosPerCustomer
	var qosList []models.QosList2
	var result models.RecapFilteredQos
	var match, cond primitive.A

	if qosParam == "throughput" {
		unit = "Mbps"
	} else if qosParam == "packet loss" {
		unit = "%"
	} else if qosParam == "jitter" || qosParam == "delay" {
		unit = "ms"
	}

	qosParam = strings.ToLower(qosParam)
	isp = strings.ToLower(isp)

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

	// Check if the value of fromDateString & toDateString equals to "no_date" or not
	if fromDateString == "no_date" && toDateString == "no_date" {
		match = bson.A{bson.D{{Key: "qos_parameter", Value: qosParam}}}
		cond = bson.A{bson.D{{Key: "$eq", Value: bson.A{"$$dataset_list.qos_parameter", qosParam}}}}
	} else {
		fromDate, err := time.ParseInLocation(dtLayout[len(dtLayout)-1], fromDateString, timeLoc)
		if err != nil {
			return result, err
		}
		toDate, err := time.ParseInLocation(dtLayout[len(dtLayout)-1], toDateString, timeLoc)
		if err != nil {
			return result, err
		}

		match = bson.A{bson.D{
			{Key: "qos_parameter", Value: qosParam},
			{Key: "date", Value: bson.D{
				{Key: "$gte", Value: fromDate}, {Key: "$lte", Value: toDate},
			}}}}
		cond = bson.A{
			bson.D{{Key: "$eq", Value: bson.A{"$$dataset_list.qos_parameter", qosParam}}},
			bson.D{{Key: "$gte", Value: bson.A{"$$dataset_list.date", fromDate}}},
			bson.D{{Key: "$lte", Value: bson.A{"$$dataset_list.date", toDate}}},
		}
	}

	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "isp", Value: isp},
		{Key: "city", Value: city},
		{Key: "service", Value: service},
		{Key: "bandwidth", Value: bandwidth},
		{Key: "dataset_list", Value: bson.D{
			{Key: "$elemMatch", Value: bson.D{
				{Key: "$and", Value: match},
			}}}}}}}
	projectStage := bson.D{{Key: "$project", Value: bson.D{
		//{Key: "isp", Value: 1},{Key: "service", Value: 1},
		//{Key: "bandwidth", Value: 1},{Key: "city", Value: 1},
		{Key: "_id", Value: 1},
		{Key: "customer_name", Value: 1},
		{Key: "upload_date", Value: 1},
		{Key: "dataset_list", Value: bson.D{
			{Key: "$filter", Value: bson.D{
				{Key: "input", Value: "$dataset_list"},
				{Key: "as", Value: "dataset_list"},
				{Key: "cond", Value: bson.D{
					{Key: "$and", Value: cond},
				}}}}}}}}}
	sortStage := bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: "upload_date", Value: 1},
		}}}
	cursor, err := collQosList.Aggregate(
		context.Background(), mongo.Pipeline{matchStage, projectStage, sortStage},
	)
	if err != nil {
		log.Fatalln(err)
	}

	if err = cursor.All(context.Background(), &qosList); err != nil {
		log.Fatalln(err)
	}

	for _, list := range qosList {
		var avg, total, max, min float64
		var num int
		for j, listDataset := range list.Dataset_List {
			total += listDataset.Total_Value
			num += listDataset.Num_Data
			if j == 0 {
				max = listDataset.Max_Value
				min = listDataset.Min_Value
				overallMax = max
				overallMin = min
			}
			if max < listDataset.Max_Value {
				max = listDataset.Max_Value
			}
			if min > listDataset.Min_Value {
				min = listDataset.Min_Value
			}
		}
		avg = total / float64(num)
		index, category := rating(qosParam, avg, bandwidth)
		filteredQos = models.RecapFilteredQosPerCustomer{
			ID_Qos:        list.ID,
			Customer_Name: list.Customer_Name,
			Average_Value: avg,
			Max_Value:     max,
			Min_Value:     min,
			Index_Rating:  index,
			Category:      category,
		}
		collFilteredQos = append(collFilteredQos, filteredQos)
		totalAvg += avg
		// totalIndex += index
		if overallMax < max {
			overallMax = max
		}
		if overallMin > min {
			overallMin = min
		}
	}

	overallAverage := totalAvg / float64(len(qosList))
	roundedOverallAverage := roundFloat(overallAverage, 3)
	// averageIndex := totalIndex / float32(len(qosList))
	// averageIndex = float32(roundFloat(float64(averageIndex), 2))

	index, category := rating(qosParam, overallAverage, bandwidth)
	result = models.RecapFilteredQos{
		Qos_Parameter:        qosParam,
		Unit:                 unit,
		Overall_Average:      roundedOverallAverage,
		Overall_Max_Value:    overallMax,
		Overall_Min_Value:    overallMin,
		Index_Rating:         index,
		Category:             category,
		Recap_F_Per_Customer: collFilteredQos,
	}
	return result, nil
}

// Get qos recap for one Customer
func recapQosCustomer(id string) (models.RecapQosCustomer, error) {
	var result models.RecapQosCustomer
	var qosList models.QosList2
	var collQDataset []models.QDatasetPerDay
	var totalIndex float32

	idQos, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return result, err
	}

	filter := bson.M{"_id": idQos}
	err = collQosList.FindOne(context.Background(), filter).Decode(&qosList)
	if err != nil {
		return result, err
	}

	filterDataset := bson.D{{Key: "id_qos", Value: idQos}}
	cursor, err := collQosDataset.Find(context.Background(), filterDataset)
	if err != nil {
		return result, err
	}

	if err = cursor.All(context.Background(), &collQDataset); err != nil {
		return result, err
	}

	for _, param := range parameters {
		var recap models.RecapQCustomerPerQParam
		var avg, total, max, min float64
		var num int
		var unit string

		recap.Qos_Parameter = param

		if param == "throughput" {
			unit = "Mbps"
		} else if param == "packet loss" {
			unit = "%"
		} else if param == "jitter" || param == "delay" {
			unit = "ms"
		}

		recap.Unit = unit

		for _, dataset := range collQDataset {
			if recap.Qos_Parameter == dataset.Qos_Parameter {
				recap.Dataset = append(recap.Dataset, dataset.Dataset...)
			}
		}
		for _, listDataset := range qosList.Dataset_List {
			if listDataset.Qos_Parameter == param {
				total += listDataset.Total_Value
				if num == 0 {
					max = listDataset.Max_Value
					min = listDataset.Min_Value
				}
				if max < listDataset.Max_Value {
					max = listDataset.Max_Value
				}
				if min > listDataset.Min_Value {
					min = listDataset.Min_Value
				}
				num += listDataset.Num_Data
			}
		}
		avg = total / float64(num)
		index, category := rating(param, avg, qosList.Bandwidth)

		recap.Average_Value = avg
		recap.Category = category
		recap.Index_Rating = index
		recap.Max_Value = max
		recap.Min_Value = min

		result.Recap_QCustomer_Per_QParam = append(result.Recap_QCustomer_Per_QParam, recap)
		totalIndex += index
	}
	averageIndex := totalIndex / float32(len(result.Recap_QCustomer_Per_QParam))
	index, category := rating("qos", float64(averageIndex), 0)

	result.Category = category
	result.Average_Index_Rating = index
	result.Bandwidth = qosList.Bandwidth
	result.Customer_Name = qosList.Customer_Name
	result.City = qosList.City
	result.Service = qosList.Service
	result.ISP = qosList.ISP
	result.ID = qosList.ID
	result.Upload_Date = qosList.Upload_Date

	return result, nil
}

// perlu diperbaiki
func getOneQosRecord(id string) models.QosRecord {
	var qosRecord models.QosRecord
	var qosInfo models.QosList
	var qosDatasets []models.QDatasetPerDay

	idQos, _ := primitive.ObjectIDFromHex(id)
	filterList := bson.M{"_id": idQos}
	cursor := collQosList.FindOne(context.Background(), filterList)
	if err := cursor.Err(); err != nil {
		panic(err)
	}

	if err := cursor.Decode(&qosInfo); err != nil {
		panic(err)
	}
	qosRecord.Qos_Info = qosInfo

	filterDataset := bson.M{"id_qos": idQos}
	cursor1, err1 := collQosDataset.Find(context.Background(), filterDataset)
	if err1 != nil {
		panic(err1)
	}

	if err1 = cursor1.All(context.Background(), &qosDatasets); err1 != nil {
		panic(err1)
	}
	qosRecord.Qos_Data = qosDatasets

	fmt.Printf("Get QosRecord from idqos: %s \n", id)

	return qosRecord
}

// bisa diedit
func deleteOneQosRecord(id string) {
	idQos, _ := primitive.ObjectIDFromHex(id)
	filterList := bson.M{"_id": idQos}
	d, err := collQosList.DeleteOne(context.TODO(), filterList)
	if err != nil {
		log.Fatal(err)
	}

	filterDataset := bson.M{"id_qos": idQos}
	d1, err1 := collQosDataset.DeleteMany(context.TODO(), filterDataset)
	if err1 != nil {
		log.Fatal(err1)
	}

	//update isp buat parameter

	log.Printf("Deleted %d QosList and %d QosDataset with idqos:%s\n", d.DeletedCount, d1.DeletedCount, id)
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

// func updateReports(idqos primitive.ObjectID, rep models.Report) {
// 	filter := bson.M{"_id": idqos}
// 	update := bson.M{"$push": bson.M{"reports": bson.M{"category": rep.Category, "average": rep.Average, "max": rep.Max, "min": rep.Min}}}
// 	result, err := collQosList.UpdateOne(context.Background(), filter, update)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Println("modified count: ", result.ModifiedCount)
// }
