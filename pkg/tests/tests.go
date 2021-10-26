package tests

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/cloud-bulldozer/ocm-api-load/pkg/helpers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/tests/handlers"
	"github.com/cloud-bulldozer/ocm-api-load/pkg/types"
	authv1 "github.com/openshift-online/ocm-sdk-go/authorizations/v1"
)

// Specify Test Cases
// They are written this way to re-use functionality where possible and
// hopefully make it easier to modify and/or extend given the declarative
// style.
var tests = []types.TestOptions{
	{
		TestName: "self-access-token",
		Path:     "/api/accounts_mgmt/v1/access_token",
		Method:   http.MethodPost,
		Handler:  handlers.TestStaticEndpoint,
	},
	{
		TestName: "list-subscriptions",
		Path:     "/api/accounts_mgmt/v1/subscriptions",
		Method:   http.MethodGet,
		Handler:  handlers.TestStaticEndpoint,
	},
	{
		TestName: "access-review",
		Path:     "/api/authorizations/v1/access_review",
		Method:   http.MethodPost,
		Handler:  handlers.TestStaticEndpoint,
		Body:     accessReviewBody(),
	},
	{
		TestName: "register-new-cluster",
		Path:     "/api/accounts_mgmt/v1/cluster_registrations",
		Method:   http.MethodPost,
		Handler:  handlers.TestRegisterNewCluster,
	},
	{
		TestName: "register-existing-cluster",
		Path:     "/api/accounts_mgmt/v1/cluster_registrations",
		Method:   http.MethodPost,
		Handler:  handlers.TestRegisterExistingCluster,
	},
	{
		TestName: "create-cluster",
		Path:     "/api/clusters_mgmt/v1/clusters",
		Method:   http.MethodPost,
		Handler:  handlers.TestCreateCluster,
	},
	{
		TestName: "list-clusters",
		Path:     "/api/clusters_mgmt/v1/clusters",
		Method:   http.MethodGet,
		Handler:  handlers.TestStaticEndpoint,
	},
	{
		TestName: "get-current-account",
		Path:     "/api/accounts_mgmt/v1/current_account",
		Method:   http.MethodGet,
		Handler:  handlers.TestStaticEndpoint,
	},
	{
		TestName: "quota-cost",
		Path:     "/api/accounts_mgmt/v1/organizations/{orgId}/quota_cost",
		Method:   http.MethodGet,
		Handler:  handlers.TestQuotaCost,
	},
	{
		TestName: "resource-review",
		Path:     "/api/authorizations/v1/resource_review",
		Method:   http.MethodPost,
		Handler:  handlers.TestStaticEndpoint,
		Body:     resourceReviewBody(),
	},
	{
		TestName: "cluster-authorizations",
		Path:     "/api/accounts_mgmt/v1/cluster_authorizations",
		Method:   http.MethodPost,
		Handler:  handlers.TestClusterAuthorizations,
	},
	{
		TestName: "self-terms-review",
		Path:     "/api/authorizations/v1/self_terms_review",
		Method:   http.MethodPost,
		Handler:  handlers.TestStaticEndpoint,
	},
	{
		TestName: "certificates",
		Path:     "/api/accounts_mgmt/v1/certificates",
		Method:   http.MethodPost,
		Handler:  handlers.TestStaticEndpoint,
		Body:     certificatesBody(),
	},
}

func accessReviewBody() []byte {
	buff := &bytes.Buffer{}
	resourceReviewReq, err := authv1.NewAccessReviewRequest().
		AccountUsername(helpers.AccountUsername).
		Action("get").
		ResourceType("Subscription").
		Build()
	if err != nil {
		log.Printf("building `access-review` request: %s", err)
		return buff.Bytes()
	}
	err = authv1.MarshalAccessReviewRequest(resourceReviewReq, buff)
	if err != nil {
		log.Printf("marshaling `access-review` request: %s", err)
	}
	return buff.Bytes()
}

func resourceReviewBody() []byte {
	buff := &bytes.Buffer{}
	resourcereviewReq, err := authv1.NewResourceReviewRequest().
		AccountUsername(helpers.AccountUsername).
		ResourceType("Cluster").
		Action("get").
		Build()
	if err != nil {
		log.Printf("building `resource-review` request: %s", err)
		return buff.Bytes()
	}
	err = authv1.MarshalResourceReviewRequest(resourcereviewReq, buff)
	if err != nil {
		log.Printf("marshaling `resource-review` request: %s", err)
	}
	return buff.Bytes()
}

func certificatesBody() []byte {
	buff := &bytes.Buffer{}
	payload := struct {
		Type string `json:"type"`
		Arch string `json:"arch"`
	}{
		"sca",
		randomCertArch(),
	}
	pl, err := json.Marshal(payload)
	if err != nil {
		log.Printf("marshaling `certificates body` request %s", err)
		return buff.Bytes()
	}
	return bytes.NewBuffer(pl).Bytes()
}

func randomCertArch() string {
	rand.Seed(time.Now().UnixNano()) //seed to randomize on each run
	arch := []string{"x86", "x86_64", "ppc", "ppc64", "ppc64le", "s390", "s390x", "ia64", "aarch64"}
	return arch[rand.Intn(len(arch))]
}
