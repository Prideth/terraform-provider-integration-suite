terraform import sapintegrationsuite_integration_flow_deployment.metering UTILITIES/metering

# A deployment on an Edge Integration Cell: prefix the runtime location ID.
terraform import sapintegrationsuite_integration_flow_deployment.metering_edge location:plant-a/UTILITIES/metering
