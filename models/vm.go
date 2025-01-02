package models

type VM struct {
	VMID   int    `json:"vmid"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type ResourceType string

const (
	QEMU ResourceType = "qemu"
	LXC  ResourceType = "lxc"
)
