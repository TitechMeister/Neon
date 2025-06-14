package altimeter

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/TitechMeister/Neon/cloudstorage"
	"github.com/labstack/echo"
)

// 新しいAltimeterの構造体を返す
func New() *Altimeter {
	return &Altimeter{
		DataHistory: []AltimeterData{},
	}
}

// Altimeterエンドポイント→現在はモックデータとしてAltimeter構造体のJSONを返す
func (handler *Altimeter) GetAltimeterData(c echo.Context) error {
	// モックデータを返す
	data := AltimeterData{
		Altitude:    100.0 + rand.Float64()*50.0, // Random altitude between 100 and 150 meters
		Pressure:    1013.25,
		Temperature: 20.0,
		Humidity:    50.0,
		Timestamp:   time.Now(), // Example timestamp
		DeviceID:    "device123",
	}
	// データを履歴に追加
	handler.addData(data)
	// JSON形式でデータを返す
	return c.JSON(200, data)
}

func (handler *Altimeter) PostAltimeterDataLog(c echo.Context) error {
	// 現在までのデータをログに追記
	err := handler.makeLogJson(handler.DataHistory)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error writing altimeter data log: %v", err))
	}
	// ログファイルのリネーム
	newName := fmt.Sprintf("altimeter_log_%s.json", time.Now().Format("20060102_150405"))
	err = os.Rename("temp_altimeter_log.json", newName)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error renaming altimeter log file: %v", err))
	}
	// ログファイルのリネームが成功したら履歴をクリア
	handler.DataHistory = []AltimeterData{}
	err = cloudstorage.UploadFile(c.Response().Writer, "25_logs", newName)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error uploading altimeter log file: %v", err))
	}
	// データのDLリンクを返す
	return c.String(200, fmt.Sprintf("Download Link for Altimeter Data: %s", "https://example.com/download/altimeter_data.json"))
}

// 現在のデータ履歴を取得する
func (handler *Altimeter) GetAltimeterHistory(c echo.Context) error {
	// 履歴データを返す
	return c.JSON(200, handler.DataHistory)
}

func (handler *Altimeter) addData(data AltimeterData) {
	// データを履歴に追加
	handler.DataHistory = append(handler.DataHistory, data)
	// 履歴が200件を超えたらjsonに書き込んで古い100件のデータを削除
	if len(handler.DataHistory) > 20 {
		handler.makeLogJson(handler.DataHistory[:10])
		handler.DataHistory = handler.DataHistory[len(handler.DataHistory)-10:]
	}
}

func (handler *Altimeter) makeLogJson(data []AltimeterData) error {
	// JSONファイルに書き込む
	file, err := os.OpenFile("temp_altimeter_log.json", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open temp_altimeter_log.json: %w", err)
	}
	defer file.Close()
	fi, _ := file.Stat()
	leng := fi.Size()

	json_, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal altimeter data: %w", err)
	}

	if leng == 0 {
		_, err = file.Write(fmt.Appendf(nil, `%s`, json_))
		if err != nil {
			return fmt.Errorf("failed to write to temp_altimeter_log.json: %w", err)
		}
	} else {
		// 頭の1文字[は削る
		json_ = json_[1:]
		_, err = file.WriteAt(fmt.Appendf(nil, `,%s`, json_), leng-1)
		if err != nil {
			return fmt.Errorf("failed to write to temp_altimeter_log.json: %w", err)
		}
	}
	return nil
}
