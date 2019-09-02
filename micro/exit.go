package main

import (
	"log"
	"os"
	"os/signal"
)

func catchTermSignal(app *App) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			log.Print("\nStopping the service....\n")
			close(app.quit)
		}
	}()
}

