package format

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"example.com/tool/models"
)

// hexToFloat32 converts a combined hexadecimal string to a float32 according to IEEE 754
func hexToFloat32(hexString string) float32 {
	intValue, err := strconv.ParseUint(hexString, 16, 32)
	if err != nil {
		fmt.Println("Error parsing hex string:", err)
		return 0
	}
	return math.Float32frombits(uint32(intValue))
}

// roundToDigits rounds a float64 number to the specified number of digits
func roundToDigits(num float64, digits int) float64 {
	pow := math.Pow(10, float64(digits))
	return math.Round(num*pow) / pow
}

// GetCurrentUnixTimestampInMilliseconds returns the current local time in milliseconds.
func getCurrentUnixTimestampInMilliseconds() int64 {
	now := time.Now()
	utcMilliseconds := now.UnixMilli()
	localMilliseconds := utcMilliseconds
	return localMilliseconds
}

// ProcessData processes the data according to settings
func ProcessData(equipmentName string, response map[string]float64, settings models.ConfigPoint) models.SentData {

	var sentData models.SentData

	timestamps := getCurrentUnixTimestampInMilliseconds()

	// 初始化 MeasurementsList 和 DataTypesList
	// measurementsList := []string{
	// 	"volt", "current", "kw", "kwh", "co2kg", "demand", "pf", "waterInTemp",
	// 	"waterOutTemp", "waterFlow", "waterFlowAcc", "waterOutPressure", "waterInPressure",
	// 	"airFlow", "airFlowAcc",
	// }
	// dataTypesList := []string{
	// 	"DOUBLE", "DOUBLE", "DOUBLE", "DOUBLE", "DOUBLE", "DOUBLE", "DOUBLE", "DOUBLE",
	// 	"DOUBLE", "DOUBLE", "DOUBLE", "DOUBLE", "DOUBLE", "DOUBLE", "DOUBLE",
	// }
	measurementsList := []string{}
	dataTypesList := []string{}
	valuesList := []float64{}

	// 讀取點位設定
	for key, setting := range settings.ChannelSetting {

		// fmt.Println("===============================================")
		// fmt.Println("key:", key)
		// fmt.Println("setting:", setting)

		measurementsList = append(measurementsList, key)
		dataTypesList = append(dataTypesList, "DOUBLE")

		addresses := setting.Value
		var values []uint16
		for _, addr := range addresses {
			val := uint16(response[addr])
			values = append(values, val)
		}

		ieee754 := setting.Ieee754
		reverse := setting.Reverse
		floatPoint := setting.FloatPoint

		if ieee754 {
			// IEEE754 float32 processing
			var hexString string

			if reverse {
				// Reverse byte order: Address3 is low part, Address2 is high part
				combined := (uint32(values[1]) << 16) | uint32(values[0])
				hexString = fmt.Sprintf("%08x", combined)
			} else {
				// Original byte order: Address2 is low part, Address3 is high part
				combined := (uint32(values[0]) << 16) | uint32(values[1])
				hexString = fmt.Sprintf("%08x", combined)
			}

			// Convert to float32
			result := hexToFloat32(hexString)
			resultRounded := roundToDigits(float64(result), floatPoint)
			valuesList = append(valuesList, resultRounded)

			// fmt.Printf("Combined Hex: %s\n", hexString)
			// fmt.Printf("Result: %.5f\n", result)
			// fmt.Printf("Rounded Result: %.2f\n", resultRounded)

		} else {
			// Non-IEEE754 processing (WORD)
			// Assuming WORD processing here is simply taking the first value as an example
			result := float32(values[0])
			resultRounded := roundToDigits(float64(result), floatPoint)
			valuesList = append(valuesList, resultRounded)
			// fmt.Printf("Result: %.2f\n", resultRounded)
		}
	}

	// 設置 SentData
	sentData = models.SentData{
		Timestamps:       timestamps,
		MeasurementsList: measurementsList,
		DataTypesList:    dataTypesList,
		ValuesList:       valuesList,
		IsAligned:        true,
		Devices:          fmt.Sprintf("root.systex.Rich19.7F.Daisy.%s", equipmentName),
	}

	// Print results
	// fmt.Println("Measurements List:", measurementsList)
	// fmt.Println("Data Types List:", dataTypesList)
	// fmt.Println("Values List:", valuesList)

	return sentData
}
