package main

import "github.com/TitechMeister/Neon/setup"

func main() {
	// This is the main entry point for the Neon application.
	// The main function is currently empty, but you can add your application logic here.
	// For example, you might want to initialize the application, set up routes, or start a server.

	e := setup.Setup()               // Call the setup function to initialize the application
	e.Logger.Fatal(e.Start(":8080")) // Start the Echo server on port 8080

}
