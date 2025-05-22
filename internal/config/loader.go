package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"sparseth/internal/log"
)

// config represents the raw YAML structure
// of the config file.
type config struct {
	Accounts []*account `yaml:"accounts"`
}

// account represents a raw YAML account entry.
type account struct {
	Address   string `yaml:"address"`
	ABI       string `yaml:"abi_path"`
	Storage   string `yaml:"storage_path"`
	HeadSlot  string `yaml:"head_slot"`
	CountSlot string `yaml:"count_slot"`
}

// isEOA checks if the account is an
// EOA or a contract account.
func (a *account) isEOA() bool {
	return a.ABI == ""
}

// Loader reads the main config file.
type Loader struct {
	log       log.Logger
	validator *Validator
	parser    *Parser
}

// NewLoader creates a new config Loader with
// the specified logging context attached.
func NewLoader(log log.Logger) *Loader {
	return &Loader{
		log:       log.With("component", "config-loader"),
		validator: NewValidator(log),
		parser:    NewParser(log),
	}
}

// Load reads the config file at the specified path.
func (l *Loader) Load(path string) (*AppConfig, error) {
	l.log.Info("load config from file", "path", path)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var raw *config
	if err = yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err = l.validator.Validate(raw); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return l.parser.Parse(raw)
}
