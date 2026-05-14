// Copyright 2026, Pulumi Corporation.

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestBuildGCPLoginStaticNode_Required(t *testing.T) {
	node := buildGCPLoginStaticNode(123456789, "ya29.token", "", "")
	out, err := yaml.Marshal(node)
	require.NoError(t, err)
	assert.YAMLEq(t, `fn::open::gcp-login:
  project: 123456789
  accessToken:
    accessToken:
      fn::secret: ya29.token
`, string(out))
}

func TestBuildGCPLoginStaticNode_WithImpersonation(t *testing.T) {
	node := buildGCPLoginStaticNode(1, "ya29.token", "sa@proj.iam.gserviceaccount.com", "1h")
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
