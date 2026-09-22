# The import ID is the three key components hex-encoded and joined by "/":
# hex("Sender_1")/hex("SenderInterface")/hex("Interface_1"). Hex text can
# never itself contain "/", so this is unambiguous regardless of what
# characters agency, scheme, and external_id contain.
terraform import sapintegrationsuite_alternative_partner.sender \
  53656e6465725f31/53656e646572496e74657266616365/496e746572666163655f31
