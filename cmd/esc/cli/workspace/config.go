// Copyright 2023, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package workspace

import (
	"errors"
	"io/fs"
)

func (w *Workspace) SetBackendConfigDefaultOrg(backendURL, defaultOrg string) error {
	return w.pulumi.SetBackendConfigDefaultOrg(backendURL, defaultOrg)
}

// Returns the default org as configured in the backend, returning an empty string if unset.
func (w *Workspace) GetBackendConfigDefaultOrg(backendURL, username string) (string, error) {
	config, err := w.pulumi.GetPulumiConfig()
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	if cfg, ok := config.BackendConfig[backendURL]; ok && cfg.DefaultOrg != "" {
		return cfg.DefaultOrg, nil
	}
	return "", nil
}
