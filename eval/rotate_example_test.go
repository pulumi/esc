// Copyright 2025, Pulumi Corporation.
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

package eval

import (
	"context"
	"github.com/pulumi/esc"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExampleRotate(t *testing.T) {
	const def = `
values:
  a:
    a:
      fn::rotate:
        rotator: swap
        inputs:
          foo: bar
        state:
          a: a
          b: b
    b:
    - c:
        fn::rotate::swap:
          foo: bar
          state:
            a: 
              fn::secret: a
            b: b
`
	env, diags, err := LoadYAMLBytes("<stdin>", []byte(def))
	require.NoError(t, err)
	require.Len(t, diags, 0)

	// rotate the environment
	execContext, err := esc.NewExecContext(nil)
	require.NoError(t, err)
	_, patches, diags := RotateEnvironment(context.Background(), "<stdin>", env, rot128{}, testProviders{}, &testEnvironments{}, execContext)
	require.Len(t, diags, 0)

	// writeback state patches
	update, err := ApplyPatches([]byte(def), patches)
	require.NoError(t, err)

	encryptedYaml, err := EncryptSecrets(context.Background(), "<stdin>", update, rot128{})
	require.NoError(t, err)

	t.Log(string(encryptedYaml))
	//	    values:
	//          a:
	//            a:
	//              fn::rotate:
	//                rotator: swap
	//                inputs:
	//                  state:
	//                    a: b
	//                    b: a
	//            b:
	//              - c:
	//                  fn::rotate::swap:
	//                    state:
	//                      a: b
	//                      b:
	//                        fn::secret:
	//                          ciphertext: ZXNjeAAAAAHhQRt8TQ==
}
