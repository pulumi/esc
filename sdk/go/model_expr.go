/*
ESC (Environments, Secrets, Config) API

Pulumi ESC allows you to compose and manage hierarchical collections of configuration and secrets and consume them in various ways.

API version: 0.1.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package esc_sdk

import (
	"encoding/json"
)

// checks if the Expr type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &Expr{}

// Expr struct for Expr
type Expr struct {
	Range *Range `json:"range,omitempty"`
	Base *Expr `json:"base,omitempty"`
	Schema interface{} `json:"schema,omitempty"`
	KeyRanges *map[string]Range `json:"keyRanges,omitempty"`
	Literal interface{} `json:"literal,omitempty"`
	Interpolate []Interpolation `json:"interpolate,omitempty"`
	Symbol []PropertyAccessor `json:"symbol,omitempty"`
	Access []Access `json:"access,omitempty"`
	List []Expr `json:"list,omitempty"`
	Object *map[string]Expr `json:"object,omitempty"`
	Builtin *ExprBuiltin `json:"builtin,omitempty"`
}

// NewExpr instantiates a new Expr object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewExpr() *Expr {
	this := Expr{}
	return &this
}

// NewExprWithDefaults instantiates a new Expr object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewExprWithDefaults() *Expr {
	this := Expr{}
	return &this
}

// GetRange returns the Range field value if set, zero value otherwise.
func (o *Expr) GetRange() Range {
	if o == nil || IsNil(o.Range) {
		var ret Range
		return ret
	}
	return *o.Range
}

// GetRangeOk returns a tuple with the Range field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Expr) GetRangeOk() (*Range, bool) {
	if o == nil || IsNil(o.Range) {
		return nil, false
	}
	return o.Range, true
}

// HasRange returns a boolean if a field has been set.
func (o *Expr) HasRange() bool {
	if o != nil && !IsNil(o.Range) {
		return true
	}

	return false
}

// SetRange gets a reference to the given Range and assigns it to the Range field.
func (o *Expr) SetRange(v Range) {
	o.Range = &v
}

// GetBase returns the Base field value if set, zero value otherwise.
func (o *Expr) GetBase() Expr {
	if o == nil || IsNil(o.Base) {
		var ret Expr
		return ret
	}
	return *o.Base
}

// GetBaseOk returns a tuple with the Base field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Expr) GetBaseOk() (*Expr, bool) {
	if o == nil || IsNil(o.Base) {
		return nil, false
	}
	return o.Base, true
}

// HasBase returns a boolean if a field has been set.
func (o *Expr) HasBase() bool {
	if o != nil && !IsNil(o.Base) {
		return true
	}

	return false
}

// SetBase gets a reference to the given Expr and assigns it to the Base field.
func (o *Expr) SetBase(v Expr) {
	o.Base = &v
}

// GetSchema returns the Schema field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *Expr) GetSchema() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.Schema
}

// GetSchemaOk returns a tuple with the Schema field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *Expr) GetSchemaOk() (*interface{}, bool) {
	if o == nil || IsNil(o.Schema) {
		return nil, false
	}
	return &o.Schema, true
}

// HasSchema returns a boolean if a field has been set.
func (o *Expr) HasSchema() bool {
	if o != nil && IsNil(o.Schema) {
		return true
	}

	return false
}

// SetSchema gets a reference to the given interface{} and assigns it to the Schema field.
func (o *Expr) SetSchema(v interface{}) {
	o.Schema = v
}

// GetKeyRanges returns the KeyRanges field value if set, zero value otherwise.
func (o *Expr) GetKeyRanges() map[string]Range {
	if o == nil || IsNil(o.KeyRanges) {
		var ret map[string]Range
		return ret
	}
	return *o.KeyRanges
}

// GetKeyRangesOk returns a tuple with the KeyRanges field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Expr) GetKeyRangesOk() (*map[string]Range, bool) {
	if o == nil || IsNil(o.KeyRanges) {
		return nil, false
	}
	return o.KeyRanges, true
}

// HasKeyRanges returns a boolean if a field has been set.
func (o *Expr) HasKeyRanges() bool {
	if o != nil && !IsNil(o.KeyRanges) {
		return true
	}

	return false
}

// SetKeyRanges gets a reference to the given map[string]Range and assigns it to the KeyRanges field.
func (o *Expr) SetKeyRanges(v map[string]Range) {
	o.KeyRanges = &v
}

// GetLiteral returns the Literal field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *Expr) GetLiteral() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.Literal
}

// GetLiteralOk returns a tuple with the Literal field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *Expr) GetLiteralOk() (*interface{}, bool) {
	if o == nil || IsNil(o.Literal) {
		return nil, false
	}
	return &o.Literal, true
}

// HasLiteral returns a boolean if a field has been set.
func (o *Expr) HasLiteral() bool {
	if o != nil && IsNil(o.Literal) {
		return true
	}

	return false
}

// SetLiteral gets a reference to the given interface{} and assigns it to the Literal field.
func (o *Expr) SetLiteral(v interface{}) {
	o.Literal = v
}

// GetInterpolate returns the Interpolate field value if set, zero value otherwise.
func (o *Expr) GetInterpolate() []Interpolation {
	if o == nil || IsNil(o.Interpolate) {
		var ret []Interpolation
		return ret
	}
	return o.Interpolate
}

// GetInterpolateOk returns a tuple with the Interpolate field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Expr) GetInterpolateOk() ([]Interpolation, bool) {
	if o == nil || IsNil(o.Interpolate) {
		return nil, false
	}
	return o.Interpolate, true
}

// HasInterpolate returns a boolean if a field has been set.
func (o *Expr) HasInterpolate() bool {
	if o != nil && !IsNil(o.Interpolate) {
		return true
	}

	return false
}

// SetInterpolate gets a reference to the given []Interpolation and assigns it to the Interpolate field.
func (o *Expr) SetInterpolate(v []Interpolation) {
	o.Interpolate = v
}

// GetSymbol returns the Symbol field value if set, zero value otherwise.
func (o *Expr) GetSymbol() []PropertyAccessor {
	if o == nil || IsNil(o.Symbol) {
		var ret []PropertyAccessor
		return ret
	}
	return o.Symbol
}

// GetSymbolOk returns a tuple with the Symbol field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Expr) GetSymbolOk() ([]PropertyAccessor, bool) {
	if o == nil || IsNil(o.Symbol) {
		return nil, false
	}
	return o.Symbol, true
}

// HasSymbol returns a boolean if a field has been set.
func (o *Expr) HasSymbol() bool {
	if o != nil && !IsNil(o.Symbol) {
		return true
	}

	return false
}

// SetSymbol gets a reference to the given []PropertyAccessor and assigns it to the Symbol field.
func (o *Expr) SetSymbol(v []PropertyAccessor) {
	o.Symbol = v
}

// GetAccess returns the Access field value if set, zero value otherwise.
func (o *Expr) GetAccess() []Access {
	if o == nil || IsNil(o.Access) {
		var ret []Access
		return ret
	}
	return o.Access
}

// GetAccessOk returns a tuple with the Access field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Expr) GetAccessOk() ([]Access, bool) {
	if o == nil || IsNil(o.Access) {
		return nil, false
	}
	return o.Access, true
}

// HasAccess returns a boolean if a field has been set.
func (o *Expr) HasAccess() bool {
	if o != nil && !IsNil(o.Access) {
		return true
	}

	return false
}

// SetAccess gets a reference to the given []Access and assigns it to the Access field.
func (o *Expr) SetAccess(v []Access) {
	o.Access = v
}

// GetList returns the List field value if set, zero value otherwise.
func (o *Expr) GetList() []Expr {
	if o == nil || IsNil(o.List) {
		var ret []Expr
		return ret
	}
	return o.List
}

// GetListOk returns a tuple with the List field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Expr) GetListOk() ([]Expr, bool) {
	if o == nil || IsNil(o.List) {
		return nil, false
	}
	return o.List, true
}

// HasList returns a boolean if a field has been set.
func (o *Expr) HasList() bool {
	if o != nil && !IsNil(o.List) {
		return true
	}

	return false
}

// SetList gets a reference to the given []Expr and assigns it to the List field.
func (o *Expr) SetList(v []Expr) {
	o.List = v
}

// GetObject returns the Object field value if set, zero value otherwise.
func (o *Expr) GetObject() map[string]Expr {
	if o == nil || IsNil(o.Object) {
		var ret map[string]Expr
		return ret
	}
	return *o.Object
}

// GetObjectOk returns a tuple with the Object field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Expr) GetObjectOk() (*map[string]Expr, bool) {
	if o == nil || IsNil(o.Object) {
		return nil, false
	}
	return o.Object, true
}

// HasObject returns a boolean if a field has been set.
func (o *Expr) HasObject() bool {
	if o != nil && !IsNil(o.Object) {
		return true
	}

	return false
}

// SetObject gets a reference to the given map[string]Expr and assigns it to the Object field.
func (o *Expr) SetObject(v map[string]Expr) {
	o.Object = &v
}

// GetBuiltin returns the Builtin field value if set, zero value otherwise.
func (o *Expr) GetBuiltin() ExprBuiltin {
	if o == nil || IsNil(o.Builtin) {
		var ret ExprBuiltin
		return ret
	}
	return *o.Builtin
}

// GetBuiltinOk returns a tuple with the Builtin field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Expr) GetBuiltinOk() (*ExprBuiltin, bool) {
	if o == nil || IsNil(o.Builtin) {
		return nil, false
	}
	return o.Builtin, true
}

// HasBuiltin returns a boolean if a field has been set.
func (o *Expr) HasBuiltin() bool {
	if o != nil && !IsNil(o.Builtin) {
		return true
	}

	return false
}

// SetBuiltin gets a reference to the given ExprBuiltin and assigns it to the Builtin field.
func (o *Expr) SetBuiltin(v ExprBuiltin) {
	o.Builtin = &v
}

func (o Expr) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o Expr) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Range) {
		toSerialize["range"] = o.Range
	}
	if !IsNil(o.Base) {
		toSerialize["base"] = o.Base
	}
	if o.Schema != nil {
		toSerialize["schema"] = o.Schema
	}
	if !IsNil(o.KeyRanges) {
		toSerialize["keyRanges"] = o.KeyRanges
	}
	if o.Literal != nil {
		toSerialize["literal"] = o.Literal
	}
	if !IsNil(o.Interpolate) {
		toSerialize["interpolate"] = o.Interpolate
	}
	if !IsNil(o.Symbol) {
		toSerialize["symbol"] = o.Symbol
	}
	if !IsNil(o.Access) {
		toSerialize["access"] = o.Access
	}
	if !IsNil(o.List) {
		toSerialize["list"] = o.List
	}
	if !IsNil(o.Object) {
		toSerialize["object"] = o.Object
	}
	if !IsNil(o.Builtin) {
		toSerialize["builtin"] = o.Builtin
	}
	return toSerialize, nil
}

type NullableExpr struct {
	value *Expr
	isSet bool
}

func (v NullableExpr) Get() *Expr {
	return v.value
}

func (v *NullableExpr) Set(val *Expr) {
	v.value = val
	v.isSet = true
}

func (v NullableExpr) IsSet() bool {
	return v.isSet
}

func (v *NullableExpr) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableExpr(val *Expr) *NullableExpr {
	return &NullableExpr{value: val, isSet: true}
}

func (v NullableExpr) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableExpr) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


