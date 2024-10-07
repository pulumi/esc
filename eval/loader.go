// Copyright 2024, Pulumi Corporation.
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

package eval

import (
	"context"

	"github.com/pulumi/esc/ast"
	"github.com/pulumi/esc/syntax"
)

type loadedEnvironment struct {
	done <-chan bool

	name  string
	env   *ast.EnvironmentDecl
	dec   Decrypter
	diags syntax.Diagnostics
	err   error
}

type loader struct {
	ctx          context.Context
	environments EnvironmentLoader
	loaded       map[string]*loadedEnvironment
}

func newLoader(ctx context.Context, environments EnvironmentLoader) *loader {
	return &loader{
		ctx:          ctx,
		environments: environments,
		loaded:       map[string]*loadedEnvironment{},
	}
}

func (l *loader) load(name string) *loadedEnvironment {
	if loaded, ok := l.loaded[name]; ok {
		return loaded
	}

	done := make(chan bool)
	result := &loadedEnvironment{done: done, name: name}
	go func() {
		defer close(done)

		bytes, dec, err := l.environments.LoadEnvironment(l.ctx, name)
		if err != nil {
			result.err = err
			return
		}
		result.dec = dec

		result.env, result.diags, result.err = LoadYAMLBytes(name, bytes)
		return
	}()

	l.loaded[name] = result
	return result
}
