package controllers

import (
	"billingo/models"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"gorm.io/gorm"
)

type ObserverBufferTask struct {
	manager *Manager
	lock    sync.Mutex
	name    string
	file    *os.File
}

func NewObserverBuffer(name string) *ObserverBufferTask {
	return &ObserverBufferTask{name: name}
}

func (t *ObserverBufferTask) Setup(db *gorm.DB, manager *Manager) {
	t.manager = manager
	var err error

	// Open or create the file to store the buffer data
	t.file, err = os.OpenFile(t.name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open file %s: %v\n", t.name, err)
	}
}

func (t *ObserverBufferTask) Main(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-t.manager.changeChan:
			// When a change is detected, process the updated data
			t.saveRowToFile(data)
		}
	}
}

func (t *ObserverBufferTask) saveRowToFile(data models.VMData) {
	// Lock to avoid race conditions when writing to the buffer
	t.lock.Lock()
	defer t.lock.Unlock()

	// Encode the buffer data to JSON and write to file
	encoder := json.NewEncoder(t.file)
	if err := encoder.Encode(data); err != nil {
		fmt.Printf("Failed to encode data to file: %v\n", err)
	}
}

func (t *ObserverBufferTask) String() string {
	return t.name
}
