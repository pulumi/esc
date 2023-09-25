// Copyright 2023, Pulumi Corporation.

package workspace

import (
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

func SetBackendConfigDefaultOrg(backendURL, defaultOrg string) error {
	return workspace.SetBackendConfigDefaultOrg(backendURL, defaultOrg)
}

func GetBackendConfigDefaultOrg(backendURL, username string) (string, error) {
	config, err := workspace.GetPulumiConfig()
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if cfg, ok := config.BackendConfig[backendURL]; ok && cfg.DefaultOrg != "" {
		return cfg.DefaultOrg, nil
	}
	return username, nil
}
