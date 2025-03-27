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
var createdSubcriptionIDs = make([]string, 0)
var validateDeletedSubcriptionIDs = make([]string, 0)
var failedDeletedSubcriptionIDs = make([]string, 0)
var createdServiceIDs = make([]string, 0)
var validateDeletedServicesIDs = make([]string, 0)
var failedDeletedServicesIDs = make([]string, 0)

func Cleanup(ctx context.Context, connection *sdk.Connection) {
	if len(createdClusterIDs) == 0 && len(createdSubcriptionIDs) == 0 && len(createdServiceIDs) == 0 {
		return
	}
	if len(createdClusterIDs) > 0 {
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
	if len(createdSubcriptionIDs) > 0 {
		connection.Logger().Info(ctx, "About to delete the following subscriptions:")
		for _, subscription := range createdSubcriptionIDs {
			connection.Logger().Info(ctx, "Subscription ID: %s", subscription)
			DeleteSubscription(ctx, subscription, connection)
		}
		for _, subscriptionID := range validateDeletedSubcriptionIDs {
			err := verifySubscriptionDeleted(ctx, subscriptionID, connection)
			if err != nil {
				failedDeletedSubcriptionIDs = append(failedDeletedSubcriptionIDs, subscriptionID)
			}
		}
		if len(failedDeletedSubcriptionIDs) > 0 {
			connection.Logger().Warn(ctx, "The following subscriptions failed archiving: %v", failedDeletedSubcriptionIDs)
		}
		createdSubcriptionIDs = make([]string, 0)
		failedDeletedSubcriptionIDs = make([]string, 0)
	}
	if len(createdServiceIDs) > 0 {
		connection.Logger().Info(ctx, "About to delete the following services:")
		for _, service := range createdServiceIDs {
			connection.Logger().Info(ctx, "Service ID: %s", service)
			DeleteService(ctx, service, connection)
		}
		for _, serviceID := range validateDeletedServicesIDs {
			err := verifyServiceDeleted(ctx, serviceID, connection)
			if err != nil {
				failedDeletedServicesIDs = append(failedDeletedServicesIDs, serviceID)
			}
		}
		if len(failedDeletedServicesIDs) > 0 {
			connection.Logger().Warn(ctx, "The following services failed to be deleted: %v", failedDeletedServicesIDs)
		}
		createdServiceIDs = make([]string, 0)
		failedDeletedServicesIDs = make([]string, 0)
	}
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

func DeleteSubscription(ctx context.Context, id string, connection *sdk.Connection) {
	connection.Logger().Info(ctx, "Deleting subscription '%s'", id)
	// Send the request to delete subscription
	response, err := connection.Delete().
		Path(SubscriptionEndpoint + id).
		Send()
	if err != nil {
		connection.Logger().Error(ctx, "Got error trying to delete subscription '%s', "+
			"adding to failed delete subscriptions", id)
		failedDeletedSubcriptionIDs = append(failedDeletedSubcriptionIDs, id)
	} else if (response.Status() != http.StatusOK) && (response.Status() != http.StatusNoContent) {
		connection.Logger().Error(ctx, "Failed to delete subscription '%s', "+
			"got http %d, marking it as failed delete subscription",
			id, response.Status())
		failedDeletedSubcriptionIDs = append(failedDeletedSubcriptionIDs, id)

	} else {
		validateDeletedSubcriptionIDs = append(validateDeletedSubcriptionIDs, id)
		connection.Logger().Info(ctx, "Subscription '%s' deleted", id)
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
		return "", nil, errors.Errorf("Failed to create cluster: expected response code %d, instead found: %d\nAdditional info: %s",
			http.StatusCreated, postResponse.Status(), postResponse.String())
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

func GetServerVersion(ctx context.Context, connection *sdk.Connection) string {
	// https://github.com/openshift-online/ocm-sdk-go/blob/main/examples/get_metadata.go
	// Get the client for the resource that manages the metadata:
	client := connection.ClustersMgmt().V1()

	// Send the request to retrieve the metadata:
	response, err := client.Get().SendContext(ctx)
	if err != nil {
		connection.Logger().Error(ctx, "Failed to get server version, got error: %v", err)
		return ""
	}
	metadata := response.Body()

	return metadata.ServerVersion()
}

func verifySubscriptionDeleted(ctx context.Context, subscriptionID string, connection *sdk.Connection) error {
	connection.Logger().Info(ctx, "verifying deleted subscription '%s'", subscriptionID)
	var forcedErr error
	var getStatus int
	err := retry.Retry(func(attempt uint) error {
		getResponse, err := connection.Get().
			Path(SubscriptionEndpoint + subscriptionID).
			Send()
		if err != nil {
			return err
		}
		if getResponse.Status() != 200 {
			getStatus = getResponse.Status()
		} else if getResponse.Status() == 200 {
			body, err := Parse(getResponse.Bytes())
			if err != nil {
				return err
			}
			status, ok := body["status"]
			if !ok {
				return err
			}
			if status == "Deprovisioned" {
				return nil
			}
			return forcedErr
		}
		return errors.Errorf("Subscription not deleted: %d", getResponse.Status())
	},
		strategy.Wait(1*time.Second),
		strategy.Limit(300))
	if err != nil {
		connection.Logger().Error(ctx, "failed to delete subscription cluster '%s': %v", subscriptionID, err)
		return err
	}
	if forcedErr != nil {
		return fmt.Errorf("failed to wait for subscription '%s' to be deleted", subscriptionID)
	}
	if getStatus != 200 {
		return fmt.Errorf("failed to wait for subscription '%s' to be deleted", subscriptionID)
	}
	connection.Logger().Info(ctx, "Subscription '%s' deleted successfully", subscriptionID)
	return nil
}

func DeleteService(ctx context.Context, id string, connection *sdk.Connection) {
	connection.Logger().Info(ctx, "Deleting service '%s'", id)
	// Send the request to delete the service
	response, err := connection.Delete().
		Path(ServiceEndpoint + id).
		Send()
	if err != nil {
		connection.Logger().Error(ctx, "Failed to delete service '%s', got error: %v", id, err)
		markFailedServiceCleanup(id)
	} else if response.Status() != 201 {
		connection.Logger().Error(ctx, "Failed to delete service '%s', got http status %d", id, response.Status())
		markFailedServiceCleanup(id)
	} else {
		validateDeletedServicesIDs = append(validateDeletedServicesIDs, id)
		connection.Logger().Info(ctx, "Service '%s' deleted", id)
	}
}

func verifyServiceDeleted(ctx context.Context, serviceID string, connection *sdk.Connection) error {
	connection.Logger().Info(ctx, "verifying deleted service '%s'", serviceID)
	var forcedErr error
	var getStatus int
	err := retry.Retry(func(attempt uint) error {
		getResponse, err := connection.Get().
			Path(ServiceEndpoint + serviceID).
			Send()
		if err != nil {
			return err
		}
		if getResponse.Status() != 200 {
			getStatus = getResponse.Status()
		} else if getResponse.Status() == 200 {
			body, err := Parse(getResponse.Bytes())
			if err != nil {
				return err
			}
			status, ok := body["service_state"]
			if !ok {
				return err
			}
			if status == "deleting service" {
				return nil
			}
			return forcedErr
		}
		return errors.Errorf("Service not deleted: %d", getResponse.Status())
	},
		strategy.Wait(1*time.Second),
		strategy.Limit(300))
	if err != nil {
		connection.Logger().Error(ctx, "failed to delete service '%s': %v", serviceID, err)
		return err
	}
	if forcedErr != nil {
		return fmt.Errorf("failed to wait for service '%s' to be deleted", serviceID)
	}
	if getStatus != 200 {
		return fmt.Errorf("failed to wait for service '%s' to be deleted", serviceID)
	}
	connection.Logger().Info(ctx, "Service '%s' deleted successfully", serviceID)
	return nil
}
