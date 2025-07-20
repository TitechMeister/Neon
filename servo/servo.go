package servo

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

// 新しいServoの構造体を返す
func New() *Servo {
	return &Servo{
		DataHistory: []ServoData{},
		Client:      &http.Client{}, // HTTPクライアントを初期化
	}
}

func (s *Servo) GetSencorName() string {
	// センサーの名前を返す
	return "servo"
}

// Servoエンドポイント→現在はモックデータとしてServoData構造体のJSONを返す
func (handler *Servo) GetData(c echo.Context) error {
	// モックデータを返す
	data := ServoData{}
	if os.Getenv("MODE") == "mock" {
		data = ServoData{
			ID:                  1,
			Status:              1, // Active status
			Timestamp:           uint32(time.Now().Unix()),
			Rudder:              -30.0 + rand.Float64()*60.0, // Random rudder angle -30 to +30 degrees
			Elevator:            -15.0 + rand.Float64()*30.0, // Random elevator angle -15 to +15 degrees
			Voltage:             11.0 + rand.Float64()*2.0,   // Random voltage 11-13V
			RudderCurrent:       1.0 + rand.Float64()*3.0,    // Random current 1-4A
			ElevatorCurrent:     1.0 + rand.Float64()*3.0,    // Random current 1-4A
			Trim:                -5.0 + rand.Float64()*10.0,  // Random trim -5 to +5 degrees
			RudderServoAngle:    -45.0 + rand.Float64()*90.0, // Random servo angle -45 to +45 degrees
			ElevatorServoAngle:  -30.0 + rand.Float64()*60.0, // Random servo angle -30 to +30 degrees
			RudderTemperature:   25.0 + rand.Float64()*20.0,  // Random temperature 25-45°C
			ElevatorTemperature: 25.0 + rand.Float64()*20.0,  // Random temperature 25-45°C
			ReceivedTime:        uint64(time.Now().UnixMilli()),
		}
	} else {
		// 実際のデータを取得する
		fmt.Println("Fetching servo data from the server...")
		req, err := http.NewRequest("GET", "http://localhost:7878/data/servo", nil)
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
			return c.String(res.StatusCode, fmt.Sprintf("Error fetching servo data: %s", res.Status))
		}
		// レスポンスボディをデコード
		if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
			return c.String(500, fmt.Sprintf("Error decoding servo data: %v", err))
		}
	}
	// データを履歴に追加
	handler.addData(data)
	// JSON形式でデータを返す
	return c.JSON(200, data)
}

func (handler *Servo) LogData(c echo.Context) error {
	// 現在までのデータをログに追記
	res := &ServoDLlink{}
	err := handler.makeLogJson(handler.DataHistory)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error writing servo data log: %v", err))
	}
	// ログファイルのリネーム
	newName := fmt.Sprintf("servo_log_%s.json", time.Now().Format("20060102_150405"))
	err = os.Rename("temp_servo_log.json", newName)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error renaming servo log file: %v", err))
	}
	// ログファイルのリネームが成功したら履歴をクリア
	handler.DataHistory = []ServoData{}
	url, err := cloudstorage.UploadFile(c.Response().Writer, "25_logs", newName)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error uploading servo log file: %v", err))
	}
	// データのDLリンクを返す
	res.DownloadLink = *url
	res.Timestamp = time.Now()
	return c.JSON(200, res)
}

// 現在のデータ履歴を取得する
func (handler *Servo) GetHistory(c echo.Context) error {
	// 履歴データを返す
	return c.JSON(200, handler.DataHistory)
}

func (handler *Servo) addData(data ServoData) {
	// データを履歴に追加
	handler.DataHistory = append(handler.DataHistory, data)
	// 履歴が20件を超えたらjsonに書き込んで古い10件のデータを削除
	if len(handler.DataHistory) > 20 {
		handler.makeLogJson(handler.DataHistory[:10])
		handler.DataHistory = handler.DataHistory[len(handler.DataHistory)-10:]
	}
}

func (handler *Servo) makeLogJson(data []ServoData) error {
	// JSONファイルに書き込む
	file, err := os.OpenFile("temp_servo_log.json", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open temp_servo_log.json: %w", err)
	}
	defer file.Close()
	fi, _ := file.Stat()
	leng := fi.Size()

	json_, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal servo data: %w", err)
	}

	if leng == 0 {
		_, err = file.Write(fmt.Appendf(nil, `%s`, json_))
		if err != nil {
			return fmt.Errorf("failed to write to temp_servo_log.json: %w", err)
		}
	} else {
		// 頭の1文字[は削る
		json_ = json_[1:]
		_, err = file.WriteAt(fmt.Appendf(nil, `,%s`, json_), leng-1)
		if err != nil {
			return fmt.Errorf("failed to write to temp_servo_log.json: %w", err)
		}
	}
	return nil
}
