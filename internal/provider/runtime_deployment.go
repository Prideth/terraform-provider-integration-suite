package provider

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/apierror"
	"github.com/Prideth/terraform-provider-sap-integration-suite/internal/client/cloudintegration"
)

const (
	deploymentPollInterval = 3 * time.Second
	deploymentPollMax      = 20 * time.Second
)

// waitForRuntimeArtifact polls the shared runtime-artifacts entity for id
// until it reaches STARTED (success) or ERROR (failure), using exponential
// backoff with full jitter, bounded by ctx's deadline. It never sleeps a
// fixed duration. Shared by every *_deployment resource (integration flow,
// value mapping, ...) since they all deploy through the same runtime status
// model.
func waitForRuntimeArtifact(ctx context.Context, client *cloudintegration.Client, id string) (*cloudintegration.RuntimeArtifact, error) {
	attempt := 0
	for {
		artifact, err := client.GetRuntimeArtifact(ctx, id)
		if err == nil {
			switch artifact.Status {
			case cloudintegration.StatusStarted:
				return artifact, nil
			case cloudintegration.StatusError:
				errInfo := artifact.ErrorInfo
				if errInfo == "" {
					errInfo = "SAP reported status ERROR without further detail on the runtime artifact"
				}
				return nil, fmt.Errorf("deployment failed: %s", errInfo)
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
	//nolint:gosec // G404: jitter timing does not need a cryptographically secure random source
	return time.Duration(rand.Int63n(int64(capped))) + deploymentPollInterval/2
}
