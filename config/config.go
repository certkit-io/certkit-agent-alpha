package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/certkit-io/certkit-agent-alpha/auth"
	"github.com/certkit-io/certkit-agent-alpha/utils"
)

var CurrentConfig Config

type Config struct {
	ApiBase      string          `json:"api_base"`
	Bootstrap    *BootstrapCreds `json:"bootstrap,omitempty"`
	Agent        *AgentCreds     `json:"agent,omitempty"`
	DesiredState json.RawMessage `json:"desired_state,omitempty"`
	Auth         *AuthCreds      `json:"auth,omitempty"`
	Version      VersionInfo     `json:"omit"`
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

type AuthCreds struct {
	KeyPair *auth.KeyPair `json:"key_pair"`
}

type VersionInfo struct {
	Version string
	Commit  string
	Date    string
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
		ApiBase: apiBase,
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

	return SaveConfig(cfg, path)
}

func SaveConfig(cfg *Config, path string) error {
	configBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	configBytes = append(configBytes, '\n')

	return utils.WriteFileAtomic(path, configBytes, 0o600)
}

func LoadConfig(path string, version VersionInfo) (Config, error) {
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

	if !hasKeyPair(&cfg) {
		log.Print("Generating new keypair...")
		keyPair, _ := auth.CreateNewKeyPair()
		cfg.Auth = &AuthCreds{
			KeyPair: keyPair,
		}
		SaveConfig(&cfg, path)
	}

	cfg.Version = version

	CurrentConfig = cfg

	return cfg, nil
}

func hasKeyPair(cfg *Config) bool {
	if cfg == nil {
		return false
	}
	if cfg.Auth == nil {
		return false
	}
	if cfg.Auth.KeyPair == nil {
		return false
	}
	if cfg.Auth.KeyPair.PublicKey == "" || cfg.Auth.KeyPair.PrivateKey == "" {
		return false
	}
	return true
}
