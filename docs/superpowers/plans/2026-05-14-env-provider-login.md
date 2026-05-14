# `esc env provider` Login Providers Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an `esc env provider` command that lets users add AWS, Azure, or GCP login-provider blocks (static credentials only) to an existing environment, in one PR.

**Architecture:** A new `provider` subcommand under `esc env` with three sibling subcommands — `aws-login`, `azure-login`, `gcp-login` — each accepting the required identifiers as positional args (per CLI Command Guidelines from epic #22959) and optional fields as flags. Each subcommand fetches the current environment definition, merges the new provider block into a configurable YAML path under `values`, wraps secret-bearing fields with `fn::secret`, and writes the result back via the existing `esc.updateEnvironment` helper. OIDC is intentionally out of scope; the command structure leaves room to add it later.

**Tech Stack:** Go, Cobra, `gopkg.in/yaml.v3`, existing `syntax/encoding.YAMLSyntax` AST helper, `github.com/pulumi/esc/cmd/esc/cli/client` API client. Tests use the existing CLI golden-file harness (`testdata/*.yaml`, run with `PULUMI_ACCEPT=true` to regenerate).

**Source references**
- Provider schemas confirmed against `schema/testdata/esc.json:348-637` (aws-login, azure-login, gcp-login `$defs`).
- Existing example AWS YAML: `eval/testdata/crypt/open.plaintext.yaml`.
- Existing patterns followed:
  - `cmd/esc/cli/env_set.go` — env fetch/modify/write pipeline (`GetEnvironment` → mutate `yaml.Node` → `updateEnvironment`).
  - `cmd/esc/cli/env_settings.go` + `env_settings_registry.go` — parent-with-subcommands pattern with a registry. **We are not using the typed-registry shim** for providers; a small per-provider file is simpler.
  - `cmd/esc/cli/cli_test.go` — golden-file driver (look for `accept()` / `PULUMI_ACCEPT`).

---

## File Structure

We create one parent file plus one file per provider. Each provider file is self-contained: cobra flags, YAML builder, and unit test in a sibling `_test.go`.

- `cmd/esc/cli/env_provider.go` — new parent `provider` command; wires the three subcommands.
- `cmd/esc/cli/env_provider_common.go` — shared helpers: `mergeProviderIntoEnv` (read env, merge provider node at `--path`, write back), `secretNode(value)` (wraps a scalar in `fn::secret`).
- `cmd/esc/cli/env_provider_aws.go` — `aws-login` subcommand, builds the `fn::open::aws-login` YAML block.
- `cmd/esc/cli/env_provider_azure.go` — `azure-login` subcommand.
- `cmd/esc/cli/env_provider_gcp.go` — `gcp-login` subcommand.
- `cmd/esc/cli/env_provider_common_test.go` — unit tests for the shared YAML builder and secret wrapping.
- `cmd/esc/cli/env_provider_aws_test.go` — unit tests for the AWS YAML builder.
- `cmd/esc/cli/env_provider_azure_test.go` — unit tests for the Azure YAML builder.
- `cmd/esc/cli/env_provider_gcp_test.go` — unit tests for the GCP YAML builder.
- `cmd/esc/cli/env.go:74` — register `newEnvProviderCmd(env)` next to the other `cmd.AddCommand` calls.
- `cmd/esc/cli/testdata/env-provider-aws-login.yaml` — golden-file end-to-end test.
- `cmd/esc/cli/testdata/env-provider-azure-login.yaml` — golden-file end-to-end test.
- `cmd/esc/cli/testdata/env-provider-gcp-login.yaml` — golden-file end-to-end test.
- `CHANGELOG_PENDING.md` — feature entry.

---

## Command surface (locked spec — every task assumes this)

```
esc env provider aws-login   [<org>/][<project>/]<env-name> <access-key-id> <secret-access-key> [--session-token TOKEN]   [--path PATH]
esc env provider azure-login [<org>/][<project>/]<env-name> <client-id> <tenant-id> <subscription-id> [--client-secret SECRET] [--path PATH]
esc env provider gcp-login   [<org>/][<project>/]<env-name> <project-number> <access-token>          [--service-account SA] [--token-lifetime LIFETIME] [--path PATH]
```

Flags shared by all three subcommands:
- `--path PATH` — Pulumi property path under `values` where the provider block is written. Defaults: `aws.login`, `azure.login`, `gcp.login`. Parsed with `resource.ParsePropertyPath`.
- `--draft[=<change-request-id>]` — identical semantics to `env set` (mirror its definition verbatim).

Provider-specific rules:
- **aws-login** — `accessKeyId` and `secretAccessKey` are required positionals. `secretAccessKey` and `sessionToken` (when set) are wrapped in `fn::secret`.
- **azure-login** — `clientId`, `tenantId`, `subscriptionId` are required positionals. `clientSecret` is optional via `--client-secret`; when set it is wrapped in `fn::secret`.
- **gcp-login** — `project` (required positional) is parsed as a positive integer and emitted as a YAML integer scalar. `accessToken` (required positional) is wrapped in `fn::secret` under `accessToken.accessToken`.

Output YAML shapes — these are normative; tests assert on them:

```yaml
# aws-login at --path aws.login
values:
  aws:
    login:
      fn::open::aws-login:
        static:
          accessKeyId: AKIA...
          secretAccessKey:
            fn::secret: ...
          sessionToken:           # only when --session-token is provided
            fn::secret: ...
```

```yaml
# azure-login at --path azure.login
values:
  azure:
    login:
      fn::open::azure-login:
        clientId: 00000000-...
        tenantId: 00000000-...
        subscriptionId: /subscriptions/...
        clientSecret:             # only when --client-secret is provided
          fn::secret: ...
```

```yaml
# gcp-login at --path gcp.login
values:
  gcp:
    login:
      fn::open::gcp-login:
        project: 123456789
        accessToken:
          accessToken:
            fn::secret: ...
          serviceAccount: sa@proj.iam.gserviceaccount.com   # optional
          tokenLifetime: 1h                                  # optional
```

Merge semantics: if the target path already contains a node, the new provider block **replaces** that node entirely. We do not attempt a deep-merge — replacement is the predictable behavior and is consistent with `esc env set`.

---

## Task 1: Wire up the parent `provider` command

**Files:**
- Create: `cmd/esc/cli/env_provider.go`
- Modify: `cmd/esc/cli/env.go` (add `cmd.AddCommand(newEnvProviderCmd(env))` after the existing `newEnvSettingsCmd(env)` line at `env.go:74`)

- [ ] **Step 1: Add the parent command file**

Create `cmd/esc/cli/env_provider.go`:

```go
// Copyright 2026, Pulumi Corporation.

package cli

import (
	"github.com/spf13/cobra"
)

func newEnvProviderCmd(env *envCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage login providers within an environment",
		Long: "Manage login providers within an environment\n" +
			"\n" +
			"Subcommands add cloud-provider login blocks (AWS, Azure, GCP) to an environment.\n" +
			"Only static credentials are supported today; OIDC support will be added later.\n",
		Args: cobra.NoArgs,
	}

	cmd.AddCommand(newEnvProviderAWSLoginCmd(env))
	cmd.AddCommand(newEnvProviderAzureLoginCmd(env))
	cmd.AddCommand(newEnvProviderGCPLoginCmd(env))

	return cmd
}
```

- [ ] **Step 2: Register the command in `env.go`**

In `cmd/esc/cli/env.go`, after the line `cmd.AddCommand(newEnvSettingsCmd(env))` (currently at line 74), add:

```go
	cmd.AddCommand(newEnvProviderCmd(env))
```

- [ ] **Step 3: Make sure it compiles (the three subcommand constructors don't exist yet — we'll stub them as part of Task 2 before this compiles cleanly)**

Run: `go build ./...`
Expected: build fails with three "undefined: newEnvProviderAWSLoginCmd / AzureLoginCmd / GCPLoginCmd" errors. This is intentional — Task 2 introduces the shared helper and Task 3 stubs the first subcommand.

- [ ] **Step 4: Commit**

```bash
git add cmd/esc/cli/env_provider.go cmd/esc/cli/env.go
git commit -m "esc env provider: scaffold parent command"
```

---

## Task 2: Shared helper `mergeProviderIntoEnv` (TDD)

**Files:**
- Create: `cmd/esc/cli/env_provider_common.go`
- Create: `cmd/esc/cli/env_provider_common_test.go`

The helper takes the current environment YAML bytes, a property path (e.g. `aws.login`), and a `*yaml.Node` representing the provider block, and returns the new YAML bytes with the node merged in under `values`. It must:
1. Round-trip an empty document into a `values:`-rooted document.
2. Replace any existing node at the path.
3. Preserve other keys at `values` and at intermediate path segments.

- [ ] **Step 1: Write failing tests**

Create `cmd/esc/cli/env_provider_common_test.go`:

```go
// Copyright 2026, Pulumi Corporation.

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func mustNode(t *testing.T, src string) *yaml.Node {
	t.Helper()
	var n yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(src), &n))
	require.Equal(t, yaml.DocumentNode, n.Kind)
	require.Len(t, n.Content, 1)
	return n.Content[0]
}

func TestMergeProviderIntoEnv_EmptyDoc(t *testing.T) {
	provider := mustNode(t, `fn::open::aws-login:
  static:
    accessKeyId: a
    secretAccessKey:
      fn::secret: s
`)

	out, err := mergeProviderIntoEnv(nil, []any{"aws", "login"}, provider)
	require.NoError(t, err)
	assert.YAMLEq(t, `values:
  aws:
    login:
      fn::open::aws-login:
        static:
          accessKeyId: a
          secretAccessKey:
            fn::secret: s
`, string(out))
}

func TestMergeProviderIntoEnv_ReplacesExisting(t *testing.T) {
	current := []byte(`values:
  aws:
    login:
      fn::open::aws-login:
        static:
          accessKeyId: old
          secretAccessKey:
            fn::secret: old
  other: keep-me
`)
	provider := mustNode(t, `fn::open::aws-login:
  static:
    accessKeyId: new
    secretAccessKey:
      fn::secret: new
`)

	out, err := mergeProviderIntoEnv(current, []any{"aws", "login"}, provider)
	require.NoError(t, err)
	assert.YAMLEq(t, `values:
  aws:
    login:
      fn::open::aws-login:
        static:
          accessKeyId: new
          secretAccessKey:
            fn::secret: new
  other: keep-me
`, string(out))
}

func TestMergeProviderIntoEnv_PreservesSiblings(t *testing.T) {
	current := []byte(`values:
  unrelated:
    foo: bar
imports:
  - default/base
`)
	provider := mustNode(t, `fn::open::gcp-login:
  project: 1
  accessToken:
    accessToken:
      fn::secret: t
`)

	out, err := mergeProviderIntoEnv(current, []any{"gcp", "login"}, provider)
	require.NoError(t, err)
	assert.YAMLEq(t, `values:
  unrelated:
    foo: bar
  gcp:
    login:
      fn::open::gcp-login:
        project: 1
        accessToken:
          accessToken:
            fn::secret: t
imports:
  - default/base
`, string(out))
}

func TestSecretNode_WrapsScalar(t *testing.T) {
	n := secretNode("hunter2")
	require.Equal(t, yaml.MappingNode, n.Kind)
	require.Len(t, n.Content, 2)
	assert.Equal(t, "fn::secret", n.Content[0].Value)
	assert.Equal(t, "hunter2", n.Content[1].Value)
	assert.Equal(t, "!!str", n.Content[1].Tag)
}
```

- [ ] **Step 2: Run the test to verify failure**

Run: `go test ./cmd/esc/cli/ -run 'TestMergeProviderIntoEnv|TestSecretNode' -count=1`
Expected: FAIL — `undefined: mergeProviderIntoEnv` and `undefined: secretNode`.

- [ ] **Step 3: Implement the helper**

Create `cmd/esc/cli/env_provider_common.go`:

```go
// Copyright 2026, Pulumi Corporation.

package cli

import (
	"fmt"

	"github.com/pulumi/esc/syntax/encoding"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"gopkg.in/yaml.v3"
)

// mergeProviderIntoEnv merges providerNode into the YAML environment definition
// at values.<path>, replacing any existing node at that path. The result is the
// new YAML document bytes.
func mergeProviderIntoEnv(envYAML []byte, path resource.PropertyPath, providerNode *yaml.Node) ([]byte, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("path must contain at least one element")
	}

	var docNode yaml.Node
	if len(envYAML) > 0 {
		if err := yaml.Unmarshal(envYAML, &docNode); err != nil {
			return nil, fmt.Errorf("unmarshaling environment definition: %w", err)
		}
	}
	if docNode.Kind != yaml.DocumentNode {
		docNode = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{}}}
	}

	valuesNode, ok := encoding.YAMLSyntax{Node: &docNode}.Get(resource.PropertyPath{"values"})
	if !ok {
		var err error
		valuesNode, err = encoding.YAMLSyntax{Node: &docNode}.Set(nil, resource.PropertyPath{"values"}, yaml.Node{
			Kind: yaml.MappingNode,
		})
		if err != nil {
			return nil, fmt.Errorf("creating values node: %w", err)
		}
	}

	if _, err := (encoding.YAMLSyntax{Node: valuesNode}).Set(nil, path, *providerNode); err != nil {
		return nil, fmt.Errorf("setting provider at %v: %w", path, err)
	}

	out, err := yaml.Marshal(docNode.Content[0])
	if err != nil {
		return nil, fmt.Errorf("marshaling definition: %w", err)
	}
	return out, nil
}

// secretNode returns a yaml mapping node of the shape `fn::secret: <value>`.
// The value is always emitted as a string scalar (tag !!str), so callers do
// not have to worry about YAML coercing tokens like "true" or "12345" into
// booleans/numbers.
func secretNode(value string) *yaml.Node {
	return &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: "fn::secret"},
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: value},
		},
	}
}
```

- [ ] **Step 4: Run the test to verify pass**

Run: `go test ./cmd/esc/cli/ -run 'TestMergeProviderIntoEnv|TestSecretNode' -count=1 -v`
Expected: PASS for all four tests.

- [ ] **Step 5: Commit**

```bash
git add cmd/esc/cli/env_provider_common.go cmd/esc/cli/env_provider_common_test.go
git commit -m "esc env provider: shared YAML merge helper"
```

---

## Task 3: `aws-login` subcommand (TDD on YAML builder, then wire to cobra)

**Files:**
- Create: `cmd/esc/cli/env_provider_aws.go`
- Create: `cmd/esc/cli/env_provider_aws_test.go`

We split the provider into a pure builder (`buildAWSLoginNode`) plus a thin cobra wrapper. The builder is unit-testable without standing up the CLI harness.

- [ ] **Step 1: Write failing test for the builder**

Create `cmd/esc/cli/env_provider_aws_test.go`:

```go
// Copyright 2026, Pulumi Corporation.

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestBuildAWSLoginNode_Required(t *testing.T) {
	node := buildAWSLoginNode("AKIAEXAMPLE", "shhh", "")
	out, err := yaml.Marshal(node)
	require.NoError(t, err)
	assert.YAMLEq(t, `fn::open::aws-login:
  static:
    accessKeyId: AKIAEXAMPLE
    secretAccessKey:
      fn::secret: shhh
`, string(out))
}

func TestBuildAWSLoginNode_WithSessionToken(t *testing.T) {
	node := buildAWSLoginNode("AKIAEXAMPLE", "shhh", "tok")
	out, err := yaml.Marshal(node)
	require.NoError(t, err)
	assert.YAMLEq(t, `fn::open::aws-login:
  static:
    accessKeyId: AKIAEXAMPLE
    secretAccessKey:
      fn::secret: shhh
    sessionToken:
      fn::secret: tok
`, string(out))
}
```

- [ ] **Step 2: Run the test to verify failure**

Run: `go test ./cmd/esc/cli/ -run TestBuildAWSLoginNode -count=1`
Expected: FAIL — `undefined: buildAWSLoginNode`.

- [ ] **Step 3: Implement the builder + cobra command**

Create `cmd/esc/cli/env_provider_aws.go`:

```go
// Copyright 2026, Pulumi Corporation.

package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

// buildAWSLoginNode returns a yaml.Node representing
// `fn::open::aws-login: { static: {...} }`. secretAccessKey and sessionToken
// are wrapped in `fn::secret`. sessionToken is omitted when empty.
func buildAWSLoginNode(accessKeyID, secretAccessKey, sessionToken string) *yaml.Node {
	staticContent := []*yaml.Node{
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "accessKeyId"},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: accessKeyID},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "secretAccessKey"},
		secretNode(secretAccessKey),
	}
	if sessionToken != "" {
		staticContent = append(staticContent,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "sessionToken"},
			secretNode(sessionToken),
		)
	}

	return &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: "fn::open::aws-login"},
			{
				Kind: yaml.MappingNode,
				Tag:  "!!map",
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Tag: "!!str", Value: "static"},
					{Kind: yaml.MappingNode, Tag: "!!map", Content: staticContent},
				},
			},
		},
	}
}

func newEnvProviderAWSLoginCmd(env *envCommand) *cobra.Command {
	var sessionToken string
	var pathStr string
	var draft string

	cmd := &cobra.Command{
		Use:   "aws-login [<org>/][<project>/]<environment-name> <access-key-id> <secret-access-key>",
		Args:  cobra.RangeArgs(2, 3),
		Short: "Add an AWS static-credentials login provider to an environment",
		Long: "Add an AWS static-credentials login provider to an environment\n" +
			"\n" +
			"Writes an `fn::open::aws-login` block with static credentials at the configured\n" +
			"path under `values`. The secret access key and session token, if any, are\n" +
			"wrapped in `fn::secret`. If a block already exists at the path it is replaced.\n",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if err := env.esc.getCachedClient(ctx); err != nil {
				return err
			}

			ref, args, err := env.getExistingEnvRef(ctx, args)
			if err != nil {
				return err
			}
			if ref.version != "" {
				return fmt.Errorf("the provider command does not accept versions")
			}
			if len(args) != 2 {
				return fmt.Errorf("expected <access-key-id> and <secret-access-key>")
			}
			accessKeyID, secretAccessKey := args[0], args[1]

			path, err := resource.ParsePropertyPath(pathStr)
			if err != nil {
				return fmt.Errorf("invalid --path: %w", err)
			}

			node := buildAWSLoginNode(accessKeyID, secretAccessKey, sessionToken)

			return applyProviderUpdate(ctx, env, ref, draft, path, node)
		},
	}

	cmd.Flags().StringVar(&sessionToken, "session-token", "", "optional AWS session token")
	cmd.Flags().StringVar(&pathStr, "path", "aws.login", "property path under `values` where the provider block is written")
	cmd.Flags().StringVar(&draft, "draft", "",
		"set flag without a value (--draft) to create a draft rather than saving changes directly. --draft=<change-request-id> to update an existing change request.")
	cmd.Flag("draft").NoOptDefVal = "new"

	return cmd
}
```

Then extend `cmd/esc/cli/env_provider_common.go` with the shared CLI-side update wrapper that all three provider subcommands call (this is the second helper, distinct from `mergeProviderIntoEnv`). Merge it into the existing file so the imports stay together — final shape of `env_provider_common.go`:

```go
// Copyright 2026, Pulumi Corporation.

package cli

import (
	"context"
	"fmt"

	"github.com/pulumi/esc/syntax/encoding"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"gopkg.in/yaml.v3"
)

// mergeProviderIntoEnv ... (unchanged from Task 2)

// secretNode ... (unchanged from Task 2)

func applyProviderUpdate(
	ctx context.Context,
	env *envCommand,
	ref environmentRef,
	draft string,
	path resource.PropertyPath,
	providerNode *yaml.Node,
) error {
	var def []byte
	var tag string
	var err error
	if draft != "" && draft != "new" {
		def, tag, err = env.esc.client.GetEnvironmentDraft(ctx, ref.orgName, ref.projectName, ref.envName, draft)
		if err != nil {
			return fmt.Errorf("getting environment draft definition: %w", err)
		}
	} else {
		def, tag, _, err = env.esc.client.GetEnvironment(ctx, ref.orgName, ref.projectName, ref.envName, "", false)
		if err != nil {
			return fmt.Errorf("getting environment definition: %w", err)
		}
	}

	newYAML, err := mergeProviderIntoEnv(def, path, providerNode)
	if err != nil {
		return err
	}

	diags, err := env.esc.updateEnvironment(ctx, ref, draft, newYAML, tag, "Provider updated.")
	if err != nil {
		return err
	}
	if len(diags) != 0 {
		return env.writePropertyEnvironmentDiagnostics(env.esc.stderr, diags)
	}
	return nil
}
```

(The `context` import is the only new one; `client` is not needed here because we go through `env.esc.client` and `env.esc.updateEnvironment`.)

- [ ] **Step 4: Run the builder test to verify pass + verify build**

Run:
```
go test ./cmd/esc/cli/ -run TestBuildAWSLoginNode -count=1 -v
```
Expected: PASS.

The full build still won't succeed yet (Azure and GCP constructors don't exist). Verify that with:
```
go build ./...
```
Expected: build fails on the two missing constructors only. If any *other* error appears, fix it before continuing.

