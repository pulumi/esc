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
