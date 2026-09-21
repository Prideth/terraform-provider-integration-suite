// Command terraform-provider-sap-integration-suite serves the
// sapintegrationsuite Terraform provider over the Terraform Plugin Protocol.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/provider"
)

// version is set at build time via -ldflags "-X main.version=..." by
// GoReleaser; it defaults to "dev" for local builds.
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "start the provider with support for debuggers")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/Prideth/sap-integration-suite",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err)
	}
}
