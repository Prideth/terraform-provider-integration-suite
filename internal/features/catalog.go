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
		Key:           "security.user_credential",
		Domain:        "security",
		Name:          "User Credential",
		Description:   "A user credential security material artifact used by integration flow adapters.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNotImplemented,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Planned:       true,
		Limitations: []string{
			"Secret values are never returned by SAP's read API; requires write-only attribute semantics not yet designed.",
		},
	},
	{
		Key:           "security.oauth2_client_credential",
		Domain:        "security",
		Name:          "OAuth2 Client Credential",
		Description:   "An OAuth2 client credential security material artifact used by integration flow adapters.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNotImplemented,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Planned:       true,
	},
	{
		Key:           "security.keystore_entry",
		Domain:        "security",
		Name:          "Keystore Entry",
		Description:   "A certificate keystore entry security material artifact.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNotImplemented,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Planned:       true,
	},
	{
		Key:           "security.certificate_user_mapping",
		Domain:        "security",
		Name:          "Certificate-User Mapping",
		Description:   "A mapping from a client certificate to an inbound user identity.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNotImplemented,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Planned:       true,
	},

	// --- Partner Directory ---
	{
		Key:           "partner_directory.entry",
		Domain:        "partner_directory",
		Name:          "Partner Directory Entry",
		Description:   "A trading-partner-style directory entry used by B2B-oriented integration flows.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNotImplemented,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
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
