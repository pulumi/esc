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

// checks if the Access type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &Access{}

// Access struct for Access
type Access struct {
	Receiver *Range `json:"receiver,omitempty"`
	Accessors []Accessor `json:"accessors,omitempty"`
}

// NewAccess instantiates a new Access object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewAccess() *Access {
	this := Access{}
	return &this
}

// NewAccessWithDefaults instantiates a new Access object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewAccessWithDefaults() *Access {
	this := Access{}
	return &this
}

// GetReceiver returns the Receiver field value if set, zero value otherwise.
func (o *Access) GetReceiver() Range {
	if o == nil || IsNil(o.Receiver) {
		var ret Range
		return ret
	}
	return *o.Receiver
}

// GetReceiverOk returns a tuple with the Receiver field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Access) GetReceiverOk() (*Range, bool) {
	if o == nil || IsNil(o.Receiver) {
		return nil, false
	}
	return o.Receiver, true
}

// HasReceiver returns a boolean if a field has been set.
func (o *Access) HasReceiver() bool {
	if o != nil && !IsNil(o.Receiver) {
		return true
	}

	return false
}

// SetReceiver gets a reference to the given Range and assigns it to the Receiver field.
func (o *Access) SetReceiver(v Range) {
	o.Receiver = &v
}

// GetAccessors returns the Accessors field value if set, zero value otherwise.
func (o *Access) GetAccessors() []Accessor {
	if o == nil || IsNil(o.Accessors) {
		var ret []Accessor
		return ret
	}
	return o.Accessors
}

// GetAccessorsOk returns a tuple with the Accessors field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Access) GetAccessorsOk() ([]Accessor, bool) {
	if o == nil || IsNil(o.Accessors) {
		return nil, false
	}
	return o.Accessors, true
}

// HasAccessors returns a boolean if a field has been set.
func (o *Access) HasAccessors() bool {
	if o != nil && !IsNil(o.Accessors) {
		return true
	}

	return false
}

// SetAccessors gets a reference to the given []Accessor and assigns it to the Accessors field.
func (o *Access) SetAccessors(v []Accessor) {
	o.Accessors = v
}

func (o Access) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o Access) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Receiver) {
		toSerialize["receiver"] = o.Receiver
	}
	if !IsNil(o.Accessors) {
		toSerialize["accessors"] = o.Accessors
	}
	return toSerialize, nil
}

type NullableAccess struct {
	value *Access
	isSet bool
}

func (v NullableAccess) Get() *Access {
	return v.value
}

func (v *NullableAccess) Set(val *Access) {
	v.value = val
	v.isSet = true
}

func (v NullableAccess) IsSet() bool {
	return v.isSet
}

func (v *NullableAccess) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableAccess(val *Access) *NullableAccess {
	return &NullableAccess{value: val, isSet: true}
}

func (v NullableAccess) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableAccess) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


