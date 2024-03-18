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

export interface OrgEnvironment {
    organization: string;
    name: string;
    created: string;
    modified: string;
}

export interface ListEnvironmentsResponse {
    environments: OrgEnvironment[];
    nextToken?: string;
}

export interface EnvironmentResponse {
    diagnostics?: EnvironmentDiagnostic[];
    exprs: Record<string, object>; // TODO: add a type for expr rather than a raw object.
    properties: Record<string, EnvironmentValue>;
    schema: object; // TODO: add a type for schema rather than a raw object.
}

export interface EnvironmentValue {
    value: any;
    secret?: boolean;
    unknown?: boolean;
    trace?: EnvironmentTrace;
}

export interface EnvironmentTrace {
    def: EnvironmentRange;
    base: EnvironmentValue;
}

export interface EnvironmentRange {
    environment: string;
    begin: EnvironmentPosition;
    end: EnvironmentPosition;
}

export interface OpenEnvironmentResponse {
    id: string;
    diagnostics?: EnvironmentDiagnostic[];
}

export interface EnvironmentPosition {
    line: number;
    column: number;
    byte: number;
}

export interface EnvironmentDiagnosticRange {
    environment?: string;
    begin: EnvironmentPosition;
    end: EnvironmentPosition;
}

export interface EnvironmentDiagnostic {
    range?: EnvironmentDiagnosticRange;
    summary?: string;
    path?: string;
}

export interface UpdateEnvironmentResponse {
    diagnostics?: EnvironmentDiagnostic[];
}

export interface ReadEnvironmentResponse {
    tag?: string;
    environmentString: string;
}

export class EnvironmentDefinition {
    // imports is a list of environments that will be imported and merged into the environment being defined.
    imports?: string[];
    // values is a map of values that will be defined within the environment.
    values?: EnvironmentDefinitionValues;

    constructor(env: EnvironmentDefinitionArgs) {
        if (env.imports && env.imports.length > 0) {
            this.imports = env.imports;
        }
        if (env.values) {
            this.values = env.values;
        }
    }
}

export interface EnvironmentDefinitionArgs {
    imports?: string[];
    values?: EnvironmentDefinitionValues;
}

export interface EnvironmentDefinitionValues {
    // pulumiConfig is a map of config keys to config values that will be passed to the Pulumi program.
    pulumiConfig?: Record<string, any>;
    // environmentVariables is a map of environment variable names to values that will be set in the terminal if running
    // `esc run` or `pulumi env run` or if running a Pulumi program.
    environmentVariables?: Record<string, string>;
    // All other values that are to be defined within the environment but not exported via either the Pulumi config or
    // environment variables.
    [key: string]: any;
}
