package controllers

import (
	"billingo/models"
	"context"
	"fmt"
	"sync"

	"gorm.io/gorm"
)

type Task interface {
	Setup(db *gorm.DB, manager *Manager)
	Main(ctx context.Context)
	String() string
}

type Manager struct {
	vmRRDData  map[int]models.RRDData
	db         *gorm.DB
	lock       sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	tasks      []Task
	changeChan chan models.VMData
}

func NewManager(ctx context.Context, db *gorm.DB) *Manager {
	cCtx, cancel := context.WithCancel(ctx)
	manager := &Manager{
		vmRRDData:  make(map[int]models.RRDData),
		db:         db,
		ctx:        cCtx,
		cancel:     cancel,
		tasks:      []Task{},
		changeChan: make(chan models.VMData, 300),
	}

	manager.LoadLatestMetrics()
	return manager
}

// LoadLatestMetrics retrieves the latest metric for each distinct VM, ordering by time descending and limiting to 1 record per VM.
func (m *Manager) LoadLatestMetrics() {
	var data []models.Data

	// Query the latest RRDData for each distinct VM
	err := m.db.
		Model(&models.Data{}).
		Select("vmid, max(time) as time").
		Group("vmid").
		Order("time desc").
		Find(&data).Error

	if err != nil {
		fmt.Printf("Error fetching latest metrics: %v\n", err)
		return
	}

	// Populate the vmRRDData map with the latest data for each VM
	for _, d := range data {
		var latest models.RRDData
		err := m.db.Where("vmid = ? AND time = ?", d.VMID, d.Time).First(&latest).Error
		if err != nil {
			fmt.Printf("Error fetching latest RRDData for VMID %d: %v\n", d.VMID, err)
			continue
		}

		// Lock the map and set the latest RRDData for each VMID
		m.lock.Lock()
		m.vmRRDData[d.VMID] = latest
		m.lock.Unlock()

		fmt.Printf("Loaded latest RRDData for VMID %d: %+v\n", d.VMID, latest)
	}
}

// AddTask registers a new task with the Manager.
func (m *Manager) AddTask(task Task) {
	m.tasks = append(m.tasks, task)
}

// StartAll starts all periodic tasks managed by the Manager.
func (m *Manager) StartAll() {
	for _, task := range m.tasks {
		task.Setup(m.db, m)
		go task.Main(m.ctx)
	}
}

// StopAll stops all periodic tasks.
func (m *Manager) StopAll() {
	m.cancel()
}

// GetMetric retrieves the metric for a given VM ID.
func (m *Manager) GetMetric(vmID int) (models.RRDData, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	value, exists := m.vmRRDData[vmID]
	return value, exists
}

// SetMetric updates the metric for a given VM ID and triggers state change actions if needed.
func (m *Manager) SetMetric(vmID int, value models.RRDData) {
	m.lock.Lock()
	defer m.lock.Unlock()

	oldValue, exists := m.vmRRDData[vmID]
	if !exists || oldValue.Time < value.Time {
		m.vmRRDData[vmID] = value
		// fmt.Printf("State %d has changed\n", vmID)
		data := models.VMData{
			RRDData: value,
			VMID:    vmID,
		}
		// Notify the observer with the updated data
		select {
		case m.changeChan <- data: // Send models.VMData to the channel
		default:
		}
	} else {
		fmt.Println(oldValue.Time, "", value.Time)
	}
}

// GetAllVMData retrieves all vmRRDData stored in the manager.
func (m *Manager) GetAllVMData() map[int]models.RRDData {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.vmRRDData
}
