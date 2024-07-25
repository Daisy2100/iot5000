package init

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"example.com/tool/models"
)

// ReadConfig reads the configuration from the config file.
func ReadConfig(filePath string) (*models.Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	var config models.Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

// readPonit reads the configuration from the config file.
func ReadPonit(filePath string) (*models.ConfigPoint, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	var config models.ConfigPoint
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

// makeAPIRequest makes an API request to the given URL and returns the response body as a string.
func MakeAPIRequest(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to make API request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read API response: %v", err)
	}

	return string(body), nil
}

// startTimer starts a timer for the given duration in minutes.
func StartTimer(minutes int) {
	duration := time.Duration(minutes) * time.Minute
	fmt.Printf("Starting timer for %d minutes...\n", minutes)
	timer := time.NewTimer(duration)
	<-timer.C
}
