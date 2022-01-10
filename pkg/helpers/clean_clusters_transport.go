package helpers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/user"
	"strings"
	"time"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/logging"
)

type CleanClustersTransport struct {
	Wrapped http.RoundTripper
	Logger  logging.Logger
}

func (t *CleanClustersTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	var err error
	manipulated := false
	ctx := request.Context()
	if t.isCreateCluster(request) {
		request, manipulated, err = t.manipulateRequest(request)
		if err != nil {
			t.Logger.Error(ctx, "Failed to manipulate request for cleanup: %v", err)
		}
	}
	response, err := t.Wrapped.RoundTrip(request)
	if err != nil {
		return response, err
	}
	if manipulated && response.StatusCode == 201 {
		response = t.addToCleanup(request, response)
	}
	if t.isDeleteCluster(request) && response.StatusCode == 204 {
		t.removeFromCleanup(request)
	}
	return response, err
}

func (t *CleanClustersTransport) addToCleanup(request *http.Request, response *http.Response) *http.Response {
	ctx := request.Context()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Logger.Error(ctx, "Failed to read body of response for request %s %s: %v", request.Method,
			request.URL.String(), err)
		return response
	}
	var cluster map[string]interface{}
	err = json.Unmarshal(body, &cluster)
	if err != nil {
		t.Logger.Error(ctx, "Failed to unmarshal body of response for request %s %s: %v", request.Method,
			request.URL.String(), err)
		return response
	}
	clusterID, ok := cluster["id"]
	if !ok {
		t.Logger.Error(ctx, "Failed to get cluster ID from body of response for request %s %s: %v", request.Method,
			request.URL.String(), err)
		return response
	}

	markClusterForCleanup(ctx, clusterID.(string), true, t.Logger)
	response.Body = ioutil.NopCloser(strings.NewReader(string(body)))
	return response
}

func (t *CleanClustersTransport) removeFromCleanup(request *http.Request) {
	ctx := request.Context()
	urlParts := strings.Split(request.URL.String(), "?")
	url := urlParts[0]
	parts := strings.Split(url, "/")
	clusterID := parts[len(parts)-1]
	t.Logger.Info(ctx, "Removing cluster '%s' from cleanup", clusterID)
	delete(createdClusterIDs, clusterID)
}

func (t *CleanClustersTransport) manipulateRequest(request *http.Request) (*http.Request, bool, error) {
	ctx := request.Context()
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		t.Logger.Error(ctx, "Failed to read body of cluster for request %s %s: %v",
			request.Method, request.URL.String(), err)
		return request, false, err
	}
	newBody, err := addTestProperties(string(body))
	if err != nil {
		t.Logger.Error(ctx, "Failed to add test properties to cluster for request %s %s: %v",
			request.Method, request.URL.String(), err)
		return request, false, err
	}
	t.Logger.Info(ctx, "%s %s: %s", request.Method, request.URL.String(), newBody)
	request.Body = ioutil.NopCloser(strings.NewReader(newBody))
	request.ContentLength = int64(len(newBody))
	return request, true, nil
}

func (t *CleanClustersTransport) isCreateCluster(request *http.Request) bool {
	url := strings.TrimSuffix(request.URL.String(), "/")
	return request.Method == "POST" && strings.HasSuffix(url, "/clusters") && request.Body != nil
}

func (t *CleanClustersTransport) isDeleteCluster(request *http.Request) bool {
	parts := strings.Split(request.URL.String(), "/")
	return parts[len(parts)-2] == "clusters" && request.Method == "DELETE"
}

func markClusterForCleanup(ctx context.Context, clusterID string, deprovision bool, logger logging.Logger) {
	logger.Info(ctx, "Marking cluster '%s' for cleanup with 'deprovision'=%v", clusterID, deprovision)
	createdClusterIDs[clusterID] = deprovision
}

func markFailedCleanup(clusterID string) {
	failedCleanupClusterIDs = append(failedCleanupClusterIDs, clusterID)
	delete(createdClusterIDs, clusterID)
}

// Parse parses the given JSON data and returns a map of strings containing the result.
func Parse(data []byte) (map[string]interface{}, error) {
	var object map[string]interface{}
	err := json.Unmarshal(data, &object)
	if err != nil {
		return nil, err
	}
	return object, nil
}

func addTestProperties(body string) (string, error) {
	cluster, err := Parse([]byte(body))
	if err != nil {
		return "", err
	}
	properties, ok := cluster["properties"].(map[string]interface{})
	if !ok {
		properties = map[string]interface{}{}
	}
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	properties["user"] = user.Name
	cluster["properties"] = properties
	if cluster["expiration_timestamp"] == "" {
		cluster["expiration_timestamp"] = time.Now().Add(time.Hour).Format(time.RFC3339)
	}
	result, err := json.Marshal(cluster)
	if err != nil {
		return "", err
	}
	return string(result), nil
}
