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
