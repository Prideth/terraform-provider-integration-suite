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
		Key:    "cloud_integration.service_endpoints",
		Domain: "cloud_integration",
		Name:   "Service Endpoints",
		Description: "Read-only discovery of the runtime service endpoints (entry point URLs and " +
			"API definition links) SAP generates for deployed Cloud Integration content.",
		SupportStatus:   StatusReadOnly,
		SupportReason:   ReasonUnsafeTerraformLifecycle,
		DataSourceTypes: []string{"sapintegrationsuite_service_endpoints"},
		PublicAPI:       true,
		APIProtocol:     "OData V2",
		Limitations: []string{
			"Discovery only, by design: SAP generates service endpoints from deployed content and " +
				"there is no create/update/delete API for them, so this provider intentionally has " +
				"no matching resource type — see docs/guides/service-endpoints.md.",
			"No single-endpoint (sapintegrationsuite_service_endpoint) data source exists: this " +
				"project could not confirm that Name uniquely and stably identifies exactly one " +
				"service endpoint, so only the collection data source " +
				"(sapintegrationsuite_service_endpoints, with optional name/protocol filters) is " +
				"implemented, to avoid a lookup data source that silently returns the wrong result " +
				"if more than one endpoint ever matches.",
			"The EntryPoint/APIDefinition Url property's JSON casing is confirmed from SAP's own " +
				"open-source Piper library parsing a live response; the ApiDefinitions entity's Url " +
				"casing specifically is inferred by consistency rather than independently confirmed " +
				"from an example touching that entity — see docs/sap-api-references.md.",
		},
		Operations: Operations{Read: true},
	},
	{
		Key:    "cloud_integration.integration_adapter",
		Domain: "cloud_integration",
		Name:   "Integration Adapter",
		Description: "A custom Integration Adapter design-time artifact (a *.esa archive built with " +
			"the SAP Adapter SDK), imported into a Cloud Integration package. Cloud Foundry " +
			"environment only.",
		SupportStatus:   StatusPartial,
		SupportReason:   ReasonPublicAPIIncomplete,
		ResourceTypes:   []string{"sapintegrationsuite_integration_adapter"},
		DataSourceTypes: []string{"sapintegrationsuite_integration_adapter"},
		PublicAPI:       true,
		APIProtocol:     "OData V2",
		Planned:         true,
		Limitations: []string{
			"This provider's evidence base for this entity is thinner than for the sibling " +
				"design-time artifact types it manages: SAP's own \"Integration Adapter Example " +
				"Requests, Cloud Foundry Environment\" documentation shows only Delete (confirming " +
				"the entity is keyed by Id alone, not the composite (Id, Version) key every other " +
				"design-time artifact type in this API uses) and the Deploy action — no Create or " +
				"Read example was found. Create is implemented by strong analogy to every sibling " +
				"artifact type's confirmed PackageId/ArtifactContent POST body shape, corroborated " +
				"by a third-party technical source, not by an SAP-published example request for " +
				"this specific entity. See docs/guides/integration-adapters.md.",
			"No in-place update: SAP documents that importing an ID that already exists on the " +
				"tenant is rejected as an error, which is positive evidence against a working " +
				"reimport-to-update flow, so every attribute is RequiresReplace rather than an " +
				"unverified PUT/PATCH.",
			"type and application are not validated against a fixed set of values: SAP's " +
				"documentation confirms both exist as UI dropdowns with example values (Analytics, " +
				"CRM, ERP, ... for type; Slack, ... for application) but does not confirm whether " +
				"the underlying property is a closed enum or free-form text.",
			"Distinct from SAP Business Accelerator Hub prebundled adapters (a different import/" +
				"auto-deploy lifecycle reached from inside the integration flow editor) and from " +
				"Integration Suite capability activation — see docs/guides/integration-adapters.md " +
				"for why these are not the same feature.",
			"There is no sapintegrationsuite_integration_adapters collection data source: no " +
				"confirmed list/filter contract for this entity set was found, unlike " +
				"ServiceEndpoints' documented Name/Protocol filters.",
		},
		Operations: Operations{Create: true, Read: true, Delete: true, Import: true},
	},
	{
		Key:    "cloud_integration.integration_adapter_deployment",
		Domain: "cloud_integration",
		Name:   "Integration Adapter Deployment",
		Description: "The runtime deployment state of a custom Integration Adapter, independent of " +
			"its design-time content lifecycle.",
		SupportStatus:   StatusPartial,
		SupportReason:   ReasonPublicAPIIncomplete,
		ResourceTypes:   []string{"sapintegrationsuite_integration_adapter_deployment"},
		DataSourceTypes: []string{},
		PublicAPI:       true,
		APIProtocol:     "OData V2",
		Planned:         true,
		Limitations: []string{
			"Deploy is confirmed directly from SAP's own example request: POST " +
				"DeployIntegrationAdapterDesigntimeArtifact?Id='...', singular \"Artifact\" (matching " +
				"every sibling deploy action in this API), with no Version query parameter (unlike " +
				"every sibling deploy action, consistent with Id being this entity's only confirmed " +
				"key).",
			"Runtime status polling and undeploy reuse the same shared IntegrationRuntimeArtifacts " +
				"entity every other *_deployment resource in this provider polls/undeploys through, " +
				"by analogy — this project could not independently confirm that a deployed custom " +
				"adapter surfaces through that same shared entity as opposed to an adapter-specific " +
				"status/undeploy mechanism (for example BuildAndDeployStatus). See " +
				"docs/guides/integration-adapters.md.",
			"Whether a fresh deployment's runtime status becomes visible immediately or after a " +
				"build/deploy delay specific to adapters (as opposed to ordinary content deployment) " +
				"was not confirmed.",
		},
		Operations: Operations{Create: true, Read: true, Delete: true, Deploy: true, Undeploy: true},
	},
	{
		Key:    "cloud_integration.custom_tag_configuration",
		Domain: "cloud_integration",
		Name:   "Custom Tag Configuration",
		Description: "The tenant-wide set of custom tags integration package owners are asked, or " +
			"required, to classify their packages with.",
		SupportStatus:   StatusPartial,
		SupportReason:   ReasonUnsafeTerraformLifecycle,
		ResourceTypes:   []string{"sapintegrationsuite_custom_tag_configuration"},
		DataSourceTypes: []string{"sapintegrationsuite_custom_tag_configuration"},
		PublicAPI:       true,
		APIProtocol:     "OData V2",
		Planned:         true,
		Limitations: []string{
			"No confirmed delete or clear operation exists for this entity anywhere in SAP's public " +
				"documentation. Destroying this resource in Terraform returns an explicit error " +
				"rather than guessing that an empty overwrite means delete, or silently dropping " +
				"Terraform state while leaving the tenant's configuration untouched — see " +
				"docs/guides/custom-tag-configurations.md.",
			"Create and Update both use the same confirmed POST .../CustomTagConfigurations?" +
				"Overwrite=true operation (SAP documents no separate plain-POST-without-Overwrite " +
				"path this provider relies on), sending the complete desired tag list every time. " +
				"Whether Overwrite=true is a full replace (removing tags not present in the new " +
				"list) is strongly implied by the word \"Overwrite\" and by the fact the documented " +
				"payload is the complete configuration, not a delta, but SAP's documentation never " +
				"uses the word \"replace\" explicitly.",
			"Whether tag names must be unique, whether permitted values are case-sensitive, and " +
				"whether SAP preserves submitted ordering are all unconfirmed by SAP's " +
				"documentation. This provider enforces tag-name uniqueness itself and treats " +
				"ordering (of both tags and permitted values) as not semantically meaningful, " +
				"modeling both as unordered Terraform sets so a reordered response never produces " +
				"a spurious diff.",
			"One documented example response shows a single-element permittedValues array " +
				"containing a comma-separated string (\"Mr. Bean, Ms. Bean\") rather than two " +
				"separate array elements; this is treated as a documentation artifact, not a " +
				"confirmed wire format, since every other array-typed field in SAP's own examples " +
				"(and everywhere else in this provider) uses one array element per value.",
		},
		Operations: Operations{Create: true, Read: true, Update: true, Delete: false, Import: true},
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
		Key:    "cloud_integration.message_stores",
		Domain: "cloud_integration",
		Name:   "Message Store Entries / JMS Resources",
		Description: "Runtime persisted-message-store entries (created by the Persist step) and " +
			"JMS queue resource metadata used by deployed integration flows. Data Stores, Data " +
			"Store Entries, Variables, and Number Ranges — all part of the same broader Message " +
			"Stores API family — each have their own dedicated catalog entry; see " +
			"cloud_integration.data_store, cloud_integration.data_store_entry, " +
			"cloud_integration.variable, and cloud_integration.number_range.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonOutOfScope,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Limitations: []string{
			"Payload/queue content is operational data, not desired state this provider manages.",
		},
	},
	{
		Key:    "cloud_integration.number_range",
		Domain: "cloud_integration",
		Name:   "Number Range",
		Description: "A Number Ranges object: generates unique interchange numbers for outbound " +
			"EDI/EDIFACT documents, with a static configuration (min/max/description/rotate/" +
			"field length) and a live runtime counter (CurrentValue, the UI's \"Next Value\") " +
			"that advances as deployed content consumes numbers.",
		SupportStatus: StatusPartial,
		SupportReason: ReasonUnsafeTerraformLifecycle,
		ResourceTypes: []string{"sapintegrationsuite_number_range"},
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Planned:       true,
		Limitations: []string{
			"SAP documents no GET operation for this entity anywhere — unlike every sibling entity " +
				"in the same Message Stores API family (DataStores, DataStoreEntries, Variables all " +
				"have documented GET examples), NumberRanges has none in SAP's curated \"Message " +
				"Stores Example Requests\" index or anywhere else this project found. Without a GET, " +
				"this resource's Read is a documented no-op that trusts local state rather than " +
				"verifying anything against the tenant: it cannot detect drift, and 'terraform " +
				"import' is rejected outright rather than silently leaving most attributes unknown.",
			"SAP documents no delete operation for this entity either; the Monitor UI shows an " +
				"\"Undeploy\" action with no confirmed REST equivalent. 'terraform destroy' returns " +
				"an explicit error rather than guessing at an unconfirmed operation — see " +
				"docs/guides/runtime-stores-and-number-ranges.md.",
			"The runtime counter is handled as a write-only, version-gated attribute " +
				"(current_value_wo / current_value_wo_version), pushed to SAP only when the " +
				"version marker changes. An ordinary Update that only changes description/" +
				"min_value/max_value/rotate/field_length omits CurrentValue from the request body " +
				"entirely, rather than resending a value this provider has no way to confirm is " +
				"still current — SAP's documentation does not confirm whether an omitted field on " +
				"this entity's PUT is preserved unchanged or reset, which is an inherent, " +
				"documented risk of this design, not a guess this provider is hiding.",
			"The UI additionally documents a \"Runtimes\" multi-select deployment field (Cloud " +
				"Integration plus any active Edge Integration Cell nodes) with no visible " +
				"counterpart in either of SAP's two documented API examples (Add, Update); this " +
				"provider's client always targets the default runtime implicitly and does not " +
				"expose runtime/location selection.",
		},
		Operations: Operations{Create: true, Update: true},
	},
	{
		Key:    "cloud_integration.variable",
		Domain: "cloud_integration",
		Name:   "Variable",
		Description: "A tenant-persisted runtime value written by an integration flow's \"Write " +
			"Variables\" step, shared across steps of the same flow (local) or across every flow " +
			"deployed on the tenant (global).",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonOutOfScope,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Limitations: []string{
			"SAP documents exactly one public operation for this entity: GET .../Variables(...)/" +
				"$value, which downloads the raw value with no structured metadata (no Visibility/ " +
				"UpdatedAt/RetainUntil fields are returned by this endpoint). There is no collection " +
				"GET, no POST, no PUT, and no confirmed DELETE — Variables are created and updated " +
				"exclusively by deployed integration flow content, an entirely different ownership " +
				"domain than Terraform-managed infrastructure.",
			"A read-only data source was deliberately not implemented: the only confirmed read " +
				"operation returns nothing but the raw runtime value itself, with no safer " +
				"metadata-only alternative available, and this provider does not place arbitrary " +
				"runtime business values into Terraform state merely because an API can return " +
				"them — see docs/provider-scope.md.",
		},
	},
	{
		Key:    "cloud_integration.data_store",
		Domain: "cloud_integration",
		Name:   "Data Store",
		Description: "A tenant-persisted runtime container of Data Store Entries, created " +
			"implicitly by an integration flow's Data Store Write step (or an XI adapter's " +
			"Temporary Storage option) the first time it writes an entry.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonOutOfScope,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Limitations: []string{
			"SAP documents exactly one public operation for this entity: GET .../DataStores?" +
				"overdueonly=true, an aggregate monitoring endpoint (NumberOfMessages/" +
				"NumberOfOverdueMessages per store) — the same class of runtime monitoring data as " +
				"cloud_integration.message_processing_logs, not configuration. There is no " +
				"independent declarative creation API: a Data Store comes into existence only as a " +
				"side effect of deployed integration flow content.",
			"The DataStores API does not support $filter, $inlinecount, $orderby, $skip, $top, " +
				"$expand, or $select (confirmed directly from SAP's own documentation).",
		},
	},
	{
		Key:    "cloud_integration.data_store_entry",
		Domain: "cloud_integration",
		Name:   "Data Store Entry",
		Description: "A single runtime message (payload and headers) persisted inside a Data " +
			"Store by an integration flow's Data Store Write step, read back by a Data Store Get " +
			"or Select step, and deleted only by a Data Store Delete step.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonOutOfScope,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Limitations: []string{
			"SAP documents only GET operations for this entity (a single entry by composite key, " +
				"all entries for a store, and all entries for a message ID) — every field (Status, " +
				"MessageId, DueAt, CreatedAt, RetainUntil) is runtime business-message state, not " +
				"infrastructure desired state. Delete exists only as a design-time integration flow " +
				"step (entry-by-entry or bulk via an XPath-derived ID list at runtime), never as a " +
				"REST call this provider could wrap in a Terraform destroy.",
			"Deliberately out_of_scope rather than not_implemented: this provider does not manage " +
				"business message payloads or place them into Terraform state, and 'terraform " +
				"destroy' semantics are not an excuse to expose operational message deletion as " +
				"desired infrastructure state — see docs/provider-scope.md.",
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
