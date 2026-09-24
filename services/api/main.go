package main

import (
	"log"
	"os"

	"dev.jevido/jevidocs/services/api/app/database"
	"dev.jevido/jevidocs/services/api/bootstrap"
)

func main() {
	app := bootstrap.Boot()

	// Serving (not an `artisan ...` invocation): bring the schema up to date
	// first, and refuse to start if that fails.
	if len(os.Args) < 2 || os.Args[1] != "artisan" {
		if err := database.MigrateOnStart(); err != nil {
			log.Fatal(err)
		}
		if err := database.PrepareOnStart(); err != nil {
			log.Fatal(err)
		}
	}

	app.Start()
}
