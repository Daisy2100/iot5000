package models

type EquipmentConfig struct {
	BaseURL      string // Base URL for the endpoints
	StartIndex   int    // Starting index for the equipment
	EndIndex     int    // Ending index for the equipment
	TotalBatches int    // Number of batches to process
}

// EquipmentData represents the structure of data returned by the equipment API
type AddressData struct {
	Address0  int `json:"Address0"`
	Address1  int `json:"Address1"`
	Address2  int `json:"Address2"`
	Address3  int `json:"Address3"`
	Address4  int `json:"Address4"`
	Address5  int `json:"Address5"`
	Address6  int `json:"Address6"`
	Address7  int `json:"Address7"`
	Address8  int `json:"Address8"`
	Address9  int `json:"Address9"`
	Address10 int `json:"Address10"`
	Address11 int `json:"Address11"`
	Address12 int `json:"Address12"`
	Address13 int `json:"Address13"`
	Address20 int `json:"Address20"`
	Address21 int `json:"Address21"`
	Address22 int `json:"Address22"`
	Address23 int `json:"Address23"`
	Address24 int `json:"Address24"`
	Address25 int `json:"Address25"`
	Address26 int `json:"Address26"`
	Address27 int `json:"Address27"`
	Address28 int `json:"Address28"`
	Address29 int `json:"Address29"`
	Address30 int `json:"Address30"`
	Address31 int `json:"Address31"`
	Address32 int `json:"Address32"`
	Address33 int `json:"Address33"`
}
