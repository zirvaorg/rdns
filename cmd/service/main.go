package main

import (
	"github.com/joho/godotenv"
	"log"
	"rdns/internal/service"
	"sync"
	"time"
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
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
		time.Sleep(30 * time.Second)
	}()

	// Scanner service
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			services.ScannerService()
			time.Sleep(30 * time.Second)
		}
	}()

	wg.Wait()
}
