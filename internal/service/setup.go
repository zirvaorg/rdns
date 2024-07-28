package service

import (
	"fmt"
	"log"
	"path/filepath"
	"rdns/internal/durable"
	"rdns/internal/model"
	"strings"
)

type Setup struct{}

func (s *Setup) ImportService() {
	importService := ImportService{}

	files, err := importService.ReadGZFiles("assets/list", ".gz")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		data, err := importService.ExtractAndReadGZ(file)
		if err != nil {
			log.Fatal(err)
		}

		dbName := strings.TrimSuffix(filepath.Base(file), ".gz") + ".db"
		err = importService.ProcessData(data, "./databases/"+dbName)
		if err != nil {
			log.Fatal(err)
		}

		err = importService.RenameGZFile(file, ".gz", ".done")
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (s *Setup) ScannerService() {
	scannerService := ScannerService{}

	dbFiles, err := filepath.Glob("./databases/*.db")
	if err != nil {
		log.Fatal("Error reading database files:", err)
	}

	for _, dbFile := range dbFiles {
		db, err := durable.ConnectDB(dbFile)
		if err != nil {
			log.Fatal("Error connecting to database:", err)
		}

		var domains []model.Domain
		result := db.Find(&domains)
		if result.Error != nil {
			log.Println("Error fetching domains from", dbFile, ":", result.Error)
		}

		scannerService.WhoIs(domains, dbFile)
		fmt.Println("ScannerService done one step")
	}

}
