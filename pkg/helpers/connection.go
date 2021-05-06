package helpers

import (
	"context"
	"net/http"

	sdk "github.com/openshift-online/ocm-sdk-go"
	"github.com/openshift-online/ocm-sdk-go/logging"
)

// BuildConnection build the vegeta connection
// that is going to be used for testing
func BuildConnection(gateway, clientID, clientSecret, token string, logger logging.Logger, ctx context.Context) (*sdk.Connection, error) {
	conn, err := sdk.NewConnectionBuilder().
		Insecure(true).
		URL(gateway).
		Client(clientID, clientSecret).
		Tokens(token).
		Logger(logger).
		TransportWrapper(func(wrapped http.RoundTripper) http.RoundTripper {
			return &CleanClustersTransport{Wrapped: wrapped, Logger: logger, Context: ctx}
		}).
		BuildContext(ctx)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
