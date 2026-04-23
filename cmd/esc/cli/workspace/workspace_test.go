// Copyright 2026, Pulumi Corporation.
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
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"testing"
	"testing/fstest"

	pulumi_workspace "github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testFS struct {
	fstest.MapFS
}

func (tfs testFS) MkdirAll(name string, perm fs.FileMode) error {
	name = strings.TrimPrefix(name, "/")
	if name == "" || name == "." {
		return nil
	}
	dir := path.Dir(name)
	if dir != "." && dir != name {
		if err := tfs.MkdirAll(dir, perm); err != nil {
			return err
		}
	}
	tfs.MapFS[name] = &fstest.MapFile{
		Mode: perm | fs.ModeDir,
	}
	return nil
}

func (tfs testFS) LockedRead(name string) ([]byte, error) {
	name = strings.TrimPrefix(name, "/")
	return tfs.ReadFile(name)
}

func (tfs testFS) LockedWrite(name string, content io.Reader, perm os.FileMode) error {
	name = strings.TrimPrefix(name, "/")
	data, err := io.ReadAll(content)
	if err != nil {
		return err
	}
	tfs.MapFS[name] = &fstest.MapFile{
		Data: data,
		Mode: perm,
	}
	return nil
}

type testPulumiWorkspace struct {
	credentials pulumi_workspace.Credentials
	config      pulumi_workspace.PulumiConfig
}

func (w *testPulumiWorkspace) DeleteAccount(backendURL string) error {
	delete(w.credentials.Accounts, backendURL)
	if w.credentials.Current == backendURL {
		w.credentials.Current = ""
	}
	return nil
}

func (w *testPulumiWorkspace) DeleteAllAccounts() error {
	w.credentials.Accounts = map[string]pulumi_workspace.Account{}
	w.credentials.Current = ""
	return nil
}

func (w *testPulumiWorkspace) SetBackendConfigDefaultOrg(backendURL, defaultOrg string) error {
	if w.config.BackendConfig == nil {
		w.config.BackendConfig = map[string]pulumi_workspace.BackendConfig{}
	}
	w.config.BackendConfig[backendURL] = pulumi_workspace.BackendConfig{DefaultOrg: defaultOrg}
	return nil
}

func (w *testPulumiWorkspace) GetPulumiConfig() (pulumi_workspace.PulumiConfig, error) {
	return w.config, nil
}

func (w *testPulumiWorkspace) GetPulumiPath(elem ...string) (string, error) {
	return path.Join(append([]string{"/pulumi"}, elem...)...), nil
}

func (w *testPulumiWorkspace) GetStoredCredentials() (pulumi_workspace.Credentials, error) {
	return w.credentials, nil
}

func (w *testPulumiWorkspace) StoreAccount(key string, account pulumi_workspace.Account, current bool) error {
	w.credentials.Accounts[key] = account
	if current {
		w.credentials.Current = key
	}
	return nil
}

func (w *testPulumiWorkspace) GetAccount(key string) (pulumi_workspace.Account, error) {
	return w.credentials.Accounts[key], nil
}

func (w *testPulumiWorkspace) NewAuthContextForTokenExchange(
	organization, team, user, token, expirationDuration string,
) (pulumi_workspace.AuthContext, error) {
	return pulumi_workspace.AuthContext{}, nil
}

func newTestWorkspace(tfs testFS, pw *testPulumiWorkspace) *Workspace {
	return New(tfs, pw)
}

func TestGetAccount(t *testing.T) {
	t.Run("returns account with default org from config", func(t *testing.T) {
		pw := &testPulumiWorkspace{
			credentials: pulumi_workspace.Credentials{
				Accounts: map[string]pulumi_workspace.Account{
					"https://api.pulumi.com": {AccessToken: "token-123", Username: "user1"},
				},
			},
			config: pulumi_workspace.PulumiConfig{
				BackendConfig: map[string]pulumi_workspace.BackendConfig{
					"https://api.pulumi.com": {DefaultOrg: "my-org"},
				},
			},
		}
		tfs := testFS{MapFS: fstest.MapFS{}}
		w := newTestWorkspace(tfs, pw)

		acct, err := w.GetAccount("https://api.pulumi.com")
		require.NoError(t, err)
		assert.Equal(t, "token-123", acct.AccessToken)
		assert.Equal(t, "user1", acct.Username)
		assert.Equal(t, "https://api.pulumi.com", acct.BackendURL)
		assert.Equal(t, "my-org", acct.DefaultOrg)
	})

	t.Run("returns empty default org when config has no entry", func(t *testing.T) {
		pw := &testPulumiWorkspace{
			credentials: pulumi_workspace.Credentials{
				Accounts: map[string]pulumi_workspace.Account{
					"https://api.pulumi.com": {AccessToken: "token-123"},
				},
			},
			config: pulumi_workspace.PulumiConfig{},
		}
		tfs := testFS{MapFS: fstest.MapFS{}}
		w := newTestWorkspace(tfs, pw)

		acct, err := w.GetAccount("https://api.pulumi.com")
		require.NoError(t, err)
		assert.Equal(t, "", acct.DefaultOrg)
	})
}

