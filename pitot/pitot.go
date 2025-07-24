package pitot

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

// 新しいPitotの構造体を返す
func New() *Pitot {
	return &Pitot{
		DataHistory: []PitotData{},
		Client:      &http.Client{}, // HTTPクライアントを初期化
	}
}

func (p *Pitot) GetSencorName() string {
	// センサーの名前を返す
	return "pitot"
}

// Pitotエンドポイント→現在はモックデータとしてPitotData構造体のJSONを返す
func (handler *Pitot) GetData(c echo.Context) error {
	// モックデータを返す
	data := PitotData{}
	if os.Getenv("MODE") == "mock" {
		data = PitotData{
			ID:           1,
			Timestamp:    uint32(time.Now().Unix()),
			Temperature:  float32(15.0 + rand.Float64()*10.0),    // Random temperature between 15-25°C
			Velocity:     float32(5.0 + rand.Float64()*5.0),   // Random velocity between 50-150 m/s
			PressureVRaw: float32(1000.0 + rand.Float64()*200.0), // Random pressure 1000-1200
			PressureARaw: float32(800.0 + rand.Float64()*300.0),  // Random pressure 800-1100
			PressureSRaw: float32(900.0 + rand.Float64()*250.0),  // Random pressure 900-1150
		}
	} else {
		// 実際のデータを取得する
		fmt.Println("Fetching pitot data from the server...")
		req, err := http.NewRequest("GET", "http://localhost:7878/data/pitot", nil)
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
			return c.String(res.StatusCode, fmt.Sprintf("Error fetching pitot data: %s", res.Status))
		}
		// レスポンスボディをデコード
		if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
			return c.String(500, fmt.Sprintf("Error decoding pitot data: %v", err))
		}
	}
	// データを履歴に追加
	handler.addData(data)
	// JSON形式でデータを返す
	return c.JSON(200, data)
}

func (handler *Pitot) LogData(c echo.Context) error {
	// 現在までのデータをログに追記
	res := &PitotDLlink{}
	err := handler.makeLogJson(handler.DataHistory)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error writing pitot data log: %v", err))
	}
	// ログファイルのリネーム
	newName := fmt.Sprintf("logs/pitot_log_%s.json", time.Now().Format("20060102_150405"))
	err = os.Rename("temp_pitot_log.json", newName)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error renaming pitot log file: %v", err))
	}
	// ログファイルのリネームが成功したら履歴をクリア
	handler.DataHistory = []PitotData{}
	url, err := cloudstorage.UploadFile(c.Response().Writer, "25_logs", newName)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error uploading pitot log file: %v", err))
	}
	// データのDLリンクを返す
	res.DownloadLink = *url
	res.Timestamp = time.Now()
	return c.JSON(200, res)
}

// 現在のデータ履歴を取得する
func (handler *Pitot) GetHistory(c echo.Context) error {
	// 履歴データを返す
	return c.JSON(200, handler.DataHistory)
}

func (handler *Pitot) addData(data PitotData) {
	// データを履歴に追加
	handler.DataHistory = append(handler.DataHistory, data)
	// 履歴が20件を超えたらjsonに書き込んで古い10件のデータを削除
	if len(handler.DataHistory) > 20 {
		handler.makeLogJson(handler.DataHistory[:10])
		handler.DataHistory = handler.DataHistory[len(handler.DataHistory)-10:]
	}
}

func (handler *Pitot) makeLogJson(data []PitotData) error {
	// JSONファイルに書き込む
	file, err := os.OpenFile("temp_pitot_log.json", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open temp_pitot_log.json: %w", err)
	}
	defer file.Close()
	fi, _ := file.Stat()
	leng := fi.Size()

	json_, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal pitot data: %w", err)
	}

	if leng == 0 {
		_, err = file.Write(fmt.Appendf(nil, `%s`, json_))
		if err != nil {
			return fmt.Errorf("failed to write to temp_pitot_log.json: %w", err)
		}
	} else {
		// 頭の1文字[は削る
		json_ = json_[1:]
		_, err = file.WriteAt(fmt.Appendf(nil, `,%s`, json_), leng-1)
		if err != nil {
			return fmt.Errorf("failed to write to temp_pitot_log.json: %w", err)
		}
	}
	return nil
}
