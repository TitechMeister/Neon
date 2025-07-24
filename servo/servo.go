package servo

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/TitechMeister/Neon/cloudstorage"
	"github.com/labstack/echo"
)

// 新しいServoの構造体を返す
func New() *Servo {
	s := &Servo{
		DataHistory:      []ServoData{},
		Client:           &http.Client{}, // HTTPクライアントを初期化
		RevElevatorValue: []float64{},    // 初期化
		RevRudderValue:   []float64{},    // 初期化
	}
	// ラダーとエレベータの逆力学モデルを計算しておく
	s.calculateServoValue()
	return s

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
			Rudder:              -15.0 + rand.Float64()*30.0, // Random rudder angle -15 to +15 degrees
			Elevator:            -15.0 + rand.Float64()*20.0, // Random elevator angle -15 to +5 degrees
			Voltage:             11.0 + rand.Float64()*2.0,   // Random voltage 11-13V
			RudderCurrent:       1.0 + rand.Float64()*3.0,    // Random current 1-4A
			ElevatorCurrent:     1.0 + rand.Float64()*3.0,    // Random current 1-4A
			Trim:                -2.5 + rand.Float64()*3.0,   // Random trim -2.5 to +0.5 degrees
			RudderServoAngle:    -15.0 + rand.Float64()*30.0, // Random rudder angle -15 to +15 degrees
			ElevatorServoAngle:  -15.0 + rand.Float64()*20.0, // Random elevator angle -15 to +5 degrees
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
	dataUI := handler.formatServoData(data)
	// JSON形式でデータを返す
	return c.JSON(200, dataUI)
}

func (handler *Servo) PostData(c echo.Context) error {
	// 現在までのデータをログに追記
	res := &ServoDLlink{}
	err := handler.makeLogJson(handler.DataHistory)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error writing servo data log: %v", err))
	}
	// ログファイルのリネーム
	newName := fmt.Sprintf("logs/servo_log_%s.json", time.Now().Format("20060102_150405"))
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

func (handler *Servo) formatServoData(data ServoData) ServoUIData {
	rudderIndex := sort.Search(len(handler.RevRudderValue), func(i int) bool {
		return handler.RevRudderValue[i] >= data.RudderServoAngle
	})
	if rudderIndex >= len(handler.RevRudderValue) {
		rudderIndex = len(handler.RevRudderValue) - 1
	}
	elevatorIndex := sort.Search(len(handler.RevElevatorValue), func(i int) bool {
		return handler.RevElevatorValue[i] <= data.ElevatorServoAngle
	})
	if elevatorIndex >= len(handler.RevElevatorValue) {
		elevatorIndex = len(handler.RevElevatorValue) - 1
	}
	return ServoUIData{
		Timestamp:           data.Timestamp,
		Rudder:              data.Rudder,
		Elevator:            data.Elevator,
		Trim:                data.Trim,
		RudderActualAngle:   handler.RevRudderValue[rudderIndex],
		ElevatorActualAngle: handler.RevElevatorValue[elevatorIndex],
		RudderTemperature:   data.RudderTemperature,
		ElevatorTemperature: data.ElevatorTemperature,
		ReceivedTime:        data.ReceivedTime,
	}
}

func (handler *Servo) calculateServoValue() {
	// 後の計算で使うように初期化時にRuddderとElevetorの値を計算しておく
	// Rudder→引数の値を-20から20で0.01区切りで計算してスライスに格納する
	for i := -20.0; i <= 20.0; i += 0.01 {
		handler.RevRudderValue = append(handler.RevRudderValue, calcRudderAngle(i))
	}
	// Elevator→引数の値を-20から20で0.01区切りで計算してスライスに格納する
	for i := -20.0; i <= 20.0; i += 0.01 {
		handler.RevElevatorValue = append(handler.RevElevatorValue, calcElevatorAngle(i))
	}
	// 広義単調増加か確認 ログ出力をする
	for i := 1; i < len(handler.RevRudderValue); i++ {
		if handler.RevRudderValue[i] < handler.RevRudderValue[i-1] {
			fmt.Printf("Rudder value is not monotonic at index %d: %f < %f\n", i, handler.RevRudderValue[i], handler.RevRudderValue[i-1])
		}
	}
	for i := 1; i < len(handler.RevElevatorValue); i++ {
		if handler.RevElevatorValue[i] > handler.RevElevatorValue[i-1] {
			fmt.Printf("Elevator value is not monotonic at index %d: %f < %f\n", i, handler.RevElevatorValue[i], handler.RevElevatorValue[i-1])
		}
	}
	fmt.Println("Servo value calculation completed.")

}

// ラダーの逆力学モデル（4次近似）
func calcRudderAngle(u float64) float64 {
	const (
		R_SERVO_COEFF_0 = 4.39
		R_SERVO_COEFF_1 = 4.21
		R_SERVO_COEFF_2 = -0.0205
		R_SERVO_COEFF_3 = 5.84e-3
		R_SERVO_COEFF_4 = 1.12e-4
	)
	return R_SERVO_COEFF_0 +
		R_SERVO_COEFF_1*u +
		R_SERVO_COEFF_2*u*u +
		R_SERVO_COEFF_3*u*u*u +
		R_SERVO_COEFF_4*u*u*u*u +
		180.0
}

// エレベータの逆力学モデル（4次近似）
func calcElevatorAngle(u float64) float64 {
	const (
		E_SERVO_COEFF_0 = -51.2
		E_SERVO_COEFF_1 = -6.52
		E_SERVO_COEFF_2 = -0.27
		E_SERVO_COEFF_3 = -0.0301
		E_SERVO_COEFF_4 = -8.65e-4
	)
	return E_SERVO_COEFF_0 +
		E_SERVO_COEFF_1*u +
		E_SERVO_COEFF_2*u*u +
		E_SERVO_COEFF_3*u*u*u +
		E_SERVO_COEFF_4*u*u*u*u +
		180.0
}
