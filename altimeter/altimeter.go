package altimeter

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

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
		// JSONファイルに書き込む
		file, _ := os.OpenFile("altimeter_log.json", os.O_RDWR|os.O_CREATE, 0600)
		defer file.Close()
		fi, _ := file.Stat()
		leng := fi.Size()

		json_, _ := json.Marshal(handler.DataHistory[:10])

		if leng == 0 {
			file.Write(fmt.Appendf(nil, `%s`, json_))
		} else {
			// 頭の1文字[は削る
			json_ = json_[1:]
			file.WriteAt(fmt.Appendf(nil, `,%s`, json_), leng-1)
		}
		handler.DataHistory = handler.DataHistory[len(handler.DataHistory)-10:]
	}
}
