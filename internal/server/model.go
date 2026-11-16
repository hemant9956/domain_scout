package server

type Info struct {
	Hostname        string `json:"hostname"`
	OperatingSystem string `json:"operating_system,omitempty"`
	Kernel          string `json:"kernel,omitempty"`
	PrimaryIP       string `json:"primary_ip,omitempty"`
}

type SystemInfo struct {
	Hostname     string     `json:"hostname"`
	IPv4         []string   `json:"ipv4"`
	IPv6         []string   `json:"ipv6"`
	MAC          string     `json:"mac"`
	ServerOS     string     `json:"server_os"`
	OSVersion    string     `json:"os_version"`
	Environment  string     `json:"environment"`
	Architecture string     `json:"architecture"`
	Cores        int        `json:"cores"`
	Threads      int        `json:"threads"`
	CPUModel     string     `json:"cpu_model"`
	Kernel       string     `json:"kernel"`
	Uptime       uint64     `json:"uptime"`
	CurrentUser  string     `json:"current_user"`
	Timezone     string     `json:"timezone"`
	Disk         DiskInfo   `json:"disk"`
	Memory       MemoryInfo `json:"memory"`
}

type DiskInfo struct {
	TotalGB float64 `json:"total_gb"`
	FreeGB  float64 `json:"free_gb"`
	UsedGB  float64 `json:"used_gb"`
}

type MemoryInfo struct {
	TotalGB     float64 `json:"total_gb"`
	AvailableGB float64 `json:"available_gb"`
	FreeGB      float64 `json:"free_gb"`
	UsedGB      float64 `json:"used_gb"`
}
