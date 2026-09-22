package features

// Operations records which lifecycle operations this provider implements
// for a feature. It is deliberately more detailed than SupportStatus alone:
// two features can both be StatusPartial for entirely different reasons,
// and a single boolean cannot say whether the gap is Update, Delete,
// Deploy, or something else. A feature with no Terraform resource or data
// source at all (for example a capability-activation concern with no
// public API) has every field false.
type Operations struct {
	Create   bool
	Read     bool
	Update   bool
	Delete   bool
	Import   bool
	Deploy   bool
	Undeploy bool
}

// Feature is one entry in the canonical support catalog: a single SAP
// Integration Suite object or capability area, and exactly how (or whether)
// this provider version manages it.
type Feature struct {
	// Key is a stable, hierarchical, machine-readable identifier, for
	// example "cloud_integration.value_mapping". Keys are never reused
	// for a different feature and never renamed once shipped in a
	// release, since practitioners' Terraform configurations reference
	// them via sapintegrationsuite_provider_feature.key.
	Key string
	// Domain groups related features, for example "cloud_integration" or
	// "security". It is the same as Key's first dot-separated segment
	// for every hierarchical key; flat, ungrouped keys (broad SAP
	// capability areas this project has not yet placed in a domain, such
	// as "integration_advisor") use "other_capability".
	Domain string
	// Name is a short, human-readable label.
	Name string
	// Description explains what the feature is and, briefly, its support
	// status in prose.
	Description   string
	SupportStatus SupportStatus
	// SupportReason explains SupportStatus when it is not StatusSupported.
	// Empty for StatusSupported features.
	SupportReason SupportReason
	// ResourceTypes lists the full Terraform resource type names (for
	// example "sapintegrationsuite_value_mapping") this provider
	// registers for this feature. Empty if none.
	ResourceTypes []string
	// DataSourceTypes lists the full Terraform data source type names
	// this provider registers for this feature. Empty if none.
	DataSourceTypes []string
	// PublicAPI reports whether SAP publishes a public, supported API for
	// this feature at all — independent of whether this provider
	// implements it. A feature can have PublicAPI true and
	// SupportStatus "unsupported" (a documented API this provider simply
	// has not gotten to yet), or PublicAPI false (no known public API
	// exists, so this provider cannot implement it no matter how much
	// engineering time is spent).
	PublicAPI bool
	// APIProtocol names the wire protocol of the public API when
	// PublicAPI is true and a single protocol applies, for example
	// "OData V2". Left empty when PublicAPI is false or the protocol is
	// not yet confirmed.
	APIProtocol string
	// Planned reports whether this feature has a concrete place on
	// ROADMAP.md, as opposed to being an open-ended possibility.
	Planned bool
	// Limitations lists specific, concrete caveats a practitioner should
	// know before relying on this feature, for example an unconfirmed
	// delete scope or a missing in-place update path. Empty for a
	// feature with no caveats beyond its SupportStatus/SupportReason.
	Limitations []string
	Operations  Operations
}

