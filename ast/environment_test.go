// Copyright 2023, Pulumi Corporation.
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

package ast

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/pulumi/esc/syntax"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/pulumi/esc/syntax/encoding"
)

func TestExample(t *testing.T) {
	t.Parallel()

	const example = `
imports:
  - green-channel
  - us-west-2
config:
  aws:
    fn::open:
      provider: aws-oidc
      inputs:
        sessionName: site-prod-session
        roleArn: some-role-arn
  pulumi:
    aws:defaultTags:
      tags:
        environment: prod
`

	syntax, diags := encoding.DecodeYAML("<stdin>", yaml.NewDecoder(strings.NewReader(example)), nil)
	require.Len(t, diags, 0)

	environment, diags := ParseEnvironment([]byte(example), syntax)
	assert.Len(t, diags, 0)

	assert.Nil(t, environment.Description)
}

func TestExample2(t *testing.T) {
	t.Parallel()

	const example = `
imports:
  - green-channel
  - us-west-2
config:
  aws:
    fn::open::aws-oidc:
      sessionName: site-prod-session
      roleArn: some-role-arn
  pulumi:
    aws:defaultTags:
      tags:
        environment: prod
`

	syntax, diags := encoding.DecodeYAML("<stdin>", yaml.NewDecoder(strings.NewReader(example)), nil)
	require.Len(t, diags, 0)

	environment, diags := ParseEnvironment([]byte(example), syntax)
	assert.Len(t, diags, 0)

	assert.Nil(t, environment.Description)
}

func TestEmptyDocument(t *testing.T) {
	t.Parallel()

	const example = ``

	syntax, diags := encoding.DecodeYAML("<stdin>", yaml.NewDecoder(strings.NewReader(example)), nil)
	require.Len(t, diags, 0)

	environment, diags := ParseEnvironment([]byte(example), syntax)
	assert.Len(t, diags, 0)

	assert.Nil(t, environment.Description)
}

type testRecordDecl struct {
	syntax syntax.Node

	Bool *BooleanExpr
	Str  *StringExpr
	Num  *NumberExpr
}

func (d *testRecordDecl) recordSyntax() *syntax.Node {
	return &d.syntax
}

func Test_parseRecord(t *testing.T) {
	type args struct {
		objName        string
		dest           recordDecl
		node           syntax.Node
		noMatchWarning bool
	}
	tests := []struct {
		name string
		args args
		want syntax.Diagnostics
	}{
		{
			name: "testDecl - valid",
			args: args{
				objName: "testDecl",
				dest:    &testRecordDecl{},
				node: syntax.Object(
					syntax.ObjectProperty(syntax.String("str"), syntax.String("world")),
					syntax.ObjectProperty(syntax.String("bool"), syntax.Boolean(true)),
					syntax.ObjectProperty(syntax.String("num"), syntax.Number(3.14)),
				),
				noMatchWarning: true,
			},
			want: nil,
		},
		{
			name: "testDecl - not an object",
			args: args{
				objName:        "testDecl",
				dest:           &testRecordDecl{},
				node:           syntax.String("hello"),
				noMatchWarning: true,
			},
			want: syntax.Diagnostics{
				{
					Diagnostic: hcl.Diagnostic{
						Summary:  "testDecl must be an object",
						Severity: 1,
					},
				},
			},
		},
		{
			name: "testDecl",
			args: args{
				objName:        "testDecl",
				dest:           &testRecordDecl{},
				node:           syntax.Object(syntax.ObjectProperty(syntax.String("hello"), syntax.String("world"))),
				noMatchWarning: true,
			},
			want: syntax.Diagnostics{
				{
					Diagnostic: hcl.Diagnostic{
						Severity: 2,
						Summary:  "Field 'hello' does not exist on Object 'testDecl'",
						Detail:   "Existing fields are: 'bool', 'num', 'str'",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, parseRecord(tt.args.objName, tt.args.dest, tt.args.node, tt.args.noMatchWarning), "parseRecord(%v, %v, %v, %v)", tt.args.objName, tt.args.dest, tt.args.node, tt.args.noMatchWarning)
		})
	}
}
