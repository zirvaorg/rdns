package service

import (
	"fmt"
	"log"
	"os"
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
		journalFileName := dbFile + "-journal"
		if _, err := os.Stat(journalFileName); err == nil {
			log.Printf("Journal file %s exists, skipping database operation for %s", journalFileName, dbFile)
			continue
		}

		db, err := durable.ConnectDB(dbFile)
		if err != nil {
			log.Printf("Error connecting to database %s: %v", dbFile, err)
			continue
		}

		table, err := durable.CreateTableIfNotExist(db, &model.WhoIs{})
		if err != nil {
			log.Println("Error creating WhoIs table:", err)
			continue
		}

		if !table {
			log.Println("WhoIs table not created")
			continue
		}

		var domains []model.Domain
		result := db.Find(&domains)
		if result.Error != nil {
			log.Println("Error fetching domains from", dbFile, ":", result.Error)
		}

		scannerService.WhoIs(domains, db)
		fmt.Println("ScannerService done one step")
	}
}
