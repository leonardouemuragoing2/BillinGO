package controllers

import (
	"billingo/models"
	"billingo/proxmox"
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

type MetricsTask struct {
	name    string
	db      *gorm.DB
	manager *Manager
}

func NewMetrics(name string) *MetricsTask {
	return &MetricsTask{name: name}
}

func (t *MetricsTask) Setup(db *gorm.DB, manager *Manager) {
	t.db = db
	t.manager = manager
}

func (t *MetricsTask) Main(ctx context.Context) {
	ticker := time.NewTicker(50 * time.Minute)
	defer ticker.Stop()

	resourceTypes := []models.ResourceType{models.QEMU, models.LXC}
	timeframe := "hour"

	// Trigger the task logic immediately
	go func() {
		fmt.Printf("Task %s is running (initial run)\n", t.name)
		nodes, err := proxmox.FetchNodes()
		if err != nil {
			fmt.Printf("Error fetching nodes: %v\n", err)
			return
		}

		for _, node := range nodes {
			go t.processNode(node.Node, timeframe, resourceTypes)
		}
	}()

	for {
		select {
		case <-ticker.C:
			// Perform task logic here
			fmt.Printf("Task %s is running\n", t.name)
			nodes, err := proxmox.FetchNodes()
			if err != nil {
				fmt.Printf("Error fetching nodes: %v\n", err)
				continue
			}

			for _, node := range nodes {
				go t.processNode(node.Node, timeframe, resourceTypes)
			}
		case <-ctx.Done():
			fmt.Printf("Task %s is stopping\n", t.name)
			return
		}
	}
}

func (t *MetricsTask) String() string {
	return t.name
}

// processNode processes a node to fetch and update VM/container metrics.
func (t *MetricsTask) processNode(node string, timeframe string, resourceTypes []models.ResourceType) {
	for _, resourceType := range resourceTypes {
		resources, err := proxmox.FetchResources(node, resourceType)
		if err != nil {
			fmt.Printf("Error fetching resources of type %s on node %s: %v\n", resourceType, node, err)
			continue
		}

		results := make(chan map[int][]models.RRDData, len(resources))
		var wg sync.WaitGroup

		for _, resource := range resources {
			wg.Add(1)
			go proxmox.RRDWorker(&wg, node, resourceType, resource.VMID, timeframe, results)
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		t.processResults(results, resourceType, node)
	}
}

func (t *MetricsTask) processResults(results chan map[int][]models.RRDData, resourceType models.ResourceType, node string) {
	rrdResults := make(map[int][]models.RRDData)
	for result := range results {
		for vmid, data := range result {
			rrdResults[vmid] = data
		}
	}

	for vmid, data := range rrdResults {
		// fmt.Printf("--- Dados históricos para %s ID %d no nó %s:\n", resourceType, vmid, node)
		for _, point := range data {
			// Atualiza o timestamp para o VMID
			vmRRDData, exists := t.manager.GetMetric(vmid)
			if exists && vmRRDData.Time >= point.Time {
				continue
			}

			if point.CPU != nil { // Verifica se CPU não é nil (ou seja, está online)
				t.manager.SetMetric(vmid, point)
				// cpuUsage := (*point.CPU / *point.MaxCPU) * 100
				// memUsage := (*point.Mem / *point.MaxMem) * 100
				// diskUsage := (*point.Disk / *point.MaxDisk) * 100
				// netInOut := (*point.NetIn / *point.NetOut) * 100
				// diskReadWrite := (*point.DiskRead / *point.DiskWrite) * 100

				// fmt.Printf("Time: %s %s ON, CPU: %.2f / %.2f | Mem: %.2f / %.2f | Disk:%.2f / %.2f | NetInOut: %.2f / %.2f |  DiskReadWrite: %.2f / %.2f \n", time.Unix(int64(point.Time), 0), resourceType, *point.CPU, *point.MaxCPU, *point.Mem, *point.MaxMem, *point.Disk, *point.MaxDisk, *point.NetIn, *point.NetOut, *point.DiskRead, *point.DiskWrite)
			}
		}
	}
}
