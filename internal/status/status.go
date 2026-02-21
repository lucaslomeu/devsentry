package status

type Options struct{}

type Status struct {
	OS        string
	CPU       CPUInfo
	Memory    MemoryInfo
	Disk      DiskInfo
	Ports     []PortInfo
	Processes []ProcessCount
	Notes     []string
}

type CPUInfo struct {
	Model  string
	Cores  int
	Load1  float64
	Load5  float64
	Load15 float64
}

type MemoryInfo struct {
	UsedBytes  uint64
	FreeBytes  uint64
	TotalBytes uint64
}

type DiskInfo struct {
	UsedBytes      uint64
	AvailableBytes uint64
	TotalBytes     uint64
}

type PortInfo struct {
	Port    int
	Command string
	PID     int
	User    string
	InUse   bool
}

type ProcessCount struct {
	Name  string
	Count int
}
