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
	"fmt"
)

// checks if the EnvironmentDiagnostic type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &EnvironmentDiagnostic{}

// EnvironmentDiagnostic struct for EnvironmentDiagnostic
type EnvironmentDiagnostic struct {
	Summary string `json:"summary"`
	Path *string `json:"path,omitempty"`
	Range *Range `json:"range,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _EnvironmentDiagnostic EnvironmentDiagnostic

// NewEnvironmentDiagnostic instantiates a new EnvironmentDiagnostic object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewEnvironmentDiagnostic(summary string) *EnvironmentDiagnostic {
	this := EnvironmentDiagnostic{}
	this.Summary = summary
	return &this
}

// NewEnvironmentDiagnosticWithDefaults instantiates a new EnvironmentDiagnostic object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewEnvironmentDiagnosticWithDefaults() *EnvironmentDiagnostic {
	this := EnvironmentDiagnostic{}
	return &this
}

// GetSummary returns the Summary field value
func (o *EnvironmentDiagnostic) GetSummary() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Summary
}

// GetSummaryOk returns a tuple with the Summary field value
// and a boolean to check if the value has been set.
func (o *EnvironmentDiagnostic) GetSummaryOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Summary, true
}

// SetSummary sets field value
func (o *EnvironmentDiagnostic) SetSummary(v string) {
	o.Summary = v
}

// GetPath returns the Path field value if set, zero value otherwise.
func (o *EnvironmentDiagnostic) GetPath() string {
	if o == nil || IsNil(o.Path) {
		var ret string
		return ret
	}
	return *o.Path
}

// GetPathOk returns a tuple with the Path field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *EnvironmentDiagnostic) GetPathOk() (*string, bool) {
	if o == nil || IsNil(o.Path) {
		return nil, false
	}
	return o.Path, true
}

// HasPath returns a boolean if a field has been set.
func (o *EnvironmentDiagnostic) HasPath() bool {
	if o != nil && !IsNil(o.Path) {
		return true
	}

	return false
}

// SetPath gets a reference to the given string and assigns it to the Path field.
func (o *EnvironmentDiagnostic) SetPath(v string) {
	o.Path = &v
}

// GetRange returns the Range field value if set, zero value otherwise.
func (o *EnvironmentDiagnostic) GetRange() Range {
	if o == nil || IsNil(o.Range) {
		var ret Range
		return ret
	}
	return *o.Range
}

// GetRangeOk returns a tuple with the Range field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *EnvironmentDiagnostic) GetRangeOk() (*Range, bool) {
	if o == nil || IsNil(o.Range) {
		return nil, false
	}
	return o.Range, true
}

// HasRange returns a boolean if a field has been set.
func (o *EnvironmentDiagnostic) HasRange() bool {
	if o != nil && !IsNil(o.Range) {
		return true
	}

	return false
}

// SetRange gets a reference to the given Range and assigns it to the Range field.
func (o *EnvironmentDiagnostic) SetRange(v Range) {
	o.Range = &v
}

func (o EnvironmentDiagnostic) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o EnvironmentDiagnostic) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["summary"] = o.Summary
	if !IsNil(o.Path) {
		toSerialize["path"] = o.Path
	}
	if !IsNil(o.Range) {
		toSerialize["range"] = o.Range
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *EnvironmentDiagnostic) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"summary",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err;
	}

	for _, requiredProperty := range(requiredProperties) {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varEnvironmentDiagnostic := _EnvironmentDiagnostic{}

	err = json.Unmarshal(data, &varEnvironmentDiagnostic)

	if err != nil {
		return err
	}

	*o = EnvironmentDiagnostic(varEnvironmentDiagnostic)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "summary")
		delete(additionalProperties, "path")
		delete(additionalProperties, "range")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableEnvironmentDiagnostic struct {
	value *EnvironmentDiagnostic
	isSet bool
}

func (v NullableEnvironmentDiagnostic) Get() *EnvironmentDiagnostic {
	return v.value
}

func (v *NullableEnvironmentDiagnostic) Set(val *EnvironmentDiagnostic) {
	v.value = val
	v.isSet = true
}

func (v NullableEnvironmentDiagnostic) IsSet() bool {
	return v.isSet
}

func (v *NullableEnvironmentDiagnostic) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnvironmentDiagnostic(val *EnvironmentDiagnostic) *NullableEnvironmentDiagnostic {
	return &NullableEnvironmentDiagnostic{value: val, isSet: true}
}

func (v NullableEnvironmentDiagnostic) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnvironmentDiagnostic) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


