package controllers

import (
	"billingo/models"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"
)

type DatabaseSaverTask struct {
	manager *Manager
	db      *gorm.DB
	name    string
	file    *os.File
}

func NewDatabaseSaverTask(name string) *DatabaseSaverTask {
	return &DatabaseSaverTask{name: name}
}

// Setup initializes the task with the database and manager instance
func (t *DatabaseSaverTask) Setup(db *gorm.DB, manager *Manager) {
	t.manager = manager
	t.db = db
	var err error

	// Open the buffer file for reading
	t.file, err = os.OpenFile(t.name, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open buffer file %s: %v\n", t.name, err)
	}
}

// Main is the main loop of the task
func (t *DatabaseSaverTask) Main(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// Cleanup and return on context cancellation
			if t.file != nil {
				t.file.Close()
			}
			return
		default:
			// Process the buffer file and save data to the database
			if err := t.processBufferFile(); err != nil {
				fmt.Printf("Error processing buffer file: %v\n", err)
			}
		}
		// Wait for an hour before processing the next batch
		<-time.After(2 * time.Minute)
	}
}

// String returns the name of the task
func (t *DatabaseSaverTask) String() string {
	return t.name
}

// processBufferFile reads the buffer file, parses the data, and performs batch insertions
func (t *DatabaseSaverTask) processBufferFile() error {
	if t.file == nil {
		return fmt.Errorf("buffer file is not open")
	}

	// Rewind the file to the beginning
	if _, err := t.file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to rewind buffer file: %v", err)
	}

	decoder := json.NewDecoder(t.file)
	var rows []models.Data
	for {
		var vmData models.VMData
		if err := decoder.Decode(&vmData); err != nil {
			// Break on EOF, return other errors
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("failed to decode buffer file: %v", err)
		}

		// Convert VMData to Data
		data := models.Data{
			RRDData: vmData.RRDData,
			VMID:    vmData.VMID,
		}
		rows = append(rows, data)

		// If batch size is reached, insert rows into the database
		if len(rows) >= 3000 {
			if err := t.insertBatch(rows); err != nil {
				return err
			}
			rows = rows[:0] // Clear the batch
		}
	}

	// Insert any remaining rows
	if len(rows) > 0 {
		if err := t.insertBatch(rows); err != nil {
			return err
		}
	}

	// Truncate the buffer file after processing
	if err := t.file.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate buffer file: %v", err)
	}
	return nil
}

// insertBatch performs a batch insertion of Data rows into the database
func (t *DatabaseSaverTask) insertBatch(rows []models.Data) error {
	if err := t.db.Create(&rows).Error; err != nil {
		return fmt.Errorf("failed to insert batch into database: %v", err)
	}
	fmt.Printf("Successfully inserted %d rows into the database\n", len(rows))
	return nil
}
