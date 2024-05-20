/*
ESC (Environments, Secrets, Config) API

Pulumi ESC allows you to compose and manage hierarchical collections of configuration and secrets and consume them in various ways.

API version: 0.1.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package esc_sdk

import (
	"encoding/json"
	"bytes"
	"fmt"
)

// checks if the OrgEnvironment type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &OrgEnvironment{}

// OrgEnvironment struct for OrgEnvironment
type OrgEnvironment struct {
	Organization *string `json:"organization,omitempty"`
	Name string `json:"name"`
	Created string `json:"created"`
	Modified string `json:"modified"`
}

type _OrgEnvironment OrgEnvironment

// NewOrgEnvironment instantiates a new OrgEnvironment object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewOrgEnvironment(name string, created string, modified string) *OrgEnvironment {
	this := OrgEnvironment{}
	this.Name = name
	this.Created = created
	this.Modified = modified
	return &this
}

// NewOrgEnvironmentWithDefaults instantiates a new OrgEnvironment object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewOrgEnvironmentWithDefaults() *OrgEnvironment {
	this := OrgEnvironment{}
	return &this
}

// GetOrganization returns the Organization field value if set, zero value otherwise.
func (o *OrgEnvironment) GetOrganization() string {
	if o == nil || IsNil(o.Organization) {
		var ret string
		return ret
	}
	return *o.Organization
}

// GetOrganizationOk returns a tuple with the Organization field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *OrgEnvironment) GetOrganizationOk() (*string, bool) {
	if o == nil || IsNil(o.Organization) {
		return nil, false
	}
	return o.Organization, true
}

// HasOrganization returns a boolean if a field has been set.
func (o *OrgEnvironment) HasOrganization() bool {
	if o != nil && !IsNil(o.Organization) {
		return true
	}

	return false
}

// SetOrganization gets a reference to the given string and assigns it to the Organization field.
func (o *OrgEnvironment) SetOrganization(v string) {
	o.Organization = &v
}

// GetName returns the Name field value
func (o *OrgEnvironment) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *OrgEnvironment) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *OrgEnvironment) SetName(v string) {
	o.Name = v
}

// GetCreated returns the Created field value
func (o *OrgEnvironment) GetCreated() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Created
}

// GetCreatedOk returns a tuple with the Created field value
// and a boolean to check if the value has been set.
func (o *OrgEnvironment) GetCreatedOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Created, true
}

// SetCreated sets field value
func (o *OrgEnvironment) SetCreated(v string) {
	o.Created = v
}

// GetModified returns the Modified field value
func (o *OrgEnvironment) GetModified() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Modified
}

// GetModifiedOk returns a tuple with the Modified field value
// and a boolean to check if the value has been set.
func (o *OrgEnvironment) GetModifiedOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Modified, true
}

// SetModified sets field value
func (o *OrgEnvironment) SetModified(v string) {
	o.Modified = v
}

func (o OrgEnvironment) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o OrgEnvironment) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Organization) {
		toSerialize["organization"] = o.Organization
	}
	toSerialize["name"] = o.Name
	toSerialize["created"] = o.Created
	toSerialize["modified"] = o.Modified
	return toSerialize, nil
}

func (o *OrgEnvironment) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"name",
		"created",
		"modified",
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

	varOrgEnvironment := _OrgEnvironment{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varOrgEnvironment)

	if err != nil {
		return err
	}

	*o = OrgEnvironment(varOrgEnvironment)

	return err
}

type NullableOrgEnvironment struct {
	value *OrgEnvironment
	isSet bool
}

func (v NullableOrgEnvironment) Get() *OrgEnvironment {
	return v.value
}

func (v *NullableOrgEnvironment) Set(val *OrgEnvironment) {
	v.value = val
	v.isSet = true
}

func (v NullableOrgEnvironment) IsSet() bool {
	return v.isSet
}

func (v *NullableOrgEnvironment) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableOrgEnvironment(val *OrgEnvironment) *NullableOrgEnvironment {
	return &NullableOrgEnvironment{value: val, isSet: true}
}

func (v NullableOrgEnvironment) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableOrgEnvironment) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

