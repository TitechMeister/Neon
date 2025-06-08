package altimeter

import (
	"math/rand"
	"time"

	"github.com/labstack/echo"
)

// Altimeterエンドポイント→現在はモックデータとしてAltimeter構造体のJSONを返す
func GetAltimeterData(c echo.Context) error {
	// モックデータを返す
	data := Altimeter{
		Altitude: 100.0 + rand.Float64()*50.0, // Random altitude between 100 and 150 meters
		Pressure:    1013.25,
		Temperature: 20.0,
		Humidity:    50.0,
		Timestamp: time.Now(), // Example timestamp
		DeviceID:    "device123",
	}

	return c.JSON(200, data)
}
