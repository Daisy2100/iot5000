package saveData

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"example.com/tool/models"
)

// SaveData sends the SentData to the database API
func SaveData(data models.SentDataByBatched, dbAPIURL string) error {

	// fmt.Printf("SaveData data: %v\n", data)

	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	req, err := http.NewRequest("POST", dbAPIURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic cm9vdDpyb290")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send data to DB: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to save data to DB, status: %v", resp.Status)
	}

	return nil
}

// AggregateAndSaveData continuously reads from the messageQueue and aggregates the data.
// Once the number of items reaches the batchSize, it sends the data to the database API.
func AggregateAndSaveData(ctx context.Context, messageQueue <-chan models.SentData, dbAPIURL string, batchSize int) {
	var batch models.SentDataByBatched

	for {
		select {
		case <-ctx.Done():
			return
		case data := <-messageQueue:
			batch.Timestamps = append(batch.Timestamps, data.Timestamps)
			batch.MeasurementsList = append(batch.MeasurementsList, data.MeasurementsList)
			batch.DataTypesList = append(batch.DataTypesList, data.DataTypesList)
			batch.ValuesList = append(batch.ValuesList, data.ValuesList)
			batch.IsAligned = data.IsAligned
			batch.Devices = append(batch.Devices, data.Devices)

			if len(batch.Timestamps) >= batchSize {
				if err := SaveData(batch, dbAPIURL); err != nil {
					fmt.Printf("failed to save batch data: %v\n", err)
				} else {
					// fmt.Println("Batch data saved successfully")
				}
				batch = models.SentDataByBatched{} // Reset batch
			}
		}
	}
}
