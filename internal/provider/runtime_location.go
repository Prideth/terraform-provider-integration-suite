package provider

import (
	"context"
	"fmt"
	"strings"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	v2 "github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/odata/v2"
)

const runtimeLocationDescription = "Runtime location ID of the Edge Integration Cell to address, " +
	"for example \"myedge\". Leave unset for the cloud runtime. SAP shows the ID in the " +
	"Integration Suite monitoring URL after selecting the Edge Integration Cell as runtime " +
	"({\"edge\":{\"runtimeLocationId\":\"myedge\"}}). Requests then go to " +
	"/location/<id>/api/v1 on the same tenant host."

// runtimeLocationResourceAttribute is the shared runtime_location_id attribute
// of resources that exist once per runtime. Moving an object to another
// runtime means creating it there, so the attribute forces replacement.
func runtimeLocationResourceAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Optional:    true,
		Description: runtimeLocationDescription + " Changing it replaces the resource.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		Validators: []validator.String{runtimeLocationValidator{}},
	}
}

// runtimeLocationDataSourceAttribute is the data source counterpart.
func runtimeLocationDataSourceAttribute() dsschema.StringAttribute {
	return dsschema.StringAttribute{
		Optional:    true,
		Description: runtimeLocationDescription,
		Validators:  []validator.String{runtimeLocationValidator{}},
	}
}

// locatedClient returns the client for the runtime named by loc: the client
// itself for the cloud runtime, or one routed through /location/<id>.
func locatedClient[T interface{ AtLocation(string) (T, error) }](c T, loc types.String, diags *diag.Diagnostics) (T, bool) {
	located, err := c.AtLocation(loc.ValueString())
	if err != nil {
		diags.AddAttributeError(path.Root("runtime_location_id"), "Invalid runtime location ID", err.Error())
		return located, false
	}
	return located, true
}

// splitLocatedImportID splits an import ID made of parts segments separated by
// "/", optionally preceded by a runtime location ID segment. It returns the
// location ("" for the cloud runtime) and the remaining segments.
func splitLocatedImportID(id string, parts int) (string, []string, error) {
	segments := strings.Split(id, "/")
	var loc string
	switch len(segments) {
	case parts:
	case parts + 1:
		loc, segments = segments[0], segments[1:]
		if err := v2.ValidateRuntimeLocationID(loc); err != nil {
			return "", nil, err
		}
	default:
		return "", nil, fmt.Errorf("expected %d segments separated by \"/\", optionally preceded by a runtime location ID, got %q", parts, id)
	}
	for _, s := range segments {
		if s == "" {
			return "", nil, fmt.Errorf("import ID %q contains an empty segment", id)
		}
	}
	return loc, segments, nil
}

func setImportedRuntimeLocation(ctx context.Context, loc string, set func(context.Context, path.Path, interface{}) diag.Diagnostics, diags *diag.Diagnostics) {
	if loc != "" {
		diags.Append(set(ctx, path.Root("runtime_location_id"), loc)...)
	}
}

type runtimeLocationValidator struct{}

func (runtimeLocationValidator) Description(context.Context) string {
	return "value must be a runtime location ID made of letters, digits, '.', '_' and '-'"
}

func (v runtimeLocationValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (runtimeLocationValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if err := v2.ValidateRuntimeLocationID(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid runtime location ID", err.Error())
	}
}
