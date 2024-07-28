package service

import (
	"compress/gzip"
	"gorm.io/gorm"
	"io"
	"log"
	"os"
	"path/filepath"
	"rdns/internal/durable"
	"rdns/internal/model"
	"strconv"
	"strings"
)

type ImportService struct{}

func (i *ImportService) createDBIfNotExists(dbName string) (*gorm.DB, error) {
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		db, err := durable.ConnectDB(dbName)
		if err != nil {
			return nil, err
		}
		if err := db.AutoMigrate(&model.Domain{}); err != nil {
			return nil, err
		}
		return db, nil
	}
	return durable.ConnectDB(dbName)
}

func (i *ImportService) ReadGZFiles(dir string, suffix string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), suffix) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func (i *ImportService) ExtractAndReadGZ(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gr.Close()

	var builder strings.Builder
	buf := make([]byte, 4096) // 4KB buffer
	for {
		n, err := gr.Read(buf)
		if err != nil && err != io.EOF {
			return "", err
		}
		if n == 0 {
			break
		}
		builder.Write(buf[:n])
	}

	return builder.String(), nil
}

func (i *ImportService) ProcessData(data string, dbName string) error {
	db, err := i.createDBIfNotExists(dbName)
	if err != nil {
		return err
	}

	var (
		batchSize, _ = strconv.Atoi(os.Getenv("BATCH_SIZE"))
		lines        = strings.Split(data, "\n")
		domains      = make([]model.Domain, 0, len(lines))
	)

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		domainParts := strings.Split(fields[0], ".")
		if len(domainParts) < 2 || domainParts[0] == "" || domainParts[1] == "" {
			continue
		}

		domains = append(domains, model.Domain{
			Name: domainParts[0],
			TLD:  domainParts[1],
		})
	}

	tx := db.Begin()

	for start := 0; start < len(domains); start += batchSize {
		end := start + batchSize
		if end > len(domains) {
			end = len(domains)
		}
		batch := domains[start:end]

		for _, domain := range batch {
			if err := tx.Create(&domain).Error; err != nil {
				if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
					log.Println("Error creating domain:", err)
				}
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (i *ImportService) RenameGZFile(originalPath, oldExtension string, newExtension string) error {
	newFilePath := strings.TrimSuffix(originalPath, oldExtension) + newExtension
	err := os.Rename(originalPath, newFilePath)
	if err != nil {
		return err
	}

	return nil
}
