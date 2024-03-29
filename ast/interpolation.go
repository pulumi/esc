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
	"strings"

	"github.com/pulumi/esc/syntax"
)

type Interpolation struct {
	Text  string
	Value *PropertyAccess
}

func parseInterpolate(node syntax.Node, value string) ([]Interpolation, syntax.Diagnostics) {
	var parts []Interpolation
	var str strings.Builder
	var diags syntax.Diagnostics
	offset := 0
	for len(value) > 0 {
		switch {
		case strings.HasPrefix(value, "$$"):
			str.WriteByte('$')
			value, offset = value[2:], offset+2
		case strings.HasPrefix(value, "${"):
			end, rest, access, accessDiags := parsePropertyAccess(node, offset+2, value[2:])

			diags.Extend(accessDiags...)
			parts = append(parts, Interpolation{
				Text:  str.String(),
				Value: access,
			})
			str.Reset()

			value, offset = rest, end
		default:
			str.WriteByte(value[0])
			value, offset = value[1:], offset+1
		}
	}
	if str.Len() != 0 {
		parts = append(parts, Interpolation{Text: str.String()})
	}
	return parts, diags
}
