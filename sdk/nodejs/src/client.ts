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

import {
    EnvironmentDefinition,
    EnvironmentResponse,
    ListEnvironmentsResponse,
    OpenEnvironmentResponse,
    OrgEnvironment,
    ReadEnvironmentResponse,
    UpdateEnvironmentResponse,
} from ".";
import axios, { AxiosRequestConfig, AxiosInstance, AxiosError, HttpStatusCode } from "axios";
import { stringify } from "yaml";

const API_URL: string = "https://api.pulumi.com/api";

export abstract class ESCApiClient {
    abstract createEnvironment(name: string): Promise<null>;
    abstract deleteEnvironment(name: string): Promise<null>;
    abstract readEnvironment(name: string): Promise<ReadEnvironmentResponse>;
    abstract checkEnvironment(name: string, tag?: string): Promise<EnvironmentResponse>;
    abstract openEnvironment(name: string): Promise<OpenEnvironmentResponse>;
    abstract updateEnvironment(
        name: string,
        body: EnvironmentDefinition,
        tag?: string,
    ): Promise<UpdateEnvironmentResponse>;
    abstract readOpenEnvironment(name: string, openSessionId: string): Promise<EnvironmentResponse>;
    abstract listEnvironments(): Promise<OrgEnvironment[]>;
}

export class ESC implements ESCApiClient {
    private readonly apiUrl: string;
    private readonly token: string;
    private readonly org: string;
    private readonly http: AxiosInstance;

    constructor(token: string, org: string, apiURL?: string) {
        this.token = token;
        this.org = org;
        this.apiUrl = apiURL || API_URL;
        this.http = axios.create(this.config);
    }

    private get config(): AxiosRequestConfig {
        return {
            headers: new axios.AxiosHeaders({
                Authorization: `token ${this.token}`,
                "Content-Type": "application/json",
                "X-Pulumi-Source": "esc-sdk-nodejs",
            }),
            baseURL: `${this.apiUrl}/preview`,
        };
    }

    public async createEnvironment(name: string): Promise<null> {
        return this.http.post(`environments/${this.org}/${name}`, {});
    }

    public async listEnvironments(): Promise<OrgEnvironment[]> {
        const response = await this.http.get<ListEnvironmentsResponse>(`environments/${this.org}`);
        return response.data.environments;
    }

    public async readEnvironment(name: string, tag?: string): Promise<ReadEnvironmentResponse> {
        const options: AxiosRequestConfig = {};
        if (tag) {
            options.headers = {
                "If-Match": tag,
            };
        }
        const resp = await this.http.get<string>(`environments/${this.org}/${name}`, options);
        const result: ReadEnvironmentResponse = {
            environmentString: resp.data,
        };
        if (resp.headers.etag) {
            result.tag = resp.headers.etag;
        }

        return result;
    }

    public async checkEnvironment(name: string): Promise<EnvironmentResponse> {
        const response = await this.http.post<EnvironmentResponse>(`environments/${this.org}/${name}/check`);
        return response.data;
    }

    public async openEnvironment(name: string): Promise<OpenEnvironmentResponse> {
        const response = await this.http.post<OpenEnvironmentResponse>(`environments/${this.org}/${name}/open`, {});
        return response.data;
    }

    public async readOpenEnvironment(name: string, openSessionId: string): Promise<EnvironmentResponse> {
        const response = await this.http.get<EnvironmentResponse>(
            `environments/${this.org}/${name}/open/${openSessionId}`,
        );
        return response.data;
    }

    public async updateEnvironment(
        name: string,
        def: EnvironmentDefinition,
        tag?: string,
    ): Promise<UpdateEnvironmentResponse> {
        const envYaml = stringify(def);
        const options: AxiosRequestConfig = {
            headers: {
                "Content-Type": "application/x-yaml",
            },
        };
        if (tag) {
            options.headers!["If-Match"] = tag;
        }

        const diags = await this.http
            .patch(`environments/${this.org}/${name}`, envYaml, options)
            .catch((err: AxiosError) => {
                if (err.response?.status === HttpStatusCode.BadRequest) {
                    return err.response?.data;
                }
                throw err;
            });

        return diags as UpdateEnvironmentResponse;
    }

    public async deleteEnvironment(name: string): Promise<null> {
        return this.http.delete(`environments/${this.org}/${name}`);
    }
}
