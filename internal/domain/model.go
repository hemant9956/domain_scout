package domain

type Domain struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Source   string `json:"source"`
	ServerID string `json:"server_id,omitempty"`
}
