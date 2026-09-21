# Capability Dependency Matrix

Only dependencies that are explicitly documented by SAP are listed. No dependency is assumed.

| Capability | Requires | Optional Dependencies | Runtime | Terraform Support |
|---|---|---|---|---|
| Cloud Integration | Integration Suite subscription | Security Content (for OAuth/credential-based adapters) | Cloud (SAP-managed) | Resource + Data Source |
| API Management (classic) | Integration Suite subscription | Cloud Integration (for orchestrating calls from iFlows) | Cloud (SAP-managed) | Data Source (v0.1.x), Resource planned |
| API Gateway / API Artifacts | Integration Suite subscription | Integration Cell or Edge Integration Cell (runtime target), a BTP Destination as backend | Cloud, Integration Cell, or Edge Integration Cell | Unsupported (v0.1.x) — planned v0.2.x |
| Integration Cell | Integration Suite subscription | API Gateway / API Artifacts (as a deployment target) | Cloud (SAP-managed cell) | Unsupported — manual bootstrap required (no public activation API found) |
| Edge Integration Cell | Integration Suite subscription, customer-managed Kubernetes cluster (provisioned outside this provider) | Integration Cell control plane, API Gateway / API Artifacts | Customer-managed | Unsupported — manual bootstrap required for activation; control-plane read/config planned once a public API is confirmed |
| Access Policies | Cloud Integration (protects Cloud Integration artifacts) | Integration Cell, Edge Integration Cell (for runtime-scoped policy replication) | Cloud, Integration Cell, Edge Integration Cell | Resource + Data Source |
| Integration Advisor | Integration Suite subscription | Trading Partner Management | Cloud (SAP-managed) | Unsupported (not yet investigated in depth) |
| Trading Partner Management | Integration Suite subscription | Integration Advisor, Partner Directory | Cloud (SAP-managed) | Unsupported (not yet investigated in depth) |
| Event Mesh | Integration Suite subscription or standalone BTP service | — | Cloud (SAP-managed) | Out of scope for this provider unless a distinct Integration-Suite-only public API surface is identified; likely belongs to a dedicated Event Mesh provider |

Capabilities not yet listed here (Integration Assessment, Migration Assessment, Data Space
Integration, Developer Hub, API Composition) have not yet had their dependency graph
confirmed against current SAP documentation and are intentionally left out rather than
guessed. They will be added once confirmed.
