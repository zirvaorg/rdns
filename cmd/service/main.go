package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"rdns/internal/durable"
	"rdns/internal/model"
	"rdns/internal/service"
	"sync"
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	if err := durable.ConnectDB(os.Getenv("DB_LOCATION")); err != nil {
		log.Fatal("Error connecting to database")
	}

	if err := durable.Connection().AutoMigrate(&model.WhoIs{}, &model.Domain{}); err != nil {
		log.Fatal(err)
	}
}

func main() {
	var wg sync.WaitGroup
	services := service.Setup{}

	// Import service
	wg.Add(1)
	go func() {
		defer wg.Done()
		services.ImportService()
	}()

	// Scanner service
	wg.Add(1)
	go func() {
		defer wg.Done()
		services.ScannerService()
	}()

	wg.Wait()
}
