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
