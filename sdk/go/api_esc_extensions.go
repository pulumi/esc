package esc_sdk

import (
	"context"
	"errors"
	"io"

	"gopkg.in/ghodss/yaml.v1"
)

type EscClient struct {
	rawClient *RawAPIClient
	EscAPI    *EscAPIService
}

func NewClient(cfg *Configuration) *EscClient {
	client := &EscClient{rawClient: NewRawAPIClient(cfg)}
	client.EscAPI = client.rawClient.EscAPI
	return client
}

func (c *EscClient) ListEnvironments(ctx context.Context, org string, continuationToken *string) (*OrgEnvironments, error) {
	request := c.EscAPI.ListEnvironments(ctx, org)
	if continuationToken != nil {
		request = request.ContinuationToken(*continuationToken)
	}

	envs, _, err := request.Execute()
	return envs, err
}

func (c *EscClient) GetEnvironment(ctx context.Context, org, envName string) (*EnvironmentDefinition, string, error) {
	env, resp, err := c.EscAPI.GetEnvironment(ctx, org, envName).Execute()
	if err != nil {
		return nil, "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return env, string(body), nil
}

func (c *EscClient) OpenEnvironment(ctx context.Context, org, envName string) (*OpenEnvironment, error) {
	openInfo, _, err := c.EscAPI.OpenEnvironment(ctx, org, envName).Execute()
	return openInfo, err
}

func (c *EscClient) ReadOpenEnvironment(ctx context.Context, org, envName, openEnvID string) (*Environment, map[string]any, error) {
	env, _, err := c.EscAPI.ReadOpenEnvironment(ctx, org, envName, openEnvID).Execute()
	if err != nil {
		return nil, nil, err
	}

	propertyMap := *env.Properties
	for k, v := range propertyMap {
		v.Value = mapValues(v.Value)
		propertyMap[k] = v
	}

	values := make(map[string]any, len(propertyMap))
	for k := range propertyMap {
		v := propertyMap[k]
		values[k] = mapValuesPrimitive(&v)
	}

	return env, values, nil
}

func (c *EscClient) OpenAndReadEnvironment(ctx context.Context, org, envName string) (*Environment, map[string]any, error) {
	openInfo, err := c.OpenEnvironment(ctx, org, envName)
	if err != nil {
		return nil, nil, err
	}

	return c.ReadOpenEnvironment(ctx, org, envName, openInfo.Id)
}

func (c *EscClient) ReadEnvironmentProperty(ctx context.Context, org, envName, openEnvID, propPath string) (*Value, any, error) {
	prop, _, err := c.EscAPI.ReadOpenEnvironmentProperty(ctx, org, envName, openEnvID).Property(propPath).Execute()
	v := mapValuesPrimitive(prop.Value)
	return prop, v, err
}

func (c *EscClient) CreateEnvironment(ctx context.Context, org, envName string) error {
	_, _, err := c.EscAPI.CreateEnvironment(ctx, org, envName).Execute()
	return err
}

func (c *EscClient) UpdateEnvironmentYaml(ctx context.Context, org, envName, yaml string) (*EnvironmentDiagnostics, error) {
	diags, _, err := c.EscAPI.UpdateEnvironmentYaml(ctx, org, envName).Body(yaml).Execute()
	return diags, err
}

func (c *EscClient) UpdateEnvironment(ctx context.Context, org, envName string, env *EnvironmentDefinition) (*EnvironmentDiagnostics, error) {
	yaml, err := MarshalEnvironmentDefinition(env)
	if err != nil {
		return nil, err
	}

	diags, _, err := c.EscAPI.UpdateEnvironmentYaml(ctx, org, envName).Body(yaml).Execute()
	return diags, err
}

func (c *EscClient) DeleteEnvironment(ctx context.Context, org, envName string) error {
	_, _, err := c.EscAPI.DeleteEnvironment(ctx, org, envName).Execute()
	return err
}

func (c *EscClient) CheckEnvironment(ctx context.Context, org string, env *EnvironmentDefinition) (*CheckEnvironment, error) {
	yaml, err := MarshalEnvironmentDefinition(env)
	if err != nil {
		return nil, err
	}

	return c.CheckEnvironmentYaml(ctx, org, yaml)
}

func (c *EscClient) CheckEnvironmentYaml(ctx context.Context, org, yaml string) (*CheckEnvironment, error) {
	check, _, err := c.EscAPI.CheckEnvironmentYaml(ctx, org).Body(yaml).Execute()
	var genericOpenApiError *GenericOpenAPIError
	if err != nil && errors.As(err, &genericOpenApiError) {
		model := genericOpenApiError.Model().(CheckEnvironment)
		return &model, err
	}

	return check, err
}

func (c *EscClient) DecryptEnvironment(ctx context.Context, org, envName string) (*EnvironmentDefinition, string, error) {
	env, resp, err := c.EscAPI.DecryptEnvironment(ctx, org, envName).Execute()

	body, bodyErr := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", bodyErr
	}

	return env, string(body), err
}

