package status

import "time"

type Options struct {
	Timeout time.Duration
}

type Status struct {
	OS        string         `json:"os"`
	CPU       CPUInfo        `json:"cpu"`
	Memory    MemoryInfo     `json:"memory"`
	Disk      DiskInfo       `json:"disk"`
	Ports     []PortInfo     `json:"ports"`
	Processes []ProcessCount `json:"processes"`
	Notes     []string       `json:"notes"`
}

type CPUInfo struct {
	Model  string  `json:"model"`
	Cores  int     `json:"cores"`
	Load1  float64 `json:"load_1"`
	Load5  float64 `json:"load_5"`
	Load15 float64 `json:"load_15"`
}

type MemoryInfo struct {
	UsedBytes  uint64 `json:"used_bytes"`
	FreeBytes  uint64 `json:"free_bytes"`
	TotalBytes uint64 `json:"total_bytes"`
}

type DiskInfo struct {
	UsedBytes      uint64 `json:"used_bytes"`
	AvailableBytes uint64 `json:"available_bytes"`
	TotalBytes     uint64 `json:"total_bytes"`
}

type PortInfo struct {
	Port    int    `json:"port"`
	Command string `json:"command"`
	PID     int    `json:"pid"`
	User    string `json:"user"`
	InUse   bool   `json:"in_use"`
}

type ProcessCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}
