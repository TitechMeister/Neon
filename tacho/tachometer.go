package tacho

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

// 新しいTachoMeterの構造体を返す
func New(logFrequency int) *TachoMeter {
	return &TachoMeter{
		DataHistory:  []TachoData{},
		Client:       &http.Client{}, // HTTPクライアントを初期化
		LogFrequency: logFrequency,   // ログ更新周波数を設定
	}
}

func (t *TachoMeter) GetSencorName() string {
	// センサーの名前を返す
	return "tachometer"
}

// TachoMeterエンドポイント→現在はモックデータとしてTachoData構造体のJSONを返す
func (handler *TachoMeter) GetData(c echo.Context) error {
	// DataHistoryの最新一件
	if len(handler.DataHistory) == 0 {
		return c.String(404, "No TachoMeter data available")
	}
	data := handler.DataHistory[len(handler.DataHistory)-1]

	// UI用JSONファイルに保存
	err := handler.makeUILogJson(data)
	if err != nil {
		// ログ保存エラーがあってもレスポンスは継続
		fmt.Printf("Warning: Failed to save UI log for TachoMeter: %v\n", err)
	}

	// JSON形式でデータを返す
	return c.JSON(200, data)
}

// localhost:7878を叩いてデータを取得して、履歴に追加する
// ゴルーチンで一定時間間隔で取得させることを想定
// UIからの操作とは独立にサーバ内で行う
// モックモードならモックデータを返す
func (handler *TachoMeter) LogData() error {
	data := TachoData{}
	if os.Getenv("MODE") == "mock" {
		data = TachoData{
			ID:           1,
			Timestamp:    uint32(time.Now().Unix()),
			RPS:          1000 + rand.Float64()*500.0,  // Random RPS between 1000 and 1500
			Strain:       uint32(500 + rand.Intn(200)), // Random strain between 500 and 700
			ReceivedTime: uint64(time.Now().UnixMilli()),
		}
	} else {
		// 実際のデータを取得する
		fmt.Println("Fetching tachometer data from the server...")
		req, err := http.NewRequest("GET", "http://localhost:7878/data/tachometer", nil)
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

func (handler *TachoMeter) GetLogFrequency() int {
	// ログ更新周波数を返す
	return handler.LogFrequency
}

func (handler *TachoMeter) PostData(c echo.Context) error {
	// 現在までのデータをログに追記
	res := &TachoDLlink{}
	err := handler.makeLogJson(handler.DataHistory)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error writing tachometer data log: %v", err))
	}
	// ログファイルのリネーム
	newName := fmt.Sprintf("logs/tachometer_log_%s.json", time.Now().Format("20060102_150405"))
	err = os.Rename("temp_tachometer_log.json", newName)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error renaming tachometer log file: %v", err))
	}

	// UI用ログファイルの処理
	uiNewName := fmt.Sprintf("logs_ui/tachometer_ui_log_%s.json", time.Now().Format("20060102_150405"))
	// logs_uiディレクトリを作成（存在しない場合）
	err = os.MkdirAll("logs_ui", 0755)
	if err != nil {
		fmt.Printf("Warning: Failed to create logs_ui directory: %v\n", err)
	} else {
		// UI用ログファイルをリネーム
		err = os.Rename("temp_tachometer_ui_log.json", uiNewName)
		if err != nil {
			fmt.Printf("Warning: Failed to rename UI log file: %v\n", err)
		}
	}

	// ログファイルのリネームが成功したら履歴をクリア
	handler.DataHistory = []TachoData{}
	url, err := cloudstorage.UploadFile(c.Response().Writer, "25_logs", newName)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error uploading tachometer log file: %v", err))
	}
	// データのDLリンクを返す
	res.DownloadLink = *url
	res.Timestamp = time.Now()
	return c.JSON(200, res)
}

// 現在のデータ履歴を取得する
func (handler *TachoMeter) GetHistory(c echo.Context) error {
	// 履歴データを返す
	return c.JSON(200, handler.DataHistory)
}

func (handler *TachoMeter) addData(data TachoData) {
	// データを履歴に追加
	handler.DataHistory = append(handler.DataHistory, data)
	// 履歴が20件を超えたらjsonに書き込んで古い10件のデータを削除
	if len(handler.DataHistory) > 20 {
		handler.makeLogJson(handler.DataHistory[:10])
		handler.DataHistory = handler.DataHistory[len(handler.DataHistory)-10:]
	}
}

func (handler *TachoMeter) makeLogJson(data []TachoData) error {
	// JSONファイルに書き込む
	file, err := os.OpenFile("temp_tachometer_log.json", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open temp_tachometer_log.json: %w", err)
	}
	defer file.Close()
	fi, _ := file.Stat()
	leng := fi.Size()

	json_, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal tachometer data: %w", err)
	}

	if leng == 0 {
		_, err = file.Write(fmt.Appendf(nil, `%s`, json_))
		if err != nil {
			return fmt.Errorf("failed to write to temp_tachometer_log.json: %w", err)
		}
	} else {
		// 頭の1文字[は削る
		json_ = json_[1:]
		_, err = file.WriteAt(fmt.Appendf(nil, `,%s`, json_), leng-1)
		if err != nil {
			return fmt.Errorf("failed to write to temp_tachometer_log.json: %w", err)
		}
	}
	return nil
}

func (handler *TachoMeter) makeUILogJson(data TachoData) error {
	// UI用JSONファイルに書き込む
	file, err := os.OpenFile("temp_tachometer_ui_log.json", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open temp_tachometer_ui_log.json: %w", err)
	}
	defer file.Close()
	fi, _ := file.Stat()
	leng := fi.Size()

	json_, err := json.Marshal([]TachoData{data})
	if err != nil {
		return fmt.Errorf("failed to marshal TachoMeter UI data: %w", err)
	}

	if leng == 0 {
		_, err = file.Write(fmt.Appendf(nil, `%s`, json_))
		if err != nil {
			return fmt.Errorf("failed to write to temp_tachometer_ui_log.json: %w", err)
		}
	} else {
		// 頭の1文字[は削る
		json_ = json_[1:]
		_, err = file.WriteAt(fmt.Appendf(nil, `,%s`, json_), leng-1)
		if err != nil {
			return fmt.Errorf("failed to write to temp_tachometer_ui_log.json: %w", err)
		}
	}
	return nil
}
