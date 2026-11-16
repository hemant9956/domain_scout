package server

type Info struct {
	Hostname        string `json:"hostname"`
	OperatingSystem string `json:"operating_system,omitempty"`
	Kernel          string `json:"kernel,omitempty"`
	PrimaryIP       string `json:"primary_ip,omitempty"`
}
