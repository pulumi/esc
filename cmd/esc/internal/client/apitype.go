// Copyright 2023, Pulumi Corporation.  All rights reserved.

package client

import (
	"fmt"
	"strings"

	"github.com/pulumi/esc"
)

type EnvironmentDiagnostic struct {
	Range   *esc.Range `json:"range,omitempty"`
	Summary string     `json:"summary,omitempty"`
	Detail  string     `json:"detail,omitempty"`
}

type EnvironmentDiagnosticsResponse struct {
	Diagnostics []EnvironmentDiagnostic `json:"diagnostics,omitempty"`
}

// Error implements the Error interface.
func (err EnvironmentDiagnosticsResponse) Error() string {
	var diags strings.Builder
	for _, d := range err.Diagnostics {
		fmt.Fprintf(&diags, "%v\n", d.Summary)
	}
	return diags.String()
}

type ListEnvironmentsResponse struct {
	Environments []string `json:"environments,omitempty"`
}

type UpdateEnvironmentResponse struct {
	EnvironmentDiagnosticsResponse
}

type CheckEnvironmentResponse struct {
	Diagnostics []EnvironmentDiagnostic `json:"diagnostics,omitempty"`
}

type OpenEnvironmentResponse struct {
	ID          string                  `json:"id"`
	Diagnostics []EnvironmentDiagnostic `json:"diagnostics,omitempty"`
}
