package models

// Config struct to hold the JSON configuration
type Config struct {
	GetDataApiHost   string `json:"getDataApiHost"`
	SentDataApiHost  string `json:"sentDataApiHost"`
	BatchSize        int    `json:"BatchSize"`
	StartMinute      int    `json:"startMinute"`
	SemaphoreForGet  int    `json:"semaphoreForGet"`
	SemaphoreForSave int    `json:"semaphoreForSave"`
}

type ConfigPoint struct {
	CommonSetting  CommonSetting    `json:"commonSetting"`
	ChannelSetting map[string]Point `json:"channelSetting"`
}

// CommonSetting holds the common configuration settings.
type CommonSetting struct {
	Company   string `json:"company"`
	Frequency int    `json:"frequency"`
	BindArea  string `json:"bindArea"`
}

// Point represents the configuration for a single channel.
type Point struct {
	Value      []string `json:"value"`
	Ieee754    bool     `json:"ieee754"`
	Reverse    bool     `json:"reverse"`
	FloatPoint int      `json:"floatPoint"`
	Type       string   `json:"Type"`
}