func TestGetCurrentCloudURL(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		account  *Account
		expected string
	}{
		{
			name:     "returns default URL when no env var and no account",
			expected: "https://api.pulumi.com",
		},
		{
			name:     "returns account backend URL when no env var",
			account:  &Account{BackendURL: "https://custom.backend.com"},
			expected: "https://custom.backend.com",
		},
		{
			name:     "env var overrides account backend URL",
			envVar:   "https://env.backend.com",
			account:  &Account{BackendURL: "https://custom.backend.com"},
			expected: "https://env.backend.com",
		},
		{
			name:     "env var overrides default URL",
			envVar:   "https://env.backend.com",
			expected: "https://env.backend.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				t.Setenv(PulumiBackendURLEnvVar, tt.envVar)
			} else {
				t.Setenv(PulumiBackendURLEnvVar, "")
			}

			pw := &testPulumiWorkspace{}
			tfs := testFS{MapFS: fstest.MapFS{}}
			w := newTestWorkspace(tfs, pw)

			result := w.GetCurrentCloudURL(tt.account)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCurrentAccount(t *testing.T) {
	t.Run("returns nil when no accounts exist", func(t *testing.T) {
		pw := &testPulumiWorkspace{
			credentials: pulumi_workspace.Credentials{},
		}
		tfs := testFS{MapFS: fstest.MapFS{}}
		w := newTestWorkspace(tfs, pw)

		acct, shared, err := w.GetCurrentAccount(false)
		require.NoError(t, err)
		assert.Nil(t, acct)
		assert.True(t, shared)
	})

	t.Run("falls back to pulumi current when no esc account set", func(t *testing.T) {
		pw := &testPulumiWorkspace{
			credentials: pulumi_workspace.Credentials{
				Current: "https://api.pulumi.com",
				Accounts: map[string]pulumi_workspace.Account{
					"https://api.pulumi.com": {AccessToken: "pulumi-token", Username: "pulumi-user"},
				},
			},
			config: pulumi_workspace.PulumiConfig{},
		}
		tfs := testFS{MapFS: fstest.MapFS{}}
		w := newTestWorkspace(tfs, pw)

		acct, shared, err := w.GetCurrentAccount(false)
		require.NoError(t, err)
		assert.True(t, shared)
		assert.Equal(t, "pulumi-token", acct.AccessToken)
		assert.Equal(t, "https://api.pulumi.com", acct.BackendURL)
	})

	t.Run("uses esc-specific account when set", func(t *testing.T) {
		pw := &testPulumiWorkspace{
			credentials: pulumi_workspace.Credentials{
				Current: "https://api.pulumi.com",
				Accounts: map[string]pulumi_workspace.Account{
					"https://api.pulumi.com":     {AccessToken: "pulumi-token"},
					"https://custom.backend.com": {AccessToken: "custom-token"},
				},
			},
			config: pulumi_workspace.PulumiConfig{},
		}
		tfs := testFS{MapFS: fstest.MapFS{}}
		w := newTestWorkspace(tfs, pw)

		// Write esc credentials pointing to the custom backend.
		credsPath := "/pulumi/.esc/credentials.json" //nolint:gosec
		escCreds := Credentials{Current: "https://custom.backend.com"}
		data, err := json.Marshal(escCreds)
		require.NoError(t, err)
		tfs.MapFS[strings.TrimPrefix(credsPath, "/")] = &fstest.MapFile{Data: data}

		acct, shared, err := w.GetCurrentAccount(false)
		require.NoError(t, err)
		assert.False(t, shared)
		assert.Equal(t, "custom-token", acct.AccessToken)
		assert.Equal(t, "https://custom.backend.com", acct.BackendURL)
	})

	t.Run("shared=true forces use of pulumi account", func(t *testing.T) {
		pw := &testPulumiWorkspace{
			credentials: pulumi_workspace.Credentials{
				Current: "https://api.pulumi.com",
				Accounts: map[string]pulumi_workspace.Account{
					"https://api.pulumi.com":     {AccessToken: "pulumi-token"},
					"https://custom.backend.com": {AccessToken: "custom-token"},
				},
			},
			config: pulumi_workspace.PulumiConfig{},
		}
		tfs := testFS{MapFS: fstest.MapFS{}}
		w := newTestWorkspace(tfs, pw)

		// Write esc credentials pointing to the custom backend.
		credsPath := "/pulumi/.esc/credentials.json" //nolint:gosec
		escCreds := Credentials{Current: "https://custom.backend.com"}
		data, err := json.Marshal(escCreds)
		require.NoError(t, err)
		tfs.MapFS[strings.TrimPrefix(credsPath, "/")] = &fstest.MapFile{Data: data}

		acct, shared, err := w.GetCurrentAccount(true)
		require.NoError(t, err)
		assert.True(t, shared)
		assert.Equal(t, "pulumi-token", acct.AccessToken)
		assert.Equal(t, "https://api.pulumi.com", acct.BackendURL)
	})

	t.Run("returns error when account not found", func(t *testing.T) {
		pw := &testPulumiWorkspace{
			credentials: pulumi_workspace.Credentials{
				Accounts: map[string]pulumi_workspace.Account{},
			},
			config: pulumi_workspace.PulumiConfig{},
		}
		tfs := testFS{MapFS: fstest.MapFS{}}
		w := newTestWorkspace(tfs, pw)

		// Write esc credentials pointing to a non-existent backend.
		credsPath := "/pulumi/.esc/credentials.json" //nolint:gosec
		escCreds := Credentials{Current: "https://missing.backend.com"}
		data, err := json.Marshal(escCreds)
		require.NoError(t, err)
		tfs.MapFS[strings.TrimPrefix(credsPath, "/")] = &fstest.MapFile{Data: data}

		_, _, err = w.GetCurrentAccount(false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("populates default org from config", func(t *testing.T) {
		pw := &testPulumiWorkspace{
			credentials: pulumi_workspace.Credentials{
				Current: "https://api.pulumi.com",
				Accounts: map[string]pulumi_workspace.Account{
					"https://api.pulumi.com": {AccessToken: "token-123"},
				},
			},
			config: pulumi_workspace.PulumiConfig{
				BackendConfig: map[string]pulumi_workspace.BackendConfig{
					"https://api.pulumi.com": {DefaultOrg: "my-org"},
				},
			},
		}
		tfs := testFS{MapFS: fstest.MapFS{}}
		w := newTestWorkspace(tfs, pw)

		acct, _, err := w.GetCurrentAccount(false)
		require.NoError(t, err)
		assert.Equal(t, "my-org", acct.DefaultOrg)
	})
}

func TestDeleteAllAccounts(t *testing.T) {
	t.Run("clears esc creds and delegates to pulumi workspace", func(t *testing.T) {
		pw := &testPulumiWorkspace{
			credentials: pulumi_workspace.Credentials{
				Current: "https://api.pulumi.com",
				Accounts: map[string]pulumi_workspace.Account{
					"https://api.pulumi.com": {AccessToken: "token-1"},
					"https://other.com":      {AccessToken: "token-2"},
				},
			},
		}
		tfs := testFS{MapFS: fstest.MapFS{}}
		w := newTestWorkspace(tfs, pw)

		// Write initial esc credentials.
		credsPath := "/pulumi/.esc/credentials.json" //nolint:gosec
		escCreds := Credentials{Current: "https://api.pulumi.com"}
		data, err := json.Marshal(escCreds)
		require.NoError(t, err)
		tfs.MapFS[strings.TrimPrefix(credsPath, "/")] = &fstest.MapFile{Data: data}

		err = w.DeleteAllAccounts()
		require.NoError(t, err)

		// Verify esc creds are cleared.
		rawCreds, err := tfs.LockedRead(credsPath)
		require.NoError(t, err)
		var creds Credentials
		require.NoError(t, json.Unmarshal(rawCreds, &creds))
		assert.Equal(t, "", creds.Current)

		// Verify pulumi accounts are cleared.
		assert.Empty(t, pw.credentials.Accounts)
		assert.Equal(t, "", pw.credentials.Current)
	})
}

func TestDeleteAccount(t *testing.T) {
	t.Run("clears esc current when deleting current account", func(t *testing.T) {
		pw := &testPulumiWorkspace{
			credentials: pulumi_workspace.Credentials{
				Current: "https://api.pulumi.com",
				Accounts: map[string]pulumi_workspace.Account{
					"https://api.pulumi.com": {AccessToken: "token-1"},
					"https://other.com":      {AccessToken: "token-2"},
				},
			},
		}
		tfs := testFS{MapFS: fstest.MapFS{}}
		w := newTestWorkspace(tfs, pw)

		// Set esc current to the account we'll delete.
		credsPath := "/pulumi/.esc/credentials.json" //nolint:gosec
		escCreds := Credentials{Current: "https://api.pulumi.com"}
		data, err := json.Marshal(escCreds)
		require.NoError(t, err)
		tfs.MapFS[strings.TrimPrefix(credsPath, "/")] = &fstest.MapFile{Data: data}

		err = w.DeleteAccount("https://api.pulumi.com")
		require.NoError(t, err)

		// Verify esc current is cleared.
		rawCreds, err := tfs.LockedRead(credsPath)
		require.NoError(t, err)
		var creds Credentials
		require.NoError(t, json.Unmarshal(rawCreds, &creds))
		assert.Equal(t, "", creds.Current)

		// Verify the account was deleted from pulumi workspace.
		_, ok := pw.credentials.Accounts["https://api.pulumi.com"]
		assert.False(t, ok)

		// Verify the other account is untouched.
		_, ok = pw.credentials.Accounts["https://other.com"]
		assert.True(t, ok)
	})

	t.Run("does not clear esc current when deleting a different account", func(t *testing.T) {
		pw := &testPulumiWorkspace{
			credentials: pulumi_workspace.Credentials{
				Current: "https://api.pulumi.com",
				Accounts: map[string]pulumi_workspace.Account{
					"https://api.pulumi.com": {AccessToken: "token-1"},
					"https://other.com":      {AccessToken: "token-2"},
				},
			},
		}
		tfs := testFS{MapFS: fstest.MapFS{}}
		w := newTestWorkspace(tfs, pw)

		// Set esc current to a different account.
		credsPath := "/pulumi/.esc/credentials.json" //nolint:gosec
		escCreds := Credentials{Current: "https://api.pulumi.com"}
		data, err := json.Marshal(escCreds)
		require.NoError(t, err)
		tfs.MapFS[strings.TrimPrefix(credsPath, "/")] = &fstest.MapFile{Data: data}

		err = w.DeleteAccount("https://other.com")
		require.NoError(t, err)

		// Verify esc current is still set to the original account.
		rawCreds, err := tfs.LockedRead(credsPath)
		require.NoError(t, err)
		var creds Credentials
		require.NoError(t, json.Unmarshal(rawCreds, &creds))
		assert.Equal(t, "https://api.pulumi.com", creds.Current)
	})

	t.Run("clears esc current when esc has no current but pulumi current matches", func(t *testing.T) {
		pw := &testPulumiWorkspace{
			credentials: pulumi_workspace.Credentials{
				Current: "https://api.pulumi.com",
				Accounts: map[string]pulumi_workspace.Account{
					"https://api.pulumi.com": {AccessToken: "token-1"},
				},
			},
		}
		tfs := testFS{MapFS: fstest.MapFS{}}
		w := newTestWorkspace(tfs, pw)

		err := w.DeleteAccount("https://api.pulumi.com")
		require.NoError(t, err)

		_, ok := pw.credentials.Accounts["https://api.pulumi.com"]
		assert.False(t, ok)
	})
}

func TestSetCurrentAccount(t *testing.T) {
	t.Run("non-shared stores backend URL in esc creds", func(t *testing.T) {
		pw := &testPulumiWorkspace{
			credentials: pulumi_workspace.Credentials{
				Current:  "https://existing.com",
				Accounts: map[string]pulumi_workspace.Account{},
			},
		}
		tfs := testFS{MapFS: fstest.MapFS{}}
		w := newTestWorkspace(tfs, pw)

		acct := Account{
			Account:    pulumi_workspace.Account{AccessToken: "new-token", Username: "new-user"},
			BackendURL: "https://new.backend.com",
		}
		err := w.SetCurrentAccount(acct, false)
		require.NoError(t, err)

		// Verify esc creds have the backend URL.
		credsPath := "/pulumi/.esc/credentials.json" //nolint:gosec
		rawCreds, err := tfs.LockedRead(credsPath)
		require.NoError(t, err)
		var creds Credentials
		require.NoError(t, json.Unmarshal(rawCreds, &creds))
		assert.Equal(t, "https://new.backend.com", creds.Current)

		// Verify account was stored in pulumi workspace.
		stored, ok := pw.credentials.Accounts["https://new.backend.com"]
		assert.True(t, ok)
		assert.Equal(t, "new-token", stored.AccessToken)

		// Verify pulumi current was NOT changed (setCurrent is false because shared=false).
		assert.Equal(t, "https://existing.com", pw.credentials.Current)
	})

	t.Run("shared clears esc current and sets pulumi current when empty", func(t *testing.T) {
		pw := &testPulumiWorkspace{
			credentials: pulumi_workspace.Credentials{
				Current:  "",
				Accounts: map[string]pulumi_workspace.Account{},
			},
		}
		tfs := testFS{MapFS: fstest.MapFS{}}
		w := newTestWorkspace(tfs, pw)

		acct := Account{
			Account:    pulumi_workspace.Account{AccessToken: "shared-token"},
			BackendURL: "https://api.pulumi.com",
		}
		err := w.SetCurrentAccount(acct, true)
		require.NoError(t, err)

		// Verify esc creds have empty current (shared mode).
		credsPath := "/pulumi/.esc/credentials.json" //nolint:gosec
		rawCreds, err := tfs.LockedRead(credsPath)
		require.NoError(t, err)
		var creds Credentials
		require.NoError(t, json.Unmarshal(rawCreds, &creds))
		assert.Equal(t, "", creds.Current)

		// Verify pulumi current was set (because it was empty and shared=true).
		assert.Equal(t, "https://api.pulumi.com", pw.credentials.Current)
	})

	t.Run("shared does not override existing pulumi current", func(t *testing.T) {
		pw := &testPulumiWorkspace{
			credentials: pulumi_workspace.Credentials{
				Current: "https://existing.com",
				Accounts: map[string]pulumi_workspace.Account{
					"https://existing.com": {AccessToken: "existing-token"},
				},
			},
		}
		tfs := testFS{MapFS: fstest.MapFS{}}
		w := newTestWorkspace(tfs, pw)

		acct := Account{
			Account:    pulumi_workspace.Account{AccessToken: "shared-token"},
			BackendURL: "https://api.pulumi.com",
		}
		err := w.SetCurrentAccount(acct, true)
		require.NoError(t, err)

		// Verify pulumi current was NOT changed (already had a value).
		assert.Equal(t, "https://existing.com", pw.credentials.Current)

		// Verify account was still stored.
		_, ok := pw.credentials.Accounts["https://api.pulumi.com"]
		assert.True(t, ok)
	})
}

func TestGetBackendConfigDefaultOrg(t *testing.T) {
	t.Run("returns default org when set", func(t *testing.T) {
		pw := &testPulumiWorkspace{
			config: pulumi_workspace.PulumiConfig{
				BackendConfig: map[string]pulumi_workspace.BackendConfig{
					"https://api.pulumi.com": {DefaultOrg: "my-org"},
				},
			},
		}
		tfs := testFS{MapFS: fstest.MapFS{}}
		w := newTestWorkspace(tfs, pw)

		org, err := w.GetBackendConfigDefaultOrg("https://api.pulumi.com", "user1")
		require.NoError(t, err)
		assert.Equal(t, "my-org", org)
	})

	t.Run("returns empty string when no config exists", func(t *testing.T) {
		pw := &testPulumiWorkspace{
			config: pulumi_workspace.PulumiConfig{},
		}
		tfs := testFS{MapFS: fstest.MapFS{}}
		w := newTestWorkspace(tfs, pw)

		org, err := w.GetBackendConfigDefaultOrg("https://api.pulumi.com", "user1")
		require.NoError(t, err)
		assert.Equal(t, "", org)
	})

	t.Run("returns empty string when backend has no default org", func(t *testing.T) {
		pw := &testPulumiWorkspace{
			config: pulumi_workspace.PulumiConfig{
				BackendConfig: map[string]pulumi_workspace.BackendConfig{
					"https://api.pulumi.com": {},
				},
			},
		}
		tfs := testFS{MapFS: fstest.MapFS{}}
		w := newTestWorkspace(tfs, pw)

		org, err := w.GetBackendConfigDefaultOrg("https://api.pulumi.com", "user1")
		require.NoError(t, err)
		assert.Equal(t, "", org)
	})

	t.Run("returns empty string for unknown backend", func(t *testing.T) {
		pw := &testPulumiWorkspace{
			config: pulumi_workspace.PulumiConfig{
				BackendConfig: map[string]pulumi_workspace.BackendConfig{
					"https://api.pulumi.com": {DefaultOrg: "my-org"},
				},
			},
		}
		tfs := testFS{MapFS: fstest.MapFS{}}
		w := newTestWorkspace(tfs, pw)

		org, err := w.GetBackendConfigDefaultOrg("https://other.com", "user1")
		require.NoError(t, err)
		assert.Equal(t, "", org)
	})
}
