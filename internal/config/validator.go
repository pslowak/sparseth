package config

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"sparseth/log"
	"strconv"
	"strings"
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
func (v *validator) validate(raw *rawConfig) error {
	for idx, acc := range raw.Accounts {
		v.log.Debug("validate account", "address", acc.Address, "index", idx)
		if err := v.validateAccount(acc); err != nil {
			return fmt.Errorf("failed to validate account at index %d: %w", idx, err)
		}
	}
	return nil
}

// validateAccount validates a single account config.
func (v *validator) validateAccount(acc *account) error {
	if acc.Address == "" {
		v.log.Error("address must not be empty")
		return fmt.Errorf("address is empty")
	}

	if !common.IsHexAddress(acc.Address) {
		v.log.Error("address must be a valid hex address", "address", acc.Address)
		return fmt.Errorf("invalid address: %s", acc.Address)
	}

	if acc.HeadSlot != "" {
		if err := isValidHexUint(acc.HeadSlot); err != nil {
			v.log.Error("head slot must be a valid hex uint", "headSlot", acc.HeadSlot)
			return fmt.Errorf("invalid head slot: %w", err)
		}
	}

	if (acc.ABI == empty && acc.HeadSlot != empty) || (acc.ABI != empty && acc.HeadSlot == empty) {
		v.log.Error("both ABI and head slot must be specified for event monitoring")
		return fmt.Errorf("invalid event config for account %s: both ABI and head slot must be specified", acc.Address)
	}

	if acc.CountSlot != "" {
		if err := isValidHexUint(acc.CountSlot); err != nil {
			v.log.Error("count slot must be a valid hex uint", "countSlot", acc.CountSlot)
			return fmt.Errorf("invalid count slot: %w", err)
		}
	}

	return nil
}

// isValidHexUint checks if the given string
// represents a valid hexadecimal unsigned integer.
func isValidHexUint(s string) error {
	trimmed := strings.TrimPrefix(s, "0x")
	if _, err := strconv.ParseUint(trimmed, 16, 64); err != nil {
		return fmt.Errorf("invalid hex number: %s", s)
	}
	return nil
}
