package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/certkit-io/certkit-agent-alpha/config"
)

type InstallRequest struct {
	PublicKey string `json:"public_key"`
	Hostname  string `json:"hostname"`
	Version   string `json:"version"`
}

type InstallResponse struct {
	AgentId string `json:"agent_id"`
}

func InstallAgent() (*InstallResponse, error) {

	hostname, _ := os.Hostname()
	payload := InstallRequest{
		PublicKey: config.CurrentConfig.Auth.KeyPair.PublicKey,
		Hostname:  hostname,
		Version:   config.CurrentConfig.Version.Version,
	}

	// Marshal payload to JSON
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}

	// Build request with raw bytes
	req, err := http.NewRequest(
		http.MethodPost,
		config.CurrentConfig.ApiBase+"/api/agent/v1/register-agent",
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	// Required for JSON
	req.Header.Set("Content-Type", "application/json")

	// (Optional) Set a timeout at the client level
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	//privKey, _ := auth.DecodePrivateKey(config.CurrentConfig.Auth.KeyPair.PrivateKey)

	// err = auth.SignRequest(req, "Eric", privKey, time.Now())
	// if err != nil {
	// 	panic(err)
	// }

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("install failed: status=%d body=%s", resp.StatusCode, body)
	}

	var installResp InstallResponse
	if err := json.Unmarshal(body, &installResp); err != nil {
		return nil, fmt.Errorf("decode install response: %w", err)
	}

	return &installResp, nil
}
