// Copyright 2023, Pulumi Corporation.  All rights reserved.

package schema

type NullBuilder struct {
	s Schema
}

func Null() *NullBuilder {
	return &NullBuilder{}
}

func (b *NullBuilder) Ref(ref string) *NullBuilder {
	return buildRef(b, ref)
}

func (b *NullBuilder) AnyOf(anyOf ...Builder) *NullBuilder {
	return buildAnyOf(b, anyOf)
}

func (b *NullBuilder) OneOf(oneOf ...Builder) *NullBuilder {
	return buildOneOf(b, oneOf)
}

func (b *NullBuilder) Title(title string) *NullBuilder {
	b.s.Title = title
	return b
}

func (b *NullBuilder) Description(description string) *NullBuilder {
	b.s.Description = description
	return b
}

func (b *NullBuilder) Deprecated(deprecated bool) *NullBuilder {
	b.s.Deprecated = deprecated
	return b
}

func (b *NullBuilder) Schema() *Schema {
	b.s.Type = "null"
	return &b.s
}