- [ ] **Step 5: Commit**

```bash
git add cmd/esc/cli/env_provider_aws.go cmd/esc/cli/env_provider_aws_test.go cmd/esc/cli/env_provider_common.go
git commit -m "esc env provider: aws-login subcommand"
```

---

## Task 4: `azure-login` subcommand (TDD on builder, then cobra)

**Files:**
- Create: `cmd/esc/cli/env_provider_azure.go`
- Create: `cmd/esc/cli/env_provider_azure_test.go`

- [ ] **Step 1: Write failing tests**

Create `cmd/esc/cli/env_provider_azure_test.go`:

```go
// Copyright 2026, Pulumi Corporation.

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestBuildAzureLoginNode_Required(t *testing.T) {
	node := buildAzureLoginNode("client-id", "tenant-id", "/subscriptions/sub", "")
	out, err := yaml.Marshal(node)
	require.NoError(t, err)
	assert.YAMLEq(t, `fn::open::azure-login:
  clientId: client-id
  tenantId: tenant-id
  subscriptionId: /subscriptions/sub
`, string(out))
}

func TestBuildAzureLoginNode_WithClientSecret(t *testing.T) {
	node := buildAzureLoginNode("client-id", "tenant-id", "/subscriptions/sub", "shhh")
	out, err := yaml.Marshal(node)
	require.NoError(t, err)
	assert.YAMLEq(t, `fn::open::azure-login:
  clientId: client-id
  tenantId: tenant-id
  subscriptionId: /subscriptions/sub
  clientSecret:
    fn::secret: shhh
