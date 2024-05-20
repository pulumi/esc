// Copyright 2024, Pulumi Corporation.  All rights reserved.
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

// checks if the EnvironmentDefinition type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &EnvironmentDefinition{}

// EnvironmentDefinition struct for EnvironmentDefinition
type EnvironmentDefinition struct {
	Imports []string `json:"imports,omitempty"`
	Values *EnvironmentDefinitionValues `json:"values,omitempty"`
}

// NewEnvironmentDefinition instantiates a new EnvironmentDefinition object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewEnvironmentDefinition() *EnvironmentDefinition {
	this := EnvironmentDefinition{}
	return &this
}

// NewEnvironmentDefinitionWithDefaults instantiates a new EnvironmentDefinition object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewEnvironmentDefinitionWithDefaults() *EnvironmentDefinition {
	this := EnvironmentDefinition{}
	return &this
}

// GetImports returns the Imports field value if set, zero value otherwise.
func (o *EnvironmentDefinition) GetImports() []string {
	if o == nil || IsNil(o.Imports) {
		var ret []string
		return ret
	}
	return o.Imports
}

// GetImportsOk returns a tuple with the Imports field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *EnvironmentDefinition) GetImportsOk() ([]string, bool) {
	if o == nil || IsNil(o.Imports) {
		return nil, false
	}
	return o.Imports, true
}

// HasImports returns a boolean if a field has been set.
func (o *EnvironmentDefinition) HasImports() bool {
	if o != nil && !IsNil(o.Imports) {
		return true
	}

	return false
}

// SetImports gets a reference to the given []string and assigns it to the Imports field.
func (o *EnvironmentDefinition) SetImports(v []string) {
	o.Imports = v
}

// GetValues returns the Values field value if set, zero value otherwise.
func (o *EnvironmentDefinition) GetValues() EnvironmentDefinitionValues {
	if o == nil || IsNil(o.Values) {
		var ret EnvironmentDefinitionValues
		return ret
	}
	return *o.Values
}

// GetValuesOk returns a tuple with the Values field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *EnvironmentDefinition) GetValuesOk() (*EnvironmentDefinitionValues, bool) {
	if o == nil || IsNil(o.Values) {
		return nil, false
	}
	return o.Values, true
}

// HasValues returns a boolean if a field has been set.
func (o *EnvironmentDefinition) HasValues() bool {
	if o != nil && !IsNil(o.Values) {
		return true
	}

	return false
}

// SetValues gets a reference to the given EnvironmentDefinitionValues and assigns it to the Values field.
func (o *EnvironmentDefinition) SetValues(v EnvironmentDefinitionValues) {
	o.Values = &v
}

func (o EnvironmentDefinition) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o EnvironmentDefinition) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Imports) {
		toSerialize["imports"] = o.Imports
	}
	if !IsNil(o.Values) {
		toSerialize["values"] = o.Values
	}
	return toSerialize, nil
}

type NullableEnvironmentDefinition struct {
	value *EnvironmentDefinition
	isSet bool
}

func (v NullableEnvironmentDefinition) Get() *EnvironmentDefinition {
	return v.value
}

func (v *NullableEnvironmentDefinition) Set(val *EnvironmentDefinition) {
	v.value = val
	v.isSet = true
}

func (v NullableEnvironmentDefinition) IsSet() bool {
	return v.isSet
}

func (v *NullableEnvironmentDefinition) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnvironmentDefinition(val *EnvironmentDefinition) *NullableEnvironmentDefinition {
	return &NullableEnvironmentDefinition{value: val, isSet: true}
}

func (v NullableEnvironmentDefinition) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnvironmentDefinition) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


