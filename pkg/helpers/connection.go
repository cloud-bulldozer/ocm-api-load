package helpers

import (
	"net/http"

	sdk "github.com/openshift-online/ocm-sdk-go"
)

// BuildConnection build the vegeta connection
// that is going to be used for testing
func BuildConnection(gateway, clientID, clientSecret, token string) (*sdk.Connection, error) {
	conn, err := sdk.NewConnectionBuilder().
		Insecure(true).
		URL(gateway).
		Client(clientID, clientSecret).
		Tokens(token).
		TransportWrapper(func(wrapped http.RoundTripper) http.RoundTripper {
			return &CleanClustersTransport{Wrapped: wrapped}
		}).
		Build()
	if err != nil {
		return nil, err
	}
	return conn, nil
}
