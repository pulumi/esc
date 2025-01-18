package eval

import (
	"github.com/pulumi/esc"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValueToSecretJSON(t *testing.T) {
	t.Run("nested secrets", func(t *testing.T) {
		actual := valueToSecretJSON(esc.NewValue(map[string]esc.Value{
			"foo": esc.NewValue(map[string]esc.Value{
				"bar": esc.NewSecret("secret"),
			}),
		}))
		expected := map[string]any{
			"foo": map[string]any{
				"bar": map[string]any{
					"fn::secret": "secret",
				},
			},
		}
		require.Equal(t, expected, actual)
	})
}
