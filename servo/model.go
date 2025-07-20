package servo

import (
	"net/http"
	"time"
)

type ServoData struct {
	ID                  uint8   `json:"id"`
	Status              uint8   `json:"status"`
	Timestamp           uint32  `json:"timestamp"`
	Rudder              float64 `json:"rudder"`
	Elevator            float64 `json:"elevator"`
	Voltage             float64 `json:"voltage"`
	RudderCurrent       float64 `json:"rudder_current"`
	ElevatorCurrent     float64 `json:"elevator_current"`
	Trim                float64 `json:"trim"`
	RudderServoAngle    float64 `json:"rudder_servo_angle"`
	ElevatorServoAngle  float64 `json:"elevator_servo_angle"`
	RudderTemperature   float64 `json:"rudder_temperature"`
	ElevatorTemperature float64 `json:"elevator_temperature"`
	ReceivedTime        uint64  `json:"received_time"`
}

type Servo struct {
	DataHistory []ServoData  `json:"data_history"`
	Client      *http.Client `json:"-"`
}

type ServoDLlink struct {
	DownloadLink string    `json:"download_link"`
	Timestamp    time.Time `json:"timestamp"`
}
