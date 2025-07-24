package tacho

import (
	"net/http"
	"time"
)

type TachoData struct {
	ID           uint8   `json:"id"`
	Timestamp    uint32  `json:"timestamp"`
	RPS          float64 `json:"rps"`
	Strain       uint32  `json:"strain"`
	ReceivedTime uint64  `json:"received_time"`
}

type TachoMeter struct {
	DataHistory  []TachoData  `json:"data_history"`
	Client       *http.Client `json:"client"`        // HTTP client for requests
	LogFrequency int          `json:"log_frequency"` // Frequency of logging data in a second
}

type TachoDLlink struct {
	DownloadLink string    `json:"download_link"`
	Timestamp    time.Time `json:"timestamp"`
}
