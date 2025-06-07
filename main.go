package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	// This is the main entry point for the Neon application.
	// The main function is currently empty, but you can add your application logic here.
	// For example, you might want to initialize the application, set up routes, or start a server.

	e := echo.New()
	// Create a new Echo instance, which is a web framework for Go.
	e.GET("/ping", ping)
	// Define a route that listens for GET requests on the /ping endpoint and calls the ping function.
	// middlewareのロガーを利用する
	e.Use(middleware.Logger())
	e.Start(":8080")
}

func ping(c echo.Context) error {
	// This is a simple handler function that responds with "pong" when the /ping endpoint is accessed.
	return c.String(200, "pong")
}
