package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"domain-scout/internal/api/response"
	"domain-scout/internal/discovery"
	"domain-scout/internal/ssh"
	"domain-scout/pkg/validator"
)

type DiscoveryHandler struct{ service *discovery.Service }

func NewDiscoveryHandler(s *discovery.Service) *DiscoveryHandler {
	return &DiscoveryHandler{service: s}
}

type discoverRequest struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	AuthType   string `json:"auth_type"`
	Password   string `json:"password,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
	Passphrase string `json:"passphrase,omitempty"`
	KnownHosts string `json:"known_hosts"`
}

func (h *DiscoveryHandler) Discover(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req discoverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}
	if err := validator.ValidateSSH(req.Host, req.Port, req.Username, req.AuthType, req.KnownHosts); err != nil {
		response.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	auth, err := ssh.NewAuth(ssh.AuthInput{
		Type: req.AuthType, Password: req.Password,
		PrivateKey: req.PrivateKey, Passphrase: req.Passphrase,
	})
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_auth", err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()

	client, err := ssh.Connect(ctx, ssh.ConnectionConfig{
		Host: req.Host, Port: req.Port, Username: req.Username,
		Auth: auth, KnownHostsPath: req.KnownHosts, Timeout: 15 * time.Second,
	})
	if err != nil {
		response.Error(w, http.StatusBadGateway, "ssh_connection_failed", err.Error())
		return
	}
	defer client.Close()

	result, err := h.service.Discover(ctx, client)
	if err != nil {
		response.Error(w, http.StatusBadGateway, "discovery_failed", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, result)
}
