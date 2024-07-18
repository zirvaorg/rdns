package service

import (
	"log"
	"os"
	"rdns/internal/durable"
	"rdns/internal/model"
	"strconv"
	"strings"
	"time"
)

type ScannerService struct{}

func (s *ScannerService) getWhoIsInfo(tld, domain string) string {
	d := domain + "." + tld
	output, err := durable.WhoisService(d, 3*time.Second)
	if err != nil {
		return ""
	}

	lines := strings.Split(output, "\n")
	var filteredLines []string

	for _, line := range lines {
		if strings.Contains(line, "last update") || strings.Contains(line, "Last Update") || strings.Contains(line, "Last update") {
			continue
		}
		filteredLines = append(filteredLines, line)
	}

	return strings.Join(filteredLines, "\n")
}

func (s *ScannerService) WhoIs(domains []model.Domain) {
	var whoisRecords []model.WhoIs
	batchSize, _ := strconv.Atoi(os.Getenv("BATCH_SIZE"))

	for _, domain := range domains {
		whoisInfo := s.getWhoIsInfo(domain.TLD, domain.Name)

		if strings.Contains(whoisInfo, "not found") {
			continue
		}

		newWhoIs := model.WhoIs{
			DomainID: domain.ID,
			Content:  whoisInfo,
			ScanTime: time.Now().Unix(),
		}
		whoisRecords = append(whoisRecords, newWhoIs)

		if len(whoisRecords) >= batchSize {
			s.batchInsertWhoIs(whoisRecords)
			whoisRecords = []model.WhoIs{}
		}
	}

	if len(whoisRecords) > 0 {
		s.batchInsertWhoIs(whoisRecords)
	}
}

func (s *ScannerService) batchInsertWhoIs(whoisRecords []model.WhoIs) {
	tx := durable.Connection().Begin()

	for _, record := range whoisRecords {
		if err := tx.Create(&record).Error; err != nil {
			log.Println("Error creating whois:", err)
			tx.Rollback()
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Println("Error committing whois:", err)
	}
}
