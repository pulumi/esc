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
        provider: swap
        inputs:
          state:
            a: a
            b: b
    b:
    - c:
        fn::rotate::swap:
          state:
            a: 
              fn::secret: a
            b: b
`
	env, diags, err := LoadYAMLBytes("<stdin>", []byte(def))
	require.NoError(t, err)
	require.Len(t, diags, 0)

	// rotate the environment
	providers := testProviders{}
	execContext, err := esc.NewExecContext(nil)
	require.NoError(t, err)
	open, diags := RotateEnvironment(context.Background(), "<stdin>", env, rot128{}, providers, &testEnvironments{}, execContext)
	require.Len(t, diags, 0)

	// writeback state patches
	update, err := esc.ApplyPatches([]byte(def), open.ExecutionContext.Patches)
	require.NoError(t, err)

	encryptedYaml, err := EncryptSecrets(context.Background(), "<stdin>", update, rot128{})
	require.NoError(t, err)

	t.Log(string(encryptedYaml))
	//	    values:
	//          a:
	//            a:
	//              fn::open:
	//                provider: swap
	//                inputs:
	//                  state:
	//                    a: b
	//                    b: a
	//            b:
	//              - c:
	//                  fn::open::swap:
	//                    state:
	//                      a: b
	//                      b:
	//                        fn::secret:
	//                          ciphertext: ZXNjeAAAAAHhQRt8TQ==
}
