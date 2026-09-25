package provider

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var designtimeVersionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

// saveAsVersionAttribute is the shared save_as_version attribute of the
// versioned design-time artifact resources.
func saveAsVersionAttribute(artifact string) schema.StringAttribute {
	return schema.StringAttribute{
		Optional: true,
		Description: "Version to save the uploaded " + artifact + " under, for example \"1.0.3\", " +
			"using SAP's SaveAsVersion function import. Without it, SAP keeps its own version " +
			"number and uploads update the current draft. The content is saved as a new version " +
			"only when this value changes, so bump it together with the content for each release; " +
			"version then reports what SAP recorded.",
		Validators: []validator.String{
			stringvalidator.RegexMatches(designtimeVersionPattern, "must be a version of the form major.minor.patch, for example 1.0.3"),
		},
	}
}

// versionToSave returns the version to save and whether a SaveAsVersion call
// is due: the planned value is set and differs from the one saved before.
func versionToSave(planned, prior types.String) (string, bool) {
	v := planned.ValueString()
	return v, v != "" && v != prior.ValueString()
}
