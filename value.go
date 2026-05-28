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

package esc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pulumi/esc/internal/util"
	"golang.org/x/exp/maps"
)

// ValueType defines the types of concrete values stored inside a Value.
type ValueType interface {
	bool | json.Number | string | []Value | map[string]Value
}

// A Value is the result of evaluating an expression within an environment definition.
type Value struct {
	// Value holds the concrete representation of the value. May be nil, bool, json.Number, string, []Value, or
	// map[string]Value.
	Value any `json:"value,omitempty"`

	// Secret is true if this value is secret.
	Secret bool `json:"secret,omitempty"`

	// Unknown is true if this value is unknown.
	Unknown bool `json:"unknown,omitempty"`

	// Trace holds information about the expression that computed this value and the value (if any) with which it was
	// merged.
	Trace Trace `json:"trace"`
}

// NewValue creates a new value with the given representation.
func NewValue[T ValueType](v T) Value {
	return Value{Value: v}
}

// NewSecret creates a new secret value with the given representation.
func NewSecret[T ValueType](v T) Value {
	switch v := any(v).(type) {
	case map[string]Value:
		for k, e := range v {
			if !e.Secret {
				v[k] = e.MakeSecret()
			}
		}
	case []Value:
		for i, e := range v {
			if !e.Secret {
				v[i] = e.MakeSecret()
			}
		}
	}
	return Value{Value: v, Secret: true}
}

func (v Value) MakeSecret() Value {
	switch vv := v.Value.(type) {
	case nil:
		copy := v
		copy.Secret = true
		return copy
	case bool:
		return NewSecret(vv)
	case json.Number:
		return NewSecret(vv)
	case string:
		return NewSecret(vv)
	case []Value:
		return NewSecret(vv)
	case map[string]Value:
		return NewSecret(vv)
	default:
		panic("invalid value")
	}
}

// Trace holds information about the expression and base of a value.
type Trace struct {
	// Def is the range of the expression that computed a value.
	Def Range `json:"def"`

	// Base is the base value with which a value was merged.
	Base *Value `json:"base,omitempty"`
}

// UnmarshalJSON decodes a Value in a single pass over the input bytes,
// reusing one json.Decoder across the entire subtree.
//
// The previous implementation json.Unmarshal'd into a RawMessage and then
// re-Unmarshal'd the captured "value" subtree, scanning each byte twice per
// nesting level — O(size × depth) on deep import-merge chains. Calling
// dec.Decode on each child has the same effect because json allocates a
// fresh Decoder when it dispatches into UnmarshalJSON. So children and
// Trace.Base are decoded by calling decodeFrom directly on the shared
// decoder, which keeps each level's bytes scanned exactly once.
//
// When json/v2 (go-json-experiment/json) stabilizes or lands in stdlib, this
// streaming logic can collapse into an UnmarshalJSONFrom(*jsontext.Decoder)
// implementation — v2 exposes the shared-decoder hook this code hand-rolls.
func (v *Value) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	return v.decodeFrom(dec)
}

// decodeFrom consumes one JSON object from dec into v.
func (v *Value) decodeFrom(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return fmt.Errorf("esc.Value: expected JSON object, got %v", tok)
	}
	return v.decodeObjectBody(dec)
}

// decodeObjectBody parses field-by-field until the matching '}', which it
// consumes. The opening '{' must already have been consumed by the caller —
// splitting it this way lets callers that have to peek the opening token
// (e.g. to distinguish object vs null) still hand the body off to this method.
func (v *Value) decodeObjectBody(dec *json.Decoder) error {
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		key, ok := keyTok.(string)
		if !ok {
			return fmt.Errorf("esc.Value: expected string field name, got %v", keyTok)
		}

		switch key {
		case "secret":
			if err := dec.Decode(&v.Secret); err != nil {
				return err
			}
		case "unknown":
			if err := dec.Decode(&v.Unknown); err != nil {
				return err
			}
		case "trace":
			if err := v.Trace.decodeFrom(dec); err != nil {
				return err
			}
		case "value":
			if err := v.decodeValueField(dec); err != nil {
				return err
			}
		default:
			// Match stdlib permissiveness: skip unknown fields so the wire
			// format can grow without breaking older readers.
			var skip json.RawMessage
			if err := dec.Decode(&skip); err != nil {
				return err
			}
		}
	}
	_, err := dec.Token() // consume '}'
	return err
}