`, string(out))
}
```

- [ ] **Step 2: Run the test to verify failure**

Run: `go test ./cmd/esc/cli/ -run TestBuildAzureLoginNode -count=1`
Expected: FAIL — `undefined: buildAzureLoginNode`.

- [ ] **Step 3: Implement the builder + cobra command**

Create `cmd/esc/cli/env_provider_azure.go`:

```go
// Copyright 2026, Pulumi Corporation.

package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

// buildAzureLoginNode returns a yaml.Node representing
// `fn::open::azure-login: { ... }`. clientSecret is wrapped in `fn::secret` and
// omitted when empty.
func buildAzureLoginNode(clientID, tenantID, subscriptionID, clientSecret string) *yaml.Node {
	loginContent := []*yaml.Node{
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "clientId"},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: clientID},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "tenantId"},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: tenantID},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "subscriptionId"},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: subscriptionID},
	}
	if clientSecret != "" {
		loginContent = append(loginContent,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "clientSecret"},
			secretNode(clientSecret),
		)
	}

	return &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: "fn::open::azure-login"},
			{Kind: yaml.MappingNode, Tag: "!!map", Content: loginContent},
		},
	}
}

func newEnvProviderAzureLoginCmd(env *envCommand) *cobra.Command {
	var clientSecret string
	var pathStr string
	var draft string

	cmd := &cobra.Command{
		Use:   "azure-login [<org>/][<project>/]<environment-name> <client-id> <tenant-id> <subscription-id>",
		Args:  cobra.RangeArgs(3, 4),
		Short: "Add an Azure static-credentials login provider to an environment",
		Long: "Add an Azure static-credentials login provider to an environment\n" +
			"\n" +
			"Writes an `fn::open::azure-login` block at the configured path under `values`.\n" +
			"`--client-secret`, if provided, is wrapped in `fn::secret`. If a block already\n" +
			"exists at the path it is replaced.\n",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if err := env.esc.getCachedClient(ctx); err != nil {
				return err
			}

			ref, args, err := env.getExistingEnvRef(ctx, args)
			if err != nil {
				return err
			}
			if ref.version != "" {
				return fmt.Errorf("the provider command does not accept versions")
			}
			if len(args) != 3 {
				return fmt.Errorf("expected <client-id> <tenant-id> <subscription-id>")
			}
			clientID, tenantID, subscriptionID := args[0], args[1], args[2]

			path, err := resource.ParsePropertyPath(pathStr)
			if err != nil {
				return fmt.Errorf("invalid --path: %w", err)
			}

			node := buildAzureLoginNode(clientID, tenantID, subscriptionID, clientSecret)

			return applyProviderUpdate(ctx, env, ref, draft, path, node)
		},
	}

	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "optional Azure client secret")
	cmd.Flags().StringVar(&pathStr, "path", "azure.login", "property path under `values` where the provider block is written")
	cmd.Flags().StringVar(&draft, "draft", "",
		"set flag without a value (--draft) to create a draft rather than saving changes directly. --draft=<change-request-id> to update an existing change request.")
	cmd.Flag("draft").NoOptDefVal = "new"

	return cmd
}
```

