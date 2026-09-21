package provider

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apierror"
	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

const (
	deploymentPollInterval = 3 * time.Second
	deploymentPollMax      = 20 * time.Second
)

// NewIntegrationFlowDeploymentResource returns a fresh resource.Resource
// implementation for sapintegrationsuite_integration_flow_deployment.
func NewIntegrationFlowDeploymentResource() resource.Resource {
	return &integrationFlowDeploymentResource{}
}

type integrationFlowDeploymentResource struct {
	client *cloudintegration.Client
}

type integrationFlowDeploymentModel struct {
	ID              types.String   `tfsdk:"id"`
	PackageID       types.String   `tfsdk:"package_id"`
	FlowID          types.String   `tfsdk:"flow_id"`
	DeployedVersion types.String   `tfsdk:"deployed_version"`
	Status          types.String   `tfsdk:"status"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
}

func (r *integrationFlowDeploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration_flow_deployment"
}

func (r *integrationFlowDeploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Expresses the desired runtime deployment state of a Cloud Integration " +
			"integration flow, independent of its design-time content lifecycle. Deploying is " +
			"asynchronous: this resource polls SAP's runtime artifact status until the deployment " +
			"reaches a terminal state (STARTED or ERROR) or the configured timeout elapses.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier in the form \"<package_id>/<flow_id>\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"package_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the integration package the flow belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"flow_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the integration flow to deploy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"deployed_version": schema.StringAttribute{
				Computed:    true,
				Description: "The design-time version that is currently deployed.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The runtime status SAP reports for this deployment (for example STARTED or ERROR).",
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(context.Background(), timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *integrationFlowDeploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*Data)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", "Expected *provider.Data")
		return
	}
	r.client = cloudintegration.New(data.HTTPClient, data.Host)
}

func (r *integrationFlowDeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan integrationFlowDeploymentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Create(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := r.client.DeployIntegrationFlow(ctx, plan.FlowID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to deploy SAP Integration Suite integration flow", diagnosticDetail(err))
		return
	}

	artifact, err := r.waitForDeployment(ctx, plan.FlowID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Deployment did not reach a ready state", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, deploymentToModel(plan.PackageID.ValueString(), artifact, plan.Timeouts))...)
}

func (r *integrationFlowDeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state integrationFlowDeploymentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	artifact, err := r.client.GetRuntimeArtifact(ctx, state.FlowID.ValueString())
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read SAP Integration Suite integration flow deployment", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, deploymentToModel(state.PackageID.ValueString(), artifact, state.Timeouts))...)
}

func (r *integrationFlowDeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan integrationFlowDeploymentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := plan.Timeouts.Update(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := r.client.DeployIntegrationFlow(ctx, plan.FlowID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to redeploy SAP Integration Suite integration flow", diagnosticDetail(err))
		return
	}

	artifact, err := r.waitForDeployment(ctx, plan.FlowID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Deployment did not reach a ready state", diagnosticDetail(err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, deploymentToModel(plan.PackageID.ValueString(), artifact, plan.Timeouts))...)
}

func (r *integrationFlowDeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state integrationFlowDeploymentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	timeout, diags := state.Timeouts.Delete(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := r.client.UndeployIntegrationFlow(ctx, state.FlowID.ValueString())
	if err != nil {
		var apiErr *apierror.Error
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return
		}
		resp.Diagnostics.AddError("Failed to undeploy SAP Integration Suite integration flow", diagnosticDetail(err))
	}
}

func (r *integrationFlowDeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	packageID, flowID, err := splitCompositeID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("package_id"), packageID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathRoot("flow_id"), flowID)...)
}

// waitForDeployment polls the runtime artifact status until it reaches
// STARTED (success) or ERROR (failure), using exponential backoff with full
// jitter, bounded by ctx's deadline. It never sleeps a fixed duration.
func (r *integrationFlowDeploymentResource) waitForDeployment(ctx context.Context, flowID string) (*cloudintegration.RuntimeArtifact, error) {
	attempt := 0
	for {
		artifact, err := r.client.GetRuntimeArtifact(ctx, flowID)
		if err == nil {
			switch artifact.Status {
			case cloudintegration.StatusStarted:
				return artifact, nil
			case cloudintegration.StatusError:
				return nil, fmt.Errorf("deployment failed: %s", artifact.ErrorInfo)
			}
		} else {
			var apiErr *apierror.Error
			if !errors.As(err, &apiErr) || !apiErr.IsNotFound() {
				return nil, err
			}
			// Not found yet: the runtime artifact has not been created by
			// SAP's asynchronous deployment pipeline. Keep polling.
		}

		delay := pollBackoff(attempt)
		attempt++

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("timed out waiting for the deployment to become ready: %w", ctx.Err())
		case <-timer.C:
		}
	}
}

// pollBackoff computes an exponential backoff with full jitter for status
// polling, capped at deploymentPollMax.
func pollBackoff(attempt int) time.Duration {
	capped := math.Min(float64(deploymentPollMax), float64(deploymentPollInterval)*math.Pow(1.5, float64(attempt)))
	return time.Duration(rand.Int63n(int64(capped))) + deploymentPollInterval/2
}

func deploymentToModel(packageID string, artifact *cloudintegration.RuntimeArtifact, tf timeouts.Value) integrationFlowDeploymentModel {
	return integrationFlowDeploymentModel{
		ID:              types.StringValue(packageID + "/" + artifact.ID),
		PackageID:       types.StringValue(packageID),
		FlowID:          types.StringValue(artifact.ID),
		DeployedVersion: types.StringValue(artifact.Version),
		Status:          types.StringValue(artifact.Status),
		Timeouts:        tf,
	}
}
