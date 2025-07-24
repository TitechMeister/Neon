package gps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/TitechMeister/Neon/cloudstorage"
	"github.com/labstack/echo"
)

// 新しいGPSの構造体を返す
func New(logFrequency int) *GPS {
	return &GPS{
		DataHistory:  []GPSData{},
		Client:       &http.Client{}, // HTTPクライアントを初期化
		LogFrequency: logFrequency,   // ログ更新周波数を設定
	}
}

func (g *GPS) GetSencorName() string {
	// センサーの名前を返す
	return "gps"
}

// GPSエンドポイント→現在はモックデータとしてGPSData構造体のJSONを返す
func (handler *GPS) GetData(c echo.Context) error {
	// DataHistoryの最新一件
	if len(handler.DataHistory) == 0 {
		return c.String(404, "No GPS data available")
	}
	data := handler.DataHistory[len(handler.DataHistory)-1]

	// JSON形式でデータを返す
	formattedData := handler.formatGPSData(data)
	return c.JSON(200, formattedData)
}

func (handler *GPS) PostData(c echo.Context) error {
	// 現在までのデータをログに追記
	res := &GPSDLlink{}
	err := handler.makeLogJson(handler.DataHistory)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error writing GPS data log: %v", err))
	}
	// ログファイルのリネーム
	newName := fmt.Sprintf("logs/gps_log_%s.json", time.Now().Format("20060102_150405"))
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

func (handler *GPS) PostTarget(c echo.Context) error {
	// リクエストボディからTargetDataを取得
	var targetData TargetData
	if err := c.Bind(&targetData); err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Error binding target data: %v", err))
	}

	// 受け取ったデータをログ表示
	fmt.Printf("Received target data: %+v\n", targetData)

	// 受け取ったデータ全体を48個の整数に変換
	dataBytes := make([]byte, 48)
	dataBytes[0] = targetData.ID
	// 1~3はパディング
	for i := 1; i < 4; i++ {
		dataBytes[i] = 0 // パディングは0で埋める
	}
	dataBytes[4] = byte(targetData.Timestamp >> 24)
	dataBytes[5] = byte(targetData.Timestamp >> 16)
	dataBytes[6] = byte(targetData.Timestamp >> 8)
	dataBytes[7] = byte(targetData.Timestamp)
	dataBytes[8] = byte(targetData.TargetLon >> 24)
	dataBytes[9] = byte(targetData.TargetLon >> 16)
	dataBytes[10] = byte(targetData.TargetLon >> 8)
	dataBytes[11] = byte(targetData.TargetLon)
	dataBytes[12] = byte(targetData.TargetLat >> 24)
	dataBytes[13] = byte(targetData.TargetLat >> 16)
	dataBytes[14] = byte(targetData.TargetLat >> 8)
	dataBytes[15] = byte(targetData.TargetLat)
	// 16~47はデータの内容を埋める
	for i := 16; i < 48; i++ {
		if i < len(targetData.Data)+16 {
			dataBytes[i] = targetData.Data[i-16]
		} else {
			dataBytes[i] = 0 // データが足りない場合は0で埋める
		}
	}
	// データバイトのログ表示
	// 48個の整数に変換したデータをログに出力
	fmt.Printf("Target data bytes: %x\n", dataBytes)

	// データを履歴に追加
	var payloadArr [48]byte
	copy(payloadArr[:], dataBytes)
	targetPayload := TargetPayload{Payload: payloadArr}
	// データをjsonとしてread
	req, err := http.NewRequest("POST", "http://localhost:7878/serial/write", nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error creating request: %v", err))
	}
	// リクエストヘッダーにContent-Typeを設定
	req.Header.Set("Content-Type", "application/json")
	// リクエストボディにtargetPayloadのjsonを設定
	jsonPayload, err := json.Marshal(targetPayload)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error marshalling target payload: %v", err))
	}
	req.ContentLength = int64(len(jsonPayload))
	req.Body = io.NopCloser(bytes.NewBuffer(jsonPayload))
	// リクエストを送信
	res, err := handler.Client.Do(req)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error sending request: %v", err))
	}
	defer res.Body.Close()

	// レスポンスボディを読み取る
	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error reading response body: %v", err))
	}
	// レスポンスの内容をログに出力
	fmt.Printf("Response from server: %s\n", responseBody)

	// レスポンスのステータスコードをチェック
	if res.StatusCode != http.StatusOK {
		return c.String(res.StatusCode, fmt.Sprintf("Error sending target data: %s", res.Status))
	}

	// 成功レスポンスを返す
	return c.JSON(http.StatusOK, map[string]string{"message": "Target data added successfully"})
}

// localhost:7878を叩いてデータを取得して、履歴に追加する
// ゴルーチンで一定時間間隔で取得させることを想定
// UIからの操作とは独立にサーバ内で行う
// モックモードならモックデータを返す
func (handler *GPS) LogData() error {
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
			Lon:          uint32(1360952000 + rand.Intn(1000000)), // 136.0952 (竹生島付近) + ランダム
			Lat:          uint32(352786000 + rand.Intn(1000000)),  // 35.2786 (竹生島付近) + ランダム
			Height:       uint32(50000 + rand.Intn(100000)),       // Random height 50-150m (in mm)
			HAcc:         uint32(1000 + rand.Intn(5000)),          // Horizontal accuracy 1-6m (in mm)
			VAcc:         uint32(2000 + rand.Intn(8000)),          // Vertical accuracy 2-10m (in mm)
			GSpeed:       uint32(rand.Intn(50000)),                // Random ground speed 0-50 m/s (in mm/s)
			HeadMot:      uint32(rand.Intn(360000000)),            // Random heading 0-360 degrees (in 1e-5 degrees)
			ReceivedTime: uint64(now.UnixMilli()),
		}
	} else {
		// 実際のデータを取得する
		fmt.Println("Fetching GPS data from the server...")
		req, err := http.NewRequest("GET", "http://localhost:7878/data/gps", nil)
		if err != nil {
			return err
		}
		res, err := handler.Client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			return fmt.Errorf("server returned status %d", res.StatusCode)
		}
		// レスポンスボディをデコード
		if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
			return err
		}
	}
	// データを履歴に追加
	handler.addData(data)
	return nil
}

func (handler *GPS) GetLogFrequency() int {
	// ログ更新周波数を返す
	return handler.LogFrequency
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

func (handler *GPS) formatGPSData(data GPSData) GPSUIData {
	// データ履歴を返す
	return GPSUIData{
		Unixtime:     data.Unixtime,
		Lon:          data.Lon,
		Lat:          data.Lat,
		ReceivedTime: data.ReceivedTime,
	}
}
