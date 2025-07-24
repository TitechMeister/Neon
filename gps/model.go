package gps

import (
	"net/http"
	"time"
)

type GPSData struct {
	ID           uint8  `json:"id"`
	FixMode      uint8  `json:"fixmode"`
	PDOP         uint16 `json:"PDOP"`
	Year         uint16 `json:"year"`
	ITow         uint32 `json:"iTow"`
	Unixtime     uint32 `json:"unixtime"`
	Lon          uint32 `json:"lon"`
	Lat          uint32 `json:"lat"`
	Height       uint32 `json:"height"`
	HAcc         uint32 `json:"hAcc"`
	VAcc         uint32 `json:"vAcc"`
	GSpeed       uint32 `json:"gSpeed"`
	HeadMot      uint32 `json:"headMot"`
	ReceivedTime uint64 `json:"received_time"`
}

type GPS struct {
	DataHistory []GPSData    `json:"data_history"`
	Client      *http.Client `json:"client"` // HTTP client for requests
}

type GPSDLlink struct {
	DownloadLink string    `json:"download_link"`
	Timestamp    time.Time `json:"timestamp"`
}

type GPSUIData struct {
	Unixtime     uint32 `json:"unixtime"`
	Lon          uint32 `json:"lon"`
	Lat          uint32 `json:"lat"`
	ReceivedTime uint64 `json:"received_time"`
}

type TargetData struct {
	ID        uint8    `json:"id"`
	Timestamp uint32   `json:"timestamp"`
	TargetLon int32    `json:"target_lon"`
	TargetLat int32    `json:"target_lat"`
	Data      [32]byte `json:"data"`
}

type TargetPayload struct {
	Payload [48]byte `json:"payload"`
}
