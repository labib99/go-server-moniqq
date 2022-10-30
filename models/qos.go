package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type QosList struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ISP           string             `json:"isp,omitempty"`
	Service       string             `json:"service,omitempty"`
	Bandwidth     float64            `json:"bandwidth,omitempty"`
	Customer_Name string             `json:"customer_name,omitempty"`
	City          string             `json:"city,omitempty"`
	Upload_Date   time.Time          `json:"upload_date,omitempty"`
	Reports       []Report           `json:"reports,omitempty"`
}

type QosList2 struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ISP           string             `json:"isp"`
	Service       string             `json:"service"`
	Bandwidth     float64            `json:"bandwidth"`
	Customer_Name string             `json:"customer_name"`
	City          string             `json:"city"`
	Upload_Date   time.Time          `json:"upload_date"`
	Dataset_List  []ListDataset      `json:"dataset_list"`
	//Reports            []Report           `json:"reports,omitempty"`
}

type ListDataset struct {
	Qos_Parameter string             `json:"qos_parameter"`
	ID_Dataset    primitive.ObjectID `json:"id_dataset" bson:"id_dataset"` //id from collqdatasetperday
	Date          time.Time          `json:"date"`
	Num_Data      int                `json:"num_data"` //number of QosData
	Total_Value   float64            `json:"total_value"`
	Min_Value     float64            `json:"min_value"`
	Max_Value     float64            `json:"max_value"`
}

type QDatasetPerDay struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ID_Qos        primitive.ObjectID `json:"id_qos,omitempty" bson:"id_qos,omitempty"`
	Qos_Parameter string             `json:"qos_parameter,omitempty"`
	Date          time.Time          `json:"date,omitempty"`
	Dataset       []QosData          `json:"dataset,omitempty"`
}

type QosData struct {
	Date_Time time.Time `json:"date_time"`
	Value     float64   `json:"value"`
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
	ID                         primitive.ObjectID        `json:"_id,omitempty" bson:"_id,omitempty"`
	ISP                        string                    `json:"isp,omitempty"`
	Service                    string                    `json:"service,omitempty"`
	Bandwidth                  float64                   `json:"bandwidth,omitempty"`
	Customer_Name              string                    `json:"customer_name,omitempty"`
	City                       string                    `json:"city,omitempty"`
	Upload_Date                time.Time                 `json:"upload_date,omitempty"`
	Average_Index_Rating       float32                   `json:"average_index_rating"`
	Category                   string                    `json:"category"`
	Recap_QCustomer_Per_QParam []RecapQCustomerPerQParam `json:"recap_qcustomer_per_qparam,omitempty"`
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

type Report struct {
	Qos_Parameter string  `json:"qos_parameter,omitempty"`
	Average       float64 `json:"average,omitempty"`
	Max           float64 `json:"max,omitempty"`
	Min           float64 `json:"min,omitempty"`
	Index_Rating  int     `json:"index_rating,omitempty"`
	Category      string  `json:"category,omitempty"`
}

type QosRecord struct {
	Qos_Info QosList          `json:"qos_info"`
	Qos_Data []QDatasetPerDay `json:"qos_data"`
}

type Identity struct {
	//ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	//Session_Token string             `json:"session_token"`
	Role     string `json:"role"`
	Isp_Name string `json:"isp_name,omitempty"`
}

// type Isp struct {
// 	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
// 	Isp_Name string             `json:"isp_name"`
// 	Services []Service          `json:"services"`
// 	Cities   []string           `json:"cities"`
// }

// type Service struct {
// 	Service_Name string `json:"service_name"`
// 	Bandwidth    int    `json:"bandwidth"`
// }