func MarshalEnvironmentDefinition(env *EnvironmentDefinition) (string, error) {
	var bs []byte
	bs, err := yaml.Marshal(env)
	if err == nil {
		return string(bs), nil
	}

	return "", err
}

func mapValuesPrimitive(value any) any {
	switch val := value.(type) {
	case *Value:
		return mapValuesPrimitive(val.Value)
	case map[string]Value:
		output := make(map[string]any, len(val))
		for k, v := range val {
			output[k] = mapValuesPrimitive(v.Value)
		}

		return output
	case []any:
		for i, v := range val {
			val[i] = mapValuesPrimitive(v)
		}
		return val
	default:
		return value
	}
}

func mapValues(value any) any {
	if val := getValue(getMapSafe(value)); val != nil {
		val.Value = mapValues(val.Value)
		return val
	}
	if mapData, isMap := value.(map[string]any); isMap {
		output := map[string]Value{}
		for key, v := range mapData {
			value := mapValues(v)
			if value == nil {
				continue
			}

			if v, ok := value.(*Value); ok && v != nil {
				output[key] = *v
			} else {
				output[key] = Value{
					Value: value,
				}
			}
		}
		return output
	} else if sliceData, isSlice := value.([]any); isSlice {
		for i, v := range sliceData {
			sliceData[i] = mapValues(v)
		}
		return sliceData
	}

	return value
}

func getValue(data map[string]any) *Value {
	_, hasValue := data["value"]
	_, hasTrace := data["trace"]
	if hasValue && hasTrace {
		return &Value{
			Value:   mapValues(data["value"]),
			Secret:  getBoolPtr(data, "secret"),
			Unknown: getBoolPtr(data, "unknown"),
			Trace:   getTrace(data["trace"].(map[string]any)),
		}
	}

	return nil
}

func getTrace(data map[string]any) Trace {
	def := getRange(getMapSafe(data["def"]))
	base := getValue(getMapSafe(data["base"]))
	if def != nil || base != nil {
		return Trace{
			Def:  def,
			Base: base,
		}
	}

	return Trace{}
}

func getMapSafe(data any) map[string]any {
	if data == nil {
		return nil
	}

	val, _ := data.(map[string]any)
	return val
}

func getRange(data map[string]any) *Range {
	begin := getPos(getMapSafe(data["begin"]))
	end := getPos(getMapSafe(data["end"]))
	environment := data["environment"].(string)
	if begin != nil && end != nil {
		return &Range{
			Environment: environment,
			Begin:       *begin,
			End:         *end,
		}
	}

	return nil
}

func getPos(data map[string]any) *Pos {
	line, hasLine := data["line"].(float64)
	column, hasColumn := data["column"].(float64)
	byteData, hasByte := data["byte"].(float64)
	if hasLine || hasColumn || hasByte {
		return &Pos{
			Line:   int32(line),
			Column: int32(column),
			Byte:   int32(byteData),
		}
	}

	return nil
}

func getBoolPtr(data map[string]any, key string) *bool {
	val, exists := data[key]
	if exists {
		v, ok := val.(bool)
		if ok {
			return &v
		}
	}

	return nil
}
