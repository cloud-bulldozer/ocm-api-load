package helpers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"

	sdk "github.com/openshift-online/ocm-sdk-go"

	errors "github.com/zgalor/weberr"
)

// createdClusterIDs maps the IDs of the cluster created by testing to a bool value for `deprovision`.
var createdClusterIDs = map[string]bool{}
var validateDeletedClusterIDs = make([]string, 0)
var failedCleanupClusterIDs = make([]string, 0)

func Cleanup(ctx context.Context, connection *sdk.Connection) {
	if len(createdClusterIDs) == 0 {
		return
	}
	connection.Logger().Info(ctx, "About to clean up the following clusters:")
	for clusterID, deprovision := range createdClusterIDs {
		connection.Logger().Info(ctx, "Cluster ID: %s, deprovision: %v", clusterID, deprovision)
		DeleteCluster(ctx, clusterID, deprovision, connection)
	}
	for _, clusterID := range validateDeletedClusterIDs {
		err := verifyClusterDeleted(ctx, clusterID, connection)
		if err != nil {
			markFailedCleanup(clusterID)
		} else {
			delete(createdClusterIDs, clusterID)
		}
	}
	if len(failedCleanupClusterIDs) > 0 {
		connection.Logger().Warn(ctx, "The following clusters failed deletion: %v", failedCleanupClusterIDs)
	}
	createdClusterIDs = make(map[string]bool)
	failedCleanupClusterIDs = make([]string, 0)
}

func DeleteCluster(ctx context.Context, id string, deprovision bool, connection *sdk.Connection) {
	connection.Logger().Info(ctx, "Deleting cluster '%s'", id)
	// Send the request to delete the cluster
	response, err := connection.Delete().
		Path(ClustersEndpoint+id).
		Parameter("deprovision", deprovision).
		Send()
	if err != nil {
		connection.Logger().Error(ctx, "Failed to delete cluster '%s', got error: %v", id, err)
		markFailedCleanup(id)
	} else if response.Status() != 204 {
		connection.Logger().Error(ctx, "Failed to delete cluster '%s', got http status %d", id, response.Status())
		markFailedCleanup(id)
	} else {
		validateDeletedClusterIDs = append(validateDeletedClusterIDs, id)
		connection.Logger().Info(ctx, "Cluster '%s' deleted", id)
	}
}

func CreateCluster(ctx context.Context, body string, gatewayConnection *sdk.Connection) (string, map[string]interface{}, error) {
	postResponse, err := gatewayConnection.Post().
		Path(ClustersEndpoint).
		String(body).
		Send()
	if err != nil {
		return "", nil, err
	}
	if postResponse.Status() != http.StatusCreated {
		return "", nil, errors.Errorf("Failed to create cluster: expected response code %d, instead found: %d",
			http.StatusCreated, postResponse.Status())
	}
	data, err := Parse(postResponse.Bytes())
	if err != nil {
		return "", nil, err
	}
	clusterID, ok := data["id"]
	if !ok {
		gatewayConnection.Logger().Error(ctx, "ClusterID not present")
	}
	gatewayConnection.Logger().Info(ctx, "Cluster '%s' created", clusterID.(string))
	return clusterID.(string), data, nil
}

func verifyClusterDeleted(ctx context.Context, clusterID string, connection *sdk.Connection) error {
	connection.Logger().Info(ctx, "verifying deleted cluster '%s'", clusterID)
	var forcedErr error
	var getStatus int
	err := retry.Retry(func(attempt uint) error {
		getResponse, err := connection.Get().
			Path(ClustersEndpoint + clusterID).
			Send()
		if err != nil {
			forcedErr = err
			return nil
		}
		if getResponse.Status() == 404 {
			getStatus = getResponse.Status()
			return nil
		}
		return errors.Errorf("Cluster still exists StatusCode: %d", getResponse.Status())
	},
		strategy.Wait(1*time.Second),
		strategy.Limit(300))
	if err != nil {
		connection.Logger().Error(ctx, "failed to delete cluster '%s': %v", clusterID, err)
		return err
	}
	if forcedErr != nil {
		return fmt.Errorf("failed to wait for cluster '%s' to be archived", clusterID)
	}
	if getStatus != 404 {
		return fmt.Errorf("failed to wait for cluster '%s' to be archived", clusterID)
	}
	connection.Logger().Info(ctx, "Cluster '%s' deleted successfully", clusterID)
	return nil
}
