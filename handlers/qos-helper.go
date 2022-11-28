package handlers

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"log"
	"moniqq/models"
	"strconv"
	"time"

	"github.com/montanaflynn/stats"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Load data from CSV file
func loadQosDatasetFromCSV(customerID primitive.ObjectID, qosParam string, r io.Reader) (models.DatasetQos, error) {
	var total, min, max float64
	var qosDataset models.DatasetQos
	var dataset []models.QosData
	var unit string

	if qosParam == "throughput" {
		unit = "Mbps"
	} else if qosParam == "packet loss" {
		unit = "%"
	} else if qosParam == "jitter" || qosParam == "delay" {
		unit = "ms"
	}

	// Read csv
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return qosDataset, err
	}

	if len(records) <= 1 {
		return qosDataset, errors.New("data <1")
	}

	for row, columns := range records {
		if row > 0 {
			var qosData models.QosData
			for column, value := range columns {
				if column == 0 { // DateTime
					var dt time.Time
					for i, layout := range dtLayout {
						dt, err = time.ParseInLocation(layout, value, timeLoc)
						if err != nil {
							i++
							if len(dtLayout) == i {
								log.Println(err)
								return qosDataset, err
							}
						} else {
							break
						}
					}
					qosData.Date_Time = dt
				} else if column == 1 { // value of qos parameter
					qosData.Value, err = strconv.ParseFloat(value, 64)
					if err != nil {
						log.Println(err)
						return qosDataset, err
					}
					total += qosData.Value
				}
			}
			if row == 1 {
				max = qosData.Value
				min = qosData.Value
			}

			if max < qosData.Value {
				max = qosData.Value
			}
			if min > qosData.Value {
				min = qosData.Value
			}
			dataset = append(dataset, qosData)
		}
	}

	qosDataset = models.DatasetQos{
		Customer_ID:   customerID,
		Qos_Parameter: qosParam,
		Unit:          unit,
		Num_Data:      len(dataset),
		Total_Value:   total,
		Max_Value:     max,
		Min_Value:     min,
		Dataset:       dataset,
	}

	return qosDataset, nil
}

// Insert data QoS to MongoDB
func insertDataQosToMongoDB(customerISP models.CustomerISP, qosDataset []models.DatasetQos) error {
	var docs []interface{}

	task1, err := collCustomerISP.InsertOne(context.Background(), customerISP)
	if err != nil {
		log.Println("ERROR: Can't insert documents to MongoDB")
		return err
	}
	log.Println("Success inserting a document with ID", task1.InsertedID, "to coll. customerISP ")

	for _, d := range qosDataset {
		docs = append(docs, d)
	}

	task2, err := collDatasetQos.InsertMany(context.Background(), docs)
	if err != nil {
		log.Println("ERROR: Can't insert documents to MongoDB")
		filter := bson.M{"_id": customerISP.ID}
		task3, _ := collCustomerISP.DeleteOne(context.TODO(), filter)
		log.Println("Deleted ", task3.DeletedCount, " document from coll. customerISP")
		return err
	}
	log.Println("Success inserting ", len(task2.InsertedIDs), " document(s) to coll. datasetQos")

	return nil
}

