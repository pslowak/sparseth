package config

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"sparseth/internal/log"
	"strconv"
)

// Validator validates monitoring configs.
type Validator struct {
	log log.Logger
}

// NewValidator creates a new Validator
// with the specified logger.
func NewValidator(log log.Logger) *Validator {
	return &Validator{
		log: log.With("component", "config-validator"),
	}
}

// Validate validates the raw config.
func (v *Validator) Validate(raw *config) error {
	for idx, acc := range raw.Accounts {
		if err := v.validateAccount(acc); err != nil {
			return fmt.Errorf("failed to validate account at index %d: %w", idx, err)
		}
	}
	return nil
}

// validateAccount validates a single account config.
func (v *Validator) validateAccount(acc *account) error {
	if acc.Address == "" {
		return fmt.Errorf("address is empty")
	}

	if !common.IsHexAddress(acc.Address) {
		return fmt.Errorf("invalid address: %s", acc.Address)
	}

	if acc.ABI == "" {
		if acc.Storage != "" || acc.HeadSlot != "" || acc.CountSlot != "" {
			return fmt.Errorf("ABI must be specified for contract accounts")
		}
	}

	if acc.HeadSlot != "" {
		if _, err := strconv.ParseUint(acc.HeadSlot, 10, 64); err != nil {
			return fmt.Errorf("invalid head slot: %w", err)
		}
	}

	return nil
}
