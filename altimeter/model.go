package altimeter

// altimeter model
// 中身は相談しながら決める
type Altimeter struct {
	// Altitude in meters
	Altitude float64 `json:"altitude"`
	// Pressure in hPa
	Pressure float64 `json:"pressure"`
	// Temperature in Celsius
	Temperature float64 `json:"temperature"`
	// Humidity in percentage
	Humidity float64 `json:"humidity"`
	// Timestamp of the measurement
	Timestamp int64 `json:"timestamp"`
	// Device ID of the altimeter
	DeviceID string `json:"device_id"`
	// Location of the altimeter
}
