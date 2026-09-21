//go:build tools

// Package tools tracks developer tool dependencies in their own module so
// that a newer Go requirement pulled in by a documentation or lint tool
// never raises the main provider module's go directive.
package tools

import (
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
