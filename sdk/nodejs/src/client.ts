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
    abstract createEnvironment(org: string, name: string): Promise<void>;
    abstract deleteEnvironment(org: string, name: string): Promise<void>;
    abstract readEnvironment(org: string, name: string): Promise<ReadEnvironmentResponse>;
    abstract checkEnvironment(org: string, name: string, tag?: string): Promise<EnvironmentResponse>;
    abstract openEnvironment(org: string, name: string): Promise<OpenEnvironmentResponse>;
    abstract updateEnvironment(
        org: string,
        name: string,
        body: EnvironmentDefinition,
        tag?: string,
    ): Promise<UpdateEnvironmentResponse>;
    abstract readOpenEnvironment(org: string, name: string, openSessionId: string): Promise<EnvironmentResponse>;
    abstract listEnvironments(org?: string): Promise<ListEnvironmentsResponse>;
}

export class ESC implements ESCApiClient {
    private readonly apiUrl: string;
    private readonly token: string;
    private readonly http: AxiosInstance;

    constructor(token?: string, apiURL?: string) {
        const accessToken = token || process.env.PULUMI_ACCESS_TOKEN;
        if (!accessToken) {
            throw new Error("No token provided");
        }
        this.token = accessToken;
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

    public async createEnvironment(org: string, name: string): Promise<void> {
        await this.http.post<null>(`environments/${org}/${name}`, {});
    }

    public async listEnvironments(org?: string): Promise<ListEnvironmentsResponse> {
        const response = await this.http.get<ListEnvironmentsResponse>(`environments/${org}`);
        return response.data;
    }

    public async readEnvironment(org: string, name: string, tag?: string): Promise<ReadEnvironmentResponse> {
        const options: AxiosRequestConfig = {};
        if (tag) {
            options.headers = {
                "If-Match": tag,
            };
        }
        const resp = await this.http.get<string>(`environments/${org}/${name}`, options);
        const result: ReadEnvironmentResponse = {
            environmentString: resp.data,
        };
        if (resp.headers.etag) {
            result.tag = resp.headers.etag;
        }

        return result;
    }

    public async checkEnvironment(org: string, name: string): Promise<EnvironmentResponse> {
        const response = await this.http.post<EnvironmentResponse>(`environments/${org}/${name}/check`);
        return response.data;
    }

    public async openEnvironment(org: string, name: string): Promise<OpenEnvironmentResponse> {
        const response = await this.http.post<OpenEnvironmentResponse>(`environments/${org}/${name}/open`, {});
        return response.data;
    }

    public async readOpenEnvironment(org: string, name: string, openSessionId: string): Promise<EnvironmentResponse> {
        const response = await this.http.get<EnvironmentResponse>(`environments/${org}/${name}/open/${openSessionId}`);
        return response.data;
    }

    public async updateEnvironment(
        org: string,
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
            .patch(`environments/${org}/${name}`, envYaml, options)
            .catch((err: AxiosError) => {
                if (err.response?.status === HttpStatusCode.BadRequest) {
                    return err.response?.data;
                }
                throw err;
            });

        return diags as UpdateEnvironmentResponse;
    }

    public async deleteEnvironment(org: string, name: string): Promise<void> {
        await this.http.delete<null>(`environments/${org}/${name}`);
    }
}
