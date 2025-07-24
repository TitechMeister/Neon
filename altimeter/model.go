package altimeter

import (
	"net/http"
	"time"
)

// altimeter model
// 中身は相談しながら決める
type AltimeterUIData struct {
	// Device ID of the altimeter
	DeviceID uint8 `json:"id"`
	// Altitude in meters
	Altitude float64 `json:"altitude"`
	// // Temperature in Celsius
	// Temperature float64 `json:"temperature"`
	// // Timestamp of the measurement
	// Timestamp int32 `json:"timestamp"`
	// Device ID of the altimeter
	ReceivedTime time.Time `json:"received_time"`
}

type AltimeterRawData struct {
	// Device ID of the altimeter
	DeviceID uint8 `json:"id"`
	// Altitude in meters
	Altitude float64 `json:"altitude"`
	// Temperature in Celsius
	Temperature float64 `json:"temperature"`
	// Timestamp of the measurement
	Timestamp int32 `json:"timestamp"`
	// Device ID of the altimeter
	ReceivedTime int64 `json:"received_time"`
}

// Altimeterのクラス
type Altimeter struct {
	// データの履歴配列
	DataHistory []AltimeterRawData `json:"data_history"`
	// httpクライアント
	Client *http.Client `json:"client"` // Uncomment if you need an HTTP client for requests
	// ログ更新周波数
	LogFrequency int `json:"log_frequency"` // Frequency of logging data in a second
}

type AltimeterDLlink struct {
	// Download link for the altimeter data
	DownloadLink string `json:"download_link"`
	// Timestamp of the download link creation
	Timestamp time.Time `json:"timestamp"`
}
