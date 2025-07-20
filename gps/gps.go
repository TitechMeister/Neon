package gps

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/TitechMeister/Neon/cloudstorage"
	"github.com/labstack/echo"
)

// 新しいGPSの構造体を返す
func New() *GPS {
	return &GPS{
		DataHistory: []GPSData{},
		Client:      &http.Client{}, // HTTPクライアントを初期化
	}
}

func (g *GPS) GetSencorName() string {
	// センサーの名前を返す
	return "gps"
}

// GPSエンドポイント→現在はモックデータとしてGPSData構造体のJSONを返す
func (handler *GPS) GetData(c echo.Context) error {
	// モックデータを返す
	data := GPSData{}
	if os.Getenv("MODE") == "mock" {
		now := time.Now()
		// 琵琶湖上の竹生島付近の座標（緯度: 35.2786, 経度: 136.0952）
		data = GPSData{
			ID:           1,
			FixMode:      3,                            // 3D fix
			PDOP:         uint16(100 + rand.Intn(200)), // Random PDOP between 100-300
			Year:         uint16(now.Year()),
			ITow:         uint32(now.Unix()),
			Unixtime:     uint32(now.Unix()),
			Lon:          uint32(1360952000 + rand.Intn(10000)), // 136.0952 (竹生島付近) + ランダム
			Lat:          uint32(352786000 + rand.Intn(10000)),  // 35.2786 (竹生島付近) + ランダム
			Height:       uint32(50000 + rand.Intn(100000)),     // Random height 50-150m (in mm)
			HAcc:         uint32(1000 + rand.Intn(5000)),        // Horizontal accuracy 1-6m (in mm)
			VAcc:         uint32(2000 + rand.Intn(8000)),        // Vertical accuracy 2-10m (in mm)
			GSpeed:       uint32(rand.Intn(50000)),              // Random ground speed 0-50 m/s (in mm/s)
			HeadMot:      uint32(rand.Intn(360000000)),          // Random heading 0-360 degrees (in 1e-5 degrees)
			ReceivedTime: uint64(now.UnixMilli()),
		}
	} else {
		// 実際のデータを取得する
		fmt.Println("Fetching GPS data from the server...")
		req, err := http.NewRequest("GET", "http://localhost:7878/data/gps", nil)
		// http.NewRequestを使ってGETリクエストを作成
		fmt.Println("Request created:", req)
		if err != nil {
			return c.String(500, fmt.Sprintf("Error creating request: %v", err))
		}
		res, err := handler.Client.Do(req)
		if err != nil {
			return c.String(500, fmt.Sprintf("Error asking request: %v", err))
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			return c.String(res.StatusCode, fmt.Sprintf("Error fetching GPS data: %s", res.Status))
		}
		// レスポンスボディをデコード
		if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
			return c.String(500, fmt.Sprintf("Error decoding GPS data: %v", err))
		}
	}
	// データを履歴に追加
	handler.addData(data)
	// JSON形式でデータを返す
	return c.JSON(200, data)
}

func (handler *GPS) LogData(c echo.Context) error {
	// 現在までのデータをログに追記
	res := &GPSDLlink{}
	err := handler.makeLogJson(handler.DataHistory)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error writing GPS data log: %v", err))
	}
	// ログファイルのリネーム
	newName := fmt.Sprintf("gps_log_%s.json", time.Now().Format("20060102_150405"))
	err = os.Rename("temp_gps_log.json", newName)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error renaming GPS log file: %v", err))
	}
	// ログファイルのリネームが成功したら履歴をクリア
	handler.DataHistory = []GPSData{}
	url, err := cloudstorage.UploadFile(c.Response().Writer, "25_logs", newName)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error uploading GPS log file: %v", err))
	}
	// データのDLリンクを返す
	res.DownloadLink = *url
	res.Timestamp = time.Now()
	return c.JSON(200, res)
}

// 現在のデータ履歴を取得する
func (handler *GPS) GetHistory(c echo.Context) error {
	// 履歴データを返す
	return c.JSON(200, handler.DataHistory)
}

func (handler *GPS) addData(data GPSData) {
	// データを履歴に追加
	handler.DataHistory = append(handler.DataHistory, data)
	// 履歴が20件を超えたらjsonに書き込んで古い10件のデータを削除
	if len(handler.DataHistory) > 20 {
		handler.makeLogJson(handler.DataHistory[:10])
		handler.DataHistory = handler.DataHistory[len(handler.DataHistory)-10:]
	}
}

func (handler *GPS) makeLogJson(data []GPSData) error {
	// JSONファイルに書き込む
	file, err := os.OpenFile("temp_gps_log.json", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open temp_gps_log.json: %w", err)
	}
	defer file.Close()
	fi, _ := file.Stat()
	leng := fi.Size()

	json_, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal GPS data: %w", err)
	}

	if leng == 0 {
		_, err = file.Write(fmt.Appendf(nil, `%s`, json_))
		if err != nil {
			return fmt.Errorf("failed to write to temp_gps_log.json: %w", err)
		}
	} else {
		// 頭の1文字[は削る
		json_ = json_[1:]
		_, err = file.WriteAt(fmt.Appendf(nil, `,%s`, json_), leng-1)
		if err != nil {
			return fmt.Errorf("failed to write to temp_gps_log.json: %w", err)
		}
	}
	return nil
}