// Catalog is the complete, canonical list of SAP Integration Suite features
// this project has evaluated. It intentionally includes unsupported and
// out-of-scope features, not just what is implemented: the whole point of
// this catalog is to let a practitioner see gaps, not just capabilities.
//
// This is derived from, and must be kept in sync with, docs/resource-design.md,
// docs/sap-api-references.md, docs/api-capability-matrix.md,
// docs/provisioning-capability-matrix.md, docs/provider-scope.md, and
// ROADMAP.md — see CONTRIBUTING.md for the rule that every feature change
// updates this catalog in the same change.
var Catalog = []Feature{
	// --- Cloud Integration ---
	{
		Key:             "cloud_integration.integration_package",
		Domain:          "cloud_integration",
		Name:            "Integration Package",
		Description:     "A Cloud Integration content package that groups integration flows and other design-time artifacts.",
		SupportStatus:   StatusSupported,
		ResourceTypes:   []string{"sapintegrationsuite_integration_package"},
		DataSourceTypes: []string{"sapintegrationsuite_integration_package"},
		PublicAPI:       true,
		APIProtocol:     "OData V2",
		Limitations: []string{
			"Metadata update only applies to customer-created packages; SAP-provided packages are read-only by SAP's own design, not a provider limitation.",
		},
		Operations: Operations{Create: true, Read: true, Update: true, Delete: true, Import: true},
	},
	{
		Key:             "cloud_integration.integration_flow",
		Domain:          "cloud_integration",
		Name:            "Integration Flow",
		Description:     "An integration flow's design-time content, managed as file-based (ZIP) content.",
		SupportStatus:   StatusSupported,
		ResourceTypes:   []string{"sapintegrationsuite_integration_flow"},
		DataSourceTypes: []string{},
		PublicAPI:       true,
		APIProtocol:     "OData V2",
		Operations:      Operations{Create: true, Read: true, Update: true, Delete: true, Import: true},
	},
	{
		Key:             "cloud_integration.integration_flow_deployment",
		Domain:          "cloud_integration",
		Name:            "Integration Flow Deployment",
		Description:     "The runtime deployment state of an integration flow, managed independently of its design-time content.",
		SupportStatus:   StatusSupported,
		ResourceTypes:   []string{"sapintegrationsuite_integration_flow_deployment"},
		DataSourceTypes: []string{},
		PublicAPI:       true,
		APIProtocol:     "OData V2",
		Operations:      Operations{Create: true, Read: true, Update: true, Delete: true, Import: true, Deploy: true, Undeploy: true},
	},
	{
		Key:           "cloud_integration.value_mapping",
		Domain:        "cloud_integration",
		Name:          "Value Mapping",
		Description:   "A value mapping design-time artifact's content, managed as file-based content.",
		SupportStatus: StatusPartial,
		SupportReason: ReasonUnsafeTerraformLifecycle,
		ResourceTypes: []string{"sapintegrationsuite_value_mapping"},
		DataSourceTypes: []string{
			"sapintegrationsuite_value_mapping",
		},
		PublicAPI:   true,
		APIProtocol: "OData V2",
		Limitations: []string{
			"No confirmed in-place update: changing name, content, or content_hash replaces the resource (create a new artifact, then delete the old one) instead of calling an unverified PUT.",
			"SAP separately documents a ValueMappingDesigntimeArtifactSaveAsVersion action this provider does not yet use.",
			"Whether Delete removes only the active version or every version of the artifact is unconfirmed against a primary source.",
		},
		Operations: Operations{Create: true, Read: true, Update: false, Delete: true, Import: true},
	},
	{
		Key:             "cloud_integration.value_mapping_deployment",
		Domain:          "cloud_integration",
		Name:            "Value Mapping Deployment",
		Description:     "The runtime deployment state of a value mapping, managed independently of its design-time content.",
		SupportStatus:   StatusSupported,
		ResourceTypes:   []string{"sapintegrationsuite_value_mapping_deployment"},
		DataSourceTypes: []string{},
		PublicAPI:       true,
		APIProtocol:     "OData V2",
		Operations:      Operations{Create: true, Read: true, Update: true, Delete: true, Import: true, Deploy: true, Undeploy: true},
	},
	{
		Key:           "cloud_integration.value_mapping_entry",
		Domain:        "cloud_integration",
		Name:          "Value Mapping Entry",
		Description:   "Individual source/target value pairs inside a value mapping scheme, managed through UpsertValMaps, UpdateDefaultValMap, and DeleteValMaps.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     true,
		APIProtocol:   "OData V2 (actions)",
		Planned:       true,
		Limitations: []string{
			"Exact request/response payload shapes and DeleteValMaps' delete granularity are not confirmed against a reachable primary source.",
			"UpsertValMaps requires an already-existing source/target agency-identifier scheme, so entries cannot be managed independently of the artifact's own content.",
		},
	},
	{
		Key:             "cloud_integration.script_collection",
		Domain:          "cloud_integration",
		Name:            "Script Collection",
		Description:     "A reusable collection of Groovy/JavaScript scripts shared across multiple integration flows.",
		SupportStatus:   StatusSupported,
		ResourceTypes:   []string{"sapintegrationsuite_script_collection"},
		DataSourceTypes: []string{"sapintegrationsuite_script_collection"},
		PublicAPI:       true,
		APIProtocol:     "OData V2",
		Operations:      Operations{Create: true, Read: true, Update: true, Delete: true, Import: true},
	},
	{
		Key:             "cloud_integration.script_collection_deployment",
		Domain:          "cloud_integration",
		Name:            "Script Collection Deployment",
		Description:     "The runtime deployment state of a script collection, managed independently of its design-time content.",
		SupportStatus:   StatusSupported,
		ResourceTypes:   []string{"sapintegrationsuite_script_collection_deployment"},
		DataSourceTypes: []string{},
		PublicAPI:       true,
		APIProtocol:     "OData V2",
		Operations:      Operations{Create: true, Read: true, Update: true, Delete: true, Import: true, Deploy: true, Undeploy: true},
	},
	{
		Key:             "cloud_integration.message_mapping",
		Domain:          "cloud_integration",
		Name:            "Message Mapping",
		Description:     "A reusable, package-level message mapping artifact's design-time content — not the inline/local message mapping step an integration flow can also define directly inside its own content.",
		SupportStatus:   StatusSupported,
		ResourceTypes:   []string{"sapintegrationsuite_message_mapping"},
		DataSourceTypes: []string{"sapintegrationsuite_message_mapping"},
		PublicAPI:       true,
		APIProtocol:     "OData V2",
		Operations:      Operations{Create: true, Read: true, Update: true, Delete: true, Import: true},
	},
	{
		Key:             "cloud_integration.message_mapping_deployment",
		Domain:          "cloud_integration",
		Name:            "Message Mapping Deployment",
		Description:     "The runtime deployment state of a message mapping, managed independently of its design-time content.",
		SupportStatus:   StatusSupported,
		ResourceTypes:   []string{"sapintegrationsuite_message_mapping_deployment"},
		DataSourceTypes: []string{},
		PublicAPI:       true,
		APIProtocol:     "OData V2",
		Operations:      Operations{Create: true, Read: true, Update: true, Delete: true, Import: true, Deploy: true, Undeploy: true},
	},
	{
		Key:           "cloud_integration.service_endpoints",
		Domain:        "cloud_integration",
		Name:          "Service Endpoints",
		Description:   "Read-only lookup of a deployed integration flow's exposed runtime service endpoint URLs.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNotImplemented,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Planned:       true,
	},
	{
		Key:           "cloud_integration.message_processing_logs",
		Domain:        "cloud_integration",
		Name:          "Message Processing Logs",
		Description:   "Runtime message processing log records for deployed integration flows.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonOutOfScope,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Limitations: []string{
			"Monitoring/operational data, not infrastructure state this provider manages — see docs/provider-scope.md.",
		},
	},
	{
		Key:           "cloud_integration.message_stores",
		Domain:        "cloud_integration",
		Name:          "Message Stores / Data Stores",
		Description:   "Runtime message queue and data store payload content used by deployed integration flows.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonOutOfScope,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Limitations: []string{
			"Payload/queue content is operational data, not desired state this provider manages.",
		},
	},

	// --- Security ---
	{
		Key:           "security.access_policy",
		Domain:        "security",
		Name:          "Access Policy",
		Description:   "An access policy restricting which artifacts a role can access.",
		SupportStatus: StatusSupported,
		ResourceTypes: []string{"sapintegrationsuite_access_policy"},
		DataSourceTypes: []string{
			"sapintegrationsuite_access_policy",
		},
		PublicAPI:   true,
		APIProtocol: "OData V2",
		Limitations: []string{
			"reconciliation_status is surfaced whenever the API returns it, but Create/Update do not " +
				"poll it to a terminal state: SAP's Manage Access Policies UI documents replicating a " +
				"policy to one or more runtimes (Cloud Integration runtime, Integration Cell, Edge " +
				"Integration Cell) with a per-runtime Fail/Success/Pending reconciliation status, but " +
				"this project could not confirm that mechanism is exposed through the public " +
				"AccessPolicies OData API this provider uses, as opposed to being UI-only. Treat the " +
				"field as informational, not something to script against.",
			"Whether RoleName refers to a BTP role collection or an individual BTP role (assigned to " +
				"users via a role collection) has not been confirmed against SAP's OData $metadata; SAP's " +
				"own UI documentation describes associating \"a role\" with the policy \"using SAP " +
				"Business Technology Platform cockpit\", which this provider treats as an opaque string " +
				"it does not interpret or manage.",
		},
		Operations: Operations{Create: true, Read: true, Update: true, Delete: true, Import: true},
	},
	{
		Key:           "security.access_policy_reference",
		Domain:        "security",
		Name:          "Access Policy Reference",
		Description:   "A single artifact reference (attribute/operator/value match rule) on an access policy.",
		SupportStatus: StatusSupported,
		ResourceTypes: []string{"sapintegrationsuite_access_policy_reference"},
		DataSourceTypes: []string{
			"sapintegrationsuite_access_policy_reference",
		},
		PublicAPI:   true,
		APIProtocol: "OData V2",
		Limitations: []string{
			"No in-place update: every attribute is part of the reference's match condition and SAP " +
				"does not document updating a reference in place, so access_policy_id, artifact_type, " +
				"attribute, operator, and value all force replacement. This is a deliberate lifecycle " +
				"choice, not a missing capability.",
			"The exact wire-format casing of the Attribute (\"Name\"/\"Id\" vs. \"NAME\"/\"ID\") and " +
				"Operator (\"EQUALS\"/\"MATCHES\" vs. \"equals\"/\"matches\") enum values has not been " +
				"confirmed against a live tenant or OData $metadata; SAP's UI documentation confirms the " +
				"two values for each but only in prose/UI-label form.",
		},
		Operations: Operations{Create: true, Read: true, Update: false, Delete: true, Import: true},
	},
	{
		Key:    "security.user_credential",
		Domain: "security",
		Name:   "User Credential",
		Description: "A \"User Credentials\" security material artifact: a username/password credential " +
			"integration flow adapters use for outbound basic or username-token authentication.",
		SupportStatus:   StatusPartial,
		SupportReason:   ReasonPublicAPIIncomplete,
		ResourceTypes:   []string{"sapintegrationsuite_user_credential"},
		DataSourceTypes: []string{"sapintegrationsuite_user_credential"},
		PublicAPI:       true,
		APIProtocol:     "OData V2",
		Planned:         true,
		Limitations: []string{
			"The password is never returned by SAP's read API; password_wo/password_wo_version are " +
				"write-only attributes (Terraform CLI 1.11+ required) and drift on the password value " +
				"itself cannot be detected — only readable metadata (user, description, kind, " +
				"company_id) is compared on Read.",
			"Update is implemented as a full PUT redeploy, matching SAP's documented \"Edit\" action " +
				"for Credentials artifacts, and resends password_wo on every apply that touches this " +
				"resource (SAP documents re-entering the secret on every edit for the sibling OAuth2 " +
				"Client Credentials artifact; this provider assumes the same requirement here since it " +
				"could not find a documented exception for User Credentials).",
			"The Kind and CompanyId field names are corroborated by a documented third-party example " +
				"payload, not by this project's own inspection of a live tenant's OData $metadata; " +
				"verify against your tenant before relying on kind=\"SuccessFactors\"/\"OpenConnectors\" " +
				"in production. See docs/guides/security-content.md.",
			"Deployment status (SAP's UI shows Stored/Deployed/Error) is not exposed: this project could " +
				"not confirm the OData property name for it, and would rather omit a computed attribute " +
				"than expose one that is silently always empty.",
		},
		Operations: Operations{Create: true, Read: true, Update: true, Delete: true, Import: true},
	},
	{
		Key:    "security.oauth2_client_credential",
		Domain: "security",
		Name:   "OAuth2 Client Credential",
		Description: "An \"OAuth2 Client Credentials\" security material artifact: the client ID, " +
			"client secret, and token service URL an integration flow adapter uses for the OAuth2 " +
			"client credentials grant (RFC 6749) on outbound requests.",
		SupportStatus:   StatusPartial,
		SupportReason:   ReasonPublicAPIIncomplete,
		ResourceTypes:   []string{"sapintegrationsuite_oauth2_client_credential"},
		DataSourceTypes: []string{"sapintegrationsuite_oauth2_client_credential"},
		PublicAPI:       true,
		APIProtocol:     "OData V2",
		Planned:         true,
		Limitations: []string{
			"The client secret is never returned by SAP's read API; client_secret_wo/" +
				"client_secret_wo_version are write-only attributes (Terraform CLI 1.11+ required) and " +
				"drift on the secret value itself cannot be detected.",
			"Only name, description, token_service_url, client_id, client_secret, and scope are " +
				"exposed: SAP's UI additionally documents Grant Type placement, Client Authentication " +
				"mode (body vs. header), Resource, Audience, and up to 20 custom parameters, but this " +
				"project could not confirm their OData property names against $metadata or a documented " +
				"example payload, so they are deliberately unimplemented rather than guessed.",
			"Update is implemented as a full PUT redeploy and resends client_secret_wo on every apply " +
				"that touches this resource, matching SAP's documented requirement to re-enter the " +
				"client secret on every edit.",
			"OAuth2 Authorization Code and OAuth2 SAML Bearer Assertion are separate SAP artifact types " +
				"this provider does not implement: Authorization Code requires interactive human " +
				"authorization (see security.oauth2_authorization_code note in " +
				"docs/guides/security-content.md), and SAML Bearer Assertion's public API contract was " +
				"not confirmed.",
		},
		Operations: Operations{Create: true, Read: true, Update: true, Delete: true, Import: true},
	},
	{
		Key:    "security.keystore_entry",
		Domain: "security",
		Name:   "Keystore Entry",
		Description: "A certificate or key pair entry in the tenant's keystore (KeystoreEntries, " +
			"Keystores, KeystoreResources, HistoryKeystoreEntries in SAP's Security Content API).",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Planned:       true,
		Limitations: []string{
			"SAP's UI documentation (Keystore Monitor) and independent technical sources confirm a " +
				"public KeystoreEntries OData entity set exists with fields resembling Alias, Type, " +
				"ValidUntil/ValidNotAfter, SubjectDN, IssuerDN, KeyType, KeySize, SerialNumber, " +
				"SignatureAlgorithm, and one or more Fingerprints, but this project could not confirm " +
				"the exact property names and casing against $metadata, so no resource or data source " +
				"is implemented yet rather than shipping a guessed field mapping. Read-only metadata " +
				"discovery (a data source, not a resource) is the recommended next step: see " +
				"docs/guides/security-content.md and ROADMAP.md.",
			"HistoryKeystoreEntries is audit/history information, not a mutable artifact, and should " +
				"only ever become a read-only data source if implemented, never a resource.",
		},
	},
	{
		Key:    "security.certificate",
		Domain: "security",
		Name:   "Certificate",
		Description: "A standalone X.509 certificate keystore entry (as opposed to a key pair), " +
			"typically an uploaded root or intermediate CA certificate.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Planned:       true,
		Limitations: []string{
			"SAP's UI documents uploading a certificate to the keystore, but this project could not " +
				"confirm the OData create/update request shape (entity set name, PEM/DER encoding " +
				"expectations, alias field) with enough confidence to implement it safely.",
		},
	},
	{
		Key:    "security.key_pair",
		Domain: "security",
		Name:   "Key Pair",
		Description: "An SAP-generated key pair keystore entry (private key plus X.509 certificate " +
			"chain), as opposed to one uploaded from outside the tenant.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Planned:       true,
		Limitations: []string{
			"SAP's UI documents key pair generation (KeyPairGenerationRequests / KeyPairResources per " +
				"the public API catalog), and private keys generated this way are never downloadable, " +
				"which would make this a safe design (the resource can manage a key pair's existence " +
				"and metadata without ever handling private key material). This project could not " +
				"confirm the generation request's exact fields (key algorithm, key size, distinguished " +
				"name) or its synchronous-vs-asynchronous lifecycle against $metadata or a documented " +
				"example, so it is not yet implemented.",
		},
	},
	{
		Key:    "security.ssh_key",
		Domain: "security",
		Name:   "SSH Key",
		Description: "An SAP-generated SSH key pair keystore entry, used for SFTP public-key " +
			"authentication (SSHKeyGenerationRequests per the public API catalog).",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Planned:       true,
		Limitations: []string{
			"Same reasoning as security.key_pair: SAP's UI documents SSH key pair creation and this " +
				"provider's standing policy prefers modeling a declarative, persistent resource over a " +
				"one-shot \"generate\" action, but this project could not confirm the request/response " +
				"contract well enough to implement it yet.",
		},
	},
	{
		Key:    "security.certificate_chain",
		Domain: "security",
		Name:   "Certificate Chain",
		Description: "A certificate chain resource associated with a key pair (CertificateChainResources " +
			"per the public API catalog).",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Planned:       true,
		Limitations: []string{
			"This project could not confirm this entity's identity, upload/download semantics, or " +
				"relationship to security.key_pair against $metadata or documented examples.",
		},
	},
	{
		Key:    "security.certificate_user_mapping",
		Domain: "security",
		Name:   "Certificate-User Mapping",
		Description: "A mapping from a client certificate to an inbound user identity, used for " +
			"inbound client certificate authentication.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
		Planned:       false,
		Limitations: []string{
			"Reverified for this feature family: SAP's own documentation (\"Managing Certificate-to-" +
				"User Mappings\", \"Client Certificate Authentication and Certificate-to-User Mapping " +
				"(Inbound)\", \"Setting Up Inbound HTTP Connections with Certificate-to-User Mapping\") " +
				"exists only under the Neo environment, with no Cloud Foundry equivalent found in SAP's " +
				"published documentation set. This provider targets the Cloud Foundry environment (its " +
				"other Security Content resources use the Cloud Foundry \"/api/v1\" OData host), so this " +
				"catalog entry is corrected from its previous \"not_implemented\"/PublicAPI:true state " +
				"to \"no_public_api\": the feature cannot be implemented for this provider's target " +
				"environment, not merely unimplemented yet. If SAP publishes a Cloud Foundry " +
				"certificate-to-user-mapping API in the future, re-open this entry.",
		},
	},
	{
		Key:    "security.secure_parameter",
		Domain: "security",
		Name:   "Secure Parameter",
		Description: "A \"Secure Parameter\" security material artifact: an opaque confidential value " +
			"(for example for a custom adapter) deployed without an associated username.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     false,
		Planned:       false,
		Limitations: []string{
			"SAP's Manage Security Material UI documents creating and deploying a Secure Parameter " +
				"artifact, but this project found third-party evidence of at least one practitioner " +
				"receiving an OData error (\"could not find an entity set or function import for " +
				"SecureParameters\") when attempting to call it through the Security Content API, " +
				"suggesting the entity set name is different from the obvious guess, is not exposed in " +
				"every API version, or is not publicly documented at all. Marked PublicAPI: false " +
				"pending confirmation, not because the UI feature doesn't exist, but because a callable " +
				"public OData contract for it was not confirmed. If a public contract is confirmed, this " +
				"would be a strong write-only-attribute candidate (value_wo/value_wo_version), the same " +
				"shape as security.user_credential's password.",
		},
	},
	{
		Key:    "security.known_hosts",
		Domain: "security",
		Name:   "Known Hosts (SSH)",
		Description: "The SSH \"known_hosts\" file artifact used to validate SFTP server host keys for " +
			"outbound SFTP connections.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     false,
		Planned:       false,
		Limitations: []string{
			"SAP's Manage Security Material UI documents uploading and downloading a Known Hosts " +
				"artifact (file content, not a structured entity with individually settable fields), " +
				"but this project could not confirm a public OData entity set or REST endpoint for it " +
				"with enough confidence to implement create/update/delete safely.",
		},
	},

	// --- Partner Directory ---
	{
		Key:           "partner_directory.partner",
		Domain:        "partner_directory",
		Name:          "Partner",
		Description:   "A Partner ID (Pid) known to the tenant's Partner Directory.",
		SupportStatus: StatusReadOnly,
		SupportReason: ReasonUnsafeTerraformLifecycle,
		DataSourceTypes: []string{
			"sapintegrationsuite_partner",
			"sapintegrationsuite_partners",
		},
		PublicAPI:   true,
		APIProtocol: "OData V2",
		Limitations: []string{
			"No resource: SAP documents no confirmed create operation for Partners — a Pid comes " +
				"into existence implicitly the first time a StringParameter, BinaryParameter, " +
				"AlternativePartner, AuthorizedUser, or UserCredentialParameter references it.",
			"Deleting a Pid is documented as cascading to every entity belonging to it, which is " +
				"the other reason this stays read-only: a Partner resource's Destroy could erase " +
				"content owned by an entirely different Terraform module.",
		},
		Operations: Operations{Read: true},
	},
	{
		Key:           "partner_directory.string_parameter",
		Domain:        "partner_directory",
		Name:          "Partner Directory String Parameter",
		Description:   "A named text value scoped to a Partner ID (Pid).",
		SupportStatus: StatusSupported,
		ResourceTypes: []string{"sapintegrationsuite_partner_string_parameter"},
		DataSourceTypes: []string{
			"sapintegrationsuite_partner_string_parameter",
			"sapintegrationsuite_partner_string_parameters",
		},
		PublicAPI:   true,
		APIProtocol: "OData V2",
		Limitations: []string{
			"Partner Directory data is stored unencrypted; do not store passwords, secrets, or " +
				"other sensitive values in a string parameter — see " +
				"docs/guides/partner-directory.md.",
		},
		Operations: Operations{Create: true, Read: true, Update: true, Delete: true, Import: true},
	},
	{
		Key:           "partner_directory.binary_parameter",
		Domain:        "partner_directory",
		Name:          "Partner Directory Binary Parameter",
		Description:   "A named binary value (for example an XSD schema or certificate) scoped to a Partner ID (Pid).",
		SupportStatus: StatusSupported,
		ResourceTypes: []string{"sapintegrationsuite_partner_binary_parameter"},
		DataSourceTypes: []string{
			"sapintegrationsuite_partner_binary_parameter",
		},
		PublicAPI:   true,
		APIProtocol: "OData V2",
		Limitations: []string{
			"Partner Directory data is stored unencrypted; do not store secrets, private keys, or " +
				"other sensitive content — see docs/guides/partner-directory.md.",
			"SAP documents a 260 KB maximum decoded value size; this provider validates it before " +
				"upload rather than letting an oversized payload fail against the live API.",
		},
		Operations: Operations{Create: true, Read: true, Update: true, Delete: true, Import: true},
	},
	{
		Key:           "partner_directory.alternative_partner",
		Domain:        "partner_directory",
		Name:          "Alternative Partner",
		Description:   "A mapping from an external identity tuple (agency, scheme, external_id) to an internal Partner ID (Pid).",
		SupportStatus: StatusSupported,
		ResourceTypes: []string{"sapintegrationsuite_alternative_partner"},
		DataSourceTypes: []string{
			"sapintegrationsuite_alternative_partner",
		},
		PublicAPI:   true,
		APIProtocol: "OData V2",
		Limitations: []string{
			"SAP's actual OData key is a hex encoding of agency/scheme/external_id, not the plain " +
				"strings; this provider computes and hides that transform, and uses the three hex " +
				"segments as the Terraform import ID so any character combination round-trips " +
				"unambiguously — see docs/guides/partner-directory.md.",
		},
		Operations: Operations{Create: true, Read: true, Update: true, Delete: true, Import: true},
	},
	{
		Key:           "partner_directory.authorized_user",
		Domain:        "partner_directory",
		Name:          "Partner Directory Authorized User",
		Description:   "A mapping from a communication user to the Partner ID (Pid) that user is authorized to act as.",
		SupportStatus: StatusSupported,
		ResourceTypes: []string{"sapintegrationsuite_partner_authorized_user"},
		DataSourceTypes: []string{
			"sapintegrationsuite_partner_authorized_user",
		},
		PublicAPI:   true,
		APIProtocol: "OData V2",
		Limitations: []string{
			"Whether SAP normalizes the User value's case internally was not confirmed against a " +
				"primary source; this provider passes it through exactly as configured, without " +
				"normalizing it.",
			"Manages only the Partner Directory mapping, never the underlying BTP user, OAuth " +
				"client, or communication user credential itself.",
		},
		Operations: Operations{Create: true, Read: true, Update: true, Delete: true, Import: true},
	},
	{
		Key:           "partner_directory.user_credential_parameter",
		Domain:        "partner_directory",
		Name:          "Partner Directory User Credential Parameter",
		Description:   "A communication username/password credential scoped to a Partner ID (Pid).",
		SupportStatus: StatusPartial,
		SupportReason: ReasonUnsafeTerraformLifecycle,
		ResourceTypes: []string{"sapintegrationsuite_partner_user_credential_parameter"},
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Limitations: []string{
			"The password is a write-only attribute (password_wo): Terraform never stores it in " +
				"plan or state, and this provider never requests or reads a password back from " +
				"SAP, which does not document returning one. Requires Terraform CLI 1.11 or later.",
			"No in-place update: no public API for changing an existing credential's password was " +
				"confirmed, so rotating it (via the paired password_wo_version attribute) replaces " +
				"the resource — delete the old credential, then create a new one.",
			"UserCredentialParameter cannot be combined with other Partner Directory entity types " +
				"in a single OData batch (ChangeSet) request; this provider always issues it " +
				"standalone.",
			"Import recovers partner_id, parameter_id, and user, but never the password: a " +
				"configuration applied right after import must still supply password_wo and a " +
				"password_wo_version, which plans as a replacement even though nothing server-side " +
				"has actually changed.",
		},
		Operations: Operations{Create: true, Read: true, Update: false, Delete: true, Import: true},
	},

	// --- Classic API Management ---
	{
		Key:           "api_management.classic.api_provider",
		Domain:        "api_management_classic",
		Name:          "API Provider (classic API Management)",
		Description:   "A classic API Management backend/API provider system definition.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNotImplemented,
		PublicAPI:     true,
		APIProtocol:   "REST/OData mixed",
		Planned:       true,
	},
	{
		Key:           "api_management.classic.api_proxy",
		Domain:        "api_management_classic",
		Name:          "API Proxy (classic API Management)",
		Description:   "A classic API Management API proxy definition.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNotImplemented,
		PublicAPI:     true,
		APIProtocol:   "REST/OData mixed",
		Planned:       true,
	},
	{
		Key:           "api_management.classic.api_product",
		Domain:        "api_management_classic",
		Name:          "API Product (classic API Management)",
		Description:   "A classic API Management API product bundling one or more API proxies.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNotImplemented,
		PublicAPI:     true,
		APIProtocol:   "REST/OData mixed",
		Planned:       true,
	},
	{
		Key:           "api_management.classic.key_value_map",
		Domain:        "api_management_classic",
		Name:          "Key Value Map (classic API Management)",
		Description:   "A classic API Management key-value map used for runtime configuration lookups.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNotImplemented,
		PublicAPI:     true,
		APIProtocol:   "REST/OData mixed",
		Planned:       true,
	},

	// --- New API Gateway model ---
	{
		Key:           "api_gateway.api_artifact",
		Domain:        "api_gateway",
		Name:          "API Artifact (API Gateway)",
		Description:   "An API-artifact-centric design-time object in SAP's newer API Gateway model.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     false,
		Planned:       true,
		Limitations: []string{
			"Existence confirmed via UI/feature documentation only; a public design-time API has not been confirmed in enough detail for a stable Terraform schema.",
		},
	},
	{
		Key:           "api_gateway.api_artifact_deployment",
		Domain:        "api_gateway",
		Name:          "API Artifact Deployment (API Gateway)",
		Description:   "The runtime deployment state of an API Gateway API artifact.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     false,
		Planned:       true,
	},
	{
		Key:           "api_gateway.api_policy",
		Domain:        "api_gateway",
		Name:          "API Policy (API Gateway)",
		Description:   "A policy attached to an API Gateway API artifact.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     false,
		Planned:       true,
	},

	// --- Integration Cell / Edge Integration Cell (content/config, once active) ---
	{
		Key:           "integration_cell.runtime",
		Domain:        "integration_cell",
		Name:          "Integration Cell Runtime",
		Description:   "Runtime status and configuration of an already-activated Integration Cell, distinct from activating the capability itself.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
		Limitations: []string{
			"No public status or configuration API was found for Integration Cell runtime content, only UI-facing operations.",
		},
	},
	{
		Key:           "edge_integration_cell.registration",
		Domain:        "edge_integration_cell",
		Name:          "Edge Integration Cell Registration",
		Description:   "SAP-side registration and runtime association of an Edge Integration Cell, never the customer-managed Kubernetes workloads themselves.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
		Limitations: []string{
			"Registration is a guided UI plus Helm-based bootstrap process; no public capability-activation or registration API was found.",
		},
	},

	// --- Capability activation/provisioning (subscription-level toggles) ---
	{
		Key:           "capabilities.cloud_integration",
		Domain:        "capability_provisioning",
		Name:          "Cloud Integration Capability Activation",
		Description:   "Activating the Cloud Integration capability itself within an Integration Suite tenant.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
		Limitations: []string{
			"Activated as part of the Integration Suite subscription/onboarding wizard; stays a manual, one-time bootstrap step.",
		},
	},
	{
		Key:           "capabilities.api_management",
		Domain:        "capability_provisioning",
		Name:          "API Management Capability Activation",
		Description:   "Activating the classic API Management capability itself within an Integration Suite tenant.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
		Limitations: []string{
			"Documented as a UI step (Integration Suite → Manage Capabilities → activate API Management); no public API found.",
		},
	},
	{
		Key:           "capabilities.api_gateway",
		Domain:        "capability_provisioning",
		Name:          "API Gateway Capability Activation",
		Description:   "Activating the newer API Gateway / API Artifacts capability itself within an Integration Suite tenant.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
	},
	{
		Key:           "capabilities.integration_cell",
		Domain:        "capability_provisioning",
		Name:          "Integration Cell Capability Activation",
		Description:   "Activating the Integration Cell capability itself within an Integration Suite tenant.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
		Limitations: []string{
			"SAP's own documentation describes activation as choosing Activate in Integration Suite → Settings → Runtime; no public API was found.",
		},
	},
	{
		Key:           "capabilities.edge_integration_cell",
		Domain:        "capability_provisioning",
		Name:          "Edge Integration Cell Capability Activation",
		Description:   "Activating the Edge Integration Cell capability itself within an Integration Suite tenant.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
	},

	// --- Other Integration Suite capability areas (research status only) ---
	{
		Key:           "integration_advisor",
		Domain:        "other_capability",
		Name:          "Integration Advisor",
		Description:   "SAP's collaborative interface-content-design capability.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     false,
	},
	{
		Key:           "trading_partner_management",
		Domain:        "other_capability",
		Name:          "Trading Partner Management",
		Description:   "SAP's B2B trading partner management capability.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     false,
	},
	{
		Key:           "event_mesh",
		Domain:        "other_capability",
		Name:          "Event Mesh",
		Description:   "SAP's event broker service for asynchronous, event-driven integration.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonOutOfScope,
		PublicAPI:     false,
		Limitations: []string{
			"Likely a separate BTP service outside this provider's Integration Suite content/capability boundary rather than a Cloud Integration design-time concern.",
		},
	},
	{
		Key:           "integration_assessment",
		Domain:        "other_capability",
		Name:          "Integration Assessment",
		Description:   "SAP's integration landscape assessment capability.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     false,
	},
	{
		Key:           "migration_assessment",
		Domain:        "other_capability",
		Name:          "Migration Assessment",
		Description:   "SAP's integration migration assessment capability.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     false,
	},
	{
		Key:           "data_space_integration",
		Domain:        "other_capability",
		Name:          "Data Space Integration",
		Description:   "SAP's data space connectivity capability within Integration Suite.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     false,
	},
	{
		Key:           "open_connectors",
		Domain:        "other_capability",
		Name:          "Open Connectors",
		Description:   "SAP's third-party SaaS connectivity capability within Integration Suite.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     false,
		Limitations: []string{
			"Not yet investigated by this project; listed so its absence is visible rather than silently omitted.",
		},
	},
}

// Lookup returns the Feature registered under key, and whether it was
// found.
func Lookup(key string) (Feature, bool) {
	for _, f := range Catalog {
		if f.Key == key {
			return f, true
		}
	}
	return Feature{}, false
}
