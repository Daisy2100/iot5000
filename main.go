package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"example.com/tool/getData"
	initSetting "example.com/tool/init"
	"example.com/tool/models"
	"example.com/tool/saveData"
)

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
	// fmt.Printf("Initial points.json: \n", points.ChannelSetting)
	// ===============================================================================================

	// 2. Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.StartMinute)*time.Minute)
	defer cancel()

	// 3. Make the initial API request
	initialAPIURL := fmt.Sprintf("http://%s:3001/setInit/Daisy", config.GetDataApiHost)
	initialResponse, err := initSetting.MakeAPIRequest(initialAPIURL)
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Printf("Initial API response: %s\n", initialResponse)
	// ===============================================================================================

	// 4. Define the semaphore with a limit of 200 concurrent requests
	// semaphoreForGetApi := make(chan struct{}, 198)
	// semaphoreForSaveApi := make(chan struct{}, 2)
	// ===============================================================================================

	// 5. Create queue
	messageQueue := make(chan models.SentData, config.MaxQueue)

	// 6. Prepare URLs for fetching data from multiple ranges
	go getData.PrepareAndFetchData(ctx, *config, *points, 1, 1000, 1, 1, messageQueue)    // Range 1-1000 on port 1
	go getData.PrepareAndFetchData(ctx, *config, *points, 1001, 2000, 2, 2, messageQueue) // Range 1001-2000 on port 2
	go getData.PrepareAndFetchData(ctx, *config, *points, 2001, 3000, 3, 3, messageQueue) // Range 2001-3000 on port 3
	go getData.PrepareAndFetchData(ctx, *config, *points, 3001, 4000, 4, 4, messageQueue) // Range 3001-4000 on port 4
	go getData.PrepareAndFetchData(ctx, *config, *points, 4001, 5000, 5, 5, messageQueue) // Range 4001-5000 on port 5

	// 8. Save data with save API
	go saveData.AggregateAndSaveData(ctx, messageQueue, fmt.Sprintf("http://%s:18080/rest/v2/insertRecords", config.SentDataApiHost), config.BatchSize)
	go saveData.AggregateAndSaveData(ctx, messageQueue, fmt.Sprintf("http://%s:18080/rest/v2/insertRecords", config.SentDataApiHost), config.BatchSize)

	// Wait for the context to expire
	<-ctx.Done()
	// ===============================================================================================

	// 8. Make the final API request before stopping
	finalAPIURL := fmt.Sprintf("http://%s:3001/setFinal/Daisy", config.GetDataApiHost)
	finalResponse, err := initSetting.MakeAPIRequest(finalAPIURL)
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Printf("Final API response: %s\n", finalResponse)
	fmt.Println("Time's up!")
}
