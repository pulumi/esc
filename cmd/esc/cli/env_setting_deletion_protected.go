// Copyright 2025, Pulumi Corporation.

package cli

import (
	"fmt"

	"github.com/pulumi/esc/cmd/esc/cli/client"
)

const settingDeletionProtected settingName = "deletion-protected"

type DeletionProtectedSetting struct{}

func (s *DeletionProtectedSetting) KebabName() string {
	return "deletion-protected"
}

func (s *DeletionProtectedSetting) HelpText() string {
	return "Enable or disable deletion protection"
}

func (s *DeletionProtectedSetting) ValidateValue(raw string) (bool, error) {
	if raw != "true" && raw != "false" {
		return false, fmt.Errorf("invalid value for %s: %s (expected true or false)", s.KebabName(), raw)
	}
	return raw == "true", nil
}

func (s *DeletionProtectedSetting) GetValue(settings *client.EnvironmentSettings) bool {
	return settings.DeletionProtected
}

func (s *DeletionProtectedSetting) SetValue(req *client.PatchEnvironmentSettingsRequest, value bool) {
	req.DeletionProtected = &value
}
