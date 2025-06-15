package setup

import (
	"github.com/TitechMeister/Neon/altimeter"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func Setup() *echo.Echo{
	altimeter := altimeter.New() // Create a new instance of the Altimeter struct
	e := echoSetup(altimeter)
	return e // Return the Echo instance with the configured routes
}

func echoSetup(altimeter *altimeter.Altimeter) *echo.Echo {
	// Echoのセットアップを行う
	// ここにEchoのルーティングやミドルウェアの設定を追加する
	// This is the main entry point for the Neon application.
	// The main function is currently empty, but you can add your application logic here.
	// For example, you might want to initialize the application, set up routes, or start a server.

	e := echo.New()

	// Create a new Echo instance, which is a web framework for Go.
	e.GET("/ping", ping)
	e.GET("/altimeter", altimeter.GetAltimeterData)
	e.POST("/altimeter/log", altimeter.PostAltimeterDataLog)
	e.GET("/altimeter/history", altimeter.GetAltimeterHistory)
	// Define a route that listens for GET requests on the /ping endpoint and calls the ping function.
	// middlewareのロガーを利用する
	e.Use(middleware.Logger())
	return e
}

func ping(c echo.Context) error {
	// This is a simple handler function that responds with "pong" when the /ping endpoint is accessed.
	return c.String(200, "pong")
}
