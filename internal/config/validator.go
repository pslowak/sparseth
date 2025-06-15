package config

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"sparseth/internal/log"
	"strconv"
)

// validator validates monitoring configs.
type validator struct {
	log log.Logger
}

// newValidator creates a new validator
// with the specified logger.
func newValidator(log log.Logger) *validator {
	return &validator{
		log: log.With("component", "config-validator"),
	}
}

// validate validates the raw config.
func (v *validator) validate(raw *config) error {
	for idx, acc := range raw.Accounts {
		if err := v.validateAccount(acc); err != nil {
			return fmt.Errorf("failed to validate account at index %d: %w", idx, err)
		}
	}
	return nil
}

// validateAccount validates a single account config.
func (v *validator) validateAccount(acc *account) error {
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