// Get qos recap based on isp name, qos_parameter, city, date
func recapQosOneParamFilteredByDate(qosParam, isp, city, service, fromDateString, toDateString string) (models.RecapFilteredQos, error) {
	var bandwidth, totalAvg, overallMax, overallMin float64
	var filteredQos models.RecapFilteredQosPerCustomer
	var collFilteredQos []models.RecapFilteredQosPerCustomer
	var customersQosData []models.CustomerQosData
	var recap models.RecapFilteredQos

	startDate, err := time.ParseInLocation(dtLayout[len(dtLayout)-1], fromDateString, timeLoc)
	if err != nil {
		return recap, err
	}
	toDate, err := time.ParseInLocation(dtLayout[len(dtLayout)-1], toDateString, timeLoc)
	if err != nil {
		return recap, err
	}
	endDate := toDate.AddDate(0, 0, 1)

	matchStage := bson.D{{
		Key: "$match", Value: bson.D{
			{Key: "isp", Value: isp},
			{Key: "city", Value: city},
			{Key: "service", Value: service},
		},
	}}

	lookUpStage := bson.D{{
		Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "datasetQos"},
			{Key: "let", Value: bson.D{{Key: "idqos", Value: "$_id"}}},
			{Key: "pipeline", Value: bson.A{
				bson.D{
					{Key: "$match", Value: bson.D{
						{Key: "$expr", Value: bson.D{
							{Key: "$eq", Value: bson.A{"$customer_id", "$$idqos"}},
						}},
						{Key: "qos_parameter", Value: qosParam},
						{Key: "dataset", Value: bson.D{
							{Key: "$elemMatch", Value: bson.D{
								{Key: "date_time", Value: bson.D{
									{Key: "$gte", Value: startDate},
									{Key: "$lt", Value: endDate},
								}},
							}},
						}},
					}},
				},
				bson.D{
					{Key: "$project", Value: bson.D{
						{Key: "qos_parameter", Value: 1},
						{Key: "unit", Value: 1},
						{Key: "customer_id", Value: 1},
						{Key: "dataset", Value: bson.D{
							{Key: "$filter", Value: bson.D{
								{Key: "input", Value: "$dataset"},
								{Key: "as", Value: "d"},
								{Key: "cond", Value: bson.D{
									{Key: "$and", Value: bson.A{
										bson.D{{Key: "$gte", Value: bson.A{"$$d.date_time", startDate}}},
										bson.D{{Key: "$lt", Value: bson.A{"$$d.date_time", endDate}}},
									}},
								}},
							}},
						}},
					}},
				},
			}},
			{Key: "as", Value: "qos_dataset"},
		},
	}}

	sortStage := bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: "upload_date", Value: 1},
		}},
	}

	cursor, err := collCustomerISP.Aggregate(
		context.Background(), mongo.Pipeline{matchStage, lookUpStage, sortStage},
	)
	if err != nil {
		log.Println(err)
		return recap, err
	}

	if err = cursor.All(context.Background(), &customersQosData); err != nil {
		log.Println(err)
		return recap, err
	}

	for i, customerQosData := range customersQosData {
		var values []float64
		var mean, max, min, stdDeviation float64

		for _, qosData := range customerQosData.Qos_Dataset[0].Dataset {
			values = append(values, qosData.Value)
		}

		mean, err = stats.Mean(values)
		if err != nil {
			log.Println(err)
			return recap, err
		}
		max, err = stats.Max(values)
		if err != nil {
			log.Println(err)
			return recap, err
		}
		min, err = stats.Min(values)
		if err != nil {
			log.Println(err)
			return recap, err
		}
		stdDeviation, err = stats.StandardDeviationSample(values)
		if err != nil {
			log.Println(err)
			return recap, err
		}

		bandwidth = customerQosData.Bandwidth
		index, category := rating(qosParam, mean, bandwidth)

		filteredQos = models.RecapFilteredQosPerCustomer{
			ID_Qos:        customerQosData.ID,
			Customer_Name: customerQosData.Customer_Name,
			Average_Value: mean,
			Std_Deviation: stdDeviation,
			Min_Value:     min,
			Max_Value:     max,
			Index_Rating:  index,
			Category:      category,
		}

		totalAvg += mean
		if i == 0 {
			overallMax = max
			overallMin = min
		}
		if overallMax < max {
			overallMax = max
		}
		if overallMin > min {
			overallMin = min
		}
		collFilteredQos = append(collFilteredQos, filteredQos)
	}

	overallAverage := totalAvg / float64(len(customersQosData))
	roundedOverallAverage, err := stats.Round(overallAverage, 3)
	if err != nil {
		log.Println(err)
		return recap, err
	}
	index, category := rating(qosParam, overallAverage, bandwidth)

	recap = models.RecapFilteredQos{
		Qos_Parameter:        qosParam,
		Unit:                 customersQosData[0].Qos_Dataset[0].Unit,
		Overall_Average:      roundedOverallAverage,
		Overall_Max_Value:    overallMax,
		Overall_Min_Value:    overallMin,
		Index_Rating:         index,
		Category:             category,
		Recap_F_Per_Customer: collFilteredQos,
	}

	return recap, nil
}

