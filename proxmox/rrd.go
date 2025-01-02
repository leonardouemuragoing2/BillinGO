package proxmox

import (
	"billingo/models"
	"billingo/utils"
	"encoding/json"
	"fmt"
	"sync"
)

func RRDWorker(wg *sync.WaitGroup, node string, resourceType models.ResourceType, vmid int, timeframe string, results chan<- map[int][]models.RRDData) {
	defer wg.Done()

	data, err := fetchRRDData(node, resourceType, vmid, timeframe)
	if err != nil {
		fmt.Printf("Erro ao buscar dados RRD para VMID %d: %v\n", vmid, err)
		return
	}
	results <- map[int][]models.RRDData{vmid: data}
}

func fetchRRDData(node string, resourceType models.ResourceType, vmid int, timeframe string) ([]models.RRDData, error) {
	url := fmt.Sprintf("/nodes/%s/%s/%d/rrddata?timeframe=%s", node, resourceType, vmid, timeframe)
	body, err := utils.FetchJSON(url)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data []models.RRDData `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}