- [ ] **Step 4: Run the test to verify pass**

Run: `go test ./cmd/esc/cli/ -run TestBuildAzureLoginNode -count=1 -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/esc/cli/env_provider_azure.go cmd/esc/cli/env_provider_azure_test.go
git commit -m "esc env provider: azure-login subcommand"
```

---

## Task 5: `gcp-login` subcommand (TDD on builder, then cobra)

**Files:**
- Create: `cmd/esc/cli/env_provider_gcp.go`
- Create: `cmd/esc/cli/env_provider_gcp_test.go`

GCP has two wrinkles vs. AWS/Azure:
1. `project` is a YAML integer (per schema `"type": "number"`), not a string.
2. The access token sits two levels deep: `accessToken.accessToken` is the actual secret; `accessToken.serviceAccount` and `accessToken.tokenLifetime` are siblings.

- [ ] **Step 1: Write failing tests**

Create `cmd/esc/cli/env_provider_gcp_test.go`:

```go
// Copyright 2026, Pulumi Corporation.

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestBuildGCPLoginNode_Required(t *testing.T) {
	node := buildGCPLoginNode(123456789, "ya29.token", "", "")
	out, err := yaml.Marshal(node)
	require.NoError(t, err)
	assert.YAMLEq(t, `fn::open::gcp-login:
  project: 123456789
  accessToken:
    accessToken:
      fn::secret: ya29.token
