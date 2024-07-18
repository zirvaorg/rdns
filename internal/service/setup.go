package service

import (
	"log"
	"rdns/internal/durable"
	"rdns/internal/model"
)

type Setup struct{}

func (s *Setup) ScannerService() {
	scannerService := ScannerService{}

	var domains []model.Domain
	result := durable.Connection().Find(&domains)
	if result.Error != nil {
		log.Fatal("Error fetching domains:", result.Error)
	}

	scannerService.WhoIs(domains)
}

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

		err = importService.ProcessData(data)
		if err != nil {
			log.Fatal(err)
		}

		err = importService.RenameGZFile(file, ".gz", ".done")
		if err != nil {
			log.Fatal(err)
		}
	}
}
