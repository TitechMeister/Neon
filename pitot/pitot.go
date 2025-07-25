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
func New(logFrequency int) *Pitot {
	return &Pitot{
		DataHistory:  []PitotData{},
		Client:       &http.Client{}, // HTTPクライアントを初期化
		LogFrequency: logFrequency,   // ログ更新周波数を設定
	}
}

func (p *Pitot) GetSencorName() string {
	// センサーの名前を返す
	return "pitot"
}

// Pitotエンドポイント→現在はモックデータとしてPitotData構造体のJSONを返す
func (handler *Pitot) GetData(c echo.Context) error {
	// DataHistoryの最新一件
	if len(handler.DataHistory) == 0 {
		return c.String(404, "No Pitot data available")
	}
	data := handler.DataHistory[len(handler.DataHistory)-1]

	// UI用JSONファイルに保存
	err := handler.makeUILogJson(data)
	if err != nil {
		// ログ保存エラーがあってもレスポンスは継続
		fmt.Printf("Warning: Failed to save UI log for Pitot: %v\n", err)
	}

	// JSON形式でデータを返す
	return c.JSON(200, data)
}

// localhost:7878を叩いてデータを取得して、履歴に追加する
// ゴルーチンで一定時間間隔で取得させることを想定
// UIからの操作とは独立にサーバ内で行う
// モックモードならモックデータを返す
func (handler *Pitot) LogData() error {
	data := PitotData{}
	if os.Getenv("MODE") == "mock" {
		data = PitotData{
			ID:           1,
			Timestamp:    uint32(time.Now().Unix()),
			Temperature:  float32(15.0 + rand.Float64()*10.0),    // Random temperature between 15-25°C
			Velocity:     float32(rand.Float64()*10.0),    // Random velocity between 5-150 m/s
			PressureVRaw: float32(1000.0 + rand.Float64()*200.0), // Random pressure 1000-1200
			PressureARaw: float32(800.0 + rand.Float64()*300.0),  // Random pressure 800-1100
			PressureSRaw: float32(900.0 + rand.Float64()*250.0),  // Random pressure 900-1150
		}
	} else {
		// 実際のデータを取得する
		fmt.Println("Fetching pitot data from the server...")
		req, err := http.NewRequest("GET", "http://localhost:7878/data/pitot", nil)
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

func (handler *Pitot) GetLogFrequency() int {
	// ログ更新周波数を返す
	return handler.LogFrequency
}

func (handler *Pitot) PostData(c echo.Context) error {
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

	// UI用ログファイルの処理
	uiNewName := fmt.Sprintf("logs_ui/pitot_ui_log_%s.json", time.Now().Format("20060102_150405"))
	// logs_uiディレクトリを作成（存在しない場合）
	err = os.MkdirAll("logs_ui", 0755)
	if err != nil {
		fmt.Printf("Warning: Failed to create logs_ui directory: %v\n", err)
	} else {
		// UI用ログファイルをリネーム
		err = os.Rename("temp_pitot_ui_log.json", uiNewName)
		if err != nil {
			fmt.Printf("Warning: Failed to rename UI log file: %v\n", err)
		}
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

func (handler *Pitot) makeUILogJson(data PitotData) error {
	// UI用JSONファイルに書き込む
	file, err := os.OpenFile("temp_pitot_ui_log.json", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open temp_pitot_ui_log.json: %w", err)
	}
	defer file.Close()
	fi, _ := file.Stat()
	leng := fi.Size()

	json_, err := json.Marshal([]PitotData{data})
	if err != nil {
		return fmt.Errorf("failed to marshal Pitot UI data: %w", err)
	}

	if leng == 0 {
		_, err = file.Write(fmt.Appendf(nil, `%s`, json_))
		if err != nil {
			return fmt.Errorf("failed to write to temp_pitot_ui_log.json: %w", err)
		}
	} else {
		// 頭の1文字[は削る
		json_ = json_[1:]
		_, err = file.WriteAt(fmt.Appendf(nil, `,%s`, json_), leng-1)
		if err != nil {
			return fmt.Errorf("failed to write to temp_pitot_ui_log.json: %w", err)
		}
	}
	return nil
}
