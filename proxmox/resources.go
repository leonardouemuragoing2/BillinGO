package proxmox

import (
	"billingo/models"
	"billingo/utils"
	"encoding/json"
	"fmt"
)

func FetchNodes() ([]models.Node, error) {
	body, err := utils.FetchJSON("/nodes")
	if err != nil {
		return nil, err
	}

	var response struct {
		Data []models.Node `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func FetchResources(node string, resourceType models.ResourceType) ([]models.VM, error) {
	url := fmt.Sprintf("/nodes/%s/%s", node, resourceType)
	body, err := utils.FetchJSON(url)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data []models.VM `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}
