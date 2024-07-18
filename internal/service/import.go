package service

import (
	"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"
	"rdns/internal/durable"
	"rdns/internal/model"
	"strconv"
	"strings"
	"sync"
)

type ImportService struct{}

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

	content, err := io.ReadAll(gr)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (i *ImportService) ProcessData(data string) error {
	var (
		batchSize, _   = strconv.Atoi(os.Getenv("BATCH_SIZE"))
		maxRoutines, _ = strconv.Atoi(os.Getenv("MAX_ROUTINES"))
		lines          = strings.Split(data, "\n")
		domains        = make([]model.Domain, 0, len(lines))
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

	var (
		tx        = durable.Connection().Begin()
		wg        sync.WaitGroup
		mu        sync.Mutex
		semaphore = make(chan struct{}, maxRoutines)
	)

	for start := 0; start < len(domains); start += batchSize {
		end := start + batchSize
		if end > len(domains) {
			end = len(domains)
		}
		batch := domains[start:end]

		wg.Add(1)
		semaphore <- struct{}{}

		go func(batch []model.Domain) {
			defer wg.Done()
			defer func() {
				<-semaphore
			}()

			for _, domain := range batch {
				mu.Lock()
				if err := tx.Create(&domain).Error; err != nil {
					if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
						log.Println("Error creating domain:", err)
					}
				}
				mu.Unlock()
			}
		}(batch)
	}
	wg.Wait()

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
