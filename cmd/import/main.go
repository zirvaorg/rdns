package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"rdns/internal/durable"
	"rdns/internal/model"
	"rdns/internal/service"
)

func init() {
	// load .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	if err := durable.ConnectDB(os.Getenv("DB_LOCATION")); err != nil {
		log.Fatal("Error connecting to database")
	}

	// migrate database
	if err := durable.Connection().AutoMigrate(
		&model.Domain{}); err != nil {
		log.Fatal(err)
	}
}

func main() {
	importService := service.ImportService{}

	fmt.Println("Reading files...")
	files, err := importService.ReadGZFiles("assets/list", ".gz")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fmt.Printf("Processing %s\n", file)
		data, err := importService.ExtractAndReadGZ(file)
		if err != nil {
			log.Fatal(err)
		}

		err = importService.ProcessData(data)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Renaming file...")
		err = importService.RenameGzFile(file, ".done")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Done!")
	}
}