// decodeValueField decodes the contents of the "value" JSON field. The
// field's concrete type is data-driven, so we peek the first token to pick a
// path rather than capturing the bytes and re-parsing them later.
func (v *Value) decodeValueField(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}

	delim, isDelim := tok.(json.Delim)
	if !isDelim {
		// Scalar: nil, bool, json.Number, or string.
		v.Value = tok
		return nil
	}

	switch delim {
	case '[':
		var arr []Value
		for dec.More() {
			var el Value
			if err := el.decodeFrom(dec); err != nil {
				return err
			}
			arr = append(arr, el)
		}
		if _, err := dec.Token(); err != nil { // consume ']'
			return err
		}
		v.Value = arr
	case '{':
		obj := map[string]Value{}
		for dec.More() {
			keyTok, err := dec.Token()
			if err != nil {
				return err
			}
			key, ok := keyTok.(string)
			if !ok {
				return fmt.Errorf("esc.Value: expected string key, got %v", keyTok)
			}
			var el Value
			if err := el.decodeFrom(dec); err != nil {
				return err
			}
			obj[key] = el
		}
		if _, err := dec.Token(); err != nil { // consume '}'
			return err
		}
		v.Value = obj
	default:
		return fmt.Errorf("esc.Value: unexpected delimiter %v", delim)
	}
	return nil
}

// decodeFrom decodes a Trace object from dec. We do this by hand rather than
// dec.Decode(&Trace) so that Trace.Base stays on the shared-decoder path
// instead of triggering a fresh NewDecoder allocation inside
// Value.UnmarshalJSON.
func (t *Trace) decodeFrom(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return fmt.Errorf("esc.Trace: expected JSON object, got %v", tok)
	}

	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		key, ok := keyTok.(string)
		if !ok {
			return fmt.Errorf("esc.Trace: expected string field name, got %v", keyTok)
		}

		switch key {
		case "def":
			if err := dec.Decode(&t.Def); err != nil {
				return err
			}
		case "base":
			// Peek so the object case can route through the shared decoder.
			tok, err := dec.Token()
			if err != nil {
				return err
			}
			if tok == nil {
				t.Base = nil
				continue
			}
			d, ok := tok.(json.Delim)
			if !ok || d != '{' {
				return fmt.Errorf("esc.Trace.Base: expected null or JSON object, got %v", tok)
			}
			base := &Value{}
			if err := base.decodeObjectBody(dec); err != nil {
				return err
			}
			t.Base = base
		default:
			var skip json.RawMessage
			if err := dec.Decode(&skip); err != nil {
				return err
			}
		}
	}
	_, err = dec.Token() // consume '}'
	return err
}

// FromJSON converts a plain-old-JSON value (i.e. a value of type nil, bool, json.Number, string, []any, or
// map[string]any) into a Value.
func FromJSON(v any) (Value, error) {
	return fromJSON("", v)
}

func fromJSON(path string, v any) (Value, error) {
	switch v := v.(type) {
	case nil:
		return Value{}, nil
	case bool:
		return NewValue(v), nil
	case json.Number:
		return NewValue(v), nil
	case string:
		return NewValue(v), nil
	case []any:
		vs := make([]Value, len(v))
		for i, v := range v {
			vv, err := fromJSON(fmt.Sprintf("[%v]", i), v)
			if err != nil {
				return Value{}, err
			}
			vs[i] = vv
		}
		return NewValue(vs), nil
	case map[string]any:
		keys := maps.Keys(v)
		sort.Strings(keys)
		vs := make(map[string]Value, len(keys))
		for _, k := range keys {
			vv, err := fromJSON(util.JoinKey(path, k), v[k])
			if err != nil {
				return Value{}, err
			}
			vs[k] = vv
		}
		return NewValue(vs), nil
	default:
		return Value{}, fmt.Errorf("%v: unsupported value of type %T", path, v)
	}
}

// ToJSON converts a Value into a plain-old-JSON value (i.e. a value of type nil, bool, json.Number, string, []any, or
// map[string]any). If redact is true, secrets are replaced with [secret].
func (v Value) ToJSON(redact bool) any {
	if v.Secret && redact {
		return "[secret]"
	}
	if v.Unknown {
		return "[unknown]"
	}

	switch pv := v.Value.(type) {
	case []Value:
		a := make([]any, len(pv))
		for i, v := range pv {
			a[i] = v.ToJSON(redact)
		}
		return a
	case map[string]Value:
		m := make(map[string]any, len(pv))
		for k, v := range pv {
			m[k] = v.ToJSON(redact)
		}
		return m
	default:
		return pv
	}
}

// ToString returns the string representation of this value. If redact is true, secrets are replaced with [secret].
func (v Value) ToString(redact bool) string {
	if v.Secret && redact {
		return "[secret]"
	}
	if v.Unknown {
		return "[unknown]"
	}

	switch pv := v.Value.(type) {
	case bool:
		if pv {
			return "true"
		}
		return "false"
	case json.Number:
		return pv.String()
	case string:
		return pv
	case []Value:
		vals := make([]string, len(pv))
		for i, v := range pv {
			vals[i] = strconv.Quote(v.ToString(redact))
		}
		return strings.Join(vals, ",")
	case map[string]Value:
		keys := maps.Keys(pv)
		sort.Strings(keys)

		pairs := make([]string, len(pv))
		for i, k := range keys {
			pairs[i] = fmt.Sprintf("%q=%q", k, pv[k].ToString(redact))
		}
		return strings.Join(pairs, ",")
	default:
		return ""
	}
}

// String is shorthand for ToString(true).
func (v Value) String() string {
	return v.ToString(true)
}
