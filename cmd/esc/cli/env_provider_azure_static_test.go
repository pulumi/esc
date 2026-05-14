// Copyright 2026, Pulumi Corporation.

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestBuildAzureLoginStaticNode_Required(t *testing.T) {
	node := buildAzureLoginStaticNode("client-id", "tenant-id", "/subscriptions/sub", "")
	out, err := yaml.Marshal(node)
	require.NoError(t, err)
	assert.YAMLEq(t, `fn::open::azure-login:
  clientId: client-id
  tenantId: tenant-id
  subscriptionId: /subscriptions/sub
`, string(out))
}

func TestBuildAzureLoginStaticNode_WithClientSecret(t *testing.T) {
	node := buildAzureLoginStaticNode("client-id", "tenant-id", "/subscriptions/sub", "shhh")
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
