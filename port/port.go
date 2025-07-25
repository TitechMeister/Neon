package port

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
)

func New() *Port {
	// Create a new instance of the Port struct.
	return &Port{
		ConnectState: "",
		BaudRate:     115200, // Default baud rate
	}
}

func (p *Port) GetAvailablePorts(c echo.Context) error {
	// ポートの一覧を取得する
	ports, err := p.getPorts()
	if err != nil {
		return c.String(500, fmt.Sprintf("Error getting available ports: %v", err))
	}
	return c.JSON(200, ports)
}

// 現在のポートの接続状況を配信

func (p *Port) GetPortState(c echo.Context) error {
	// pをJSONにして配信
	return c.JSON(200, p)
}

func (p *Port) ConnectPort(c echo.Context) error {
	// ポートに接続する
	portName := c.QueryParam("port")
	// baudrateをクエリパラメータで
	str := c.QueryParam("baudrate")
	if str == "" {
		str = "115200"
	}
	fmt.Printf("Connecting to port: %s with baudrate: %s\n", portName, str)
	baudrate, err := strconv.Atoi(str)
	if err != nil {
		return c.String(400, "Invalid baudrate parameter")
	}
	if portName == "" {
		return c.String(400, "Port name is required")
	}

	// 接続リクエストに使う構造体を整備
	reqinfo := &ConnectionRequest{
		PortName: portName,
		BaudRate: baudrate}

	// http://localhost:7878/serial/connect にreqをボディとしたPOST

	reqBody, err := json.Marshal(reqinfo)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error marshalling request body: %v", err))
	}
	resp, err := http.Post("http://localhost:7878/serial/connect", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return c.String(500, fmt.Sprintf("Error connecting to port: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.String(resp.StatusCode, fmt.Sprintf("Error connecting to port: %s", resp.Status))
	}

	// 接続成功時にログ出力
	fmt.Printf("Successfully connected to port: %s with baudrate: %d\n", portName, baudrate)

	// ここでポートに接続する処理を実装
	// 例: err := connectToPort(portName)
	// if err != nil {
	//     return c.String(500, fmt.Sprintf("Error connecting to port: %v", err))
	// }

	p.ConnectState = "READING" // Update the connected state
	p.BaudRate = baudrate

	return c.String(200, fmt.Sprintf("Connected to port: %s", portName))
}

// disconnectするエンドポイント
func (p *Port) DisconnectPort(c echo.Context) error {
	// ポートから切断する
	// http://localhost:7878/serial/disconnect にPOSTリクエストを送る
	resp, err := http.Post("http://localhost:7878/serial/disconnect", "application/json", nil)
	if err != nil {
		return c.String(500, fmt.Sprintf("Error disconnecting from port: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.String(resp.StatusCode, fmt.Sprintf("Error disconnecting from port: %s", resp.Status))
	}

	p.ConnectState = "DISCONNECTED" // ポート名を空にする
	p.BaudRate = 0

	return c.String(200, "Disconnected from port")
}

// 10秒に一回呼び出され、状態がERRORならDisconnectする
func (p *Port) CheckPort() (string, error) {
	// ポートの接続状態を確認する
	// if p.Connected == "" {
	// 	return "No port connected", nil
	// }
	resp, err := http.Get("http://localhost:7878/serial/state")
	if err != nil {
		return "", fmt.Errorf("failed to check port state: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	var state PortState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return "", fmt.Errorf("failed to decode port state: %w", err)
	}
	if state.State == "ERROR" {
		_, err = http.Post("http://localhost:7878/serial/disconnect", "application/json", nil)
		if err != nil {
			return "", fmt.Errorf("failed to disconnect port: %w", err)
		}
	}
	p.ConnectState = state.State // Update the connected state
	return state.State, nil
}

func (p *Port) getPorts() (*AvailablePortsResponse, error) {
	// localhost:7878のAPIからポート情報を取得する
	resp, err := http.Get("http://localhost:7878/serial/available_ports")
	if err != nil {
		return nil, fmt.Errorf("failed to request available ports: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var portsResponse AvailablePortsResponse
	if err := json.NewDecoder(resp.Body).Decode(&portsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &portsResponse, nil
}
