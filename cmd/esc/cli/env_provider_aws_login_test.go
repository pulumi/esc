// Copyright 2026, Pulumi Corporation.

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestBuildAWSLoginStaticNode_Required(t *testing.T) {
	node := buildAWSLoginStaticNode("AKIAEXAMPLE", "shhh", "")
	out, err := yaml.Marshal(node)
	require.NoError(t, err)
	assert.YAMLEq(t, `fn::open::aws-login:
  static:
    accessKeyId: AKIAEXAMPLE
    secretAccessKey:
      fn::secret: shhh
`, string(out))
}

func TestBuildAWSLoginStaticNode_WithSessionToken(t *testing.T) {
	node := buildAWSLoginStaticNode("AKIAEXAMPLE", "shhh", "tok")
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
