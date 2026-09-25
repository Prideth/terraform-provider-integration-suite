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
			"Runtime targeting is not managed. SAP stores the runtimes a policy is replicated to " +
				"(Cloud Integration runtime, Integration Cell, Edge Integration Cell) in the " +
				"AccessPolicyRuntimeAssignments navigation property, whose schema and write semantics " +
				"are not publicly documented; see edge_integration_cell.access_policy_replication. " +
				"Earlier releases exposed a reconciliation_status attribute; it was removed because the " +
				"AccessPolicies entity has no such property.",
			"role_name is matched by SAP against the Values attribute of a BTP custom role. That role " +
				"and the role collection granting it belong to the SAP/btp provider; this provider only " +
				"passes the string through.",
			"role_name forces replacement: SAP's API accepts RoleName in the PUT payload, but no " +
				"documentation confirms that changing it renames the policy rather than being rejected. " +
				"Replacing a policy also deletes its artifact references on SAP's side.",
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
			"No in-place update: SAP's UI can edit a reference, but the only public contract found " +
				"(SAP's own CI/CD tooling) creates and deletes references and never updates one, so every " +
				"attribute forces replacement. Replacement deletes the old reference before creating the " +
				"new one, which briefly leaves the matched artifacts unprotected by it.",
			"artifact_type, attribute and operator are passed through verbatim. Confirmed wire values " +
				"are Type INTEGRATION_FLOW, ConditionAttribute Name and ConditionType exactString; SAP " +
				"publishes no complete list of the other constants (for example for Integration Package, " +
				"Message Queue or regular-expression matching). Read a UI-created reference with the data " +
				"source to learn them.",
			"Releases before this correction sent invented field names (ArtifactType, Attribute, " +
				"Operator, Value) and could not have worked against a real tenant.",
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
		Description: "Any entry (certificate, SAP-generated key pair, or other RSA/DSA/EC-keyed " +
			"entry) in the tenant's keystore, read-only.",
		SupportStatus: StatusReadOnly,
		SupportReason: ReasonUnsafeTerraformLifecycle,
		DataSourceTypes: []string{
			"sapintegrationsuite_keystore_entry",
			"sapintegrationsuite_keystore_entries",
		},
		PublicAPI:   true,
		APIProtocol: "OData V2",
		Limitations: []string{
			"Confirmed fields (verbatim from SAP's own \"Get All Keystore Entries\"/\"Get Keystore " +
				"Entry by Alias\" documented example): Hexalias, Alias, KeyType, KeySize, " +
				"ValidNotBefore, ValidNotAfter. SAP's own example response is truncated (\"....\"), " +
				"and prose elsewhere mentions Subject DN/Issuer DN/last-modified information existing " +
				"without giving their exact JSON property names, so those are not exposed as raw SAP " +
				"fields — sapintegrationsuite_certificate derives subject/issuer/serial/fingerprint " +
				"locally instead, by parsing the certificate bytes with Go's crypto/x509.",
			"No API field distinguishes SAP-owned from tenant-administrator-owned entries. This " +
				"provider does not guess at ownership; sapintegrationsuite_certificate and " +
				"sapintegrationsuite_key_pair rely on and surface SAP's own server-side protection " +
				"when an Update/Delete is attempted against a protected entry, rather than trying to " +
				"detect it in advance — see docs/guides/security-content.md.",
			"No resource: this entity represents fundamentally different object types (plain " +
				"certificates, generated key pairs) with different lifecycles, so a single mutable " +
				"sapintegrationsuite_keystore_entry resource was deliberately not created — see " +
				"security.certificate and security.key_pair instead.",
		},
	},
	{
		Key:    "security.certificate",
		Domain: "security",
		Name:   "Certificate",
		Description: "A standalone X.509 certificate keystore entry (as opposed to a key pair), " +
			"typically an uploaded root or intermediate CA certificate.",
		SupportStatus: StatusSupported,
		ResourceTypes: []string{"sapintegrationsuite_certificate"},
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Planned:       true,
		Limitations: []string{
			"Create/Update both confirmed via SAP's own \"Import and Update Certificate\" " +
				"documentation: PUT CertificateResources('<hexalias>')/$value, raw PEM body — SAP's " +
				"documentation explicitly flags the PUT-creates-an-entity quirk. SAP's documented " +
				"example body is enclosed in literal square brackets; this project treats those as " +
				"documentation formatting, not literal bytes to send, since no other example in SAP's " +
				"documentation uses that convention.",
			"Delete uses the documented keystore mass-deletion operation (PUT " +
				"KeystoreResources('system')?deleteEntries=true) with exactly the one alias this " +
				"resource owns — there is no documented per-entry DELETE for this entity.",
			"Drift detection compares a locally-computed SHA-256 fingerprint of the certificate's DER " +
				"bytes, not raw PEM text, so line-wrapping/line-ending differences between your " +
				"configuration and SAP's own re-serialization never produce a spurious diff; a " +
				"genuinely different certificate does surface as drift.",
			"certificate is intentionally not Sensitive: public X.509 certificate content is not " +
				"confidential. This resource never handles a private key.",
		},
		Operations: Operations{Create: true, Read: true, Update: true, Delete: true, Import: true},
	},
	{
		Key:    "security.key_pair",
		Domain: "security",
		Name:   "Key Pair",
		Description: "An SAP-generated key pair keystore entry (private key plus X.509 certificate), " +
			"as opposed to one uploaded from outside the tenant.",
		SupportStatus: StatusPartial,
		SupportReason: ReasonUnsafeTerraformLifecycle,
		ResourceTypes: []string{"sapintegrationsuite_key_pair"},
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Planned:       true,
		Limitations: []string{
			"Create confirmed field-for-field via SAP's own \"Generate a Key Pair\" documentation: " +
				"POST KeyPairGenerationRequests. The private key never leaves SAP — this resource has " +
				"no field for it and never requests one.",
			"No update operation is documented for a generated key pair: every attribute that defines " +
				"the generated key material is RequiresReplace.",
			"Only the subset SAP confirms KeystoreEntries returns (key_type, key_size, " +
				"valid_not_before, valid_not_after) is read back and refreshed on every plan; " +
				"generation-only parameters SAP does not confirm returning (signature_algorithm, " +
				"key_algorithm_parameter, the subject DN fields) are trusted from the last successful " +
				"write, not re-verified — the same reason this is `partial`, not `supported`.",
			"Delete uses the same documented keystore mass-deletion operation as " +
				"sapintegrationsuite_certificate, with exactly the one alias this resource owns.",
		},
		Operations: Operations{Create: true, Read: true, Delete: true, Import: true},
	},
	{
		Key:           "security.ssh_key",
		Domain:        "security",
		Name:          "SSH Key",
		Description:   "An SSH-capable key pair used for SFTP public-key authentication.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonOutOfScope,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Limitations: []string{
			"Reverified for this feature family and corrected: SAP's Security Content API overview " +
				"lists no independent \"SSH Key\" resource, and the tenant keystore's own \"Creating a " +
				"Key Pair/SSH Key Pair\" UI documentation uses the identical Key Pair attribute set " +
				"(alias, key type, key size, signature algorithm, subject DN fields, validity) for " +
				"both — \"Create > Key Pair\" and \"Create > SSH Key\" are the same underlying " +
				"mechanism with a different label. No separate SSHKeyGenerationRequests field " +
				"contract (mandatory/optional fields, example body) was found documented anywhere.",
			"An RSA or DSA sapintegrationsuite_key_pair's public key can be exported in OpenSSH " +
				"format via public_key_openssh, backed by SAP's confirmed " +
				"KeystoreEntries('<hexalias>')/Sshkey/$value — this covers the SSH use case without a " +
				"separate resource. EC key pairs are documented as unsupported for this export.",
		},
	},
	{
		Key:           "security.certificate_chain",
		Domain:        "security",
		Name:          "Certificate Chain",
		Description:   "A certificate chain associated with a key pair.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Planned:       true,
		Limitations: []string{
			"Reverified for this feature family: SAP's Security Content API overview describes " +
				"certificate chain import/export as part of the Key Pair resource's own capabilities " +
				"(\"create a certificate signing request, or import and export the related certificate " +
				"chain\"), not an independently documented CertificateChainResources contract — no " +
				"example request, field table, or worked response was found for it anywhere in SAP's " +
				"published documentation, unlike Certificate and Key Pair. Not implemented this phase; " +
				"if a concrete contract is confirmed, this would likely be scoped by key-pair alias " +
				"(for example sapintegrationsuite_key_pair_certificate_chain) rather than a standalone " +
				"global resource, matching that ownership relationship.",
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
			"Reverified for this feature family: SAP's own Security Content API overview lists " +
				"\"Secure Parameter\" as a resource conceptually covered by the same OData API as " +
				"User Credentials/OAuth2 Client Credentials, but the curated \"Security Content " +
				"Example Requests\" index (the same authoritative per-entity page this project used " +
				"to confirm every other Security Content operation) lists zero example requests for " +
				"it, and its own \"Deploying a Secure Parameter Artifact\" page describes only the " +
				"Eclipse/Node-Explorer deployment wizard, not a REST contract. Combined with prior " +
				"third-party evidence of an OData error resolving a SecureParameters entity set, this " +
				"remains unconfirmed rather than either \"no public API\" (the overview page does " +
				"list it) or \"public API\" (nothing about it is actually callable-confirmed). If a " +
				"public contract is confirmed, this would be a strong write-only-attribute candidate " +
				"(value_wo/value_wo_version), the same shape as security.user_credential's password.",
		},
	},
	{
		Key:    "security.known_hosts",
		Domain: "security",
		Name:   "Known Hosts (SSH)",
		Description: "The SSH \"known_hosts\" file artifact used to validate SFTP server host keys for " +
			"outbound SFTP connections.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
		Planned:       false,
		Limitations: []string{
			"Reverified for this feature family and strengthened from research_required to " +
				"no_public_api: unlike Secure Parameter (at least conceptually listed in SAP's " +
				"Security Content API overview's resource table), Known Hosts does not appear in " +
				"that table at all. SAP's \"Deploying an SSH Known Hosts Artifact\" documentation " +
				"describes only the Manage Security Material UI (Create > Known Hosts (SSH), " +
				"Browse/Add/Deploy), with no REST endpoint mentioned anywhere. No public OData " +
				"entity set or REST endpoint was found for it in any documentation this project could " +
				"reach.",
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
		Key:    "api_management.classic.api_provider",
		Domain: "api_management_classic",
		Name:   "API Provider (classic API Management)",
		Description: "A classic API Management backend/API provider system definition — the " +
			"connection an API Proxy targets.",
		SupportStatus:   StatusPartial,
		SupportReason:   ReasonUnsafeTerraformLifecycle,
		ResourceTypes:   []string{"sapintegrationsuite_api_provider"},
		DataSourceTypes: []string{"sapintegrationsuite_api_provider", "sapintegrationsuite_api_providers"},
		PublicAPI:       true,
		APIProtocol:     "OData V2 (Management.svc)",
		Limitations: []string{
			"Only the \"Internet\" connection type is supported (direct host/port, optionally over " +
				"SSL). SAP documents three further connection types (On Premise via Cloud Connector, " +
				"Open Connectors, Cloud Integration) with distinct field sets this provider could not " +
				"confirm a field-level JSON mapping for from a reachable primary source.",
			"No Update operation: SAP's own official Piper apiProviderUpload tooling documents that " +
				"only Create is supported through this API; every attribute is RequiresReplace.",
			"Eventual consistency: SAP documents up to approximately 20 seconds of caching before a " +
				"just-created or just-deleted provider is reliably visible to GET; Create polls with " +
				"bounded, jittered backoff to reduce (not eliminate) this window.",
		},
		Operations: Operations{Create: true, Read: true, Delete: true, Import: true},
	},
	{
		Key:    "api_management.classic.api_proxy",
		Domain: "api_management_classic",
		Name:   "API Proxy (classic API Management)",
		Description: "A classic API Management API proxy definition: the ZIP-bundled design-time " +
			"content (proxy endpoint, target endpoint, policies, resources) deployed as a callable API.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonPublicAPIIncomplete,
		PublicAPI:     true,
		APIProtocol:   "OData V2 (Management.svc/APIProxies)",
		Planned:       true,
		Limitations: []string{
			"The APIProxies entity set, its GET, and its DELETE are confirmed (SAP's own " +
				"documentation and worked examples reference \"Management.svc/APIProxies\" and " +
				"\"APIProxies('<name>')\" directly), and the proxy content bundle's ZIP structure is " +
				"confirmed field-for-field from SAP's own public sample repository " +
				"(SAP/apibusinesshub-api-recipes) — but the exact wire mechanism for uploading that " +
				"ZIP content through a Create/Update REST call (multipart form data, a base64 JSON " +
				"field, or something else) is not confirmed from any reachable primary source. SAP's " +
				"own official user guide describes only the UI-based import wizard for this operation.",
			"Depends on api_management.classic.api_provider already existing: SAP's own sample " +
				"repository documents that importing a proxy fails if the API Provider it references " +
				"does not already exist on the target tenant by name.",
		},
	},
	{
		Key:    "api_management.classic.api_proxy_deployment",
		Domain: "api_management_classic",
		Name:   "API Proxy Deployment (classic API Management)",
		Description: "The runtime deployment state of a classic API Proxy, potentially independent " +
			"of its design-time content.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonPublicAPIIncomplete,
		PublicAPI:     true,
		Limitations: []string{
			"Depends on api_management.classic.api_proxy, which this provider does not implement in " +
				"this phase — see that entry. SAP's own documentation states that a proxy transported " +
				"or exported, individually or as part of a product, \"by default gets imported to the " +
				"target in the deployed state,\" suggesting deployment may be a Create-time side effect " +
				"rather than an independent action, but this was not confirmed further given the " +
				"unconfirmed Create mechanism itself.",
		},
	},
	{
		Key:    "api_management.classic.policy",
		Domain: "api_management_classic",
		Name:   "Policy (classic API Management)",
		Description: "An individual mediation policy (for example VerifyAPIKey, Quota, " +
			"AssignMessage) attached to a classic API Proxy's proxy or target endpoint flow.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonPublicAPIIncomplete,
		PublicAPI:     true,
		Limitations: []string{
			"Confirmed to be XML content embedded inside the API Proxy ZIP bundle (a <policies> " +
				"element in the proxy's root XML, referencing named files under a Policy/ folder), not " +
				"an independently addressable OData entity with its own Create/Read/Update/Delete — so " +
				"individual policies are not a separate resource candidate; they would be managed as " +
				"part of api_management.classic.api_proxy's opaque content, once that entity's own " +
				"Create mechanism is confirmed.",
		},
	},
	{
		Key:    "api_management.classic.api_product",
		Domain: "api_management_classic",
		Name:   "API Product (classic API Management)",
		Description: "A classic API Management API product bundling one or more API proxies for " +
			"subscription, with optional custom attributes and request quotas.",
		SupportStatus:   StatusSupported,
		ResourceTypes:   []string{"sapintegrationsuite_api_product"},
		DataSourceTypes: []string{"sapintegrationsuite_api_product"},
		PublicAPI:       true,
		APIProtocol:     "OData V2 (Management.svc)",
		Limitations: []string{
			"api_proxy_names is only sent on Create: SAP's own documented Update (PUT) worked " +
				"example never includes the apiProxies association, so this provider treats it as " +
				"RequiresReplace rather than guess at an unconfirmed way to add or remove proxies " +
				"from an existing product.",
			"Referenced API proxies are expected to already exist through some other means (the SAP " +
				"Integration Suite UI, or a future sapintegrationsuite_api_proxy once its Create " +
				"mechanism is confirmed — see api_management.classic.api_proxy).",
			"DELETE is inferred from the consistent key-predicate DELETE convention confirmed " +
				"directly for APIProviders and CertificateStoreReferences within this same " +
				"Management.svc API family, not independently verified for APIProducts specifically.",
		},
		Operations: Operations{Create: true, Read: true, Update: true, Delete: true, Import: true},
	},
	{
		Key:    "api_management.classic.certificate_store_reference",
		Domain: "api_management_classic",
		Name:   "Certificate Store Reference (classic API Management)",
		Description: "A named pointer to an already-existing API Management keystore or " +
			"truststore, used so a virtual host's TLS configuration can be repointed at a new store " +
			"(for certificate rotation) without editing the virtual host itself.",
		SupportStatus: StatusSupported,
		ResourceTypes: []string{"sapintegrationsuite_api_management_certificate_store_reference"},
		DataSourceTypes: []string{
			"sapintegrationsuite_api_management_certificate_store_reference",
		},
		PublicAPI:   true,
		APIProtocol: "OData V2 (Management.svc)",
		Limitations: []string{
			"Manages only the reference (a name pointing at a store name); the keystore/truststore " +
				"and its certificate content are confirmed to be a UI-only upload with no accompanying " +
				"REST API, so this provider requires the referenced store to already exist.",
			"This is the best-confirmed Classic API Management resource in this provider: full " +
				"Create/Read/Update/Delete, request and response bodies, and error codes are all " +
				"confirmed verbatim from SAP's own official documentation.",
		},
		Operations: Operations{Create: true, Read: true, Update: true, Delete: true, Import: true},
	},
	{
		Key:    "api_management.classic.key_value_map",
		Domain: "api_management_classic",
		Name:   "Key Value Map (classic API Management)",
		Description: "A classic API Management key-value map used for runtime configuration " +
			"lookups, readable through the Key Value Map Operations policy.",
		SupportStatus:   StatusPartial,
		SupportReason:   ReasonUnsafeTerraformLifecycle,
		ResourceTypes:   []string{"sapintegrationsuite_api_key_value_map"},
		DataSourceTypes: []string{"sapintegrationsuite_api_key_value_map"},
		PublicAPI:       true,
		APIProtocol:     "OData V2 (Management.svc/GenericKeyMapEntries)",
		Limitations: []string{
			"No Update: SAP's documentation confirms a full Create/Update-entries/Delete UI " +
				"lifecycle exists, but only Create's REST payload is shown verbatim anywhere " +
				"reachable; every attribute, including entries, is RequiresReplace.",
			"Encrypted maps are not supported: SAP's isEncrypted field is real and documented, but " +
				"this provider could not confirm whether GET returns an encrypted entry's plaintext " +
				"value back, a masked placeholder, or nothing at all. This resource always sends " +
				"isEncrypted = false and rejects a configuration that sets encrypted = true.",
		},
		Operations: Operations{Create: true, Read: true, Delete: true, Import: true},
	},

	// --- Current API Management (API Artifacts / Integration Cell) — distinct from Classic
	// API Management (API Providers/Proxies/Products), which remains a separate, later phase.
	// Keys keep the "api_gateway" prefix from earlier research passes (never renamed once
	// shipped), but names/descriptions now use SAP's current "API Artifact"/"Integration
	// Cell" terminology throughout — see docs/guides/current-api-management.md.
	{
		Key:    "api_gateway.api_artifact",
		Domain: "api_gateway",
		Name:   "API Artifact — Current API Management",
		Description: "A design-time API configuration (endpoints, policies, security, runtime " +
			"behavior) in SAP's current, API-centric integration model, created under Design > " +
			"Integrations and APIs.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
		Planned:       true,
		Limitations: []string{
			"Reverified thoroughly for this feature family, not merely re-asserted: the canonical " +
				"Integration Content OData API's own exhaustive resource table (the same one this " +
				"provider already relies on for IntegrationDesigntimeArtifacts and every sibling " +
				"entity) does not list API Artifacts at all, ruling out the hypothesis that they are " +
				"just a new IntegrationDesigntimeArtifacts Type value. Over twenty documentation " +
				"pages covering every stage of the artifact lifecycle (four creation methods, " +
				"configuration, versioning, access management, deletion, deployment, monitoring) " +
				"were read in full; none mentions a REST/OData endpoint, entity set, or SAP Business " +
				"Accelerator Hub link. Corrected from research_required to no_public_api: this is a " +
				"confirmed absence, not an open question.",
		},
	},
	{
		Key:    "api_gateway.api_artifact_deployment",
		Domain: "api_gateway",
		Name:   "API Artifact Deployment — Integration Cell",
		Description: "The runtime deployment state of an API Artifact on Integration Cell or " +
			"Edge Integration Cell.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
		Planned:       true,
		Limitations: []string{
			"Same reverification and same conclusion as api_gateway.api_artifact: no deployment or " +
				"undeploy API was found documented anywhere. Confirmed from SAP's own documentation, " +
				"for whenever a public API does exist: runtime profile is effectively immutable once " +
				"an API artifact is created (RequiresReplace territory), with a documented exception " +
				"only for reassigning the target Edge Integration Cell node; and the virtual host " +
				"selected at deployment time can differ from the one configured at design time, with " +
				"the deployed endpoint URL reflecting whichever was actually used to deploy — see " +
				"docs/sap-api-references.md and docs/guides/current-api-management.md.",
		},
	},
	{
		Key:    "api_gateway.api_policy",
		Domain: "api_gateway",
		Name:   "API Artifact Policy",
		Description: "A policy or mediation step (authentication, authorization, quota, rate " +
			"limiting, transformation, and so on) attached to an API Artifact.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
		Planned:       true,
		Limitations: []string{
			"Whether policies are persisted as opaque content nested inside the artifact or as " +
				"independently addressable entities could not be determined, since no public API " +
				"exists for API Artifacts at all to inspect either representation against.",
		},
	},
	{
		Key:    "api_gateway.reusable_api_artifact",
		Domain: "api_gateway",
		Name:   "Reusable API Artifact",
		Description: "A modular, internal-only API Artifact variant with no external HTTP " +
			"endpoint of its own, invoked by other API Artifacts through the API Direct adapter.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
		Limitations: []string{
			"SAP's own documentation confirms this is a variant of the general API Artifact " +
				"concept (not accessible via HTTP endpoint, invoked only via API Direct, cannot " +
				"recursively call other reusable APIs, requires a unique base path), which is why " +
				"this is its own catalog entry rather than folded silently into api_gateway." +
				"api_artifact — but the underlying blocker is identical: no public API exists for " +
				"API Artifacts of any kind.",
		},
	},
	{
		Key:    "api_gateway.runtime_profile",
		Domain: "api_gateway",
		Name:   "Runtime Profile",
		Description: "The target integration platform (Cloud Integration, Integration Cell, " +
			"Edge Integration Cell, SAP Process Orchestration) an API Artifact or integration " +
			"flow is designed and deployed for.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
		Limitations: []string{
			"Configured and displayed only under Settings > Integrations; no API was found for " +
				"reading or managing this list. Even setting the missing API aside, this provider " +
				"judges the profile list a weak data-source candidate on its own merits: it is " +
				"small, stable, effectively enum-like platform metadata, better served by " +
				"documentation than a live API call. SAP's own \"Runtime Profiles\" reference page " +
				"does not list a distinct \"Integration Cell\" row at all, despite Integration Cell " +
				"being offered as a runtime profile choice elsewhere in the same documentation set — " +
				"an inconsistency in SAP's own documentation this provider records rather than " +
				"resolves by guessing.",
		},
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
			"Reconfirmed for this feature family: Integration Cell activation is a Settings > " +
				"Runtime UI step, and runtime content/status is visible only through Monitor > " +
				"Integrations and APIs, its own monitoring surface distinct from Cloud Integration's. " +
				"No public status or configuration API was found for Integration Cell runtime " +
				"content, only UI-facing operations.",
		},
	},
	{
		Key:    "integration_cell.virtual_host",
		Domain: "integration_cell",
		Name:   "Integration Cell Virtual Host",
		Description: "The public-facing host name and base path through which API Artifacts " +
			"(and MCP Servers) deployed to Integration Cell are exposed.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
		Limitations: []string{
			"Managed under Monitor > Manage Virtual Host by an administrator holding the " +
				"PI_Administrator role collection; no Create/Read/Update/Delete API was found " +
				"documented anywhere, despite dedicated UI-procedure pages existing for configuring, " +
				"editing, and deleting an eligible virtual host. If a public API is ever confirmed, " +
				"the default virtual host would need read-only treatment rather than an ordinary " +
				"mutable resource — SAP documents it as having restricted editability compared to an " +
				"administrator-created additional virtual host.",
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
			"Reconfirmed: an Edge Node is added and removed exclusively through the Edge Lifecycle " +
				"Management (ELM) UI, and onboarding is completed by running the standalone Edge " +
				"Lifecycle Management Bridge executable against the target Kubernetes cluster's " +
				"kubeconfig — a local CLI/Kubernetes handshake, not an HTTP API this provider could call. " +
				"No public registration, runtime-association, or deregistration API was found.",
		},
	},
	{
		Key:    "edge_integration_cell.local_api",
		Domain: "edge_integration_cell",
		Name:   "Edge Integration Cell Local API Access",
		Description: "SAP's public /local/api/v1 and /local/api/eic/v1 REST endpoints, reachable directly " +
			"against a running Edge Integration Cell node without going through the cloud UI: Message " +
			"Processing Logs, Message Stores/JMS, DataStores/Variables (OData V2), and the Operations " +
			"Cockpit's Component/Job/RuntimeParameter resources (OData V4).",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonOutOfScope,
		PublicAPI:     true,
		APIProtocol:   "OData V2 (Message Processing Logs, Message Stores) / OData V4 (Operations Cockpit)",
		Limitations: []string{
			"Confirmed reachable and documented (SAP Business Accelerator Hub packages " +
				"sap-int-eic-eic-operations, sap-int-eic-message-processing-logs-v1, " +
				"sap-int-eic-message-store-v1), authenticated with the same certificate or " +
				"clientId/clientsecret mechanisms as the rest of this provider, and CSRF-token gated for " +
				"modifying calls — but every entity behind it is monitoring/operational runtime data " +
				"(message processing records, message store/JMS contents, data store/variable values), " +
				"the same category this provider already excludes for Cloud Integration's own " +
				"MessageProcessingLogs/DataStores/Variables. See edge_integration_cell.runtime for the " +
				"separate Operations Cockpit control-plane objects reachable through this same /local " +
				"prefix.",
		},
	},
	{
		Key:    "edge_integration_cell.runtime",
		Domain: "edge_integration_cell",
		Name:   "Edge Integration Cell Runtime Operations",
		Description: "The Operations Cockpit API's Component, Pod, RuntimeParameter, and Job/JobSchedule " +
			"entities: per-component status, pod resource limits/replica counts, log levels, and " +
			"scheduled-job configuration for an Edge Integration Cell's own SAP-operator-managed pods.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonOutOfScope,
		PublicAPI:     true,
		APIProtocol:   "OData V4",
		Limitations: []string{
			"Confirmed reachable at /local/api/eic/v1 (also documented on SAP Business Accelerator Hub " +
				"as the Edge Integration Cell package's OData V4 API), and RuntimeParameter is explicitly " +
				"described as changeable (\"Save to run the changes in the back end\") — a genuinely " +
				"writable-looking configuration object, not read-only monitoring data.",
			"Every RuntimeParameter documented (LOG_LEVEL, MIN_REPLICAS, MAX_REPLICAS, CPU_LIMIT, " +
				"MEMORY_LIMIT, EPHEMERAL_STORAGE_LIMIT) configures Kubernetes-level pod resource requests, " +
				"replica counts, or log verbosity for SAP-operator-managed components — the same category " +
				"of concern docs/provider-scope.md already excludes as Kubernetes/Helm infrastructure, " +
				"just fronted by an EIC-specific API instead of the raw Kubernetes API.",
			"The API's own documentation states the Component resource lets a caller \"restart " +
				"components\" — an imperative runtime operation this provider's standing policy refuses " +
				"to model as a Terraform resource (see the retry/restart/cancel/purge rule this provider " +
				"already applies elsewhere).",
			"The exact entity keys and PATCH/POST payload shapes for RuntimeParameter and JobSchedule are " +
				"not confirmed from a reachable primary source: the package's $metadata/EDMX and worked " +
				"examples live behind SAP Business Accelerator Hub's authenticated catalog pages, which " +
				"redirect unauthenticated requests to a login page, the same access limitation this " +
				"project has documented repeatedly for other api.sap.com packages.",
		},
	},
	{
		Key:    "edge_integration_cell.deployment_target",
		Domain: "edge_integration_cell",
		Name:   "Edge Integration Cell Deployment Targeting",
		Description: "Selecting which Edge Integration Cell node (as opposed to Cloud Integration or " +
			"Integration Cell) a design-time artifact deploys to.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
		Limitations: []string{
			"SAP's Operations UI documents a \"Runtimes\" field (\"one or more runtime nodes to deploy " +
				"the artifact to, including Cloud Integration and any active Edge Integration Cell " +
				"nodes\") on Number Ranges specifically, but neither of its two documented API examples " +
				"(Add, Update) shows a runtime/location parameter — see docs/sap-api-references.md. " +
				"The same gap holds for every confirmed Deploy action this provider already calls " +
				"(IntegrationDesigntimeArtifacts, MessageMappingDesigntimeArtifacts, " +
				"ScriptCollectionDesigntimeArtifacts, ValueMappingDesigntimeArtifacts): none documents a " +
				"runtime-location/node-ID parameter, so this provider always targets the implicit default " +
				"runtime and never guesses at an undocumented one.",
		},
	},
	{
		Key:    "edge_integration_cell.access_policy_replication",
		Domain: "edge_integration_cell",
		Name:   "Edge Integration Cell Access Policy Replication",
		Description: "Selecting which runtimes (Integration Cell, specific Edge Integration Cell nodes) " +
			"an Access Policy replicates to.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonPublicAPIIncomplete,
		PublicAPI:     true,
		APIProtocol:   "OData V2",
		Limitations: []string{
			"The public AccessPolicies entity carries runtime assignments in its " +
				"AccessPolicyRuntimeAssignments navigation property (SAP's own CI/CD tooling strips it " +
				"from downloaded policies), and SAP Help documents the per-runtime Fail/Success/Pending " +
				"reconciliation status. Neither SAP Help nor any public sample documents the entity's " +
				"properties, how a runtime is identified, or whether assignments can be written through " +
				"the API, so the provider does not guess at them.",
			"Unblocking this needs the AccessPolicyRuntimeAssignments entity type from the Security " +
				"Content API specification on api.sap.com or from a tenant's $metadata document.",
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
		Name:          "Current API Management Capability Activation",
		Description:   "Activating SAP's current, API-centric API Management capability (API Artifacts, Integration Cell) itself within an Integration Suite tenant.",
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
		Key:    "api_composition.business_data_graph",
		Domain: "other_capability",
		Name:   "API Composition Business Data Graph",
		Description: "A composed, unified GraphQL/OData business data model (a \"Business Data " +
			"Graph\") spanning multiple backend systems (S/4HANA, SuccessFactors, custom OData/" +
			"REST sources), activated as a sub-capability of API Management alongside Developer " +
			"Hub and current API Management.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNotImplemented,
		PublicAPI:     true,
		Planned:       true,
		Limitations: []string{
			"The strongest confirmed-but-unimplemented finding from this provider's capability " +
				"audit: a genuine, separately entitled \"API Composition\" service (plan " +
				"configuration) exposes a Configuration API at " +
				"{region-specific host}/configuration/v1/sap.graph/GraphConfiguration, confirmed " +
				"with verbatim worked examples for Create (POST, full sample body with " +
				"businessDataGraphIdentifier/dataSources/locatingPolicy), Read (GET by ID), and " +
				"Update (PATCH by ID); Delete is explicitly stated to exist (\"Whether you need to " +
				"create, update, or delete business data graphs, this API provides an automated " +
				"option\") but no verbatim DELETE example was captured during this audit pass.",
			"Authentication reuses this provider's existing OAuth 2.0 client-credentials pattern, " +
				"but through a third distinct credential set: a Process Integration Runtime service " +
				"instance on the integration-flow plan (explicitly documented as NOT the same api " +
				"plan this provider's own oauth block already uses), separate again from " +
				"provider.api_management's apiportal-apiaccess credentials.",
			"Deliberately not implemented in this audit-only phase: this phase's scope is " +
				"classification, not implementation. Recommended as the top candidate for a future " +
				"dedicated phase, given the unusually strong Create/Read/Update evidence already on " +
				"record — stronger than most objects this provider has implemented to date.",
		},
	},
	{
		Key:    "odata_provisioning",
		Domain: "other_capability",
		Name:   "OData Provisioning",
		Description: "A capability that exposes SAP Business Suite backend OData services " +
			"(SAP Gateway back-end-enablement) through SAP Integration Suite, without requiring an " +
			"on-premise SAP Gateway hub.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     true,
		Limitations: []string{
			"A role literally named ODPAPIAccess is documented as a prerequisite for registering " +
				"OData services, a positive signal that a management API exists, but this audit pass " +
				"did not locate a worked request/response example or a dedicated API-access page for " +
				"it — the same kind of gap already documented for Integration Assessment, requiring " +
				"the same depth of follow-up research before any implementation decision.",
		},
	},
	{
		Key:           "developer_hub",
		Domain:        "other_capability",
		Name:          "Developer Hub",
		Description:   "SAP's API/Event/MCP Server catalog, publication, and subscription capability for Integration Suite, reachable through its own /api/1.0 REST API and its own devportal-apiaccess OAuth credentials, separate from every Cloud Integration and current API Management endpoint this provider otherwise talks to.",
		SupportStatus: StatusSeparateProvider,
		SupportReason: ReasonOutOfScope,
		PublicAPI:     true,
		Limitations: []string{
			"Developer Hub has its own API boundary, its own OAuth client credentials, and a consumer/catalog object lifecycle (Products, Applications, Subscriptions) distinct in shape from this provider's Integration Suite content and capability model.",
			"Planned as a separate, independently versioned Terraform provider (working name Prideth/terraform-provider-sap-developer-hub) rather than a domain inside this one, so its release cadence and credential surface never entangle with this provider's.",
			"This entry intentionally represents the whole Developer Hub capability as a single scope statement; individual Developer Hub objects (Product, Application, Subscription, and so on) are not separately cataloged here.",
		},
	},
	{
		Key:    "integration_advisor.design_time_content",
		Domain: "other_capability",
		Name:   "Integration Advisor Design-Time Content",
		Description: "Message Implementation Guidelines (MIGs), Mapping Guidelines (MAGs, " +
			"standard/overlay/XSLT), custom Type Systems, Codelists, Shared Code, and Global " +
			"Parameters — SAP's collaborative B2B interface-content-design objects.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
		Limitations: []string{
			"No API-access documentation page (the kind that, when present, confirmed a real public " +
				"API for Classic API Management and Integration Assessment) exists anywhere in " +
				"Integration Advisor's roughly eighty-five-page documentation tree. The one OAuth-" +
				"credential page found in this area (\"Creating OAuth Client Credentials for Cloud " +
				"Foundry Environment\") turned out, on full reading, to describe authenticating " +
				"against the destination Cloud Integration tenant for artifact injection (Process " +
				"Integration Runtime service, plan api), not a credential for Integration Advisor's " +
				"own design-time content management.",
		},
	},
	{
		Key:    "integration_advisor.runtime_artifact_injection",
		Domain: "other_capability",
		Name:   "Integration Advisor Runtime Artifact Injection",
		Description: "Injecting generated runtime artifacts (from a Mapping Guideline) directly " +
			"into an integration flow's resources on a target Cloud Integration tenant.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonOutOfScope,
		PublicAPI:     false,
		Limitations: []string{
			"Confirmed as a UI wizard (Mapping Guideline > Inject > SAP Cloud Integration Flow " +
				"Resources > choose tenant/package/integration flow > Inject), not a documented REST " +
				"call, and — independent of that — an imperative one-shot action rather than " +
				"desired-state configuration this provider's plan/apply model could represent even if " +
				"an API for it were confirmed.",
		},
	},
	{
		Key:    "trading_partner_management.company_profile",
		Domain: "other_capability",
		Name:   "Trading Partner Management Company Profile",
		Description: "The tenant's own company profile and its subsidiaries, the initiator side " +
			"of every trading partner agreement.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
		Limitations: []string{
			"Reconfirmed: no \"accessing APIs programmatically\", API reference, or service-instance/" +
				"service-key page exists anywhere in SAP's entire Trading Partner Management " +
				"documentation tree (roughly 90 pages checked) — the same kind of page that, when " +
				"present, confirmed a real public API for Classic API Management and Integration " +
				"Assessment. Content is downloadable as JSON through a UI Download button only " +
				"(company.json), never through a documented REST endpoint.",
		},
	},
	{
		Key:    "trading_partner_management.partner_profile",
		Domain: "other_capability",
		Name:   "Trading Partner Management Partner Profile",
		Description: "Trading partner profiles and communication partner profiles: the " +
			"counterparty side of a trading partner agreement.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
		Limitations: []string{
			"Same evidence as trading_partner_management.company_profile: UI-only (Design > B2B " +
				"Scenarios), downloadable as JSON, no documented REST/OData endpoint found.",
		},
	},
	{
		Key:           "trading_partner_management.agreement_template",
		Domain:        "other_capability",
		Name:          "Trading Partner Management Agreement Template",
		Description:   "A reusable template defining the shape of trading partner agreements created from it.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
	},
	{
		Key:    "trading_partner_management.agreement",
		Domain: "other_capability",
		Name:   "Trading Partner Management Agreement",
		Description: "A trading partner agreement: business transaction activities, identifiers, " +
			"and integration/message flow configuration between the tenant's company profile and a " +
			"trading partner.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
	},
	{
		Key:    "trading_partner_management.partner_directory_generation",
		Domain: "other_capability",
		Name:   "Trading Partner Management Partner Directory Generation",
		Description: "The generated Partner Directory entries (String/Binary Parameters, prefixed " +
			"\"SAP_TPM\") an agreement's activation pushes into Partner Directory for runtime use by " +
			"generic integration flows.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonOutOfScope,
		PublicAPI:     false,
		Limitations: []string{
			"Confirmed, verbatim: \"When a trading partner agreement gets activated, the complete " +
				"agreement information gets pushed into the partner directory.\" This is a UI-triggered " +
				"(Activate action), one-way generation side effect, not an independent create/update " +
				"operation this provider's already-implemented Partner Directory resources " +
				"(sapintegrationsuite_partner_string_parameter and siblings) could safely manage or " +
				"even represent — out of scope for the same reason this provider does not model other " +
				"imperative generation/replication actions as resources, regardless of whether a " +
				"public API for the trigger itself is ever confirmed.",
		},
	},
	{
		Key:    "event_mesh",
		Domain: "other_capability",
		Name:   "Event Mesh",
		Description: "SAP's Solace PubSub+-based event broker service for asynchronous, " +
			"event-driven integration: queues, topic subscriptions, and webhook subscriptions.",
		SupportStatus: StatusSeparateProvider,
		SupportReason: ReasonOutOfScope,
		PublicAPI:     true,
		Limitations: []string{
			"Reconfirmed, not just assumed: Event Mesh is activated through the same generic " +
				"\"Activating and Managing Capabilities\" mechanism as every other Integration Suite " +
				"capability (no dedicated public activation API, consistent with every other " +
				"capability audited), but once active it exposes a genuine, well-documented broker " +
				"management surface (service-key-based channel/queue/subscription creation, " +
				"AMQP/MQTT/REST messaging APIs) confirmed across roughly forty-five documentation " +
				"pages.",
			"Event Mesh predates, and is usable entirely independently of, SAP Integration Suite — " +
				"it is a general-purpose BTP messaging service consumed by many unrelated SAP " +
				"products, with its own Solace PubSub+-derived API family fundamentally different in " +
				"shape from the OData-centric model this provider is built around. Community " +
				"Terraform support for Solace PubSub+ already exists independently. Rather than " +
				"absorbing a broker-management surface into this Integration-Suite-scoped provider, " +
				"this belongs in a separate, independently versioned provider, the same reasoning " +
				"already applied to Developer Hub.",
		},
	},
	{
		Key:    "integration_assessment.master_data",
		Domain: "integration_assessment",
		Name:   "Integration Assessment Master Data",
		Description: "SAP Integration Solution Advisory Methodology (ISA-M) taxonomy: Domain, " +
			"Style, Use Case Pattern, Integration Pattern, Key Characteristic (and its Group/Value/" +
			"Recommendation), Deployment Model, Domain Determination — largely SAP-maintained " +
			"reference content a tenant can review and adjust.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     true,
		Limitations: []string{
			"A confirmed, separate BTP service (\"Integration Assessment APIs\", entitlement " +
				"integration-assessment) exposes these entities through two OAuth 2.0-secured base " +
				"URLs (\"entities\" and \"management\") from its own service key — the general shape " +
				"and full entity list are confirmed from SAP's own \"Integration Assessment APIs\" " +
				"documentation page — but no worked request/response example was found in any " +
				"reachable primary source (the SAP-docs mirror, an official PDF user guide, and an " +
				"SAP TechEd hands-on sample repository were all checked and cover only UI procedures), " +
				"so no field-level JSON schema could be confirmed for any entity in this group.",
		},
	},
	{
		Key:    "integration_assessment.landscape_configuration",
		Domain: "integration_assessment",
		Name:   "Integration Assessment Landscape Configuration",
		Description: "Tenant-owned integration landscape inventory: Application, Application " +
			"Instance, Technology, Technology Instance, Vendor, and their association entities " +
			"(Technology Domain, Technology Style, Technology Key Characteristic).",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     true,
		Limitations: []string{
			"The strongest Terraform-candidate family in this capability: these are practitioner-" +
				"authored configuration, not workflow or reporting data, and SAP documents concrete " +
				"per-tenant limits confirming real, bounded storage (maximum 20000 Applications, 20000 " +
				"Application Instances, 50 Technologies, 150 Technology Instances, 10000 Vendors) — " +
				"but, as with integration_assessment.master_data, no field-level JSON schema for any " +
				"Create/Read/Update/Delete operation was found in any reachable primary source.",
		},
	},
	{
		Key:    "integration_assessment.assessment_workflow",
		Domain: "integration_assessment",
		Name:   "Integration Assessment Requests and Assessment Workflow",
		Description: "Business solution/interface Request and Request Line Item workflow objects, " +
			"the Integration Flow/Message Flow content they reference, and Request Line Item " +
			"Technology Instance Decision.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonOutOfScope,
		PublicAPI:     true,
		Limitations: []string{
			"Confirmed to be workflow/project state, not desired-state configuration: SAP documents " +
				"an explicit Request status machine (draft -> new -> in progress -> completed, with a " +
				"Reopen action), matching this provider's established Message Processing Log/" +
				"Developer Hub Subscription category of exclusion — out of scope regardless of whether " +
				"a field-level API contract is ever confirmed for it.",
		},
	},
	{
		Key:    "migration_assessment.source_system",
		Domain: "other_capability",
		Name:   "Migration Assessment Source System",
		Description: "A registered SAP Process Orchestration system (7.31 SP28+, 7.40 SP23+, or " +
			"7.50 SP06+) Migration Assessment extracts integration scenario data from.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonNoPublicAPI,
		PublicAPI:     false,
		Limitations: []string{
			"No API-access or service-key documentation page exists anywhere in Migration " +
				"Assessment's documentation tree (a small, roughly fifteen-page tree, entirely " +
				"checked). Its own documentation instead describes Migration Assessment as an API " +
				"*consumer*: it reaches into a registered source system's own SAP Process " +
				"Orchestration APIs (via Cloud Connector/Destination service) to extract data — the " +
				"opposite direction from a public API this provider could manage Migration " +
				"Assessment's own objects through.",
		},
	},
	{
		Key:    "migration_assessment.extraction_and_evaluation",
		Domain: "other_capability",
		Name:   "Migration Assessment Extraction and Scenario Evaluation",
		Description: "Data Extraction Requests, Scenario Evaluation Requests, and their resulting " +
			"assessment-category/migration-readiness/effort-estimate reports.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonOutOfScope,
		PublicAPI:     false,
		Limitations: []string{
			"Confirmed to be action-triggered workflow and reporting data, not desired-state " +
				"configuration: SAP's own documentation describes choosing \"Create\" on a Data " +
				"Extraction Request as starting an extraction process with a resulting Completed/" +
				"Completed with warnings/Completed with errors status, and Scenario Evaluation " +
				"results are assessment-category classifications, migration-readiness scores, and " +
				"effort estimates — reporting output, the same category this provider already " +
				"excludes for Message Processing Logs. Out of scope regardless of whether a public " +
				"API is ever confirmed for triggering these actions.",
		},
	},
	{
		Key:    "data_space_integration",
		Domain: "other_capability",
		Name:   "Data Space Integration",
		Description: "SAP's Dataspace-Protocol-based data space connectivity capability " +
			"(Connectors, Assets, Policies, Contract Definitions, Contract Negotiations/" +
			"Agreements) within Integration Suite, initially scoped to the Catena-X data space.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonResearchRequired,
		PublicAPI:     true,
		Limitations: []string{
			"Confirmed real: a dedicated \"Data Space Integration API Access\" service instance " +
				"(plan api, roles AuthGroup_DataspaceConsumer/AuthGroup_DataspaceProvider, " +
				"client_credentials grant) is documented, and its OData REST APIs are confirmed " +
				"listed on SAP Business Accelerator Hub at api.sap.com/package/dataspaceintegration/" +
				"rest — unreachable to this project without an SAP support login, the same " +
				"limitation hit repeatedly for other packages, so no field-level schema was " +
				"confirmed during this audit pass.",
			"A genuinely complex, multi-persona object model (per-connector service instances, " +
				"Assets, Policies, Contract Definitions, Contract Negotiations, Contract Agreements) " +
				"built on an external standard (the Dataspace Protocol / International Data Spaces " +
				"initiative) — this audit intentionally did not attempt a shallow implementation and " +
				"instead flags this as a candidate needing its own dedicated future research phase, " +
				"the same treatment Classic API Management and Integration Assessment each received " +
				"as standalone phases.",
		},
	},
	{
		Key:    "open_connectors",
		Domain: "other_capability",
		Name:   "Open Connectors",
		Description: "SAP's third-party SaaS connectivity capability within Integration Suite: a " +
			"catalog of 170+ third-party connector types (each with its own normalized REST API and " +
			"OpenAPI-documented instance configuration), originally the standalone Cloud Elements " +
			"product.",
		SupportStatus: StatusUnsupported,
		SupportReason: ReasonOutOfScope,
		PublicAPI:     true,
		Limitations: []string{
			"A deliberate suitability judgment, not a research gap: Open Connectors is not one " +
				"coherent API this provider could model with a handful of resources, the pattern " +
				"every other capability in this catalog follows. It is a catalog of 170+ independent " +
				"third-party connector types, each with its own authentication scheme, configuration " +
				"schema, and normalized-but-still-connector-specific REST surface. Implementing even " +
				"connector-instance management generically would mean either an unbounded, " +
				"per-connector-type schema explosion, or an opaque untyped-JSON resource that gives " +
				"up the type safety and validation this provider's schema-first design otherwise " +
				"provides everywhere else.",
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
