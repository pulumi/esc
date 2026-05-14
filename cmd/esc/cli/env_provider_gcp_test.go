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
