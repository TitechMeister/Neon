package setup

import (
	"fmt"
	"os"
	"time"

	"github.com/TitechMeister/Neon/altimeter"
	"github.com/TitechMeister/Neon/gps"
	"github.com/TitechMeister/Neon/pitot"
	"github.com/TitechMeister/Neon/port"
	"github.com/TitechMeister/Neon/servo"
	"github.com/TitechMeister/Neon/tacho"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func Setup() *echo.Echo {
	app := &Neon{}
	app.Port = port.New()         // Create a new instance of the Port struct
	app.portChecker()             // Start checking the port status periodically
	altimeter := altimeter.New(2) // Create a new instance of the Altimeter struct
	gps := gps.New(1)             // Create a new instance of the GPS struct
	pitot := pitot.New(2)         // Create a new instance of the Pitot struct
	tacho := tacho.New(1)         // Create a new instance of the TachoMeter struct
	servo := servo.New(2)         // Create a new instance of the Servo struct
	// Initialize the Altimeter instance, which is a struct that handles altimeter data.
	app.AddSencor(altimeter) // Add the Altimeter instance to the Neon application
	app.AddSencor(gps)       // Add the GPS instance to the Neon application
	app.AddSencor(pitot)     // Add the Pitot instance to the Neon application
	app.AddSencor(tacho)     // Add the TachoMeter instance to the Neon application
	app.AddSencor(servo)     // Add the Servo instance to the Neon
	// すべてのセンサーのロガーをセットアップ
	for _, sensor := range app.Sencors {
		app.loggerSetup(*sensor)
	}
	e := app.echoSetup()
	err := os.MkdirAll("logs/test", 0755)
	if err != nil {
		fmt.Printf("Warning: Failed to create logs directory: %v\n", err)
	}
	err = os.MkdirAll("logs_ui/test", 0755)
	if err != nil {
		fmt.Printf("Warning: Failed to create logs_ui directory: %v\n", err)
	}
	return e // Return the Echo instance with the configured routes
}

func (app *Neon) AddSencor(sencor Sencor) {
	// Add a new sensor to the Neon application.
	// This function currently does not do anything, but you can implement logic to add sensors if needed.
	app.Sencors = append(app.Sencors, &sencor)
}

func (app *Neon) loggerSetup(sencor Sencor) {
	// altimeter内の関数LogDataをgoroutineを用いて周波数に合わせて実行
	go func() {
		// This function sets up a logger for the Altimeter instance.
		fmt.Println("Setting up logger for sencor:", sencor.GetSencorName())

		// Tickerを作成
		ticker := time.NewTicker(time.Second / time.Duration(sencor.GetLogFrequency()))
		defer ticker.Stop() // goroutineが終了する際にTickerを停止

		// Tickerのチャンネルから定期的にシグナルを受信
		for range ticker.C {
			sencor.LogData()
		}
	}()
}

func (app *Neon) portChecker() {
	// This function checks the port status periodically.
	// It can be used to monitor the connection state of the port and handle reconnections if necessary.
	go func() {
		ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds
		defer ticker.Stop()

		for range ticker.C {
			state, err := app.Port.CheckPort()
			if err != nil {
				fmt.Printf("Error checking port: %v\n", err)
			} else {
				fmt.Printf("Port state: %s\n", state)
			}
		}
	}()
}

func (app *Neon) echoSetup() *echo.Echo {
	// Echoのセットアップを行う
	// ここにEchoのルーティングやミドルウェアの設定を追加する
	// This is the main entry point for the Neon application.
	// The main function is currently empty, but you can add your application logic here.
	// For example, you might want to initialize the application, set up routes, or start a server.

	e := echo.New()

	// Create a new Echo instance, which is a web framework for Go.
	e.GET("/ping", ping)
	e.GET("/port", app.Port.GetAvailablePorts)
	e.GET("/port/state", app.Port.GetPortState) // Add a route to get the current port state
	e.POST("/port/connect", app.Port.ConnectPort)
	e.POST("/port/disconnect", app.Port.DisconnectPort)
	for _, sencor := range app.Sencors {
		s := sencor // Create a local variable to avoid closure issues in the loop
		// Loop through all sensors in the Neon application and set up their routes.
		// Each sensor has its own routes for getting data, logging data, and getting history.
		e.GET("/data/"+(*s).GetSencorName(), (*s).GetData)
		e.POST("/data/"+(*s).GetSencorName()+"/log", (*s).PostData)
		e.GET("/data/"+(*s).GetSencorName()+"/history", (*s).GetHistory)
		// sencorタイプがGPSの場合
		if g, ok := (*s).(*gps.GPS); ok {
			// GPSセンサーの特定のルートを設定
			e.POST("/data/gps/target", (*g).PostTarget)
		}
	}

	// e.GET("/altimeter", altimeter.GetAltimeterData)
	// e.POST("/altimeter/log", altimeter.PostAltimeterDataLog)
	// e.GET("/altimeter/history", altimeter.GetAltimeterHistory)

	// Define a route that listens for GET requests on the /ping endpoint and calls the ping function.
	// middlewareのロガーを利用する
	e.Use(middleware.Logger())
	return e
}

func ping(c echo.Context) error {
	// This is a simple handler function that responds with "pong" when the /ping endpoint is accessed.
	return c.String(200, "pong")
}

func postLogNotOnUI(c echo.Context) error {
	// This function is a placeholder for handling POST requests to log data on the UI.
	// Currently, it does not perform any actions but can be implemented later.
	return c.String(200, "Log data received")
}
