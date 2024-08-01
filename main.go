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
	workerpool "github.com/gammazero/workerpool"
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

	// 4. Create worker pools
	wpGet1 := workerpool.New(config.SemaphoreForGet)
	wpGet2 := workerpool.New(config.SemaphoreForGet)
	wpGet3 := workerpool.New(config.SemaphoreForGet)
	wpGet4 := workerpool.New(config.SemaphoreForGet)
	wpGet5 := workerpool.New(config.SemaphoreForGet)
	// wpSave := workerpool.New(config.SemaphoreForSave)

	// 5. Create queue
	// var apiRequestCount int32
	// var apiSaveCount int32
	messageQueue := make(chan models.SentData, config.MaxQueue)

	// 6. Prepare and fetch data
	// go getData.PrepareAndFetchData(ctx, *config, *points, 1, 1000, 1, 1, messageQueue, wpGet1, &apiRequestCount)
	// go getData.PrepareAndFetchData(ctx, *config, *points, 1001, 2000, 2, 2, messageQueue, wpGet2, &apiRequestCount)
	// go getData.PrepareAndFetchData(ctx, *config, *points, 2001, 3000, 3, 3, messageQueue, wpGet3, &apiRequestCount)
	// go getData.PrepareAndFetchData(ctx, *config, *points, 3001, 4000, 4, 4, messageQueue, wpGet4, &apiRequestCount)
	// go getData.PrepareAndFetchData(ctx, *config, *points, 4001, 5000, 5, 5, messageQueue, wpGet5, &apiRequestCount)
	go getData.PrepareAndFetchDataNoCount(ctx, *config, *points, 1, 1000, 1, 1, messageQueue, wpGet1)
	go getData.PrepareAndFetchDataNoCount(ctx, *config, *points, 1001, 2000, 2, 2, messageQueue, wpGet2)
	go getData.PrepareAndFetchDataNoCount(ctx, *config, *points, 2001, 3000, 3, 3, messageQueue, wpGet3)
	go getData.PrepareAndFetchDataNoCount(ctx, *config, *points, 3001, 4000, 4, 4, messageQueue, wpGet4)
	go getData.PrepareAndFetchDataNoCount(ctx, *config, *points, 4001, 5000, 5, 5, messageQueue, wpGet5)

	// 7. Submit task to worker pool for saving data
	// go saveData.AggregateAndSaveData(ctx, messageQueue, fmt.Sprintf("http://%s:18080/rest/v2/insertRecords", config.SentDataApiHost), config.BatchSize, wpSave, &apiSaveCount)
	// go saveData.AggregateAndSaveData(ctx, messageQueue, fmt.Sprintf("http://%s:18080/rest/v2/insertRecords", config.SentDataApiHost), config.BatchSize, wpSave, &apiSaveCount)
	// go saveData.AggregateAndSaveDataByGoRoutine(ctx, messageQueue, fmt.Sprintf("http://%s:18080/rest/v2/insertRecords", config.SentDataApiHost), config.BatchSize, &apiSaveCount)
	// go saveData.AggregateAndSaveDataByGoRoutine(ctx, messageQueue, fmt.Sprintf("http://%s:18080/rest/v2/insertRecords", config.SentDataApiHost), config.BatchSize, &apiSaveCount)
	go saveData.AggregateAndSaveDataByGoRoutineNoCount(ctx, messageQueue, fmt.Sprintf("http://%s:18080/rest/v2/insertRecords", config.SentDataApiHost), config.BatchSize)
	go saveData.AggregateAndSaveDataByGoRoutineNoCount(ctx, messageQueue, fmt.Sprintf("http://%s:18080/rest/v2/insertRecords", config.SentDataApiHost), config.BatchSize)

	// Wait for the context to be done
	<-ctx.Done()

	// 8. Make the final API request before stopping
	finalAPIURL := fmt.Sprintf("http://%s:3001/setFinal/Daisy", config.GetDataApiHost)
	finalResponse, err := initSetting.MakeAPIRequest(finalAPIURL)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Wait for all tasks to complete
	wpGet1.StopWait()
	wpGet2.StopWait()
	wpGet3.StopWait()
	wpGet4.StopWait()
	wpGet5.StopWait()
	// wpSave.StopWait()

	// Close the messageQueue after all tasks are done
	close(messageQueue)

	// totalSeconds := config.StartMinute * 60
	// averageRequestsPerSecond := float64(apiRequestCount) / float64(totalSeconds)
	// averageSavePerSecond := float64(apiSaveCount) / float64(totalSeconds)

	fmt.Printf("Final API response: %s\n", finalResponse)
	fmt.Println("Time's up!")
	// fmt.Printf("Average API requests per second: %.2f\n", averageRequestsPerSecond)
	// fmt.Printf("Average API save per second: %.2f\n", averageSavePerSecond)
}