`, string(out))
}

func TestBuildGCPLoginNode_WithImpersonation(t *testing.T) {
	node := buildGCPLoginNode(1, "ya29.token", "sa@proj.iam.gserviceaccount.com", "1h")
	out, err := yaml.Marshal(node)
	require.NoError(t, err)
	assert.YAMLEq(t, `fn::open::gcp-login:
  project: 1
  accessToken:
    accessToken:
      fn::secret: ya29.token
    serviceAccount: sa@proj.iam.gserviceaccount.com
    tokenLifetime: 1h
`, string(out))
}
```

- [ ] **Step 2: Run the test to verify failure**

Run: `go test ./cmd/esc/cli/ -run TestBuildGCPLoginNode -count=1`
Expected: FAIL — `undefined: buildGCPLoginNode`.

- [ ] **Step 3: Implement the builder + cobra command**

Create `cmd/esc/cli/env_provider_gcp.go`:

```go
// Copyright 2026, Pulumi Corporation.

package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

// buildGCPLoginNode returns a yaml.Node representing
// `fn::open::gcp-login: { project, accessToken: { accessToken: {fn::secret}, ... } }`.
// serviceAccount and tokenLifetime are omitted when empty.
func buildGCPLoginNode(project int64, accessToken, serviceAccount, tokenLifetime string) *yaml.Node {
	accessTokenContent := []*yaml.Node{
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "accessToken"},
		secretNode(accessToken),
	}
	if serviceAccount != "" {
		accessTokenContent = append(accessTokenContent,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "serviceAccount"},
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: serviceAccount},
		)
	}
	if tokenLifetime != "" {
		accessTokenContent = append(accessTokenContent,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "tokenLifetime"},
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: tokenLifetime},
		)
	}

	return &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: "fn::open::gcp-login"},
			{
				Kind: yaml.MappingNode,
				Tag:  "!!map",
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Tag: "!!str", Value: "project"},
					{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.FormatInt(project, 10)},
					{Kind: yaml.ScalarNode, Tag: "!!str", Value: "accessToken"},
					{Kind: yaml.MappingNode, Tag: "!!map", Content: accessTokenContent},
				},
			},
		},
	}
}

func newEnvProviderGCPLoginCmd(env *envCommand) *cobra.Command {
	var serviceAccount string
	var tokenLifetime string
	var pathStr string
	var draft string

	cmd := &cobra.Command{
		Use:   "gcp-login [<org>/][<project>/]<environment-name> <project-number> <access-token>",
		Args:  cobra.RangeArgs(2, 3),
		Short: "Add a GCP static-credentials login provider to an environment",
		Long: "Add a GCP static-credentials login provider to an environment\n" +
			"\n" +
			"Writes an `fn::open::gcp-login` block at the configured path under `values`. The\n" +
			"access token is wrapped in `fn::secret`. <project-number> must be the numerical\n" +
			"GCP project ID. If a block already exists at the path it is replaced.\n",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if err := env.esc.getCachedClient(ctx); err != nil {
				return err
			}

			ref, args, err := env.getExistingEnvRef(ctx, args)
			if err != nil {
				return err
			}
			if ref.version != "" {
				return fmt.Errorf("the provider command does not accept versions")
			}
			if len(args) != 2 {
				return fmt.Errorf("expected <project-number> and <access-token>")
			}
			project, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid project number %q: must be a positive integer", args[0])
			}
			if project <= 0 {
				return fmt.Errorf("invalid project number %q: must be a positive integer", args[0])
			}
			accessToken := args[1]

			path, err := resource.ParsePropertyPath(pathStr)
			if err != nil {
				return fmt.Errorf("invalid --path: %w", err)
			}

			node := buildGCPLoginNode(project, accessToken, serviceAccount, tokenLifetime)

			return applyProviderUpdate(ctx, env, ref, draft, path, node)
		},
	}

	cmd.Flags().StringVar(&serviceAccount, "service-account", "", "optional GCP service account to impersonate")
	cmd.Flags().StringVar(&tokenLifetime, "token-lifetime", "", "optional lifetime for impersonated credentials, e.g. 1h30m")
	cmd.Flags().StringVar(&pathStr, "path", "gcp.login", "property path under `values` where the provider block is written")
	cmd.Flags().StringVar(&draft, "draft", "",
		"set flag without a value (--draft) to create a draft rather than saving changes directly. --draft=<change-request-id> to update an existing change request.")
	cmd.Flag("draft").NoOptDefVal = "new"

	return cmd
}
```

