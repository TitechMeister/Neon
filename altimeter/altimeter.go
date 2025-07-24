package altimeter

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

// 新しいAltimeterの構造体を返す
func New(logflequenty int) *Altimeter {
	return &Altimeter{
		DataHistory:  []AltimeterRawData{},
		Client:       &http.Client{}, // HTTPクライアントを初期化
		LogFrequency: logflequenty,   // ログ更新周波数を設定
	}
}

func (a *Altimeter) GetSencorName() string {
	// センサーの名前を返す
	return "altimeter"
}

// Altimeterエンドポイント→現在はモックデータとしてAltimeter構造体のJSONを返す
func (handler *Altimeter) GetData(c echo.Context) error {
	// DataHistoryの最新一件
	if len(handler.DataHistory) == 0 {
		return c.String(404, "No Altimeter data available")
	}
	data := handler.DataHistory[len(handler.DataHistory)-1]

	// UI用JSONファイルに保存
	err := handler.makeUILogJson(data)
	if err != nil {
		// ログ保存エラーがあってもレスポンスは継続
		fmt.Printf("Warning: Failed to save UI log for altimeter: %v\n", err)
	}

	// JSON形式でデータを返す
	formattedData := handler.formatData(data)
	return c.JSON(200, formattedData)
}

// localhost:7878を叩いてデータを取得して、履歴に追加する
// ゴルーチンで一定時間間隔で取得させることを想定
// UIからの操作とは独立にサーバ内で行う
// モックモードならモックデータを返す
func (handler *Altimeter) LogData() error {
	data := AltimeterRawData{}
	if os.Getenv("MODE") == "mock" {
		data = AltimeterRawData{
			DeviceID:     1,
			Altitude:     10 - rand.Float64()*10.0, // Random altitude between 100 and 150 meters
			Temperature:  20.0,
			Timestamp:    1234567890,    // Example timestamp
			ReceivedTime: 1622547800000, // Example received time
		}
	} else {
		// 実際のデータを取得する
		fmt.Println("Fetching altimeter data from the server...")
		req, err := http.NewRequest("GET", "http://localhost:7878/data/ultrasonic", nil)
		// http.NewRequestを使ってGETリクエストを作成
		fmt.Println("Request created:", req)
		if err != nil {
			return err
		}
		res, err := handler.Client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			return err
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

func (handler *Altimeter) GetLogFrequency() int {
	// ログ更新周波数を返す
	return handler.LogFrequency
}

func (handler *Altimeter) PostData(c echo.Context) error {
	// 現在までのデータをログに追記
	res := &AltimeterDLlink{}
	err := handler.makeLogJson(handler.DataHistory)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error writing altimeter data log: %v", err))
	}
	// ログファイルのリネーム
	newName := fmt.Sprintf("logs/altimeter_log_%s.json", time.Now().Format("20060102_150405"))
	err = os.Rename("temp_altimeter_log.json", newName)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error renaming altimeter log file: %v", err))
	}

	// UI用ログファイルの処理
	uiNewName := fmt.Sprintf("logs_ui/altimeter_ui_log_%s.json", time.Now().Format("20060102_150405"))
	// logs_uiディレクトリを作成（存在しない場合）
	err = os.MkdirAll("logs_ui", 0755)
	if err != nil {
		fmt.Printf("Warning: Failed to create logs_ui directory: %v\n", err)
	} else {
		// UI用ログファイルをリネーム
		err = os.Rename("temp_altimeter_ui_log.json", uiNewName)
		if err != nil {
			fmt.Printf("Warning: Failed to rename UI log file: %v\n", err)
		}
	}

	// ログファイルのリネームが成功したら履歴をクリア
	handler.DataHistory = []AltimeterRawData{}
	url, err := cloudstorage.UploadFile(c.Response().Writer, "25_logs", newName)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error uploading altimeter log file: %v", err))
	}
	// データのDLリンクを返す
	res.DownloadLink = *url
	res.Timestamp = time.Now()
	return c.JSON(200, res)
}

// 現在のデータ履歴を取得する
func (handler *Altimeter) GetHistory(c echo.Context) error {
	// 履歴データを返す
	return c.JSON(200, handler.DataHistory)
}

func (handler *Altimeter) addData(data AltimeterRawData) {
	// データを履歴に追加
	handler.DataHistory = append(handler.DataHistory, data)
	// 履歴が200件を超えたらjsonに書き込んで古い100件のデータを削除
	if len(handler.DataHistory) > 20 {
		handler.makeLogJson(handler.DataHistory[:10])
		handler.DataHistory = handler.DataHistory[len(handler.DataHistory)-10:]
	}
}

func (handler *Altimeter) makeLogJson(data []AltimeterRawData) error {
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

func (handler *Altimeter) makeUILogJson(data AltimeterRawData) error {
	// UI用JSONファイルに書き込む
	file, err := os.OpenFile("temp_altimeter_ui_log.json", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open temp_altimeter_ui_log.json: %w", err)
	}
	defer file.Close()
	fi, _ := file.Stat()
	leng := fi.Size()

	// フォーマットされたデータを使用
	formattedData := handler.formatData(data)
	json_, err := json.Marshal([]AltimeterUIData{formattedData})
	if err != nil {
		return fmt.Errorf("failed to marshal altimeter UI data: %w", err)
	}

	if leng == 0 {
		_, err = file.Write(fmt.Appendf(nil, `%s`, json_))
		if err != nil {
			return fmt.Errorf("failed to write to temp_altimeter_ui_log.json: %w", err)
		}
	} else {
		// 頭の1文字[は削る
		json_ = json_[1:]
		_, err = file.WriteAt(fmt.Appendf(nil, `,%s`, json_), leng-1)
		if err != nil {
			return fmt.Errorf("failed to write to temp_altimeter_ui_log.json: %w", err)
		}
	}
	return nil
}

func (handler *Altimeter) formatData(data AltimeterRawData) AltimeterUIData {
	// AltimeterRawDataをAltimeterDataに変換する
	return AltimeterUIData{
		DeviceID:     data.DeviceID,
		Altitude:     data.Altitude,
		ReceivedTime: time.UnixMilli(time.Now().UnixMilli()), // ミリ秒から秒に変換
	}

}
