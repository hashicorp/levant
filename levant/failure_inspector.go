package levant

import (
	"strings"
	"sync"

	nomad "github.com/hashicorp/nomad/api"
	"github.com/jrasell/levant/logging"
)

// checkFailedDeployment helps log information about deployment failures.
func (c *nomadClient) checkFailedDeployment(depID *string) {

	var allocIDS []string

	allocs, _, err := c.nomad.Deployments().Allocations(*depID, nil)
	if err != nil {
		logging.Error("levant/failure_inspector: unable to query deployment allocations for deployment %s",
			depID)
	}

	// Iterate the allocations on the deployment and create a list of each allocID
	// to inspect.
	for _, alloc := range allocs {
		allocIDS = append(allocIDS, alloc.ID)
	}

	// Setup a waitgroup so the function doesn't return until all allocations have
	// been inspected.
	var wg sync.WaitGroup
	wg.Add(len(allocIDS))

	// Inspect each allocation.
	for _, id := range allocIDS {
		logging.Debug("levant/failure_inspector: launching allocation inspector for alloc %v", id)
		go c.allocInspector(&id, &wg)
	}

	wg.Wait()
}

// allocInspector inspects an allocations events to log any useful information
// which may help debug deployment failures.
func (c *nomadClient) allocInspector(allocID *string, wg *sync.WaitGroup) {

	// Inform the wait group we have finished our task upon completion.
	defer wg.Done()

	resp, _, err := c.nomad.Allocations().Info(*allocID, nil)
	if err != nil {
		logging.Error("levant/failure_inspector: unable to query alloc %v: %v", allocID, err)
	}

	// Iterate each each Task and Event to log any relevant information which may
	// help debug deployment failures.
	for _, task := range resp.TaskStates {
		for _, event := range task.Events {
			switch event.Type {
			case nomad.TaskDriverFailure:
				logging.Info("levant/failure_inspector: allocation %v incurred %v due to %s",
					*allocID, nomad.TaskDriverFailure, strings.TrimSpace(event.DriverError))
			case nomad.TaskGenericMessage:
				logging.Info("levant/failure_inspector: allocation %v incurred %v due to %s",
					*allocID, nomad.TaskGenericMessage, strings.TrimSpace(event.Message))
			case nomad.TaskArtifactDownloadFailed:
				logging.Info("levant/failure_inspector: allocation %v incurred %v due to %s",
					*allocID, nomad.TaskArtifactDownloadFailed, strings.TrimSpace(event.DownloadError))
			}
		}
	}
}
