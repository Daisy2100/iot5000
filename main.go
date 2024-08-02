package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	initSetting "example.com/tool/init"
	"example.com/tool/models"
)

type AddressFormatData struct {
	Timestamps       int64
	MeasurementsList []string
	DataTypesList    []string
	ValuesList       []float64
	Devices          string
}

const (
	Duration    = 600 * time.Second // 運行時間
	BatchSize   = 100
	WorkerCount = 12 // 並發請求的數量
)

var urls = []string{
	// http://10.41.1.58:3001 ~ 3005/equipment4999
	"http://10.41.1.58:3001/equipment",
	"http://10.41.1.58:3002/equipment",
	"http://10.41.1.58:3003/equipment",
	"http://10.41.1.58:3004/equipment",
	"http://10.41.1.58:3005/equipment",
}

func main() {

	// 1-1. Read the config
	config, err := initSetting.ReadConfig("./config.json")
	if err != nil {
		log.Fatalf(err.Error())
	}

	// 1-2. Read the points
	points, err := initSetting.ReadPonit("./points.json")
	if err != nil {
		log.Fatalf(err.Error())
	}

	responseChan := make(chan AddressFormatData, BatchSize)
	var wg sync.WaitGroup

	// 3. Make the initial API request
	initialAPIURL := fmt.Sprintf("http://%s:3001/setInit/Daisy", config.GetDataApiHost)
	initialResponse, err := initSetting.MakeAPIRequest(initialAPIURL)
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Printf("Initial API response: %s\n", initialResponse)

	// 使用 context 來控制停止信號
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.StartMinute)*time.Minute)
	defer cancel()

	// 這個 goroutine 是用於處理 read chan 資料
	for i := 0; i < config.SemaphoreForSave; i++ {
		go processChannel(responseChan, ctx.Done())
	}

	for i := 0; i < config.SemaphoreForGet; i++ {
		for _, url := range urls {
			wg.Add(1)
			go func(url string) {
				equipmentCount := 0
				defer wg.Done()
				for {
					equipmentCount++
					if equipmentCount > 2 {
						equipmentCount = 1
					}

					resultUrl := fmt.Sprintf("%s%d", url, equipmentCount)

					select {
					case <-ctx.Done():
						return
					default:
						saveData := sendHTTPRequest(resultUrl, points)
						if saveData != nil {
							responseChan <- *saveData // 寫入 chan
						}
					}
				}
			}(url)
		}
	}

	// 等待所有 workers 執行完畢
	wg.Wait()

	// close() 意思是禁止 chan 後續寫入但是可以讀取剩餘資料
	close(responseChan)

	fmt.Println("Final process remaining items in the channel...")
	// 關閉前最後剩下的一波資料(沒來得及搞完)
	finalBatch := make([]AddressFormatData, 0)
	for apiResp := range responseChan {
		finalBatch = append(finalBatch, apiResp)
	}
	flushBatchData(&finalBatch)
	fmt.Println("All done!")

	// 8. Make the final API request before stopping
	finalAPIURL := fmt.Sprintf("http://%s:3001/setFinal/Daisy", config.GetDataApiHost)
	finalResponse, err := initSetting.MakeAPIRequest(finalAPIURL)
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Printf("Final API response: %s\n", finalResponse)
	fmt.Println("Time's up!")
}

func sendHTTPRequest(url string, points *models.ConfigPoint) *AddressFormatData {
	// 設置請求超時
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return nil
	}

	var apiResp map[string]float64
	if err := json.Unmarshal(body, &apiResp); err != nil {
		fmt.Println("Error parsing response:", err)
		return nil
	}

	equipmentName, err := extractEquipmentName(url)
	if err != nil {
		fmt.Println("Error parsing response:", err)
		return nil
	}

	// TODO 整理 AddressData to SaveData 資料
	saveData := addressDataToSaveData(apiResp, equipmentName, *points)
	return &saveData
}

