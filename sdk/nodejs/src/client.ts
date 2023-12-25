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

import { EnvironmentResponse, ListEnvironmentsResponse, OpenEnvironmentResponse, OrgEnvironment } from ".";
import axios, { AxiosRequestConfig, AxiosInstance } from "axios";

const API_URL: string = "https://api.pulumi.com/api";

export abstract class ESCApiClient {
    abstract createEnvironment(name: string): Promise<null>;
    abstract deleteEnvironment(name: string): Promise<null>;
    abstract getEnvironment(name: string): Promise<EnvironmentResponse>;
    abstract openEnvironment(name: string): Promise<OpenEnvironmentResponse>;
    abstract updateEnvironment(name: string, body: string, tag?: string): Promise<null>;
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
        this.http = axios.create(this.getConfig());
    }

    private getConfig(): AxiosRequestConfig {
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

    public async getEnvironment(name: string): Promise<EnvironmentResponse> {
        const response = await this.http.get<EnvironmentResponse>(`environments/${this.org}/${name}`);
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

    public async updateEnvironment(name: string, body: string, tag?: string): Promise<null> {
        return this.http.put(`environments/${this.org}/${name}`, {
            body,
            tag,
        });
    }

    public async deleteEnvironment(name: string): Promise<null> {
        return this.http.delete(`environments/${this.org}/${name}`);
    }
}