- [ ] **Step 4: Run all unit tests + a full build**

Run:
```
go build ./...
go test ./cmd/esc/cli/ -run 'TestBuildAWSLoginNode|TestBuildAzureLoginNode|TestBuildGCPLoginNode|TestMergeProviderIntoEnv|TestSecretNode' -count=1 -v
```
Expected: build PASS; all listed tests PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/esc/cli/env_provider_gcp.go cmd/esc/cli/env_provider_gcp_test.go
git commit -m "esc env provider: gcp-login subcommand"
```

---

## Task 6: Golden-file integration tests

**Files:**
- Create: `cmd/esc/cli/testdata/env-provider-aws-login.yaml`
- Create: `cmd/esc/cli/testdata/env-provider-azure-login.yaml`
- Create: `cmd/esc/cli/testdata/env-provider-gcp-login.yaml`

These exercise the full command through the test harness (`cmd/esc/cli/cli_test.go`). The harness regenerates expected output when `PULUMI_ACCEPT=true` is set. We generate them rather than hand-write, but the input scripts and post-conditions are specified here so the engineer doesn't accept arbitrary output.

- [ ] **Step 1: Write the AWS test script and post-conditions**

Create `cmd/esc/cli/testdata/env-provider-aws-login.yaml` with the input block only (everything before the second `---` will be filled by `PULUMI_ACCEPT`):

```yaml
run: |
  esc env init default/test
  esc env provider aws-login default/test AKIAEXAMPLE shhh && esc env get default/test
  esc env provider aws-login default/test AKIAEXAMPLE shhh --session-token tok && esc env get default/test
  esc env provider aws-login default/test AKIANEW newshhh --path other.creds && esc env get default/test