func processChannel(responseChan <-chan AddressFormatData, stop <-chan struct{}) {
	batch := make([]AddressFormatData, 0, BatchSize)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case apiResp, ok := <-responseChan:
			if !ok {
				// Channel 以關閉, 最後一次處理 batch 剩餘資料
				flushBatchData(&batch)
				return
			}
			// 將資料放入 batch 中
			batch = append(batch, apiResp)
			if len(batch) == BatchSize { // 達到 batch 最大水位資料時開閘放水
				flushBatchData(&batch)
				batch = batch[:0] // 釋放記憶體以防 OOM
			}
		case <-ticker.C: // ticker 觸發，做滿水位同樣行為
			flushBatchData(&batch)
			batch = batch[:0]
		case <-stop: // 停止運行，做滿水位同樣行為
			flushBatchData(&batch)
			return
		}
	}
}

func flushBatchData(batchResponseData *[]AddressFormatData) {
	if len(*batchResponseData) > 0 {
		timestampInMilliseconds := getCurrentUnixTimestampInMilliseconds()

		// 構造要發送的 JSON
		var timestamps []int64
		var measurementsList [][]string
		var dataTypesList [][]string
		var valuesList [][]float64
		var devices []string

		for _, data := range *batchResponseData {
			timestamps = append(timestamps, timestampInMilliseconds)
			measurementsList = append(measurementsList, data.MeasurementsList)
			dataTypesList = append(dataTypesList, data.DataTypesList)
			valuesList = append(valuesList, data.ValuesList)
			devices = append(devices, data.Devices)
		}

		// 構造請求體
		requestBody := map[string]interface{}{
			"timestamps":        timestamps,
			"measurements_list": measurementsList,
			"data_types_list":   dataTypesList,
			"values_list":       valuesList,
			"is_aligned":        false, // TODO ?
			"devices":           devices,
		}

		// 將 JSON 數據轉換為字串
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			fmt.Println("Error marshalling JSON:", err)
			return
		}

		// 發送 HTTP POST 請求
		url := "http://10.41.1.58:18080/rest/v2/insertRecords"
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}
		req.Header.Set("Authorization", "Basic cm9vdDpyb290")
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return
		}
		defer resp.Body.Close()

		// fmt.Println("Response status:", resp.Status)
	}
}

func addressDataToSaveData(data map[string]float64, equipmentName string, point models.ConfigPoint) AddressFormatData {
	//TODO
	var formatData AddressFormatData

	timestamps := getCurrentUnixTimestampInMilliseconds()

	measurementsList := []string{}
	dataTypesList := []string{}
	valuesList := []float64{}

	// 讀取點位設定
	for key, setting := range point.ChannelSetting {

		measurementsList = append(measurementsList, key)
		dataTypesList = append(dataTypesList, "DOUBLE")

		addresses := setting.Value
		var values []uint16
		for _, addr := range addresses {
			val := uint16(data[addr])
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

	// 設置 formatData
	formatData = AddressFormatData{
		Timestamps:       timestamps,
		MeasurementsList: measurementsList,
		DataTypesList:    dataTypesList,
		ValuesList:       valuesList,
		Devices:          fmt.Sprintf("root.systex.Rich19.7F.Daisy.%s", equipmentName),
	}

	// Print results
	// fmt.Println("Measurements List:", measurementsList)
	// fmt.Println("Data Types List:", dataTypesList)
	// fmt.Println("Values List:", valuesList)

	return formatData
}

// --------------------------

func getCurrentUnixTimestampInMilliseconds() int64 {
	now := time.Now()
	utcMilliseconds := now.UnixMilli()
	localMilliseconds := utcMilliseconds
	return localMilliseconds
}
func extractEquipmentName(url string) (string, error) {
	re := regexp.MustCompile(`equipment(\d+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 2 {
		return "", fmt.Errorf("unable to extract equipment name from URL: %s", url)
	}
	return "equipment" + matches[1], nil
}

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
