package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomerISP struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ISP           string             `json:"isp"`
	Service       string             `json:"service"`
	Bandwidth     float64            `json:"bandwidth"`
	Customer_Name string             `json:"customer_name"`
	City          string             `json:"city"`
	Upload_Date   time.Time          `json:"upload_date"`
}

type DatasetQos struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Customer_ID   primitive.ObjectID `json:"customer_id,omitempty" bson:"customer_id,omitempty"`
	Qos_Parameter string             `json:"qos_parameter"`
	Unit          string             `json:"unit"`
	Num_Data      int                `json:"num_data"` //number of QosData
	Total_Value   float64            `json:"total_value"`
	Min_Value     float64            `json:"min_value"`
	Max_Value     float64            `json:"max_value"`
	Dataset       []QosData          `json:"dataset"`
}

type QosData struct {
	Date_Time time.Time `json:"date_time"`
	Value     float64   `json:"value"`
}

type CustomerQosData struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ISP           string             `json:"isp"`
	Service       string             `json:"service"`
	Bandwidth     float64            `json:"bandwidth"`
	Customer_Name string             `json:"customer_name"`
	City          string             `json:"city"`
	Upload_Date   time.Time          `json:"upload_date"`
	Qos_Dataset   []DatasetQos       `json:"qos_dataset"`
}

type RecapFilteredQos struct {
	Qos_Parameter        string                        `json:"qos_parameter"`
	Unit                 string                        `json:"unit"`
	Overall_Average      float64                       `json:"overall_average"`
	Std_Deviation        float64                       `json:"std_deviation"`
	Overall_Min_Value    float64                       `json:"overall_min_value"`
	Overall_Max_Value    float64                       `json:"overall_max_value"`
	Index_Rating         float32                       `json:"index_rating"`
	Category             string                        `json:"category"`
	Recap_F_Per_Customer []RecapFilteredQosPerCustomer `json:"recap_f_per_customer"`
}

type RecapFilteredQosPerCustomer struct {
	ID_Qos        primitive.ObjectID `json:"id_qos,omitempty" bson:"id_qos,omitempty"`
	Customer_Name string             `json:"customer_name,omitempty"`
	Average_Value float64            `json:"average_value,omitempty"`
	Std_Deviation float64            `json:"std_deviation,omitempty"`
	Min_Value     float64            `json:"min_value"`
	Max_Value     float64            `json:"max_value"`
	Index_Rating  float32            `json:"index_rating,omitempty"`
	Category      string             `json:"category,omitempty"`
}

type RecapQosCustomer struct {
	ID                         primitive.ObjectID        `json:"_id" bson:"_id"`
	ISP                        string                    `json:"isp"`
	Service                    string                    `json:"service"`
	Bandwidth                  float64                   `json:"bandwidth"`
	Customer_Name              string                    `json:"customer_name"`
	City                       string                    `json:"city"`
	Upload_Date                time.Time                 `json:"upload_date"`
	Average_Index_Rating       float32                   `json:"average_index_rating"`
	Category                   string                    `json:"category"`
	Recap_QCustomer_Per_QParam []RecapQCustomerPerQParam `json:"recap_qcustomer_per_qparam"`
}

type RecapQCustomerPerQParam struct {
	Qos_Parameter string    `json:"qos_parameter,omitempty"`
	Unit          string    `json:"unit"`
	Average_Value float64   `json:"average_value,omitempty"`
	Std_Deviation float64   `json:"std_deviation,omitempty"`
	Min_Value     float64   `json:"min_value"`
	Max_Value     float64   `json:"max_value"`
	Index_Rating  float32   `json:"index_rating,omitempty"`
	Category      string    `json:"category,omitempty"`
	Dataset       []QosData `json:"dataset,omitempty"`
}
