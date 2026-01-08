package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/certkit-io/certkit-agent-alpha/utils"
)

var CurrentConfig Config

type Config struct {
	APIBASE      string          `json:"api_base"`
	Bootstrap    *BootstrapCreds `json:"bootstrap,omitempty"`
	Agent        *AgentCreds     `json:"agent,omitempty"`
	DesiredState json.RawMessage `json:"desired_state,omitempty"`
}

type BootstrapCreds struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

type AgentCreds struct {
	AgentID      string `json:"agent_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

const (
	defaultAPIBase = "https://app.certkit.io"
)

func CreateInitialConfig(path string) error {
	access := os.Getenv("ACCESS_KEY")
	secret := os.Getenv("SECRET_KEY")

	if access == "" || secret == "" {
		return fmt.Errorf("ACCESS_KEY and SECRET_KEY are required for first install")
	}

	apiBase := os.Getenv("CERTKIT_API_BASE")
	if apiBase == "" {
		apiBase = defaultAPIBase
	}

	cfg := &Config{
		APIBASE: apiBase,
		Bootstrap: &BootstrapCreds{
			AccessKey: access,
			SecretKey: secret,
		},
		Agent:        nil,
		DesiredState: nil,
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	configBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	configBytes = append(configBytes, '\n')

	return utils.WriteFileAtomic(path, configBytes, 0o600)
}

func LoadConfig(path string) (Config, error) {
	var cfg Config

	if path == "" {
		return cfg, fmt.Errorf("config path is empty")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, fmt.Errorf("config file does not exist: %s", path)
		}
		return cfg, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	if len(bytes.TrimSpace(b)) == 0 {
		return cfg, fmt.Errorf("config file %s is empty", path)
	}

	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	// // Exactly one of Bootstrap or Agent should be present
	// if cfg.Bootstrap == nil && cfg.Agent == nil {
	// 	return cfg, fmt.Errorf(
	// 		"config %s: either bootstrap or agent credentials must be present",
	// 		path,
	// 	)
	// }

	CurrentConfig = cfg

	return cfg, nil
}
