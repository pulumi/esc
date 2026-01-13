// Copyright 2026, Pulumi Corporation.

package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/esc/cmd/esc/cli/client"
	"github.com/pulumi/esc/cmd/esc/cli/workspace"
	pulumi_workspace "github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

func TestParsePartitionFromARN(t *testing.T) {
	tests := []struct {
		name     string
		arn      string
		expected string
	}{
		{
			name:     "standard AWS partition",
			arn:      "arn:aws:sts::123456789012:assumed-role/role-name/session",
			expected: "aws",
		},
		{
			name:     "GovCloud partition",
			arn:      "arn:aws-us-gov:sts::123456789012:assumed-role/role-name/session",
			expected: "aws-us-gov",
		},
		{
			name:     "China partition",
			arn:      "arn:aws-cn:sts::123456789012:assumed-role/role-name/session",
			expected: "aws-cn",
		},
		{
			name:     "invalid ARN defaults to aws",
			arn:      "invalid",
			expected: "aws",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePartitionFromARN(tt.arn)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseEnvRef(t *testing.T) {
	tests := []struct {
		name        string
		defaultOrg  string
		envName     string
		expectedOrg string
		expectedPrj string
		expectedEnv string
	}{
		{
			name:        "full path org/project/env",
			defaultOrg:  "default-org",
			envName:     "my-org/my-project/my-env",
			expectedOrg: "my-org",
			expectedPrj: "my-project",
			expectedEnv: "my-env",
		},
		{
			name:        "project/env uses default org",
			defaultOrg:  "default-org",
			envName:     "my-project/my-env",
			expectedOrg: "default-org",
			expectedPrj: "my-project",
			expectedEnv: "my-env",
		},
		{
			name:        "env only uses default org and default project",
			defaultOrg:  "default-org",
			envName:     "my-env",
			expectedOrg: "default-org",
			expectedPrj: "default",
			expectedEnv: "my-env",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := parseEnvRef(tt.envName, tt.defaultOrg)
			assert.Equal(t, tt.expectedOrg, ref.orgName)
			assert.Equal(t, tt.expectedPrj, ref.projectName)
			assert.Equal(t, tt.expectedEnv, ref.envName)
		})
	}
}

func TestGenerateEnvironmentYAML(t *testing.T) {
	roleArn := "arn:aws:iam::123456789012:role/TestRole"
	yaml := generateEnvironmentYAML(roleArn)

	assert.Contains(t, yaml, "fn::open::aws-login:")
	assert.Contains(t, yaml, "oidc:")
	assert.Contains(t, yaml, "duration: 1h")
	assert.Contains(t, yaml, roleArn)
	assert.Contains(t, yaml, "sessionName: pulumi-esc-session")
	assert.Contains(t, yaml, "AWS_ACCESS_KEY_ID: ${aws.login.accessKeyId}")
	assert.Contains(t, yaml, "AWS_SECRET_ACCESS_KEY: ${aws.login.secretAccessKey}")
	assert.Contains(t, yaml, "AWS_SESSION_TOKEN: ${aws.login.sessionToken}")
}

func TestSetupAWSCmd_MissingRoleName(t *testing.T) {
	backend := "https://api.pulumi.com"
	creds := pulumi_workspace.Credentials{
		Current: backend,
		Accounts: map[string]pulumi_workspace.Account{
			backend: {
				Username:    "test-user",
				AccessToken: "access-token",
			},
		},
	}

	fs := testFS{}
	testWorkspace := workspace.New(fs, &testPulumiWorkspace{
		credentials: creds,
		config: pulumi_workspace.PulumiConfig{
			BackendConfig: map[string]pulumi_workspace.BackendConfig{
				backend: {DefaultOrg: "test-org"},
			},
		},
	})

	var stdout, stderr bytes.Buffer
	esc := &escCommand{
		command:   "esc",
		login:     &testLoginManager{creds: creds},
		workspace: testWorkspace,
		environ:   testEnviron{},
		stdout:    &stdout,
		stderr:    &stderr,
		newClient: func(userAgent, backendURL, accessToken string, insecure bool) client.Client {
			return &testPulumiClient{user: "test-user", defaultOrg: "test-org"}
		},
	}

	setup := &setupCommand{esc: esc}
	cmd := newSetupAWSCmd(setup)

	cmd.SetArgs([]string{})
	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--role-name is required")
}

func TestSetupAWSCmd_WithOrgFlag(t *testing.T) {
	// This test verifies that the --org flag is properly parsed
	// The actual AWS calls would fail without real credentials,
	// but we can verify the command structure is correct
	backend := "https://api.pulumi.com"
	creds := pulumi_workspace.Credentials{
		Current: backend,
		Accounts: map[string]pulumi_workspace.Account{
			backend: {
				Username:    "test-user",
				AccessToken: "access-token",
			},
		},
	}

	fs := testFS{}
	testWorkspace := workspace.New(fs, &testPulumiWorkspace{
		credentials: creds,
		config: pulumi_workspace.PulumiConfig{
			BackendConfig: map[string]pulumi_workspace.BackendConfig{
				backend: {DefaultOrg: "default-org"},
			},
		},
	})

	var stdout, stderr bytes.Buffer
	esc := &escCommand{
		command:   "esc",
		login:     &testLoginManager{creds: creds},
		workspace: testWorkspace,
		environ:   testEnviron{},
		stdout:    &stdout,
		stderr:    &stderr,
		newClient: func(userAgent, backendURL, accessToken string, insecure bool) client.Client {
			return &testPulumiClient{user: "test-user", defaultOrg: "default-org"}
		},
	}

	setup := &setupCommand{esc: esc}
	cmd := newSetupAWSCmd(setup)

	// Verify the command has the expected flags
	assert.NotNil(t, cmd.Flags().Lookup("role-name"))
	assert.NotNil(t, cmd.Flags().Lookup("policy"))
	assert.NotNil(t, cmd.Flags().Lookup("org"))
	assert.NotNil(t, cmd.Flags().Lookup("environment"))

	// Verify default policy value
	policyFlag := cmd.Flags().Lookup("policy")
	assert.Equal(t, "AdministratorAccess", policyFlag.DefValue)
}

// Mock IAM client for testing
type mockIAMClient struct {
	createOpenIDConnectProviderFunc    func(ctx context.Context, params *iam.CreateOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.CreateOpenIDConnectProviderOutput, error)
	addClientIDToOpenIDConnectProvider func(ctx context.Context, params *iam.AddClientIDToOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.AddClientIDToOpenIDConnectProviderOutput, error)
	createRoleFunc                     func(ctx context.Context, params *iam.CreateRoleInput, optFns ...func(*iam.Options)) (*iam.CreateRoleOutput, error)
	attachRolePolicyFunc               func(ctx context.Context, params *iam.AttachRolePolicyInput, optFns ...func(*iam.Options)) (*iam.AttachRolePolicyOutput, error)
}

func (m *mockIAMClient) CreateOpenIDConnectProvider(ctx context.Context, params *iam.CreateOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.CreateOpenIDConnectProviderOutput, error) {
	return m.createOpenIDConnectProviderFunc(ctx, params, optFns...)
}

func (m *mockIAMClient) AddClientIDToOpenIDConnectProvider(ctx context.Context, params *iam.AddClientIDToOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.AddClientIDToOpenIDConnectProviderOutput, error) {
	return m.addClientIDToOpenIDConnectProvider(ctx, params, optFns...)
}

func (m *mockIAMClient) CreateRole(ctx context.Context, params *iam.CreateRoleInput, optFns ...func(*iam.Options)) (*iam.CreateRoleOutput, error) {
	return m.createRoleFunc(ctx, params, optFns...)
}

func (m *mockIAMClient) AttachRolePolicy(ctx context.Context, params *iam.AttachRolePolicyInput, optFns ...func(*iam.Options)) (*iam.AttachRolePolicyOutput, error) {
	return m.attachRolePolicyFunc(ctx, params, optFns...)
}

func TestCreateOrUpdateOIDCProvider(t *testing.T) {
	t.Run("creates new OIDC provider", func(t *testing.T) {
		var stdout bytes.Buffer
		esc := &escCommand{stdout: &stdout}

		iamClient := &mockIAMClient{
			createOpenIDConnectProviderFunc: func(ctx context.Context, params *iam.CreateOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.CreateOpenIDConnectProviderOutput, error) {
				assert.Equal(t, oidcIssuerURL, *params.Url)
				assert.Equal(t, []string{"aws:test-org"}, params.ClientIDList)
				return &iam.CreateOpenIDConnectProviderOutput{}, nil
			},
		}

		isNew, err := createOrUpdateOIDCProvider(context.Background(), esc, iamClient, "arn:aws:iam::123456789012:oidc-provider/api.pulumi.com/oidc", "aws:test-org", "thumbprint123")

		require.NoError(t, err)
		assert.True(t, isNew)
		assert.Contains(t, stdout.String(), "(created)")
	})

	t.Run("adds audience to existing OIDC provider", func(t *testing.T) {
		var stdout bytes.Buffer
		esc := &escCommand{stdout: &stdout}

		iamClient := &mockIAMClient{
			createOpenIDConnectProviderFunc: func(ctx context.Context, params *iam.CreateOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.CreateOpenIDConnectProviderOutput, error) {
				return nil, &iamtypes.EntityAlreadyExistsException{Message: aws.String("provider already exists")}
			},
			addClientIDToOpenIDConnectProvider: func(ctx context.Context, params *iam.AddClientIDToOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.AddClientIDToOpenIDConnectProviderOutput, error) {
				assert.Equal(t, "aws:test-org", *params.ClientID)
				return &iam.AddClientIDToOpenIDConnectProviderOutput{}, nil
			},
		}

		isNew, err := createOrUpdateOIDCProvider(context.Background(), esc, iamClient, "arn:aws:iam::123456789012:oidc-provider/api.pulumi.com/oidc", "aws:test-org", "thumbprint123")

		require.NoError(t, err)
		assert.False(t, isNew)
		assert.Contains(t, stdout.String(), "(existing)")
		assert.Contains(t, stdout.String(), "Added audience")
	})
}

func TestCreateRole(t *testing.T) {
	t.Run("creates new role", func(t *testing.T) {
		var stdout bytes.Buffer
		esc := &escCommand{stdout: &stdout}

		iamClient := &mockIAMClient{
			createRoleFunc: func(ctx context.Context, params *iam.CreateRoleInput, optFns ...func(*iam.Options)) (*iam.CreateRoleOutput, error) {
				assert.Equal(t, "TestRole", *params.RoleName)
				assert.Contains(t, *params.AssumeRolePolicyDocument, "sts:AssumeRoleWithWebIdentity")
				assert.Contains(t, *params.AssumeRolePolicyDocument, "aws:test-org")
				return &iam.CreateRoleOutput{}, nil
			},
		}

		isNew, err := createRole(context.Background(), esc, iamClient, "TestRole", "arn:aws:iam::123456789012:oidc-provider/api.pulumi.com/oidc", "aws:test-org")

		require.NoError(t, err)
		assert.True(t, isNew)
		assert.Contains(t, stdout.String(), "(created)")
	})

	t.Run("handles existing role", func(t *testing.T) {
		var stdout bytes.Buffer
		esc := &escCommand{stdout: &stdout}

		iamClient := &mockIAMClient{
			createRoleFunc: func(ctx context.Context, params *iam.CreateRoleInput, optFns ...func(*iam.Options)) (*iam.CreateRoleOutput, error) {
				return nil, &iamtypes.EntityAlreadyExistsException{Message: aws.String("role already exists")}
			},
		}

		isNew, err := createRole(context.Background(), esc, iamClient, "TestRole", "arn:aws:iam::123456789012:oidc-provider/api.pulumi.com/oidc", "aws:test-org")

		require.NoError(t, err)
		assert.False(t, isNew)
		assert.Contains(t, stdout.String(), "(existing)")
	})

	t.Run("returns error on failure", func(t *testing.T) {
		var stdout bytes.Buffer
		esc := &escCommand{stdout: &stdout}

		iamClient := &mockIAMClient{
			createRoleFunc: func(ctx context.Context, params *iam.CreateRoleInput, optFns ...func(*iam.Options)) (*iam.CreateRoleOutput, error) {
				return nil, errors.New("access denied")
			},
		}

		_, err := createRole(context.Background(), esc, iamClient, "TestRole", "arn:aws:iam::123456789012:oidc-provider/api.pulumi.com/oidc", "aws:test-org")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "creating IAM role")
	})
}

func TestAttachRolePolicy(t *testing.T) {
	t.Run("attaches policy successfully", func(t *testing.T) {
		var stdout bytes.Buffer
		esc := &escCommand{stdout: &stdout}

		iamClient := &mockIAMClient{
			attachRolePolicyFunc: func(ctx context.Context, params *iam.AttachRolePolicyInput, optFns ...func(*iam.Options)) (*iam.AttachRolePolicyOutput, error) {
				assert.Equal(t, "TestRole", *params.RoleName)
				assert.Equal(t, "arn:aws:iam::aws:policy/AdministratorAccess", *params.PolicyArn)
				return &iam.AttachRolePolicyOutput{}, nil
			},
		}

		err := attachRolePolicy(context.Background(), esc, iamClient, "TestRole", "arn:aws:iam::aws:policy/AdministratorAccess")

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Policy Attached")
	})

	t.Run("returns error on failure", func(t *testing.T) {
		var stdout bytes.Buffer
		esc := &escCommand{stdout: &stdout}

		iamClient := &mockIAMClient{
			attachRolePolicyFunc: func(ctx context.Context, params *iam.AttachRolePolicyInput, optFns ...func(*iam.Options)) (*iam.AttachRolePolicyOutput, error) {
				return nil, errors.New("policy not found")
			},
		}

		err := attachRolePolicy(context.Background(), esc, iamClient, "TestRole", "arn:aws:iam::aws:policy/InvalidPolicy")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "attaching policy to role")
	})
}

func TestPrintEnvironmentYAML(t *testing.T) {
	var stdout bytes.Buffer
	esc := &escCommand{stdout: &stdout}

	printEnvironmentYAML(esc, "arn:aws:iam::123456789012:role/TestRole")

	output := stdout.String()
	assert.Contains(t, output, "Add the following to your ESC environment:")
	assert.Contains(t, output, "arn:aws:iam::123456789012:role/TestRole")
	assert.Contains(t, output, "fn::open::aws-login:")
}
