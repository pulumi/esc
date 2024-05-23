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

import { after, before, describe, it } from "node:test";
import assert from "assert";
import * as esc from "../esc";

const ENV_PREFIX = "sdk-ts-test";
describe("ESC", async () => {

    const PULUMI_ACCESS_TOKEN = process.env.PULUMI_ACCESS_TOKEN;
    const PULUMI_ORG = process.env.PULUMI_ORG;
    if (!PULUMI_ACCESS_TOKEN) {
        throw new Error("PULUMI_ACCESS_TOKEN not set");
    }
    if (!PULUMI_ORG) {
        throw new Error("PULUMI_ORG not set");
    }
    let config = new esc.Configuration();
    config.apiKey = `token ${PULUMI_ACCESS_TOKEN}`;
    const client = new esc.EscApi(config);
    const baseEnvName = `${ENV_PREFIX}-base-${Date.now()}`;

    before(async () => {
    
        const envDef: esc.EnvironmentDefinition = {
            values: {
                base: baseEnvName,
            },
        }
        await removeAllTestEnvs(client, PULUMI_ORG);
        await client.createEnvironment(PULUMI_ORG, baseEnvName);
        await client.updateEnvironment(PULUMI_ORG, baseEnvName, envDef);
    });

    after(async () => {
        await client.deleteEnvironment(PULUMI_ORG, baseEnvName);
    });

    it("should create, list, update, get, decrypt, open and delete an environment", async () => {
        const name = `${ENV_PREFIX}-${Date.now()}`;
    
        await assert.doesNotReject(client.createEnvironment(PULUMI_ORG, name));
        const orgs = await client.listEnvironments(PULUMI_ORG);
        assert.notEqual(orgs, undefined);
        assert(orgs?.environments?.some((e) => e.name === name))

        const envDef: esc.EnvironmentDefinition = {
            imports: [baseEnvName],
            values: {
                foo: "bar",
                my_secret: {
                    "fn::secret": "shh! don't tell anyone",
                },
                my_array: [1, 2, 3],
                pulumiConfig: {
                    foo: "${foo}",
                },
                environmentVariables: {
                    FOO: "${foo}",
                },
            },
        }
        const diags = await client.updateEnvironment(PULUMI_ORG, name, envDef)
        assert.notEqual(diags, undefined);
        assert.equal(diags?.diagnostics, undefined);

        const env = await client.getEnvironment(PULUMI_ORG, name)

        assert.notEqual(env, undefined);
        assertEnvDef(env!, baseEnvName);
        assert.notEqual(env?.definition?.values?.my_secret, undefined);

        const decryptEnv = await client.decryptEnvironment(PULUMI_ORG, name);

        assert.notEqual(decryptEnv, undefined);
        assertEnvDef(decryptEnv!, baseEnvName);
        assert.equal(decryptEnv?.definition?.values?.my_secret["fn::secret"], "shh! don't tell anyone");
                
        const openEnv = await client.openAndReadEnvironment(PULUMI_ORG, name);
        
        assert.equal(openEnv?.values?.base, baseEnvName);
        assert.equal(openEnv?.values?.foo, "bar");
        assert.deepEqual(openEnv?.values?.my_array, [1, 2, 3]);
        assert.deepEqual(openEnv?.values?.my_secret, "shh! don't tell anyone");
        assert.equal(openEnv?.values?.pulumiConfig?.foo, "bar");
        assert.equal(openEnv?.values?.environmentVariables?.FOO, "bar");
        console.log("test");

        const openInfo = await client.openEnvironment(PULUMI_ORG, name);
        assert.notEqual(openInfo, undefined);

        const value = await client.readOpenEnvironmentProperty(PULUMI_ORG, name, openInfo?.id!, "pulumiConfig.foo");
        assert.equal(value?.value, "bar");

        await client.deleteEnvironment(PULUMI_ORG, name);
    });

    it("check environment valid", async () => {
        const envDef: esc.EnvironmentDefinition = {
            values: {
                foo: "bar",
            },
        }

        const diags = await client.checkEnvironment(PULUMI_ORG, envDef)
        assert.notEqual(diags, undefined);
        assert.equal(diags?.diagnostics?.length, 0);
    });

    it("check environment invalid", async () => {
        const envDef: esc.EnvironmentDefinition = {
            values: {
                foo: "bar",
                pulumiConfig: {
                    foo: "${bad_ref}",
                },
            },
        }
        const diags = await client.checkEnvironment(PULUMI_ORG, envDef)
        assert.notEqual(diags, undefined);
        assert.equal(diags?.diagnostics?.length, 1)
        assert.equal(diags?.diagnostics?.[0].summary, "unknown property \"bad_ref\"")
    });
});

function assertEnvDef(env: esc.EnvironmentDefinitionResponse, baseEnvName: string) {
    assert.equal(env.definition?.imports?.length, 1);
    assert.equal(env.definition?.imports?.[0], baseEnvName);
    assert.equal(env.definition?.values?.foo, "bar");
    assert.deepEqual(env.definition?.values?.my_array, [1, 2, 3]);
    assert.equal(env.definition?.values?.pulumiConfig?.foo, "${foo}");
    assert.equal(env.definition?.values?.environmentVariables?.FOO, "${foo}");
}

async function removeAllTestEnvs(client: esc.EscApi, orgName: string): Promise<any> {
    let continuationToken: string | undefined = undefined;
    do {
        const orgs: esc.OrgEnvironments | undefined = await client.listEnvironments(orgName, continuationToken);

        assert.notEqual(orgs, undefined);
        orgs?.environments?.forEach(async (e: esc.OrgEnvironment) => {
            if (e.name.startsWith(ENV_PREFIX)) {
                await client.deleteEnvironment(orgName, e.name);
            }
        });

        continuationToken = orgs?.nextToken;
    } while (continuationToken !== undefined && continuationToken !== "");
}
