package altimeter

import "time"

// altimeter model
// 中身は相談しながら決める
type AltimeterData struct {
	// Altitude in meters
	Altitude float64 `json:"altitude"`
	// Pressure in hPa
	Pressure float64 `json:"pressure"`
	// Temperature in Celsius
	Temperature float64 `json:"temperature"`
	// Humidity in percentage
	Humidity float64 `json:"humidity"`
	// Timestamp of the measurement
	Timestamp time.Time `json:"timestamp"`
	// Device ID of the altimeter
	DeviceID string `json:"device_id"`
	// Location of the altimeter
}

// Altimeterのクラス
type Altimeter struct {
	// データの履歴配列
	DataHistory []AltimeterData `json:"data_history"`
}

