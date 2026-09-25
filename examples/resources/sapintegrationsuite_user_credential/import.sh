terraform import sapintegrationsuite_user_credential.backend BACKEND_BASIC

# A credential on an Edge Integration Cell: prefix the runtime location ID.
terraform import sapintegrationsuite_user_credential.backend_edge location:plant-a/BACKEND_BASIC