// Rate the qos parameter using TIPHON standard
func rating(param string, value, bandwidth float64) (index float32, category string) {
	if param == qosParameters[0] {
		// Percent = (Throughput/bandwidth) x 100
		// Bandwidth and throughput in Mbps
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
	} else if param == qosParameters[1] {
		if value < 3 {
			index = 4
			category = "Sangat Bagus"
		} else if value >= 3 && value < 15 {
			index = 3
			category = "Bagus"
		} else if value >= 15 && value < 25 {
			index = 2
			category = "Sedang"
		} else if value >= 25 {
			index = 1
			category = "Buruk"
		}
	} else if param == qosParameters[2] {
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
	} else if param == qosParameters[3] {
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

// Get All docs from a collection in Mongodb
func getAllDocs(coll *mongo.Collection, results any) any {
	opt := options.Find().SetSort(bson.D{{Key: "upload_date", Value: -1}})
	cursor, err := coll.Find(context.Background(), bson.D{{}}, opt)
	if err != nil {
		log.Fatalln(err)
	}

	if err = cursor.All(context.Background(), &results); err != nil {
		log.Fatalln(err)
	}

	return results
}

// Get qos recap for one customer
func recapQosCustomer(id string) (models.RecapQosCustomer, error) {
	var recapQCustomer models.RecapQosCustomer
	var collQDataset []models.DatasetQos
	var totalIndex float32

	idQos, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return recapQCustomer, err
	}

	filter := bson.M{"_id": idQos}
	err = collCustomerISP.FindOne(context.Background(), filter).Decode(&recapQCustomer)
	if err != nil {
		return recapQCustomer, err
	}

	filterDataset := bson.D{{Key: "customer_id", Value: idQos}}
	cursor, err := collDatasetQos.Find(context.Background(), filterDataset)
	if err != nil {
		return recapQCustomer, err
	}

	if err = cursor.All(context.Background(), &collQDataset); err != nil {
		return recapQCustomer, err
	}

	for _, dataset := range collQDataset {
		var recap models.RecapQCustomerPerQParam
		var values []float64

		recap.Qos_Parameter = dataset.Qos_Parameter
		recap.Unit = dataset.Unit

		avg := dataset.Total_Value / float64(dataset.Num_Data)
		recap.Average_Value, err = stats.Round(avg, 3)
		if err != nil {
			return recapQCustomer, err
		}

		recap.Index_Rating, recap.Category = rating(
			recap.Qos_Parameter, recap.Average_Value, recapQCustomer.Bandwidth)
		recap.Max_Value = dataset.Max_Value
		recap.Min_Value = dataset.Min_Value
		recap.Dataset = dataset.Dataset

		for _, data := range dataset.Dataset {
			values = append(values, data.Value)
		}
		recap.Std_Deviation, err = stats.StandardDeviation(values)
		if err != nil {
			return recapQCustomer, err
		}

		recapQCustomer.Recap_QCustomer_Per_QParam = append(recapQCustomer.Recap_QCustomer_Per_QParam, recap)
		totalIndex += recap.Index_Rating
	}
	averageIndex := totalIndex / float32(len(recapQCustomer.Recap_QCustomer_Per_QParam))
	recapQCustomer.Average_Index_Rating, recapQCustomer.Category = rating(
		"qos", float64(averageIndex), 0)

	return recapQCustomer, nil
}

// Delete QoS data from one customer
func deleteOneQosRecord(id string) {
	idQos, _ := primitive.ObjectIDFromHex(id)
	filterList := bson.M{"_id": idQos}
	d, err := collCustomerISP.DeleteOne(context.TODO(), filterList)
	if err != nil {
		log.Fatal(err)
	}

	filterDataset := bson.M{"customer_id": idQos}
	d1, err1 := collDatasetQos.DeleteMany(context.TODO(), filterDataset)
	if err1 != nil {
		log.Fatal(err1)
	}

	log.Printf(
		"Deleted %d QosList and %d QosDataset with idqos:%s\n", d.DeletedCount, d1.DeletedCount, id)
}
