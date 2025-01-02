package controllers

import (
	"billingo/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/gorm"
)

type BatchSaveToDatabaseTask struct {
	name      string
	filePath  string
	batchSize int
	db        *gorm.DB
	lock      sync.Mutex
}

// NewBatchSaveToDatabaseTask initializes a new BatchSaveToDatabaseTask.
func NewBatchSaveToDatabaseTask(name, filePath string, batchSize int) *BatchSaveToDatabaseTask {
	return &BatchSaveToDatabaseTask{
		name:      name,
		filePath:  filePath,
		batchSize: batchSize,
	}
}

func (t *BatchSaveToDatabaseTask) Setup(db *gorm.DB, manager *Manager) {
	t.db = db
}

func (t *BatchSaveToDatabaseTask) Main(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			t.processFileData()
			// Sleep for a while to avoid continuous processing
			time.Sleep(2 * time.Minute)
		}
	}
}

func (t *BatchSaveToDatabaseTask) processFileData() {
	log.Println("running batch save...")
	t.lock.Lock()
	defer t.lock.Unlock()

	// Open the file for reading
	file, err := os.Open(t.filePath)
	if err != nil {
		fmt.Printf("Failed to open file %s: %v\n", t.filePath, err)
		return
	}
	defer file.Close()

	// Decode data from the file
	var buffer []models.VMData
	decoder := json.NewDecoder(file)
	for {
		var data models.VMData
		if err := decoder.Decode(&data); err != nil {
			if err.Error() != "EOF" {
				fmt.Printf("Failed to decode data from file: %v\n", err)
			}
			break
		}
		buffer = append(buffer, data)
	}

	if len(buffer) == 0 {
		return
	}

	// Convert VMData to Data
	convertedBuffer := make([]models.Data, len(buffer))
	for i, vmData := range buffer {
		convertedBuffer[i] = models.Data{
			RRDData: vmData.RRDData,
			VMID:    vmData.VMID,
		}
	}

	// Save the data to the database in batches
	for i := 0; i < len(convertedBuffer); i += t.batchSize {
		end := i + t.batchSize
		if end > len(convertedBuffer) {
			end = len(convertedBuffer)
		}
		batch := convertedBuffer[i:end]
		if err := t.db.Create(&batch).Error; err != nil {
			fmt.Printf("Failed to save batch to database: %v\n", err)
			return
		}
	}

	// Clear the file after successful save
	t.clearFile()
}

func (t *BatchSaveToDatabaseTask) clearFile() {
	file, err := os.OpenFile(t.filePath, os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to clear file %s: %v\n", t.filePath, err)
		return
	}
	defer file.Close()
}

func (t *BatchSaveToDatabaseTask) String() string {
	return t.name
}
