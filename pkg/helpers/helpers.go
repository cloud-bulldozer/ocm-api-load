package helpers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	sdk "github.com/openshift-online/ocm-sdk-go"
	log "github.com/sirupsen/logrus"

	errors "github.com/zgalor/weberr"

	"gitlab.cee.redhat.com/service/uhc-clusters-service/test/api"
)

// createdClusterIDs maps the IDs of the cluster created by testing to a bool value for `deprovision`.
var createdClusterIDs = map[string]bool{}
var validateDeletedClusterIDs = make([]string, 0)
var failedCleanupClusterIDs = make([]string, 0)

func Cleanup(connection *sdk.Connection) {
	if len(createdClusterIDs) == 0 {
		return
	}
	log.Infof("About to clean up the following clusters:")
	for clusterID, deprovision := range createdClusterIDs {
		log.Infof("Cluster ID: %s, deprovision: %v", clusterID, deprovision)
		DeleteCluster(clusterID, deprovision, connection)
	}
	for _, clusterID := range validateDeletedClusterIDs {
		err := verifyClusterDeleted(clusterID, connection)
		if err != nil {
			markFailedCleanup(clusterID)
		} else {
			delete(createdClusterIDs, clusterID)
		}
	}
	if len(failedCleanupClusterIDs) > 0 {
		log.Errorf("The following clusters failed deletion: %v", failedCleanupClusterIDs)
	}
	createdClusterIDs = make(map[string]bool)
	failedCleanupClusterIDs = make([]string, 0)
}

func DeleteCluster(id string, deprovision bool, connection *sdk.Connection) {
	log.Infof("Deleting cluster '%s'", id)
	// Send the request to delete the cluster
	response, err := connection.Delete().
		Path(fmt.Sprintf("%s/%s", ClustersEndpoint, id)).
		Parameter("deprovision", deprovision).
		Send()
	if err != nil {
		log.Errorf("Failed to delete cluster '%s', got error: %v", id, err)
		markFailedCleanup(id)
	} else if response.Status() != 204 {
		log.Errorf("Failed to delete cluster '%s', got http status %d", id, response.Status())
		markFailedCleanup(id)
	} else {
		validateDeletedClusterIDs = append(validateDeletedClusterIDs, id)
		log.Infof("Cluster '%s' deleted", id)
	}
}

func CreateCluster(body string, gatewayConnection *sdk.Connection) (string, map[string]interface{}, error) {
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
	clusterID := api.DigString(data, "id")
	log.Infof("Cluster '%s' created", clusterID)
	return clusterID, data, nil
}

func verifyClusterDeleted(clusterID string, connection *sdk.Connection) error {
	log.Infof("verifying deleted cluster '%s'", clusterID)
	getStatus, err := api.RunAttempt(func() (interface{}, bool) {
		getResponse, err := connection.Get().
			Path(api.ClustersEndpoint + clusterID).
			Send()
		if getResponse.Status() == 404 || err != nil {
			return getResponse.Status(), false
		}
		return getResponse.Status(), true
	}, 300, time.Second)
	if err != nil {
		log.Errorf("failed to delete cluster '%s': %v", clusterID, err)
		return err
	}
	if getStatus != 404 {
		return fmt.Errorf("failed to wait for cluster '%s' to be archived", clusterID)
	}
	log.Infof("Cluster '%s' deleted successfully", clusterID)
	return nil
}

// CreateFolder creates folder in the system
func CreateFolder(path string) error {
	log.Infof("Creating '%s' directory", path)
	folder, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	err = os.MkdirAll(folder, os.FileMode(0755))
	if err != nil {
		return err
	}
	return nil
}

// CreateFile creates the file with the given name
func CreateFile(name, path string) (*os.File, error) {
	resultPath := filepath.Join(path, name)
	out, err := os.Create(resultPath)
	if err != nil {
		// Silently ignore pre-existing file.
		if err == os.ErrExist {
			return out, nil
		}
		return nil, fmt.Errorf("writing result: %v", err)
	}
	return out, nil
}
