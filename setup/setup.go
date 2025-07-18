package setup

import (
	"github.com/TitechMeister/Neon/altimeter"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func Setup() *echo.Echo {
	app := &Neon{}
	altimeter := altimeter.New() // Create a new instance of the Altimeter struct
	// Initialize the Altimeter instance, which is a struct that handles altimeter data.
	app.AddSencor(altimeter) // Add the Altimeter instance to the Neon application
	e := app.echoSetup()
	return e // Return the Echo instance with the configured routes
}

func (app *Neon) AddSencor(sencor Sencor) {
	// Add a new sensor to the Neon application.
	// This function currently does not do anything, but you can implement logic to add sensors if needed.
	app.Sencors = append(app.Sencors, &sencor)
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
	for _, sencor := range app.Sencors {
		s := sencor // Create a local variable to avoid closure issues in the loop
		// Loop through all sensors in the Neon application and set up their routes.
		// Each sensor has its own routes for getting data, logging data, and getting history.
		e.GET("/data/"+(*s).GetSencorName(), (*s).GetData)
		e.POST("/data/"+(*s).GetSencorName()+"/log", (*s).LogData)
		e.GET("/data/"+(*s).GetSencorName()+"/history", (*s).GetHistory)
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