---
```

- [ ] **Step 2: Write Azure script**

Create `cmd/esc/cli/testdata/env-provider-azure-login.yaml`:

```yaml
run: |
  esc env init default/test
  esc env provider azure-login default/test cid tid /subscriptions/sub && esc env get default/test
  esc env provider azure-login default/test cid tid /subscriptions/sub --client-secret shhh && esc env get default/test

---
```

- [ ] **Step 3: Write GCP script**

Create `cmd/esc/cli/testdata/env-provider-gcp-login.yaml`:

```yaml
run: |
  esc env init default/test
  esc env provider gcp-login default/test 123456789 ya29.token && esc env get default/test
  esc env provider gcp-login default/test 123456789 ya29.token --service-account sa@proj.iam.gserviceaccount.com --token-lifetime 1h && esc env get default/test

---
```

- [ ] **Step 4: Generate expected output and verify it manually**

Run: `PULUMI_ACCEPT=true go test ./cmd/esc/cli/ -run TestCLI -count=1`

Then open each `testdata/env-provider-*.yaml` file and **manually verify** the expected output matches the YAML shapes described in the command-surface section above. Pay particular attention to:
- `fn::secret` wrapping appears on `secretAccessKey`, `sessionToken`, `clientSecret`, and the inner GCP `accessToken`.
- `gcp-login` `project` is rendered as an unquoted integer.
- The `--path other.creds` AWS case writes under `values.other.creds`, not `values.aws.login`.

If any of these are wrong, the bug is in the builders (Tasks 3–5), not in the golden file. Fix the builder and re-run with `PULUMI_ACCEPT=true`.

- [ ] **Step 5: Run the harness in verification mode**

Run: `go test ./cmd/esc/cli/ -run TestCLI -count=1`
Expected: PASS (the harness now diffs actual output against the recorded golden output).

- [ ] **Step 6: Commit**

```bash
git add cmd/esc/cli/testdata/env-provider-aws-login.yaml cmd/esc/cli/testdata/env-provider-azure-login.yaml cmd/esc/cli/testdata/env-provider-gcp-login.yaml
git commit -m "esc env provider: golden-file end-to-end tests"
```

---

## Task 7: Final verification + changelog

**Files:**
- Modify: `CHANGELOG_PENDING.md`

- [ ] **Step 1: Add changelog entry**

In `CHANGELOG_PENDING.md`, under the appropriate "Improvements" / "Features" section (read the file first to match its existing style), add:

```
- Add `esc env provider {aws-login,azure-login,gcp-login}` commands to add static-credentials login providers to an environment.
```

- [ ] **Step 2: Run the full pre-commit gate**

Run: `make format && make lint && make test`
Expected: clean across the board.

- [ ] **Step 3: Run the race-enabled coverage suite**

Per `CLAUDE.md`: if any root-level `.go` files are touched, `make test_cover` is mandatory. We have not touched root-level files, but the new code paths go through `client` and `eval` indirectly, so run it anyway as a safety net.

Run: `make test_cover`
Expected: PASS.

- [ ] **Step 4: Smoke-test the CLI manually**

```bash
make build
esc env init <org>/test-provider-plan
esc env provider aws-login <org>/test-provider-plan AKIAEXAMPLE shhh
esc env get <org>/test-provider-plan
# Verify: values.aws.login holds an fn::open::aws-login block with the secret wrapped in fn::secret.
esc env rm <org>/test-provider-plan
```

If the user does not have credentials handy, skip this step and note it in the PR description.

- [ ] **Step 5: Commit and open PR**

```bash
git add CHANGELOG_PENDING.md
git commit -m "esc env provider: changelog entry"
```

PR title: `cmd/esc: add env provider login subcommands for AWS, Azure, GCP`
PR description should link to pulumi/pulumi#23043 and call out:
- OIDC is **not** implemented; the command structure leaves room for it.
- Secret-bearing fields are wrapped in `fn::secret`.
- Existing blocks at the target path are replaced wholesale (not deep-merged).
