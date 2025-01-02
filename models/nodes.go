package models

type Node struct {
	SSLFingerprint string  `json:"ssl_fingerprint"`
	MaxDisk        int64   `json:"maxdisk"`
	Uptime         int64   `json:"uptime"`
	Type           string  `json:"type"`
	ID             string  `json:"id"`
	Mem            int64   `json:"mem"`
	Disk           int64   `json:"disk"`
	Node           string  `json:"node"`
	CPU            float64 `json:"cpu"`
	Status         string  `json:"status"`
	Level          string  `json:"level"`
	MaxMem         int64   `json:"maxmem"`
	MaxCPU         int     `json:"maxcpu"`
}
