
from esc import api, models, ApiException
from pydantic import StrictBytes
from typing import Mapping, Any
import inspect
import yaml

class EscClient:
    esc_api: api.EscApi = None
    """EscClient

    :param esc_api: EscApi (required)
    """
    def __init__(self, esc_api: api.EscApi) -> None:
        self.esc_api = esc_api

    def list_environments(self, org_name: str, continuation_token: str = None) -> models.OrgEnvironments:
        return self.esc_api.list_environments(org_name, continuation_token)
    
    def get_environment(self, org_name: str, env_name: str) -> tuple[models.EnvironmentDefinition, StrictBytes]:
        response = self.esc_api.get_environment_with_http_info(org_name, env_name)
        return response.data, response.raw_data

    def open_environment(self, org_name: str, env_name: str) -> models.Environment:
        return self.esc_api.open_environment(org_name, env_name)
    
    def read_open_environment(self, org_name: str, env_name: str, open_session_id: str) -> tuple[models.Environment, Mapping[str, Any], str]:
        response = self.esc_api.read_open_environment_with_http_info(org_name, env_name, open_session_id)
        values = convertEnvPropertiesToValues(response.data.properties)
        return response.data, values, response.raw_data.decode('utf-8')
    
    def open_and_read_environment(self, org_name: str, env_name: str) -> tuple[models.Environment, Mapping[str, any], str]:
        openEnv = self.open_environment(org_name, env_name)
        return self.read_open_environment(org_name, env_name, openEnv.id)
    
    def read_open_environment_property(self, org_name: str, env_name: str, open_session_id: str, property_name: str) -> tuple[models.Value, Any]:
        v = self.esc_api.read_open_environment_property(org_name, env_name, open_session_id, property_name)
        return v, convertPropertyToValue(v.value)
    
    def create_environment(self, org_name: str, env_name: str) -> models.Environment:
        return self.esc_api.create_environment(org_name, env_name)
    
    def update_environment_yaml(self, org_name: str, env_name: str, yaml_body: str) -> models.EnvironmentDiagnostics:
        return self.esc_api.update_environment_yaml(org_name, env_name, yaml_body)
    
    def update_environment(self, org_name: str, env_name: str, env: models.EnvironmentDefinition) -> models.Environment:
        yaml_body = yaml.dump(env)
        return self.update_environment_yaml(org_name, env_name, yaml_body)
    
    def delete_environment(self, org_name: str, env_name: str) -> None:
        self.esc_api.delete_environment(org_name, env_name)

    def check_environment_yaml(self, org_name: str, yaml_body: str) -> models.CheckEnvironment:
        try:
            response = self.esc_api.check_environment_yaml_with_http_info(org_name, yaml_body)
            return response.data
        except ApiException as e:
            return e.data
    
    def check_environment(self, org_name: str, env: models.EnvironmentDefinition) -> models.CheckEnvironment:
        yaml_body = yaml.safe_dump(env.to_dict())
        return self.check_environment_yaml(org_name, yaml_body)
    
    def decrypt_environment(self, org_name: str, env_name: str) -> tuple[models.EnvironmentDefinition, str]:
        response = self.esc_api.decrypt_environment_with_http_info(org_name, env_name)
        return response.data, response.raw_data.decode('utf-8')
    

def convertEnvPropertiesToValues(env: Mapping[str, models.Value]) -> Any:
    if env is None:
        return env

    values = {}
    for key in env:
        value = env[key]
        
        values[key] = convertPropertyToValue(value.value)
    
    return values

def convertPropertyToValue(property: Any) -> Any:
    if property is None:
        return property

    value = property
    if isinstance(property, dict) and "value" in property:
        value = convertPropertyToValue(property["value"])
        return value

    if value is None:
        return value

    if type(value) is list:
        result = []
        for item in value:
            result.append(convertPropertyToValue(item))
        return result

    if isObject(value):
        result = {}
        for key in value:
            result[key] = convertPropertyToValue(value[key])
        return result
    
    return value

def isObject(obj):
    return inspect.isclass(obj) or isinstance(obj, dict)