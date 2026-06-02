// Copyright 2026, Pulumi Corporation.
//
// Hand-written shims for generic functions. See shim_gen.go for the rest.
// Deprecated: import github.com/pulumi/pulumi/sdk/v3/go/esc/syntax directly.

package syntax

import (
	"encoding/json"

	canonical "github.com/pulumi/pulumi/sdk/v3/go/esc/syntax"
)

// AsNumber converts a NumberValue to a json.Number.
func AsNumber[T NumberValue](v T) json.Number { return canonical.AsNumber(v) }

// NumberSyntax creates a new NumberNode with the given syntax and value.
func NumberSyntax[T NumberValue](syntax Syntax, value T) *NumberNode {
	return canonical.NumberSyntax(syntax, value)
}

// Number creates a new NumberNode with the given value.
func Number[T NumberValue](value T) *NumberNode { return canonical.Number(value) }
