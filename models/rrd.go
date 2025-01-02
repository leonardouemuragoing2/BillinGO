package models

type RRDData struct {
	Time      int      `json:"time"`
	MaxCPU    *float64 `json:"maxcpu,omitempty"`
	MaxDisk   *float64 `json:"maxdisk,omitempty"`
	MaxMem    *float64 `json:"maxmem,omitempty"`
	Disk      *float64 `json:"disk,omitempty"`
	CPU       *float64 `json:"cpu,omitempty"`
	Mem       *float64 `json:"mem,omitempty"`
	NetOut    *float64 `json:"netout,omitempty"`
	NetIn     *float64 `json:"netin,omitempty"`
	DiskRead  *float64 `json:"diskread,omitempty"`
	DiskWrite *float64 `json:"diskwrite,omitempty"`
}

type VMData struct {
	RRDData
	VMID int `json:"vmid"`
}

// SyncStatusEnum defines the possible values for the SyncStatus field.
type SyncStatusEnum string

const (
	SyncStatusPending SyncStatusEnum = "pending"
	SyncStatusSuccess SyncStatusEnum = "success"
)

type Data struct {
	BaseModel
	RRDData
	VMID       int            `json:"vmid"`
	SyncStatus SyncStatusEnum `json:"sync_status" gorm:"type:enum('pending', 'success');default:'pending'"`
}

type DataRaw struct {
	BaseModel
	RRDData
	VMID  int    `json:"vmid"`
	Topic string `json:"topic"`
}
