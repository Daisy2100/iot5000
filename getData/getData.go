package getData

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	format "example.com/tool/format"
	"example.com/tool/models"
)

func extractEquipmentName(url string) (string, error) {
	re := regexp.MustCompile(`equipment(\d+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 2 {
		return "", fmt.Errorf("unable to extract equipment name from URL: %s", url)
	}
	return "equipment" + matches[1], nil
}

// fetchEquipmentData fetches data from a single endpoint.
func fetchEquipmentData(ctx context.Context, url string) (map[string]float64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %v", url, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK response from %s: %v", url, resp.Status)
	}

	var data map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response from %s: %v", url, err)
	}

	return data, nil
}

// GetData fetches data from a list of equipment endpoints and processes them.
func GetData(ctx context.Context, urls []string, points models.ConfigPoint) ([]models.SentData, []error) {
	var results []models.SentData
	var errors []error

	for _, url := range urls {
		equipmentName, err := extractEquipmentName(url)

		data, err := fetchEquipmentData(ctx, url)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		// Process the data according to points
		processedData := format.ProcessData(equipmentName, data, points)
		results = append(results, processedData)
	}

	return results, errors
}

// PrepareAndFetchData prepares URLs based on given parameters and fetches data using concurrent goroutines.
func PrepareAndFetchData(ctx context.Context, config models.Config, points models.ConfigPoint, startRange, endRange, portStart, portEnd int, messageQueue chan<- models.SentData) {
	// Prepare URLs
	urls := make([]string, 0)
	host := config.GetDataApiHost

	for portOffset := portStart; portOffset <= portEnd; portOffset++ {
		for i := startRange; i <= endRange; i++ {
			url := fmt.Sprintf("http://%s:%d/equipment%d", host, 3000+portOffset, i)
			urls = append(urls, url)
		}
	}

	// Fetch data with concurrency control
	func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				data, errs := GetData(ctx, urls, points)
				if len(errs) > 0 {
					log.Printf("Errors occurred while fetching data: %v", errs)
				}

				for _, item := range data {
					messageQueue <- item
				}
			}
		}
	}()
}
