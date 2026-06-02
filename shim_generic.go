// Copyright 2026, Pulumi Corporation.
//
// Hand-written shims for generic functions, which cannot be re-exported as
// package-level vars. See shim_gen.go for the rest of the engine re-exports.
// Deprecated: import github.com/pulumi/pulumi/sdk/v3/go/esc directly.

package esc

import canonical "github.com/pulumi/pulumi/sdk/v3/go/esc"

// NewValue creates a new Value from the given representation.
func NewValue[T ValueType](v T) Value { return canonical.NewValue(v) }

// NewSecret creates a new secret Value from the given representation.
func NewSecret[T ValueType](v T) Value { return canonical.NewSecret(v) }
