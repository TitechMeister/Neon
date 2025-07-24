package pitot

import (
	"net/http"
	"time"
)

type PitotData struct {
	ID           uint8   `json:"id"`             // デバイス識別子
	Timestamp    uint32  `json:"timestamp"`      // 時刻
	Temperature  float32 `json:"temperature"`    // 温度
	Velocity     float32 `json:"velocity"`       // 対気速度
	PressureVRaw float32 `json:"pressure_v_raw"` // 対気圧力生データ
	PressureARaw float32 `json:"pressure_a_raw"` // 迎角圧力生データ
	PressureSRaw float32 `json:"pressure_s_raw"` // 横滑り圧力生データ
}

type Pitot struct {
	DataHistory  []PitotData  `json:"data_history"`  // データ履歴
	Client       *http.Client `json:"client"`        // HTTPクライアント
	LogFrequency int          `json:"log_frequency"` // Frequency of logging data in a second
}

type PitotDLlink struct {
	DownloadLink string    `json:"download_link"` // ダウンロードリンク
	Timestamp    time.Time `json:"timestamp"`     // リンクの生成時刻
}
