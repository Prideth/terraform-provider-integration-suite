package cloudintegration

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

// integrationFlowConfigurationsEntitySet is the entity set behind the
// Configurations navigation of IntegrationDesigntimeArtifacts.
const integrationFlowConfigurationsEntitySet = "Configurations"

// IntegrationFlowConfiguration is one externalized parameter of an integration
// flow, following the Configuration entity type of the tenant $metadata.
type IntegrationFlowConfiguration struct {
	ParameterKey   string `json:"ParameterKey"`
	ParameterValue string `json:"ParameterValue"`
	DataType       string `json:"DataType,omitempty"`
	Description    string `json:"Description,omitempty"`
}

// integrationFlowConfigurationWrite is the body of SAP's documented update
// request: PUT IntegrationDesigntimeArtifacts(Id,Version)/$links/
// Configurations('<key>') with {"ParameterValue": ..., "DataType": ...}.
type integrationFlowConfigurationWrite struct {
	ParameterValue string `json:"ParameterValue"`
	DataType       string `json:"DataType,omitempty"`
}

// ListIntegrationFlowConfigurations reads every externalized parameter of the
// given design-time version of an integration flow, sorted by key.
func (c *Client) ListIntegrationFlowConfigurations(ctx context.Context, flowID, version string) ([]IntegrationFlowConfiguration, error) {
	key, err := designtimeArtifactKey(flowID, version)
	if err != nil {
		return nil, err
	}

	body, err := c.odata.Get(ctx, integrationDesigntimeArtifactsEntitySet+key+"/"+integrationFlowConfigurationsEntitySet)
	if err != nil {
		return nil, err
	}

	var configs []IntegrationFlowConfiguration
	if err := v2.DecodeCollection(body, &configs); err != nil {
		return nil, err
	}
	sort.Slice(configs, func(i, j int) bool { return configs[i].ParameterKey < configs[j].ParameterKey })
	return configs, nil
}

// UpdateIntegrationFlowConfiguration sets one externalized parameter. dataType
// should be the parameter's current DataType so that SAP keeps its type.
func (c *Client) UpdateIntegrationFlowConfiguration(ctx context.Context, flowID, version, parameterKey, value, dataType string) error {
	key, err := designtimeArtifactKey(flowID, version)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(integrationFlowConfigurationWrite{ParameterValue: value, DataType: dataType})
	if err != nil {
		return fmt.Errorf("cloudintegration: encoding integration flow configuration: %w", err)
	}

	path := integrationDesigntimeArtifactsEntitySet + key + "/$links/" + integrationFlowConfigurationsEntitySet + v2.KeyPredicate(parameterKey)
	_, err = c.odata.Put(ctx, path, payload)
	return err
}
