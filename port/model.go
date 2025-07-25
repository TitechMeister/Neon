package port

type ConnectionRequest struct {
	// Connect to the port
	PortName string `json:"portname"`
	// Baud rate for the connection
	BaudRate int `json:"baudrate"`
}

type AvailablePort struct {
	Description string `json:"description"`
	Device      string `json:"device"`
	Hwid        string `json:"hwid"`
}

type AvailablePortsResponse struct {
	AvailablePorts []AvailablePort `json:"available_ports"`
}

type Port struct {
	ConnectState string `json:"connected"` // Port name if connected, empty string if not
	BaudRate     int    `json:"baudrate"`  // Baud rate of the connection
}

type PortState struct {
	State string `json:"state"` // State of the port (e.g., "connected", "disconnected", "error")
}
