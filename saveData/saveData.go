package saveData

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"example.com/tool/models"
	"github.com/gammazero/workerpool"
)

// Custom HTTP client with increased timeout and connection pooling
// https://www.cnblogs.com/JulianHuang/p/15950624.html
var httpClient = &http.Client{
	Timeout: 5 * time.Second, // Increase timeout as needed
	Transport: &http.Transport{
		// 同一個主機 最大的連接數
		// MaxIdleConnsPerHost: 50,
		// MaxConnsPerHost:     1,
		// IdleConnTimeout:     10 * time.Second,
		DisableKeepAlives: false, // Enable keep-alive connections
	},
}

// SaveData sends the SentData to the database API with retry logic and increased timeout
func SaveData(data models.SentDataByBatched, dbAPIURL string) error {
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
	req.ContentLength = int64(len(payload)) // Ensure Content-Length is set correctly

	// Implementing retry logic
	const maxRetries = 2
	for i := 0; i < maxRetries; i++ {
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Printf("failed to send data to DB (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(2 * time.Second) // Exponential backoff can be implemented here
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil // Success
		}

		log.Printf("failed to save data to DB, status: %v (attempt %d/%d)", resp.Status, i+1, maxRetries)
		time.Sleep(10 * time.Millisecond) // Exponential backoff can be implemented here
	}

	return fmt.Errorf("failed to send data to DB after %d attempts: %w", maxRetries, err)
}

// AggregateAndSaveData continuously reads from the messageQueue and aggregates the data.
// Once the number of items reaches the batchSize, it sends the data to the database API.
func AggregateAndSaveData(ctx context.Context, messageQueue <-chan models.SentData, dbAPIURL string, batchSize int, wp *workerpool.WorkerPool,
	apiSaveCount *int32) {
	var batch models.SentDataByBatched

	for {
		select {
		case <-ctx.Done():
			// 時間結束時，送出最後一次請求
			if len(batch.Timestamps) > 0 {
				if err := SaveData(batch, dbAPIURL); err != nil {
					fmt.Printf("failed to save batch data: %v\n", err)
				}
			}
			return

		case data := <-messageQueue:
			wp.Submit(func() {
				batch.Timestamps = append(batch.Timestamps, data.Timestamps)
				batch.MeasurementsList = append(batch.MeasurementsList, data.MeasurementsList)
				batch.DataTypesList = append(batch.DataTypesList, data.DataTypesList)
				batch.ValuesList = append(batch.ValuesList, data.ValuesList)
				batch.IsAligned = data.IsAligned
				batch.Devices = append(batch.Devices, data.Devices)

				if len(batch.Timestamps) >= batchSize {
					if err := SaveData(batch, dbAPIURL); err != nil {
						fmt.Printf("failed to save batch data: %v\n", err)
					}
					batch = models.SentDataByBatched{} // Reset batch

					atomic.AddInt32(apiSaveCount, int32(batchSize)) // Increment the counter
				}
			})

			// Add a sleep interval to prevent resource exhaustion
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// AggregateAndSaveData continuously reads from the messageQueue and aggregates the data.
// Once the number of items reaches the batchSize, it sends the data to the database API.
func AggregateAndSaveDataByGoRoutine(ctx context.Context, messageQueue <-chan models.SentData, dbAPIURL string, batchSize int, apiSaveCount *int32) {
	var batch models.SentDataByBatched

	for {
		select {
		case <-ctx.Done():
			// 時間結束時，送出最後一次請求
			if len(batch.Timestamps) > 0 {
				if err := SaveData(batch, dbAPIURL); err != nil {
					fmt.Printf("failed to save batch data: %v\n", err)
				}
			}
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
				}
				batch = models.SentDataByBatched{} // Reset batch

				atomic.AddInt32(apiSaveCount, int32(batchSize)) // Increment the counter
			}
		}
	}
}

func AggregateAndSaveDataByGoRoutineNoCount(ctx context.Context, messageQueue <-chan models.SentData, dbAPIURL string, batchSize int) {
	var batch models.SentDataByBatched

	for {
		select {
		case <-ctx.Done():
			// 時間結束時，送出最後一次請求
			if len(batch.Timestamps) > 0 {
				if err := SaveData(batch, dbAPIURL); err != nil {
					fmt.Printf("failed to save batch data: %v\n", err)
				}
			}
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
				}
				batch = models.SentDataByBatched{} // Reset batch
			}
		}
	}
}

func AggregateAndSaveDataNoCount(ctx context.Context, messageQueue <-chan models.SentData, dbAPIURL string, batchSize int, wp *workerpool.WorkerPool) {
	var batch models.SentDataByBatched

	for {
		select {
		case <-ctx.Done():
			// 時間結束時，送出最後一次請求
			if len(batch.Timestamps) > 0 {
				if err := SaveData(batch, dbAPIURL); err != nil {
					fmt.Printf("failed to save batch data: %v\n", err)
				}
			}
			return

		case data := <-messageQueue:
			wp.Submit(func() {
				batch.Timestamps = append(batch.Timestamps, data.Timestamps)
				batch.MeasurementsList = append(batch.MeasurementsList, data.MeasurementsList)
				batch.DataTypesList = append(batch.DataTypesList, data.DataTypesList)
				batch.ValuesList = append(batch.ValuesList, data.ValuesList)
				batch.IsAligned = data.IsAligned
				batch.Devices = append(batch.Devices, data.Devices)

				if len(batch.Timestamps) >= batchSize {
					if err := SaveData(batch, dbAPIURL); err != nil {
						fmt.Printf("failed to save batch data: %v\n", err)
					}
					batch = models.SentDataByBatched{} // Reset batch
				}
			})

			// Add a sleep interval to prevent resource exhaustion
			time.Sleep(1 * time.Millisecond)
		}
	}
}
