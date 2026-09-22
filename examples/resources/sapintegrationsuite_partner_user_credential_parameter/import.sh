# Import recovers partner_id, parameter_id, and user, but never the
# password: apply a matching configuration right afterward, supplying
# password_wo and a password_wo_version. That first apply plans as a
# replacement even though nothing server-side has actually changed - an
# inherent limitation of importing a write-only-secret resource.
terraform import sapintegrationsuite_partner_user_credential_parameter.receiver Receiver_1/USER
