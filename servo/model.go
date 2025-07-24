package servo

import (
	"net/http"
	"time"
)

// hoge無印→入力された値
// Current→流れる電流
// hogeServoAngle→実際のサーボの角度→頑張って対応させるしか...
// 矢印→
// 温度表示...
// rad -15~15 trim + elv 3.6~-14 trim -5~+2
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
	DataHistory      []ServoData  `json:"data_history"`
	Client           *http.Client `json:"-"`
	RevElevatorValue []float64
	RevRudderValue   []float64
}

type ServoDLlink struct {
	DownloadLink string    `json:"download_link"`
	Timestamp    time.Time `json:"timestamp"`
}

type ServoUIData struct {
	Rudder              float64 `json:"rudder"`
	Elevator            float64 `json:"elevator"`
	Trim                float64 `json:"trim"`
	RudderActualAngle    float64 `json:"rudder_actual_angle"`
	ElevatorActualAngle  float64 `json:"elevator_actual_angle"`
	RudderTemperature   float64 `json:"rudder_temperature"`
	ElevatorTemperature float64 `json:"elevator_temperature"`
	ReceivedTime        uint64  `json:"received_time"`
	Timestamp           uint32  `json:"timestamp"`
}
