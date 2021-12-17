package main

import (
	"go-evm-indexer/app"
	"go-evm-indexer/config"
	"log"
	"path/filepath"
)

func main() {
	configFile, err := filepath.Abs(".env")
	if err != nil {
		log.Fatalf("❌ failed to find `.env` file : %s\n", err.Error())
	}
	config.Read(configFile)

	log.Println("running...")
	app.Run()
}
