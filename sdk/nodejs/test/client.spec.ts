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

import { describe, it } from "node:test";
import assert from "assert";
import { EnvironmentDefinition, ESC } from "../src";
import { stringify } from "yaml";

describe("ESC", () => {
    it("should create, update, check, open and delete an environment", async () => {
        const PULUMI_ACCESS_TOKEN = process.env.PULUMI_ACCESS_TOKEN;
        const PULUMI_ORG = process.env.PULUMI_ORG;
        if (!PULUMI_ACCESS_TOKEN) {
            throw new Error("PULUMI_ACCESS_TOKEN not set");
        }
        if (!PULUMI_ORG) {
            throw new Error("PULUMI_ORG not set");
        }
        const client = new ESC(PULUMI_ACCESS_TOKEN, PULUMI_ORG);
        const name = `test-${Date.now()}`;

        await assert.doesNotReject(client.createEnvironment(name));
        const envs = await client.listEnvironments();
        assert(envs.some((e) => e.name === name));

        // Add some configuration to the environment.
        const foo = "bar";
        const envDef = new EnvironmentDefinition({
            values: {
                blah: "blah",
                pulumiConfig: {
                    foo,
                },
                environmentVariables: {
                    FOO: foo,
                },
            },
        });
        let updateResp = await client.updateEnvironment(name, envDef);
        assert.strictEqual(updateResp.diagnostics, undefined);

        const getResp = await client.readEnvironment(name);
        assert.strictEqual(getResp.environmentString, stringify(envDef));
        const { tag } = getResp;
        assert.ok(tag);

        let checkResult = await client.checkEnvironment(name);
        assert.strictEqual(checkResult.diagnostics, undefined);

        let session = await client.openEnvironment(name);
        assert.ok(session.id);

        let openEnv = await client.readOpenEnvironment(name, session.id);
        assert.strictEqual(openEnv.diagnostics, undefined);

        envDef.values!.pulumiConfig!.haha = "business";
        updateResp = await client.updateEnvironment(name, envDef, tag);
        assert.strictEqual(updateResp.diagnostics, undefined);

        checkResult = await client.checkEnvironment(name);
        assert.strictEqual(checkResult.diagnostics, undefined);

        session = await client.openEnvironment(name);
        assert.ok(session.id);

        openEnv = await client.readOpenEnvironment(name, session.id);
        assert.strictEqual(openEnv.diagnostics, undefined);

        await assert.doesNotReject(client.deleteEnvironment(name));
    });
});
