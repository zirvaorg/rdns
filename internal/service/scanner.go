package service

import (
	"fmt"
	"gorm.io/gorm"
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

func (s *ScannerService) CreateTableIfNotExist(dbName string) (*gorm.DB, error) {
	db, err := durable.ConnectDB(dbName)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(dbName); os.IsNotExist(err) || !db.Migrator().HasTable(&model.WhoIs{}) {
		if err := db.AutoMigrate(&model.WhoIs{}); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func (s *ScannerService) WhoIs(domains []model.Domain, dbName string) {
	db, err := s.CreateTableIfNotExist(dbName)
	if err != nil {
		return
	}

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
			s.batchInsertWhoIs(db, whoisRecords)
			whoisRecords = []model.WhoIs{}
		}
	}

	if len(whoisRecords) > 0 {
		s.batchInsertWhoIs(db, whoisRecords)
	}
}

func (s *ScannerService) batchInsertWhoIs(db *gorm.DB, whoisRecords []model.WhoIs) {
	tx := db.Begin()

	for _, record := range whoisRecords {
		if err := tx.Create(&record).Error; err != nil {
			if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
				fmt.Println("Error creating whois record:", err)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return
	}
}
