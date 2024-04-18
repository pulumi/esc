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

package client

import (
	"fmt"
	"strings"
	"time"

	"github.com/pulumi/esc"
)

type EnvironmentDiagnostic struct {
	Range   *esc.Range `json:"range,omitempty"`
	Summary string     `json:"summary,omitempty"`
	Detail  string     `json:"detail,omitempty"`
}

type EnvironmentErrorResponse struct {
	Code        int                     `json:"code,omitempty"`
	Message     string                  `json:"message,omitempty"`
	Diagnostics []EnvironmentDiagnostic `json:"diagnostics,omitempty"`
}

func (err EnvironmentErrorResponse) Error() string {
	errString := fmt.Sprintf("[%d] %s", err.Code, err.Message)
	if len(err.Diagnostics) > 0 {
		errString += fmt.Sprintf("\nDiags: %s", diagsErrorString(err.Diagnostics))
	}
	return errString
}

type EnvironmentDiagnosticError struct {
	Diagnostics []EnvironmentDiagnostic `json:"diagnostics,omitempty"`
}

// Error implements the Error interface.
func (err EnvironmentDiagnosticError) Error() string {
	return diagsErrorString(err.Diagnostics)
}

func diagsErrorString(envDiags []EnvironmentDiagnostic) string {
	var diags strings.Builder
	for _, d := range envDiags {
		fmt.Fprintf(&diags, "%v\n", d.Summary)
	}
	return diags.String()
}

type EnvironmentRevision struct {
	Number       int       `json:"number"`
	Created      time.Time `json:"created"`
	CreatorLogin string    `json:"creatorLogin"`
	CreatorName  string    `json:"creatorName"`
	Tags         []string  `json:"tags"`
}

type CreateEnvironmentRevisionTagRequest struct {
	Revision *int `json:"revision,omitempty"`
}

type UpdateEnvironmentRevisionTagRequest struct {
	Revision *int `json:"revision,omitempty"`
}

type EnvironmentRevisionTag struct {
	Name        string    `json:"name"`
	Revision    int       `json:"revision"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
	EditorLogin string    `json:"editorLogin"`
	EditorName  string    `json:"editorName"`
}

type ListEnvironmentRevisionTagsResponse struct {
	Tags      []EnvironmentRevisionTag `json:"tags"`
	NextToken string                   `json:"nextToken"`
}
type OrgEnvironment struct {
	Organization string `json:"organization,omitempty"`
	Name         string `json:"name,omitempty"`
}

type ListEnvironmentsResponse struct {
	Environments []OrgEnvironment `json:"environments,omitempty"`
	NextToken    string           `json:"nextToken,omitempty"`
}

type UpdateEnvironmentResponse struct {
	EnvironmentDiagnosticError
}

type CheckEnvironmentResponse struct {
	Diagnostics []EnvironmentDiagnostic `json:"diagnostics,omitempty"`
}

type OpenEnvironmentResponse struct {
	ID          string                  `json:"id"`
	Diagnostics []EnvironmentDiagnostic `json:"diagnostics,omitempty"`
}
