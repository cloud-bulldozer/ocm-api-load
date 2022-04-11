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

type CleanTestTransport struct {
	Wrapped http.RoundTripper
	Logger  logging.Logger
}

func (t *CleanTestTransport) RoundTrip(request *http.Request) (*http.Response, error) {
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
		response = t.addToCleanup(request, response, true)
	}
	if t.isDeleteCluster(request) && response.StatusCode == 204 {
		t.removeFromCleanup(request)
	}
	if t.isClusterRegistration(request) && response.StatusCode == 201 {
		response = t.addToCleanup(request, response, false)
	}
	if t.isClusterAuthorization(request) && response.StatusCode == 200 {
		response = t.addToArchive(request, response, false)
	}
	return response, err
}

func (t *CleanTestTransport) addToCleanup(request *http.Request, response *http.Response, deprovision bool) *http.Response {
	ctx := request.Context()

	var cluster map[string]interface{}
	err := json.NewDecoder(response.Body).Decode(&cluster)
	if err != nil {
		t.Logger.Error(ctx, "Failed to unmarshal body of response for request %s %s: %v", request.Method,
			request.URL.String(), err)
		return response
	}
	clusterID, ok := cluster["id"]
	if !ok {
		// Testing for cluster registration body
		clusterID, ok = cluster["cluster_id"]
		if !ok {
			t.Logger.Error(ctx, "Failed to get cluster ID from body of response for request %s %s: %v", request.Method,
				request.URL.String(), err)
			return response
		}
	}
	markClusterForCleanup(ctx, clusterID.(string), deprovision, t.Logger)

	body, err := json.Marshal(cluster)
	if err != nil {
		t.Logger.Error(ctx, "Failed to marshall body of response for request %s %s: %v", request.Method,
			request.URL.String(), err)
		return response
	}
	response.Body = ioutil.NopCloser(strings.NewReader(string(body)))
	return response
}

func (t *CleanTestTransport) addToArchive(request *http.Request, response *http.Response, deprovision bool) *http.Response {
	ctx := request.Context()
	var authorization map[string]interface{}
	err := json.NewDecoder(response.Body).Decode(&authorization)
	if err != nil {
		t.Logger.Error(ctx, "Failed to unmarshal body of response for request %s %s: %v", request.Method,
			request.URL.String(), err)
		return response
	}
	subscription, ok := authorization["subscription"].(map[string]interface{})
	if !ok {
		t.Logger.Error(ctx, "Failed to get subscription from body of response for request %s %s: %v", request.Method,
			request.URL.String(), err)
		return response
	}
	subscriptionID, ok := subscription["id"]
	if !ok {
		t.Logger.Error(ctx, "Failed to get subscription ID from body of response for request %s %s: %v", request.Method,
			request.URL.String(), err)
		return response
	}

	markSubscriptionForArchiving(ctx, subscriptionID.(string), t.Logger)
	body, err := json.Marshal(authorization)
	if err != nil {
		t.Logger.Error(ctx, "Failed to marshall body of response for request %s %s: %v", request.Method,
			request.URL.String(), err)
		return response
	}
	response.Body = ioutil.NopCloser(strings.NewReader(string(body)))
	return response
}

func (t *CleanTestTransport) removeFromCleanup(request *http.Request) {
	ctx := request.Context()
	urlParts := strings.Split(request.URL.String(), "?")
	url := urlParts[0]
	parts := strings.Split(url, "/")
	clusterID := parts[len(parts)-1]
	t.Logger.Info(ctx, "Removing cluster '%s' from cleanup", clusterID)
	delete(createdClusterIDs, clusterID)
}

func (t *CleanTestTransport) manipulateRequest(request *http.Request) (*http.Request, bool, error) {
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

func (t *CleanTestTransport) isCreateCluster(request *http.Request) bool {
	url := strings.TrimSuffix(request.URL.String(), "/")
	return request.Method == "POST" && strings.HasSuffix(url, "/clusters") && request.Body != nil
}

func (t *CleanTestTransport) isClusterRegistration(request *http.Request) bool {
	url := strings.TrimSuffix(request.URL.String(), "/")
	return request.Method == "POST" && strings.HasSuffix(url, "/cluster_registrations") && request.Body != nil
}

func (t *CleanTestTransport) isClusterAuthorization(request *http.Request) bool {
	url := strings.TrimSuffix(request.URL.String(), "/")
	return request.Method == "POST" && strings.HasSuffix(url, "/cluster_authorizations") && request.Body != nil
}

func (t *CleanTestTransport) isDeleteCluster(request *http.Request) bool {
	parts := strings.Split(request.URL.String(), "/")
	return parts[len(parts)-2] == "clusters" && request.Method == "DELETE"
}

func markClusterForCleanup(ctx context.Context, clusterID string, deprovision bool, logger logging.Logger) {
	logger.Info(ctx, "Marking cluster '%s' for cleanup with 'deprovision'=%v", clusterID, deprovision)
	createdClusterIDs[clusterID] = deprovision
}

func markSubscriptionForArchiving(ctx context.Context, subscriptionID string, logger logging.Logger) {
	logger.Info(ctx, "Marking subscription '%s' for archiving", subscriptionID)
	createdSubcriptionIDs = append(createdSubcriptionIDs, subscriptionID)
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
