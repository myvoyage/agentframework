// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025 Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package markdown

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"AgentFramework/agent/skills"
)

// SkillValidator interface for validating skill definitions
type SkillValidator interface {
	Validate(def *skills.SkillDefinition) error
}

// EligibilityValidator validates skill eligibility
type EligibilityValidator struct {
	osDetector     OSDetector
	binaryChecker BinaryChecker
	envChecker    EnvChecker
}

// NewEligibilityValidator creates a new EligibilityValidator
func NewEligibilityValidator() *EligibilityValidator {
	return &EligibilityValidator{
		osDetector:     NewOSDetector(),
		binaryChecker: NewBinaryChecker(),
		envChecker:    NewEnvChecker(),
	}
}

// Validate checks if a skill is eligible to be used on the current system
func (v *EligibilityValidator) Validate(def *skills.SkillDefinition) error {
	// Check if skill is disabled (from metadata)
	if disabled, ok := def.Metadata["disabled"].(bool); ok && disabled {
		return fmt.Errorf("skill %s is disabled", def.ID)
	}

	// Check OS compatibility (from metadata)
	if osList, ok := def.Metadata["os"].([]interface{}); ok && len(osList) > 0 {
		var osStrings []string
		for _, os := range osList {
			if osStr, isStr := os.(string); isStr {
				osStrings = append(osStrings, osStr)
			}
		}
		if !v.osDetector.Contains(osStrings) {
			return fmt.Errorf("skill %s not supported on current OS: %s", def.ID, runtime.GOOS)
		}
	}

	// Check required binaries (from metadata)
	if bins, ok := def.Metadata["bins"].([]interface{}); ok {
		for _, bin := range bins {
			if binStr, isStr := bin.(string); isStr && binStr != "" && !v.binaryChecker.Exists(binStr) {
				return fmt.Errorf("required binary not found: %s (for skill %s)", binStr, def.ID)
			}
		}
	}

	// Check required environment variables (from metadata)
	if envVars, ok := def.Metadata["env"].([]interface{}); ok {
		for _, env := range envVars {
			if envStr, isStr := env.(string); isStr && envStr != "" && !v.envChecker.Exists(envStr) {
				return fmt.Errorf("required environment variable not set: %s (for skill %s)", envStr, def.ID)
			}
		}
	}

	// Check prerequisites
	for _, prereq := range def.Prerequisites {
		if prereq.Required {
			switch prereq.Type {
			case "command":
				if prereq.Check != "" {
					cmd := exec.Command("sh", "-c", prereq.Check)
					if err := cmd.Run(); err != nil {
						return fmt.Errorf("prerequisite check failed: %s (for skill %s)", prereq.Description, def.ID)
					}
				}
			case "env_var":
				if prereq.Check != "" && !v.envChecker.Exists(prereq.Check) {
					return fmt.Errorf("prerequisite check failed: %s (for skill %s)", prereq.Description, def.ID)
				}
			case "file_exists":
				if prereq.Check != "" {
					if _, err := os.Stat(prereq.Check); os.IsNotExist(err) {
						return fmt.Errorf("prerequisite check failed: %s (for skill %s)", prereq.Description, def.ID)
					}
				}
			}
		}
	}

	return nil
}

// OSDetector detects the current operating system
type OSDetector interface {
	Contains(osList []string) bool
	GetCurrent() string
}

// DefaultOSDetector implements OSDetector
type DefaultOSDetector struct{}

// NewOSDetector creates a new DefaultOSDetector
func NewOSDetector() OSDetector {
	return &DefaultOSDetector{}
}

// Contains checks if the current OS is in the list
func (d *DefaultOSDetector) Contains(osList []string) bool {
	if len(osList) == 0 {
		return true // No OS restriction means all are supported
	}

	currentOS := d.GetCurrent()
	for _, os := range osList {
		if strings.EqualFold(os, currentOS) {
			return true
		}
	}

	return false
}

// GetCurrent returns the current OS
func (d *DefaultOSDetector) GetCurrent() string {
	return runtime.GOOS
}

// BinaryChecker checks if a binary exists
type BinaryChecker interface {
	Exists(binary string) bool
}

// DefaultBinaryChecker implements BinaryChecker
type DefaultBinaryChecker struct{}

// NewBinaryChecker creates a new DefaultBinaryChecker
func NewBinaryChecker() BinaryChecker {
	return &DefaultBinaryChecker{}
}

// Exists checks if a binary is available in PATH
func (c *DefaultBinaryChecker) Exists(binary string) bool {
	if binary == "" {
		return false
	}

	cmd := exec.Command("which", binary)
	if runtime.GOOS == "windows" {
		cmd = exec.Command("where", binary)
	}

	err := cmd.Run()
	return err == nil
}

// EnvChecker checks if environment variables are set
type EnvChecker interface {
	Exists(env string) bool
	Get(env string) string
}

// DefaultEnvChecker implements EnvChecker
type DefaultEnvChecker struct{}

// NewEnvChecker creates a new DefaultEnvChecker
func NewEnvChecker() EnvChecker {
	return &DefaultEnvChecker{}
}

// Exists checks if an environment variable is set
func (c *DefaultEnvChecker) Exists(env string) bool {
	if env == "" {
		return false
	}

	val := os.Getenv(env)
	return strings.TrimSpace(val) != ""
}

// Get returns the value of an environment variable
func (c *DefaultEnvChecker) Get(env string) string {
	return os.Getenv(env)
}
