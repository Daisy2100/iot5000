package models

// SentData represents the structure of data to be sent to the database API
type SentData struct {
	Timestamps       int64     `json:"timestamps"`
	MeasurementsList []string  `json:"measurements_list"`
	DataTypesList    []string  `json:"data_types_list"`
	ValuesList       []float64 `json:"values_list"`
	IsAligned        bool      `json:"is_aligned"`
	Devices          string    `json:"devices"`
}

type SentDataByBatched struct {
	Timestamps       []int64     `json:"timestamps"`
	MeasurementsList [][]string  `json:"measurements_list"`
	DataTypesList    [][]string  `json:"data_types_list"`
	ValuesList       [][]float64 `json:"values_list"`
	IsAligned        bool        `json:"is_aligned"`
	Devices          []string    `json:"devices"`
}
