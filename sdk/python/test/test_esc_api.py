# coding: utf-8

import unittest
import os
from datetime import datetime

import esc

ENV_PREFIX = "sdk-python-test-"

class TestEscApi(unittest.TestCase):
    """EscApi unit test stubs"""

    def setUp(self) -> None:
        self.accessToken = os.getenv("PULUMI_ACCESS_TOKEN")
        self.assertIsNotNone(self.accessToken, "PULUMI_ACCESS_TOKEN must be set")
	    
        self.orgName = os.getenv("PULUMI_ORG")
        self.assertIsNotNone(self.orgName, "PULUMI_ORG must be set")
	
        configuration = esc.Configuration()
        configuration.api_key['Authorization'] = "token " + self.accessToken
        self.apiClient = esc.ApiClient(configuration)
        self.client = esc.EscClient(esc.EscApi(self.apiClient))
        
        self.remove_all_python_test_envs()

        self.baseEnvName = ENV_PREFIX + "base-" + datetime.now().strftime("%Y%m%d%H%M%S")
        self.client.create_environment(self.orgName, self.baseEnvName)
        self.envName = None

    def tearDown(self) -> None:
        if self.baseEnvName != None:
            self.client.delete_environment(self.orgName, self.baseEnvName)
        if self.envName != None:
            self.client.delete_environment(self.orgName, self.envName)

    def test_environment_end_to_end(self) -> None:

        self.envName = ENV_PREFIX + "end-to-end-" + datetime.now().strftime("%Y%m%d%H%M%S")
        self.client.create_environment(self.orgName, self.envName)

        envs = self.client.list_environments(self.orgName)
        self.assertFindEnv(envs)

        fooReference = "${foo}"
        yaml = f"""
imports:
  - {self.baseEnvName}
values:
  foo: bar
  my_secret:
    fn::secret: "shh! don't tell anyone"
  my_array: [1, 2, 3]
  pulumiConfig:
    foo: {fooReference}
  environmentVariables:
    FOO: {fooReference}
"""    
        self.client.update_environment_yaml(self.orgName, self.envName, yaml)

        env, new_yaml = self.client.get_environment(self.orgName, self.envName)
        self.assertIsNotNone(env)
        self.assertIsNotNone(new_yaml)

        self.assertEnvDef(env)
        self.assertIsNotNone(env.values.additional_properties["my_secret"])

        decrypted_env, _ = self.client.decrypt_environment(self.orgName, self.envName)
        self.assertIsNotNone(decrypted_env)
        self.assertEnvDef(decrypted_env)
        self.assertIsNotNone(decrypted_env.values.additional_properties["my_secret"])

        _, values, yaml = self.client.open_and_read_environment(self.orgName, self.envName)
        self.assertIsNotNone(yaml)

        self.assertEqual(values["foo"], "bar")
        self.assertEqual(values["my_array"], [1, 2, 3])
        self.assertEqual(values["my_secret"], "shh! don't tell anyone")
        self.assertIsNotNone(values["pulumiConfig"])
        self.assertEqual(values["pulumiConfig"]["foo"], "bar")
        self.assertIsNotNone(values["environmentVariables"])
        self.assertEqual(values["environmentVariables"]["FOO"], "bar")

        openInfo = self.client.open_environment(self.orgName, self.envName)
        self.assertIsNotNone(openInfo)

        v, value = self.client.read_open_environment_property(self.orgName, self.envName, openInfo.id, "foo")
        self.assertIsNotNone(v)
        self.assertEqual(v.value, "bar")
        self.assertEqual(value, "bar")

    def test_check_environment_valid(self):
        envDef = esc.EnvironmentDefinition(values=esc.EnvironmentDefinitionValues(additional_properties={"foo": "bar"}))

        diags = self.client.check_environment(self.orgName, envDef)
        self.assertNotEqual(diags, None)
        self.assertEqual(diags.diagnostics, None)

    def test_check_environment_invalid(self):
        envDef = esc.EnvironmentDefinition(values=esc.EnvironmentDefinitionValues(additional_properties={"foo": "bar"}, pulumi_config={"foo": "${bad_ref}"}))
        diags = self.client.check_environment(self.orgName, envDef)
        self.assertNotEqual(diags, None)
        self.assertNotEqual(diags.diagnostics, None)
        self.assertEqual(len(diags.diagnostics), 1)
        self.assertEqual(diags.diagnostics[0].summary, "unknown property \"bad_ref\"")

    def assertEnvDef(self, env):
        self.assertListEqual(env.imports, [self.baseEnvName])
        self.assertEqual(env.values.additional_properties["foo"], "bar")
        self.assertEqual(env.values.additional_properties["my_array"], [1, 2, 3])
        self.assertIsNotNone(env.values.pulumi_config)
        self.assertEqual(env.values.pulumi_config["foo"], "${foo}")
        self.assertIsNotNone(env.values.environment_variables)
        self.assertEqual(env.values.environment_variables["FOO"], "${foo}")

    def assertFindEnv(self, envs):
        self.assertIsNotNone(envs)
        self.assertGreater(len(envs.environments), 0)
        for env in envs.environments:
            if env.name == self.envName:
                return
            
        self.fail("Environment {envName} not found".format(envName=self.envName))

    def remove_all_python_test_envs(self) -> None:
        continuationToken = None
        while True:
            envs = self.client.list_environments(self.orgName, continuationToken)
            for env in envs.environments:
                if env.name.startswith(ENV_PREFIX):
                    self.client.delete_environment(self.orgName, env.name)
            
            continuationToken = envs.next_token
            if continuationToken == None or continuationToken == "":
                break

if __name__ == '__main__':
    unittest.main()
