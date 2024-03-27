import { string } from "yaml/dist/schema/common/string";
import { Environment, EnvironmentDefinitionValues, OpenEnvironment, OrgEnvironments, OrgEnvironment, EnvironmentDefinition, EscApi as EscRawApi, Configuration, Value, EnvironmentDiagnostics, CheckEnvironment, Pos, Range, Trace } from "./raw/index";
import * as yaml from "js-yaml";
import { AxiosError } from "axios";

export { Configuration, Environment, EnvironmentDefinitionValues, OpenEnvironment, OrgEnvironments, OrgEnvironment, EnvironmentDefinition, EscRawApi, Value, EnvironmentDiagnostics, CheckEnvironment, Pos, Range, Trace };

export interface EnvironmentDefinitionResponse {
    definition: EnvironmentDefinition;
    yaml: string;
}

export interface EnvironmentResponse {
    environment?: Environment;
    values?: EnvironmentDefinitionValues;
}

export interface EnvironmentPropertyResponse {
    property: Value;
    value: any;
}

type KeyValueMap = {[key: string]: string};
/**
 * 
 * @export
 * @class EscApi
 */
export class EscApi {
    rawApi: EscRawApi;
    config: Configuration;
    constructor(config: Configuration) {
        this.config = config;
        this.rawApi = new EscRawApi(config);
    }
    async listEnvironments(org: string, continuationToken?: string | undefined): Promise<OrgEnvironments | undefined> {
        const resp = await this.rawApi.listEnvironments(org, continuationToken);
        if (resp.status === 200) {
            return resp.data;
        }

        throw new Error(`Failed to list environments: ${resp.statusText}`);
    }
    async getEnvironment(org: string, name: string): Promise<EnvironmentDefinitionResponse | undefined> {
        const resp = await this.rawApi.getEnvironment(org, name);
        if (resp.status === 200) {
            const doc = yaml.load(resp.data as string);
            return {
                definition: doc as EnvironmentDefinition,
                yaml: resp.data as string,
            };
        }

        throw new Error(`Failed to get environment: ${resp.statusText}`);
    }

    async openEnvironment(org: string, name: string): Promise<OpenEnvironment | undefined> {
        const resp = await this.rawApi.openEnvironment(org, name);
        if (resp.status === 200) {
            return resp.data;
        }

        throw new Error(`Failed to open environment: ${resp.statusText}`);
    }

    async readOpenEnvironment(org: string, name: string, openSessionID: string): Promise<EnvironmentResponse | undefined> {
        const resp = await this.rawApi.readOpenEnvironment(org, name, openSessionID);
        if (resp.status === 200) {
            return {
                environment: resp.data,
                values: convertEnvPropertiesToValues(resp.data.properties),
            }
        }

        throw new Error(`Failed to read environment: ${resp.statusText}`);
    }

    async openAndReadEnvironment(org: string, name: string): Promise<EnvironmentResponse | undefined> {
        const open = await this.openEnvironment(org, name);
        if (open?.id) {
            return await this.readOpenEnvironment(org, name, open.id);        
        }

        throw new Error(`Failed to open and read environment: ${open}`);
    }

    async readOpenEnvironmentProperty(org: string, name: string, openSessionID: string, property: string): Promise<EnvironmentPropertyResponse | undefined> {
        const resp = await this.rawApi.readOpenEnvironmentProperty(org, name, openSessionID, property);
        if (resp.status === 200) {
            return {
                property: resp.data,
                value: convertPropertyToValue(resp.data),
            }
        }

        throw new Error(`Failed to read environment property: ${resp.statusText}`);
    }

    async createEnvironment(org: string, name: string): Promise<void> {
        const resp = await this.rawApi.createEnvironment(org, name);
        if (resp.status === 200) {
            return;
        }

        throw new Error(`Failed to create environment: ${resp.statusText}`);
    }

    async updateEnvironmentYaml(org: string, name: string, yaml: string): Promise<EnvironmentDiagnostics | undefined> {
        const resp = await this.rawApi.updateEnvironmentYaml(org, name, yaml);
        if (resp.status === 200) {
            return resp.data;
        }

        throw new Error(`Failed to update environment: ${resp.statusText}`);
    }
    
    async updateEnvironment(org: string, name: string, values: EnvironmentDefinition): Promise<EnvironmentDiagnostics | undefined> {
        const body = yaml.dump(values);
        const resp = await this.rawApi.updateEnvironmentYaml(org, name, body);
        if (resp.status === 200) {
            return resp.data;
        }

        throw new Error(`Failed to update environment: ${resp.statusText}`);
    }

    async deleteEnvironment(org: string, name: string): Promise<void> {
        const resp = await this.rawApi.deleteEnvironment(org, name);
        if (resp.status === 200) {
            return;
        }

        throw new Error(`Failed to delete environment: ${resp.statusText}`);
    }

    async checkEnvironmentYaml(org: string, yaml: string): Promise<CheckEnvironment | undefined> {
        try {
            const resp = await this.rawApi.checkEnvironmentYaml(org, yaml);
            if (resp.status === 200) {
                return resp.data;
            }

            throw new Error(`Failed to check environment: ${resp.statusText}`);
        } catch (err: any) {
            if (err instanceof  AxiosError) {
                if (err.response?.status === 400) {
                    return err.response?.data;
                }
            }
            throw err;
        }    
    }

    async checkEnvironment(org: string, env: EnvironmentDefinition): Promise<CheckEnvironment | undefined> {
        const body = yaml.dump(env);
        return await this.checkEnvironmentYaml(org, body);
    }

    async decryptEnvironment(org: string, name: string): Promise<EnvironmentDefinitionResponse | undefined> {
        const resp = await this.rawApi.decryptEnvironment(org, name);
        if (resp.status === 200) {
            const doc = yaml.load(resp.data as string);
            return {
                definition: doc as EnvironmentDefinition,
                yaml: resp.data as string,
            };
        }

        throw new Error(`Failed to decrypt environment: ${resp.statusText}`);
    }

}

function convertEnvPropertiesToValues(env: {[key:string]: Value} | undefined): {[key:string]: any} {
    if (!env) {
        return {};
    }

    const values: {[key:string]: any} = {};
    for (const key in env) {
        const value = env[key];
        
        values[key] = convertPropertyToValue(value);
    }
    
    return values;
}

function convertPropertyToValue(property: any): any {
    if (!property) {
        return property;
    }

    let value = property;
    if ("value" in property) {
        value = convertPropertyToValue(property.value);
    }

    if (!value) {
        return value;
    }

    if (Array.isArray(value)) {
        const array = value as Value[];
        return array.map((v) => convertPropertyToValue(v));
    }

    if (typeof value === "object") {
        const result: any = {}
        for (const key in value) {
            result[key] = convertPropertyToValue(value[key]);
        }

        return result
    }
    
    return value;
}
