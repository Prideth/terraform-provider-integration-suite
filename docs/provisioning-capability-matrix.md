# Provisioning Capability Matrix

This matrix tracks, per Integration Suite capability, whether SAP publishes an official
public API to activate, read, configure, check the status of, or deactivate it. It is the
result of researching the SAP Help Portal (via its official `SAP-docs` GitHub mirror, since
`help.sap.com` itself was not reachable from this development environment), SAP Business
Accelerator Hub search results, SAP Community blog posts authored by SAP, and official SAP
sample repositories. **No internal/UI-only endpoint was used as a substitute for a missing
public API.**

Because SAP evolves Integration Suite continuously, treat every "Unsupported" row as a
snapshot: re-check SAP Business Accelerator Hub before assuming it still holds.

| Capability | Public Activation API | Read API | Configuration API | Status API | Deactivation API | Async | Terraform |
|---|---|---|---|---|---|---|---|
| Cloud Integration | Unsupported – manual bootstrap required (activated as part of the Integration Suite subscription/onboarding wizard) | N/A (implicitly available once subscribed) | N/A | N/A | Unsupported | N/A | N/A (out of scope — this is a subscription-time concern) |
| API Management (classic) | Unsupported – manual bootstrap required (documented as a UI step: Integration Suite → Manage Capabilities → activate API Management) | Unsupported | Unsupported | Unsupported | Unsupported | N/A | Unsupported |
| Current API Management / API Artifacts capability | Unsupported – manual bootstrap required (UI-driven capability activation) | Unsupported (capability-level; individual API Artifacts were reverified with a thorough documentation sweep and confirmed to have no public design-time API either — see `docs/guides/current-api-management.md`) | Unsupported | Unsupported | Unsupported | N/A | Unsupported |
| Integration Cell | Unsupported – manual bootstrap required. SAP's own documentation ("Activate Integration Cell") describes activation as choosing *Activate* in Integration Suite → Settings → Runtime; no public API was found. | Unsupported (capability-level) | Unsupported | Unsupported | Unsupported | Unknown | Unsupported |
| Edge Integration Cell | Unsupported – manual bootstrap required. Activation/registration is a guided UI + Helm-based bootstrap process (SAP-provided installer running on customer Kubernetes); no public capability-activation API was found. | Unsupported (capability-level; runtime Operations APIs exist for components/pods/jobs but are documented as UI-facing and are not treated as Terraform data sources per this provider's monitoring boundary) | Unsupported | Unsupported | Unsupported | N/A | Unsupported |
| Integration Advisor | Not yet confirmed | Not yet confirmed | Not yet confirmed | Not yet confirmed | Not yet confirmed | Not yet confirmed | Unsupported (not yet investigated) |
| Trading Partner Management | Not yet confirmed | Not yet confirmed | Not yet confirmed | Not yet confirmed | Not yet confirmed | Not yet confirmed | Unsupported (not yet investigated) |
| Event Mesh | Not yet confirmed | Not yet confirmed | Not yet confirmed | Not yet confirmed | Not yet confirmed | Not yet confirmed | Likely out of scope for this provider (see provider-scope.md) |
| Integration Assessment | Not yet confirmed | Not yet confirmed | Not yet confirmed | Not yet confirmed | Not yet confirmed | Not yet confirmed | Unsupported (not yet investigated) |
| Migration Assessment | Not yet confirmed | Not yet confirmed | Not yet confirmed | Not yet confirmed | Not yet confirmed | Not yet confirmed | Unsupported (not yet investigated) |
| Data Space Integration | Not yet confirmed | Not yet confirmed | Not yet confirmed | Not yet confirmed | Not yet confirmed | Not yet confirmed | Unsupported (not yet investigated) |

## Conclusion for `sapintegrationsuite_capability`

No currently confirmed public API activates, deactivates, or reconfigures an Integration
Suite capability. A generic `sapintegrationsuite_capability` resource is therefore **not**
implemented in v0.1.0. Capability activation remains a documented manual bootstrap step
(performed once, out-of-band, by an administrator) before this provider's content and
runtime resources can be used. This will be revisited if/when SAP publishes a capability
management API; the gap is documented rather than worked around through an internal
endpoint, per this project's API policy.
