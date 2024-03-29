// Copyright 2023, Pulumi Corporation.

package cli

import (
	"context"
	"errors"
	"testing"

	"github.com/pulumi/esc/cmd/esc/cli/workspace"
	"github.com/pulumi/pulumi/pkg/v3/backend/display"
	pulumi_workspace "github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/stretchr/testify/assert"
)

type noCredsLoginManager int

// Current returns the current cloud backend if one is already logged in.
func (noCredsLoginManager) Current(ctx context.Context, cloudURL string, insecure, setCurrent bool) (*pulumi_workspace.Account, error) {
	return nil, nil
}

// Login logs into the target cloud URL and returns the cloud backend for it.
func (noCredsLoginManager) Login(
	ctx context.Context,
	cloudURL string,
	insecure bool,
	command string,
	message string,
	welcome func(display.Options),
	current bool,
	opts display.Options,
) (*pulumi_workspace.Account, error) {
	return nil, errors.New("unauthorized")
}

func TestNoCreds(t *testing.T) {
	fs := testFS{}
	esc := &escCommand{
		workspace: workspace.New(fs, &testPulumiWorkspace{}),
		login:     noCredsLoginManager(0),
	}
	err := esc.getCachedClient(context.Background())
	assert.ErrorContains(t, err, "could not determine current cloud")
}

func TestInvalidSelfHostedBackend(t *testing.T) {
	fs := testFS{}
	esc := &escCommand{workspace: workspace.New(fs, &testPulumiWorkspace{
		credentials: pulumi_workspace.Credentials{
			Current: "http://pulumi.com",
			Accounts: map[string]pulumi_workspace.Account{
				"http://pulumi.com": {},
			},
		},
	})}
	err := esc.getCachedClient(context.Background())
	assert.ErrorContains(t, err, "not a valid self-hosted backend")

	esc.command = "pulumi"
	err = esc.getCachedClient(context.Background())
	assert.ErrorContains(t, err, "pulumi login")
}

func TestFilestateBackend(t *testing.T) {
	fs := testFS{}
	esc := &escCommand{workspace: workspace.New(fs, &testPulumiWorkspace{
		credentials: pulumi_workspace.Credentials{
			Current: "gs://foo",
			Accounts: map[string]pulumi_workspace.Account{
				"gs://foo": {},
			},
		},
	})}
	err := esc.getCachedClient(context.Background())
	assert.ErrorContains(t, err, "does not support Pulumi ESC")
}
