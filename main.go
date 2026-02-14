package main

import (
	"log"

	"smsups-dashboard/config"
	"smsups-dashboard/gui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Config load warning: %v (using defaults)", err)
	}

	dash := gui.New(cfg)
	dash.Run(cfg)
}
