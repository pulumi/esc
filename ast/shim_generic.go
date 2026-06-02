// Copyright 2026, Pulumi Corporation.
//
// Hand-written shims for generic functions. See shim_gen.go for the rest.
// Deprecated: import github.com/pulumi/pulumi/sdk/v3/go/esc/ast directly.

package ast

import (
	canonical "github.com/pulumi/pulumi/sdk/v3/go/esc/ast"
	"github.com/pulumi/pulumi/sdk/v3/go/esc/syntax"
)

// Number creates a new NumberExpr with the given value.
func Number[T syntax.NumberValue](value T) *NumberExpr { return canonical.Number(value) }
